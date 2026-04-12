# Guía de Deployment de Workers en GCP

Esta guía documenta todos los errores encontrados y sus soluciones durante el deployment del worker `climate_aggregate` para facilitar el deployment de otros workers.

## Tabla de Contenidos
1. [Configuración Inicial de GCP](#configuración-inicial-de-gcp)
2. [Errores de golangci-lint](#errores-de-golangci-lint)
3. [Errores de Permisos GCP](#errores-de-permisos-gcp)
4. [Errores de Docker Build](#errores-de-docker-build)
5. [Checklist Pre-Deployment](#checklist-pre-deployment)

---

## Configuración Inicial de GCP

### 1. Crear Service Accounts

#### Runtime Service Account (para los workers en Cloud Run)
```bash
export PROJECT_ID="phosboard"
export SA_NAME="phosboard-runtime-sa"
export SA_EMAIL="${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"

# Crear la Service Account
gcloud iam service-accounts create ${SA_NAME} \
  --display-name="Phosboard Runtime Service Account" \
  --project=${PROJECT_ID}
```

#### Permisos para Runtime SA
```bash
# Pub/Sub subscriber
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/pubsub.subscriber"

# Pub/Sub publisher
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/pubsub.publisher"

# Cloud Storage admin (para bucket GCS)
gcloud storage buckets add-iam-policy-binding gs://phosboard-documents \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/storage.objectAdmin"

# Secret Manager accessor
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/secretmanager.secretAccessor"

# Cloud SQL client
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/cloudsql.client"
```

#### GitHub Actions Service Account
```bash
export GH_SA="github-actions-sa@phosboard.iam.gserviceaccount.com"

# Artifact Registry writer
gcloud projects add-iam-policy-binding phosboard \
  --member="serviceAccount:${GH_SA}" \
  --role="roles/artifactregistry.writer"

# Cloud Run developer
gcloud projects add-iam-policy-binding phosboard \
  --member="serviceAccount:${GH_SA}" \
  --role="roles/run.developer"

# Service Account User
gcloud projects add-iam-policy-binding phosboard \
  --member="serviceAccount:${GH_SA}" \
  --role="roles/iam.serviceAccountUser"

# Secret Manager accessor
gcloud projects add-iam-policy-binding phosboard \
  --member="serviceAccount:${GH_SA}" \
  --role="roles/secretmanager.secretAccessor"

# Service Account Token Creator (IMPORTANTE)
gcloud projects add-iam-policy-binding phosboard \
  --member="serviceAccount:${GH_SA}" \
  --role="roles/iam.serviceAccountTokenCreator"
```

### 2. Configurar Workload Identity para cada repositorio

**IMPORTANTE**: Cada repositorio de worker necesita su propio binding de Workload Identity.

```bash
export REPO_NAME="phosboard-worker-climate-aggregate"  # Cambiar según el worker

gcloud iam service-accounts add-iam-policy-binding github-actions-sa@phosboard.iam.gserviceaccount.com \
  --project=phosboard \
  --role="roles/iam.workloadIdentityUser" \
  --member="principalSet://iam.googleapis.com/projects/544990213867/locations/global/workloadIdentityPools/github-pool/attribute.repository/SamuelAnjel/${REPO_NAME}"
```

### 3. Crear Secretos en Secret Manager

```bash
# DATABASE_URL (solo necesario crearlo una vez)
echo -n "postgresql://user:password@host:5432/db" | \
  gcloud secrets create phosboard-database-url \
  --data-file=- \
  --project=phosboard

# Verificar que existe
gcloud secrets list --project=phosboard --filter="name:phosboard-database-url"
```

---

## Errores de golangci-lint

### Error 1: Formato inválido de configuración

#### Síntoma
```
jsonschema: "version" does not validate with "/properties/version/type": got number, want string
jsonschema: "issues" does not validate with "/properties/issues/additionalProperties": additional properties 'exclude-rules' not allowed
jsonschema: "" does not validate with "/additionalProperties": additional properties 'linters-settings' not allowed
```

#### Causa
golangci-lint v2.5.0 tiene un esquema muy estricto que no permite ciertas propiedades.

#### Solución
Usar una configuración minimalista en `.golangci.yml`:

```yaml
version: "2"  # String, no number

run:
  go: "1.25"
  timeout: 5m

linters:
  enable:
    - errcheck
    - unused
    - govet
    - ineffassign
  disable:
    - staticcheck  # Para evitar SA1019 deprecation warnings

# NO usar: linters-settings, exclude-rules, exclude-use-default
```

### Error 2: Linter no disponible

#### Síntoma
```
Error: can't load config: typecheck is not a linter, it cannot be enabled or disabled
```

#### Causa
`typecheck` no es un linter que se pueda habilitar manualmente en v2.5.0.

#### Solución
Remover `typecheck` de la lista de linters habilitados. Solo usar: `errcheck`, `unused`, `govet`, `ineffassign`.

### Error 3: Errores de errcheck

#### Síntoma
```
Error return value of `reader.Close` is not checked (errcheck)
Error return value of `os.Setenv` is not checked (errcheck)
Error return value of `client.Close` is not checked (errcheck)
```

#### Solución
Manejar todos los errores explícitamente:

```go
// Para Close() en defer
defer func() {
    if closeErr := reader.Close(); closeErr != nil {
        fmt.Printf("warning: failed to close reader: %v\n", closeErr)
    }
}()

// Para os.Setenv
if err := os.Setenv("KEY", "value"); err != nil {
    return fmt.Errorf("set env var: %w", err)
}
```

### Error 4: SA1019 deprecation warning de pubsub

#### Síntoma
```
SA1019: "cloud.google.com/go/pubsub" is deprecated: Please use cloud.google.com/go/pubsub/v2.
```

#### Causa
pubsub v1 está deprecado, pero v2 tiene breaking API changes significativos.

#### Solución
**Opción 1 (recomendada)**: Deshabilitar staticcheck temporalmente:
```yaml
linters:
  disable:
    - staticcheck
```

**Opción 2**: Migrar a pubsub/v2 (requiere cambios mayores en el código).

---

## Errores de Permisos GCP

### Error 1: Permission 'iam.serviceAccounts.getAccessToken' denied

#### Síntoma
```
Error: google-github-actions/get-secretmanager-secrets failed with: 
failed to access secret "projects/544990213867/secrets/phosboard-database-url/versions/latest": 
permission 'iam.serviceAccounts.getAccessToken' denied on resource (or it may not exist).
```

#### Causas Posibles
1. Falta el rol `roles/iam.serviceAccountTokenCreator`
2. No existe el binding de Workload Identity para el repositorio específico
3. El secreto no existe en Secret Manager

#### Solución 1: Agregar Service Account Token Creator
```bash
gcloud projects add-iam-policy-binding phosboard \
  --member="serviceAccount:github-actions-sa@phosboard.iam.gserviceaccount.com" \
  --role="roles/iam.serviceAccountTokenCreator"
```

#### Solución 2: Verificar Workload Identity binding
```bash
# Ver bindings actuales
gcloud iam service-accounts get-iam-policy \
  github-actions-sa@phosboard.iam.gserviceaccount.com \
  --project=phosboard

# Agregar si falta
gcloud iam service-accounts add-iam-policy-binding \
  github-actions-sa@phosboard.iam.gserviceaccount.com \
  --project=phosboard \
  --role="roles/iam.workloadIdentityUser" \
  --member="principalSet://iam.googleapis.com/projects/544990213867/locations/global/workloadIdentityPools/github-pool/attribute.repository/SamuelAnjel/[REPO-NAME]"
```

#### Solución 3: Verificar que el secreto existe
```bash
gcloud secrets list --project=phosboard --filter="name:phosboard-database-url"
```

---

## Errores de Docker Build

### Error 1: Directory not found en Docker build

#### Síntoma
```
#18 [builder 7/7] RUN CGO_ENABLED=0 GOOS=linux go build -o /worker ./cmd/worker
#18 0.042 stat /app/cmd/worker: directory not found
```

#### Causa
El contexto de Docker no está copiando correctamente los archivos, o hay un problema con el `.dockerignore`.

#### Solución
1. Verificar que no exista `.dockerignore` que excluya `cmd/`
2. Agregar un paso de verificación en el Dockerfile:

```dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Verify cmd/worker exists (debugging step)
RUN ls -la cmd/worker/

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker ./cmd/worker

FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /worker .

CMD ["./worker"]
```

3. Probar build localmente antes de hacer push:
```bash
docker build -t test-worker .
```

---

## Checklist Pre-Deployment

### Antes de crear el primer tag:

#### 1. Configuración de GCP
- [ ] Runtime service account creada (`phosboard-runtime-sa`)
- [ ] Runtime SA tiene roles: pubsub.subscriber, pubsub.publisher, storage.objectAdmin, secretmanager.secretAccessor
- [ ] GitHub Actions SA (`github-actions-sa`) tiene roles: artifactregistry.writer, run.developer, iam.serviceAccountUser, secretmanager.secretAccessor, iam.serviceAccountTokenCreator
- [ ] Workload Identity binding creado para el repositorio específico
- [ ] Secreto `phosboard-database-url` existe en Secret Manager
- [ ] Bucket GCS `phosboard-documents` existe

#### 2. Código del Worker
- [ ] Migrado de MinIO a GCS
- [ ] `go.mod` incluye `cloud.google.com/go/storage v1.47.0`
- [ ] Todos los `Close()` tienen manejo de errores
- [ ] `os.Setenv()` tiene manejo de errores
- [ ] Código compila: `go build ./cmd/worker`
- [ ] `go vet ./...` pasa sin errores

#### 3. Configuración de golangci-lint
- [ ] `.golangci.yml` tiene configuración minimalista válida para v2.5.0
- [ ] No usa propiedades no soportadas: `linters-settings`, `exclude-rules`, `exclude-use-default`
- [ ] `staticcheck` está deshabilitado (para evitar warnings de pubsub v1)

#### 4. Dockerfile
- [ ] Usa `golang:1.25-alpine` como base
- [ ] Copia `go.mod` y `go.sum` primero (para cache)
- [ ] Copia todo el código con `COPY . .`
- [ ] Build comando correcto: `CGO_ENABLED=0 GOOS=linux go build -o /worker ./cmd/worker`
- [ ] Build local exitoso: `docker build -t test .`

#### 5. GitHub Actions Workflow
- [ ] Archivo en `.github/workflows/build-deploy.yml`
- [ ] Variables de entorno correctas: `GCP_PROJECT_ID`, `GAR_LOCATION`, `RUNTIME_SERVICE_ACCOUNT`
- [ ] Workflow tiene 3 jobs: lint, security, build
- [ ] Deploy flags incluyen: `--service-account`, `--min-instances=0`, `--max-instances=3`

#### 6. Testing Local
- [ ] `go build ./...` exitoso
- [ ] `go vet ./...` sin errores
- [ ] `docker build .` exitoso

### Al crear el tag:

```bash
# Verificar que todo está committeado
git status

# Crear y pushear tag
git tag v1.0.0
git push origin v1.0.0
```

### Monitorear Deployment:
```bash
# Ver workflow en GitHub
https://github.com/SamuelAnjel/[REPO-NAME]/actions

# Ver logs de Cloud Run (después del deploy)
gcloud logs read --resource-type=cloud_run_revision \
  --filter="resource.serviceName=[SERVICE-NAME]" \
  --limit=50 \
  --project=phosboard

# Ver servicio
gcloud run services describe [SERVICE-NAME] \
  --region=us-east1 \
  --project=phosboard
```

---

## Resumen de Errores Comunes y Soluciones Rápidas

| Error | Solución Rápida |
|-------|-----------------|
| `typecheck is not a linter` | Remover `typecheck` de `.golangci.yml` |
| `exclude-rules not allowed` | Usar configuración minimalista, remover sección `issues` |
| `version must be string` | Cambiar `version: 2` a `version: "2"` |
| `Error return value not checked` | Agregar manejo de error explícito con `if err :=` |
| `SA1019 pubsub deprecated` | Deshabilitar `staticcheck` en linters |
| `permission getAccessToken denied` | Agregar rol `iam.serviceAccountTokenCreator` |
| `secret not found` | Crear secreto con `gcloud secrets create` |
| `Workload Identity failed` | Agregar binding para el repositorio específico |
| `cmd/worker not found` | Verificar Dockerfile, agregar `RUN ls -la cmd/worker/` |

---

## Plantillas

### .golangci.yml
```yaml
version: "2"

run:
  go: "1.25"
  timeout: 5m

linters:
  enable:
    - errcheck
    - unused
    - govet
    - ineffassign
  disable:
    - staticcheck
```

### Dockerfile
```dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN ls -la cmd/worker/

RUN CGO_ENABLED=0 GOOS=linux go build -o /worker ./cmd/worker

FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /worker .

CMD ["./worker"]
```

### GitHub Actions Workflow (snippet)
```yaml
env:
  GCP_PROJECT_ID: phosboard
  GCP_PROJECT_NUMBER: '544990213867'
  RUNTIME_SERVICE_ACCOUNT: phosboard-runtime-sa@phosboard.iam.gserviceaccount.com
  GAR_LOCATION: us-east1

jobs:
  lint:
    # golangci-lint v2.5.0
  
  security:
    # Trivy scan
  
  build:
    steps:
      - name: Deploy to Cloud Run
        uses: google-github-actions/deploy-cloudrun@v2
        with:
          service: worker-[NAME]
          region: us-east1
          flags: '--allow-unauthenticated --min-instances=0 --max-instances=3 --service-account=${{ env.RUNTIME_SERVICE_ACCOUNT }}'
          env_vars: |
            DATABASE_URL=${{ env.DATABASE_URL }}
            GOOGLE_PROJECT_ID=${{ env.GCP_PROJECT_ID }}
```

---

## Notas Finales

1. **Siempre verificar localmente antes de pushear**: `go build`, `go vet`, `docker build`
2. **Cada worker necesita su propio Workload Identity binding**
3. **golangci-lint v2.5.0 es muy restrictivo** - usar configuración minimalista
4. **Documentar cada error nuevo** en este archivo para referencia futura
5. **Crear tags para deployments**, no push directo a main

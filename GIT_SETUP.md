# Git Setup - Phosboard

## Repositorios

| Proyecto | URL SSH | Rama Principal |
|----------|---------|----------------|
| phosboard-api | git@github.com:SamuelAnjel/phosboard-api.git | main |

## Esquema de Ramas

```
main                    # Producción - solo merges desde dev
dev                     # Desarrollo estable - merges desde features/fixes
feature/*               # Nuevas funcionalidades
fix/*                   # Correcciones de bugs
```

## Inicialización (ya realizada)

```bash
# Backend
cd backend
git init
git remote add origin git@github.com:SamuelAnjel/phosboard-api.git

# Crear ramas
git checkout -b main
git branch dev
git branch feature/initial

# Push inicial
git push -u origin main dev feature/initial
```

## Workflow de Desarrollo

### 1. Iniciar nueva feature
```bash
git checkout dev
git pull origin dev
git checkout -b feature/nombre-feature
```

### 2. Trabajar en la feature
```bash
# Hacer commits
git add .
git commit -m "feat: descripción"
```

### 3. Mergear a dev
```bash
git checkout dev
git merge feature/nombre-feature
git push origin dev
```

### 4. Release a producción
```bash
# Crear tag
git checkout main
git merge dev
git tag v1.0.0
git push origin main --tags
```

## GitHub Actions

El workflow `build-deploy.yml` se ejecuta al crear un tag:
- Lint con golangci-lint
- Security scan con Trivy
- Build de imagen Docker
- Push a Artifact Registry (us-east1-docker.pkg.dev/phosboard/phosboard-images)
- Deploy a Cloud Run

### Permisos Service Account

La service account `phosboard-sa@phosboard.iam.gserviceaccount.com` necesita:
```bash
# Permisos requeridos
gcloud projects add-iam-policy-binding phosboard \
  --member="serviceAccount:phosboard-sa@phosboard.iam.gserviceaccount.com" \
  --role="roles/artifactregistry.writer"

gcloud projects add-iam-policy-binding phosboard \
  --member="serviceAccount:phosboard-sa@phosboard.iam.gserviceaccount.com" \
  --role="roles/cloudrun.developer"
```

### Workload Identity
- Pool: github-pool
- Provider: github-provider

## Comandos Útiles

```bash
# Ver ramas locales
git branch

# Ver ramas remotas
git branch -r

# Ver todas las ramas
git branch -a

# Eliminar rama local
git branch -d nombre-rama

# Eliminar rama remota
git push origin --delete nombre-rama

# Ver estado
git status

# Ver diff
git diff

# Deshacer último commit (sin push)
git reset --soft HEAD~1
```

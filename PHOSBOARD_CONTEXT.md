# Phosboard - Contexto del Proyecto

## Descripción General

Phosboard es un sistema de gestión de documentos con control de acceso basado en roles (RBAC).

## Credenciales y Variables de Entorno

### PostgreSQL
| Variable | Valor |
|----------|-------|
| HOST | localhost:5432 |
| USER | phos_user |
| PASSWORD | phos_password |
| DATABASE | phosboard |
| CONNECTION STRING | postgres://phos_user:phos_password@localhost:5432/phosboard |

### RBAC System (Implementado Abril 2026)
- **Roles**: super-admin (dueño), tenant-admin (cliente líder), tenant-editor (cuentas asociadas)
- **Autorización**: Por endpoint + método HTTP (GET, POST, PUT, DELETE)
- **Tablas**: `users`, `permissions`, `user_roles`, `role_permissions`
- **Middleware**: `AuthWithAuthorization()` verifica permisos por endpoint
- **Eliminado**: Todas las credenciales hardcodeadas del backend

### MinIO
| Variable | Valor |
|----------|-------|
| ENDPOINT | localhost:9000 |
| ACCESS KEY | minioadmin |
| SECRET KEY | minioadmin |
| CONSOLE | localhost:9001 |

### Pub/Sub Emulator
| Variable | Valor |
|----------|-------|
| PROJECT ID | phosboard |
| HOST | localhost:8085 |
| DASHBOARD | localhost:8086 |

### Variables por Componente
```
# Backend
DATABASE_URL=postgres://phos_user:phos_password@localhost:5432/phosboard
GOOGLE_PROJECT_ID=phosboard
# PUBSUB_EMULATOR_HOST=localhost:8085  # Uncomment for local emulator
DISPATCHER_INTERVAL_SECONDS=900
MINIO_ENDPOINT=localhost:9000        # Solo para desarrollo local
MINIO_ACCESS_KEY=minioadmin          # Solo para desarrollo local
MINIO_SECRET_KEY=minioadmin          # Solo para desarrollo local

# Workers (v2.0.0+ con Bronze/Silver)
# Discovery Worker
DATABASE_URL=postgres://phos_user:phos_password@localhost:5432/phosboard
GOOGLE_PROJECT_ID=phosboard
# PUBSUB_EMULATOR_HOST=localhost:8085  # Uncomment for local emulator

# Scraper Worker (Bronze/Silver)
DATABASE_URL=postgres://phos_user:phos_password@localhost:5432/phosboard
GOOGLE_PROJECT_ID=phosboard
GCS_BUCKET=phosboard-documents        # Para producción (GCS)
# PUBSUB_EMULATOR_HOST=localhost:8085  # Uncomment for local emulator

# Semantic Worker
DATABASE_URL=postgres://phos_user:phos_password@localhost:5432/phosboard
GOOGLE_PROJECT_ID=phosboard
GOOGLE_LOCATION=us-central1           # Para Vertex AI
# PUBSUB_EMULATOR_HOST=localhost:8085  # Uncomment for local emulator
```

---

### global_documents (actualizada PHOS-27, PHOS-30)
| Columna | Tipo | Descripción |
|---------|------|-------------|
| ... | | |
| semantic_analysis | JSONB | Análisis semántico de Vertex AI |
| social_temperature | NUMERIC(5,2) | Temperatura social (0-100) |

---

## Stack Tecnológico (v2.0.0+)

- **Backend**: Go 1.22+ con pgx/v5 (sin ORMs)
- **Framework HTTP**: Gin (github.com/gin-gonic/gin)
- **Base de Datos**: PostgreSQL 16+ con pgvector
- **Frontend**: Vue 3 + Vite + TypeScript + TailwindCSS + Vuetify
- **Workers**: Go con Pub/Sub Push, Google Cloud Storage
- **AI/ML**: Vertex AI Gemini 1.5 Flash
- **Infra**: 
  - **Local**: Docker Compose (PostgreSQL, MinIO, Pub/Sub Emulator)
  - **Producción**: Cloud Run, Cloud SQL, GCS, Pub/Sub, Secret Manager
- **Logging**: slog en formato JSON
- **Patrón Data Lake**: Bronze/Silver para almacenamiento de HTML

## Reglas de Desarrollo

1. **SQL**: snake_case, UUID como PKs con `uuid_generate_v4()`, scripts idempotentes, siempre incluir `created_at` y `updated_at`
2. **Go**: Sin ORMs, usar SQL crudo con pgx, `context.Context` como primer parámetro, slog JSON, error wrapping con `fmt.Errorf("contexto: %w", err)`

---

## Estructura del Proyecto

```
phosboard/
├── docker-compose.yaml          # Servicios: postgres, minio, pubsub
├── data/
│   └── migrations/
│       ├── 00001_initial_schema.sql
│       ├── 00002_seed_sources.sql
│       ├── 00003_sources_orchestration.sql
│       ├── 00004_social_climate.sql
│       ├── 00005_tenant_concepts.sql
│       ├── 00006_add_semantic_analysis.sql
│       ├── 00007_add_temperature_to_documents.sql
│       └── 00008_add_sources_config.sql
├── backend/
│   ├── .env
│   ├── .env.example
│   ├── Dockerfile              # Para producción
│   ├── go.mod
│   ├── go.sum
│   ├── .github/
│   │   └── workflows/
│   │       └── build-deploy.yml  # CI/CD para GCP
│   ├── cmd/
│   │   ├── api/main.go
│   │   ├── setup-pubsub/main.go
│   │   └── setup-minio/main.go
│   └── internal/
│       ├── config/
│       ├── db/
│       ├── models/
│       ├── repository/
│       ├── dispatcher/
│       ├── handler/
│       └── publisher/
├── frontend/
│   └── src/
│       ├── plugins/
│       │   ├── axios.ts        (PHOS-36)
│       │   └── vuetify.ts      (PHOS-37)
│       ├── stores/auth.ts       (PHOS-36)
│       ├── router/index.ts      (PHOS-36, PHOS-37)
│       ├── views/
│       │   ├── Login.vue        (PHOS-36)
│       │   ├── Dashboard.vue    (PHOS-37)
│       │   ├── Documents.vue     (PHOS-37)
│       │   └── Concepts.vue     (PHOS-37)
│       ├── components/
│       │   └── layouts/
│       │       └── DefaultLayout.vue (PHOS-37)
│       ├── composables/
│       └── types/
└── workers/
    ├── discovery/          (PHOS-21) - v2.0.0+
    │   ├── cmd/worker/main.go
    │   └── internal/
    │       ├── discovery/collector.go
    │       ├── publisher/publisher.go      # Publica con document_id
    │       ├── repository/discovery_task.go # Método CreateDocument
    │       └── subscriber/subscriber.go    # Crea documentos antes de publicar
    ├── scraper/            (PHOS-22) - v2.0.0+ Bronze/Silver
    │   ├── cmd/worker/main.go
    │   └── internal/
    │       ├── scraper/collector.go        # HTTP client, HTML extractor, GCS storage
    │       ├── publisher/publisher.go
    │       ├── repository/discovery_task.go # UpdateGlobalDocument
    │       └── handler/pubsub.go           # HTTP handler para Pub/Sub push
    ├── semantic/           (PHOS-27, PHOS-28) - v2.0.0+
    │   ├── cmd/worker/main.go
    │   └── internal/
    │       ├── analyzer/analyzer.go        # Vertex AI Gemini integration
    │       ├── publisher/publisher.go
    │       ├── repository/document_repository.go
    │       └── handler/pubsub.go
    ├── social_probe/       (PHOS-29)
    │   ├── cmd/worker/main.go
    │   └── internal/
    │       ├── scraper/scraper.go          # Mock scraper
    │       ├── publisher/publisher.go
    │       ├── storage/gcs.go
    │       └── handler/pubsub.go
    ├── climate_aggregate/  (PHOS-30) - v2.0.0+
    │   ├── cmd/worker/main.go
    │   └── internal/
    │       ├── calculator/calculator.go    # Temperature calculation
    │       ├── repository/document_repository.go
    │       ├── storage/gcs.go
    │       └── handler/pubsub.go
    ├── fetcher/            (DEPRECADO) - Reemplazado por scraper
    │   ├── .env
    │   ├── cmd/worker/main.go
    │   └── internal/
    │       ├── fetcher/rss_collector.go
    │       ├── models/raw_payload.go
    │       └── storage/parquet_writer.go
    └── parser/             (DEPRECADO) - Reemplazado por patrón Bronze/Silver
```

---

## Migración Inicial (00001_initial_schema.sql)

### Extensiones
- `uuid-ossp` - Generación de UUIDs
- `vector` - Vectores para embeddings (dimensión 1536)

### Tablas

#### tenants
| Columna | Tipo | Descripción |
|---------|------|-------------|
| id | UUID | PK |
| name | VARCHAR(255) | Nombre del tenant |
| slug | VARCHAR(100) | Identificador único |
| settings | JSONB | Configuración |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP | |

#### roles
| Columna | Tipo | Descripción |
|---------|------|-------------|
| id | UUID | PK |
| tenant_id | UUID | FK a tenants |
| name | VARCHAR(100) | Nombre del rol |
| permissions | JSONB | Permisos del rol |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP | |

#### tenant_users
| Columna | Tipo | Descripción |
|---------|------|-------------|
| id | UUID | PK |
| tenant_id | UUID | FK a tenants |
| user_id | UUID | FK al usuario (externo) |
| role_id | UUID | FK a roles |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP | |

#### sources
| Columna | Tipo | Descripción |
|---------|------|-------------|
| id | UUID | PK |
| name | VARCHAR(255) | Nombre de la fuente |
| type | VARCHAR(50) | Tipo de fuente |
| url | TEXT | URL del RSS/Sitemap |
| fetch_strategy | VARCHAR(50) | Estrategia de scraping |
| interval_minutes | INTEGER | Intervalo de fetch |
| last_run_at | TIMESTAMP | Último fetch |
| config | JSONB | Configuración (PHOS-40: max_links) |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP | |

#### global_documents
| Columna | Tipo | Descripción |
|---------|------|-------------|
| id | UUID | PK |
| source_id | UUID | FK a sources |
| title | VARCHAR(500) | Título |
| content_text | TEXT | Contenido |
| url | TEXT | URL (UNIQUE) |
| raw_payload | JSONB | Datos crudos |
| content_embedding | VECTOR(1536) | Embedding |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP | |

#### tenant_documents
| Columna | Tipo | Descripción |
|---------|------|-------------|
| tenant_id | UUID | FK a tenants (parte de PK compuesta) |
| document_id | UUID | FK a global_documents (parte de PK compuesta) |
| matched_keywords | TEXT[] | Keywords asociadas |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP | |
| PRIMARY KEY (tenant_id, document_id) |

#### discovery_tasks (PHOS-17)
| Columna | Tipo | Descripción |
|---------|------|-------------|
| id | UUID | PK |
| url | TEXT | URL única (UNIQUE) |
| source_type | VARCHAR(50) | Tipo de fuente |
| status | VARCHAR(20) | pending/processing/completed |
| priority | INTEGER | Prioridad |
| retry_count | INTEGER | Intentos |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP | |

#### social_mentions (PHOS-17)
Tabla particionada por `posted_at` (RANGE)
| Columna | Tipo | Descripción |
|---------|------|-------------|
| id | UUID | PK (+ posted_at) |
| document_id | UUID | FK a global_documents |
| platform | VARCHAR(50) | Plataforma |
| text_content | TEXT | Contenido |
| engagement_score | INTEGER | Score de engagement |
| sentiment_score | DECIMAL(5,4) | Score de sentimiento |
| posted_at | TIMESTAMP | Fecha de publicación |

#### document_temperatures (PHOS-17)
| Columna | Tipo | Descripción |
|---------|------|-------------|
| id | UUID | PK |
| document_id | UUID | FK a global_documents (UNIQUE) |
| total_mentions | INTEGER | Total menciones |
| total_engagement | INTEGER | Engagement total |
| temperature_score | DECIMAL(5,4) | Score de temperatura |
| velocity_metrics | JSONB | Métricas de velocidad |
| calculated_at | TIMESTAMP |Último cálculo |

---

## Backend Go

### PHOS-32: Refactor a Gin Framework
- **Framework**: `github.com/gin-gonic/gin`
- **Middleware**: `internal/http/middleware/auth.go`
- **Handler**: `internal/handler/auth_handler.go`

### Autenticación JWT (PHOS-31, PHOS-32)
- **Paquete**: `internal/auth/auth.go`
- **Middleware**: `internal/http/middleware/auth.go`
- **Librería**: `github.com/golang-jwt/jwt/v5`
- **Payload del token**: user_id, tenant_id, role_id
- **Expiración**: 24 horas
- **Endpoint de login**: POST /api/auth/login

### cmd/api/main.go
- Servidor HTTP en puerto 8080
- Desplegado en **Cloud Run** (us-east1)
- Endpoint GET `/health` para health check
- Endpoint GET `/api/v1/tenants/{tenant_id}/documents` para documentos
- Logger slog en formato JSON
- Middleware de logging con tiempo de respuesta
- Graceful shutdown
- **Variable PORT**: Cloud Run la maneja automáticamente (8080)

### internal/config/config.go
- Carga `DATABASE_URL` desde variables de entorno
- Retorna error si no está definida o está vacía

### internal/db/db.go
- Conexión a PostgreSQL usando pgxpool
- Funciones: Connect, Ping, Close

### internal/models/document.go
- Structs: GlobalDocument, DocumentWithSource, DocumentWithAnalysis

### internal/repository/document_repository.go
- Interfaz: DocumentRepository
- PostgresDocumentRepository con pool de pgx
- InsertGlobalDocument(ctx, doc) - Inserta/actualiza por URL冲突
- LinkDocumentToTenant(ctx, tenantID, docID, keywords) - Vincula documento
- GetLatestByTenant(ctx, tenantID) - Últimos 10 documentos
- GetDocumentsByTenant(ctx, tenantID, limit, offset) - Documentos con análisis y temperatura
- GetOrCreateSource(ctx, name, sourceType) - Get or create con ON CONFLICT
- TrackDocument(ctx, tenantID, url, sourceType, priority) - Transacción: global_documents + tenant_documents + discovery_tasks

### internal/repository/source_repository.go (PHOS-40)
- Interfaz: SourceRepository
- PostgresSourceRepository con pool de pgx
- CreateSource(ctx, name, sourceType) - Crea fuente
- GetSources(ctx) - Lista todas las fuentes
- GetSourceByID(ctx, id) - Obtiene fuente por ID
- UpdateSourceConfig(ctx, id, config) - Actualiza configuración JSON
- DeleteSource(ctx, id) - Elimina fuente

### PHOS-38: Automatización de Ingesta (Dispatcher + Worker Discovery)
- **Dispatcher**: `internal/dispatcher/dispatcher.go`
- Ticker configurable (default 15 minutos, configurable via DISPATCHER_INTERVAL_SECONDS)
- Consulta fuentes donde last_run_at + interval_minutes < NOW()
- Publica a topic `source-discovery` con payload {source_id, url, timestamp}
- Actualiza last_run_at solo si publicación exitosa
- **Worker Discovery**: Proceso completo del pipeline
  - Consume de `source-discovery-sub`
  - Descarga RSS/XML y extrae URLs
  - Inserta en `discovery_tasks` (ON CONFLICT DO NOTHING)
  - Publica en `url-scrape` por cada URL nueva
  - Logging con slog para trackear URLs descubiertas
- **Fuente de prueba**: Diario El Día (https://www.diarioeldia.cl/rss.xml, 15 min)

### tenant_concepts (PHOS-26)
Tabla para conceptos semilla por tenant
| Columna | Tipo | Descripción |
|---------|------|-------------|
| id | UUID | PK |
| tenant_id | UUID | FK a tenants |
| concept_term | VARCHAR(255) | Término del concepto |
| is_active | BOOLEAN | Soft delete |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP | |
| UNIQUE (tenant_id, concept_term) | | |

### .env (Backend)
- DATABASE_URL
- GOOGLE_PROJECT_ID
- PUBSUB_EMULATOR_HOST
- DISPATCHER_INTERVAL_SECONDS (default: 900 = 15 minutos)
- JWT_SECRET (secret para firmar tokens JWT)
- **RBAC**: Sistema completo implementado, sin credenciales hardcodeadas

### PHOS-31: Autenticación JWT con RBAC (Actualizado Abril 2026)
- **Paquete**: `internal/auth`
- **Middleware**: `internal/http/middleware/authorization.go` (RBAC completo)
- **Librería**: `github.com/golang-jwt/jwt/v5`
- **Payload del token**: user_id, tenant_id, role_ids (array)
- **Expiración**: 24 horas
- **Endpoint de login**: POST /api/v1/auth/login (usa base de datos con bcrypt)

### Login (RBAC - Base de datos)
```json
POST /api/v1/auth/login
{
  "email": "superadmin@phosboard.cl",
  "password": "password"
}
```
Respuesta:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid",
    "email": "superadmin@phosboard.cl",
    "roles": ["super-admin"]
  }
}
```

### Uso del token con RBAC
```bash
# Super-admin requiere tenant_id en request body
curl -X POST http://localhost:8080/api/v1/documents/track \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com", "tenant_id": "uuid"}'

# Tenant-admin/editor (tenant_id extraído del token)
curl -X POST http://localhost:8080/api/v1/documents/track \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'
```

### Workers .env
Cada worker (discovery, scraper, semantic) tiene su propio `.env` con:
- DATABASE_URL
- GOOGLE_PROJECT_ID
- GOOGLE_LOCATION (semantic: us-central1)
- PUBSUB_EMULATOR_HOST
- MINIO_ENDPOINT (scraper, semantic)
- MINIO_ACCESS_KEY (scraper, semantic)
- MINIO_SECRET_KEY (scraper, semantic)

### internal/models/
- document.go: GlobalDocument, DocumentWithSource
- source.go: SourceForFetch

### Tests
| Paquete | Cobertura |
|---------|-----------|
| config | 100% |
| db | ~60% (tests refactored) |
| repository | 100% |
| dispatcher | 100% |
| handler | 100% |

### Worker Dependencies
- `godotenv` - Carga de variables de entorno desde `.env`

---

## Worker Fetcher (PHOS-3 / PHOS-5 / PHOS-12)

### Estructura
```
workers/fetcher/
├── cmd/worker/main.go
└── internal/
    ├── fetcher/
    │   ├── rss_collector.go      (PHOS-12)
    │   └── rss_collector_test.go
    ├── models/
    │   └── raw_payload.go
    └── storage/
        ├── parquet_writer.go
        └── minio_uploader.go
```

### internal/fetcher/rss_collector.go (PHOS-12)
- RSS Parser: `github.com/mmcdole/gofeed`
- Feed URL: `https://www.diarioeldia.cl/rss.xml`
- SourceID: `diario-el-dia` (hardcoded)
- Max items: 20
- Mapeo: GUID → ID, Link → URL, Description → HTMLContent
- Fallback ID: SHA256 hash de URL si no hay GUID

### internal/subscriber/subscriber.go (PHOS-14)
- Pub/Sub Consumer Mode
- Escucha `fetcher-tasks` subscription
- Ack solo si MinIO upload exitoso
- Variables: GOOGLE_PROJECT_ID, PUBSUB_SUBSCRIPTION

### Dependencias
- `github.com/parquet-go/parquet-go` - Parquet
- `github.com/minio/minio-go/v7` - MinIO
- `github.com/mmcdole/gofeed` - RSS parsing (PHOS-12)
- `github.com/minio/minio-go/v7` - SDK de MinIO

---

## Worker Discovery (PHOS-21)

### Estructura
```
workers/discovery/
├── cmd/worker/main.go
└── internal/
    ├── discovery/collector.go
    ├── publisher/publisher.go
    ├── repository/discovery_task.go
    └── subscriber/subscriber.go
```

### Flujo (PHOS-38, PHOS-40) - v2.0.0+ con document_id
1. Consume de `source-discovery-sub` con mensaje: `{source_id, url, timestamp}`
2. Lee configuración de `sources.config` (PHOS-40: max_links, default: 20)
3. Descarga RSS y extrae URLs
4. Inserta en `discovery_tasks` (patrón Outbox con ON CONFLICT)
5. **Nuevo**: Crea documento en `global_documents` con `source_id` y URL (contenido vacío inicialmente)
6. **PHOS-40**: Publica en `url-scrape` con `document_id` (no `source_id`) hasta alcanzar max_links
7. Logging con slog para trackear progreso

### Cambios en v2.0.0+
- **Mensaje de salida**: Ahora usa `document_id` en lugar de `source_id`
- **Creación temprana**: Documentos creados en `global_documents` antes del scraping
- **Integración con Bronze/Silver**: Prepara documentos para el patrón data lake

### Tests
- `internal/discovery/collector_test.go` - 3 tests (Discover, isAbsoluteURL, InvalidURL)
- `internal/publisher/publisher_test.go` - 1 test (JSONMarshal)

---

## Worker Scraper (PHOS-22) - v2.0.0 Bronze/Silver Pattern

### Estructura
```
workers/scraper/
├── cmd/worker/main.go
└── internal/
    ├── scraper/collector.go    # HTTP client, HTML extractor, GCS storage
    ├── publisher/publisher.go
    ├── repository/discovery_task.go
    └── handler/pubsub.go       # HTTP handler for Pub/Sub push
```

### Arquitectura: Bronze/Silver Data Lake Pattern
- **Versión**: v2.0.0+ (consolidación de fetcher y parser workers)
- **Trigger**: Pub/Sub Push subscription (`url-scrape-sub`)
- **Storage**: Google Cloud Storage bucket `phosboard-documents`
- **Patrón**: Bronze (raw) → Silver (cleaned) → PostgreSQL (text)

### Flujo (Bronze/Silver Pattern Actualizado)
1. Recibe HTTP POST de Pub/Sub con mensaje `URLScrapeTask`:
   ```json
   {
     "document_id": "uuid",
     "url": "https://example.com"
   }
   ```
2. **Paso A**: HTTP GET para obtener HTML crudo de la URL
3. **Paso B (Bronze Layer)**: Guarda HTML crudo en GCS: `raw-html/{document_id}.html` (sin límites de tamaño)
4. **Paso C**: Limpia HTML en memoria (remueve scripts, estilos, anuncios, navegación)
5. **Paso D (Silver Layer)**: Guarda HTML limpio en GCS: `clean-html/{document_id}.html`
6. **Paso E**: Extrae texto plano completo
7. **Paso F**: Guarda texto plano completo en GCS: `plain-text/{document_id}.txt` (para Vertex AI)
8. **Paso G**: Guarda primeros 15,000 caracteres en PostgreSQL `global_documents.content_text` (para eficiencia)
9. **Paso H**: Publica `DocumentAnalyzeTask` a topic `document-analyze` con clave de texto plano

### Storage Pattern (Data Lake Actualizado)
- **Bronze (Raw)**: `raw-html/{document_id}.html` - HTML original, sin límites de tamaño
- **Silver (Cleaned)**: `clean-html/{document_id}.html` - HTML procesado para análisis
- **Text Layer**: `plain-text/{document_id}.txt` - Texto plano completo para Vertex AI
- **PostgreSQL**: Primeros 15,000 caracteres de texto - Para búsqueda de texto completo y vista previa
- **Nunca** almacenar HTML en PostgreSQL

### Filosofía de Diseño Actualizada
- **Eficiencia en PostgreSQL**: Guardar solo primeros 15,000 caracteres para documentos largos
- **Texto completo en GCS**: Vertex AI accede a texto completo desde GCS
- **Vertex AI maneja textos largos** (ventana de contexto de 1M tokens)
- **Optimizar basado en datos de uso real**, no suposiciones

### Consolidación de Workers
- **Worker Fetcher deprecado**: Scraper ahora maneja todo el fetching de contenido
- **Worker Parser deprecado**: Patrón Bronze/Silver reemplaza procesamiento por lotes
- **MinIO eliminado**: Usando SDK nativo de GCS con autenticación ADC

## Worker Climate Aggregate (PHOS-30)

### Arquitectura: Cloud Run Service + Pub/Sub Push (v2.0.0)

El worker se despliega como un **Cloud Run Service** que recibe mensajes vía **Pub/Sub Push Subscription**:
- Escala a 0 cuando no hay mensajes (sin costo cuando idle)
- Escala automáticamente cuando llegan mensajes
- Recibe HTTP POST de Pub/Sub en endpoint `/`

### Estructura
```
workers/climate_aggregate/
├── cmd/worker/main.go            # HTTP server (port 8080)
├── internal/
│   ├── handler/pubsub.go        # HTTP handler for Pub/Sub push
│   ├── calculator/calculator.go # Temperature calculation algorithm
│   ├── repository/              # PostgreSQL operations
│   └── storage/gcs.go           # GCS reads (migrated from MinIO)
└── .github/workflows/build-deploy.yml
```

### Flujo
1. Pub/Sub envía HTTP POST a `/` con mensaje push
2. Handler valida y decodifica mensaje (base64)
3. Descarga JSON de menciones desde GCS (bucket phosboard-documents)
4. Calcula temperatura usando algoritmo determinista
5. Actualiza `global_documents.social_temperature` en PostgreSQL
6. Responde HTTP 200 (ACK) o 500 (retry)

### Endpoints

| Endpoint | Method | Descripción |
|----------|--------|-------------|
| `/` | POST | Recibe mensajes de Pub/Sub push subscription |
| `/health` | GET | Health check endpoint |

### Pub/Sub Configuration

**Subscription Type**: Push Subscription
```bash
gcloud pubsub subscriptions update climate-aggregate-sub \
  --push-endpoint=https://worker-climate-aggregate-544990213867.us-east1.run.app/ \
  --push-auth-service-account=phosboard-runtime-sa@phosboard.iam.gserviceaccount.com
```

**IAM**: Solo `phosboard-runtime-sa` puede invocar el servicio (no público)

### Algoritmo de Temperatura
- Suma engagement_score total de menciones
- Calcula promedio: total_engagement / cantidad_menciones
- Normaliza a 0-100: (avg_engagement / 500) * 100
- Diseño modular para futura integración con Vertex AI

### Migración (00007)
- Columna `social_temperature NUMERIC(5,2)` en `global_documents`

### Migration: MinIO → GCS (v1.0.14)
- ✅ Storage migrado de MinIO a Google Cloud Storage
- ✅ Bucket: `phosboard-documents`
- ✅ Archivo renombrado: `internal/storage/minio.go` → `internal/storage/gcs.go`
- ✅ Removida dependencia de `github.com/minio/minio-go/v7`

### Dependencies
- `github.com/jackc/pgx/v5` - PostgreSQL
- `cloud.google.com/go/storage` - Google Cloud Storage
- `github.com/joho/godotenv` - Environment

### Deployment
- **Repository**: `SamuelAnjel/phosboard-worker-climate-aggregate`
- **Service**: `worker-climate-aggregate`
- **Region**: us-east1
- **Min Instances**: 0 (scales to zero)
- **Max Instances**: 3
- **Service Account**: `phosboard-runtime-sa@phosboard.iam.gserviceaccount.com`
- **URL**: https://worker-climate-aggregate-544990213867.us-east1.run.app

---

## Worker Social Probe (PHOS-29)

### Estructura
```
workers/social_probe/
├── cmd/worker/main.go
└── internal/
    ├── scraper/scraper.go     # Mock scraper (generates fake mentions)
    ├── publisher/publisher.go # Publishes to climate-aggregate
    ├── storage/minio.go       # Uploads to social-payloads
    └── subscriber/subscriber.go
```

### Flujo
1. Consume de `social-probe-sub`
2. Por cada query, ejecuta scraper mock (goroutines + WaitGroup)
3. Agrupa todas las menciones en un JSON
4. Sube a MinIO bucket `social-payloads` en ruta `mentions/{document_id}/{timestamp}.json`
5. Publica en `climate-aggregate` con document_id y minio_mentions_key

### Mock Scraper
- Genera 3-5 menciones ficticias por query
- Plataformas: twitter, bluesky, facebook, instagram
- Campos: text, author, date, platform, engagement_score
- Concurrente: usa goroutines con sync.WaitGroup

### Dependencies
- `cloud.google.com/go/pubsub` - Pub/Sub
- `github.com/minio/minio-go/v7` - MinIO
- `github.com/joho/godotenv` - Environment

---

## Worker Semantic Analyzer (PHOS-27, PHOS-28)

### Estructura
```
workers/semantic/
├── cmd/worker/main.go
└── internal/
    ├── analyzer/analyzer.go    # Vertex AI Gemini integration
    ├── publisher/publisher.go
    ├── repository/
    ├── storage/minio.go
    └── subscriber/subscriber.go
```

### Flujo Actualizado
1. Consume de `document-analyze-sub` (pull subscription)
2. Descarga texto plano desde GCS: `plain-text/{document_id}.txt` (ya limpio del scraper)
3. Obtiene conceptos activos desde `tenant_concepts`
4. Analiza con Vertex AI Gemini 1.5 Flash (PHOS-28: integración real)
5. Guarda resultado en `global_documents.semantic_analysis`
6. Si hay search_queries, publica en `social-probe`

### PHOS-28: Integración con Vertex AI Gemini
- **Cliente**: `cloud.google.com/go/vertexai/genai`
- **Modelo**: `gemini-2.5-flash`
- **Ubicación**: `us-central1` (configurable via `GOOGLE_LOCATION`)
- **Configuración**: ResponseMIMEType = "application/json" forzar respuesta JSON
- **Prompt**: Construido con conceptos activos y texto completo desde GCS (sin truncamiento)
- **Error handling**: Si falla la conexión a Vertex AI, el worker termina con `os.Exit(1)` (sin fallback)
- **Dependencias**: `cloud.google.com/go/vertexai v0.17.0`

### PHOS-34: Procesamiento de Texto (Simplificado)
- **Texto ya limpio**: Scraper envía texto plano sin HTML tags
- **Normalización**: Espacios múltiples reducidos a uno solo con `strings.Fields`
- **Sanitización UTF-8**: `strings.ToValidUTF8` reemplaza bytes inválidos con espacio
- **Sin límites artificiales**: Vertex AI soporta 1M tokens (~750k palabras)
- **Eficiencia**: No necesita limpieza HTML (ya hecho por scraper)

### PHOS-25: Upsert para evitar colisiones de URL
- Query con `ON CONFLICT (url)` para manejar duplicados
- Usa `uuid_generate_v4()` para el id
- Actualiza title, content_text y updated_at en conflicto

### Tests
- `internal/scraper/collector_test.go` - 3 tests (HTTPClient_Get, InvalidURL, Extractor_Extract)
- `internal/publisher/publisher_test.go` - 1 test (JSONMarshal)

### Plan de Testing para Pipeline Bronze/Silver

#### 1. Testing Unitario (Completado)
- ✅ **Discovery**: 4/4 tests pasando
- ✅ **Scraper**: 4/4 tests pasando
- ✅ **Builds**: Todos los workers compilan exitosamente

#### 2. Testing de Integración Local
- **Flujo end-to-end**: `backend → discovery → scraper → semantic`
- **Validación Bronze/Silver**: Almacenamiento en GCS y PostgreSQL
- **Mensajes entre workers**: Verificar uso correcto de `document_id`

#### 3. Testing de Deployment
- **Tags v2.0.1+**: Workflows solo se ejecutan en tags (condición `if: startsWith(github.ref, 'refs/tags/v')`)
- **Cloud Run services**: Health checks y logs
- **Integración GCP**: Pub/Sub, GCS, Secret Manager

#### 4. Escenarios de Testing
- **Documentos largos**: Sin límites artificiales
- **HTML complejo**: Limpieza efectiva en Silver layer
- **URLs múltiples**: Discovery con límite configurable
- **Recuperación de errores**: URLs inválidas, timeouts

### Errores y Soluciones

#### Error: Subscription does not exist
```
rpc error: code = NotFound desc = Subscription does not exist (resource=source-discovery-sub)
```
- **Causa**: Las suscripciones Pub/Sub no existían
- **Solución**: Ejecutar `cd backend && go run ./cmd/setup-pubsub`

#### Error: failed to connect to database (unix socket)
```
failed to connect to `user=samuel database=`: /private/tmp/.s.PGSQL.5432: dial error
```
- **Causa**: Los workers no cargaban el archivo `.env`
- **Solución**: Agregar `godotenv.Load(".env")` en `main.go` de cada worker

#### Error: invalid input syntax for type uuid
```
ERROR: invalid input syntax for type uuid: "manual" (SQLSTATE 22P02)
```
- **Causa**: Se pasaba "manual" (string) como source_id que requiere UUID
- **Solución**: Crear función `getOrCreateSource()` que obtiene o crea la fuente en la tabla `sources`

#### Error: no unique or exclusion constraint
```
ERROR: there is no unique or exclusion constraint matching the ON CONFLICT specification (SQLSTATE 42P10)
```
- **Causa**: La tabla `sources` no tiene restricción UNIQUE en la columna `name`
- **Solución**: Cambiar approach: SELECT primero, luego INSERT si no existe

#### Error: duplicate key value violates unique constraint (URL)
```
ERROR: duplicate key value violates unique constraint "global_documents_url_key"
```
- **Causa**: La tabla `global_documents` tiene constraint UNIQUE en `url`
- **Solución**: Implementar upsert con `ON CONFLICT (url) DO UPDATE SET` (PHOS-25)

#### Error: no unique or exclusion constraint matching the ON CONFLICT specification
```
ERROR: there is no unique or exclusion constraint matching the ON CONFLICT specification (SQLSTATE 42P10)
```
- **Causa**: Faltaba índice único en columna `url` de `global_documents`
- **Solución**: `CREATE UNIQUE INDEX idx_global_documents_url ON global_documents(url);`

#### Error: scanning User struct (RBAC Implementation)
```
ERROR: scanning User struct: cannot scan into dest[1]: cannot assign string to time.Time
```
- **Causa**: Los campos `CreatedAt` y `UpdatedAt` en struct `User` eran `string` pero PostgreSQL devuelve `timestamptz`
- **Solución**: Cambiar tipo de `string` a `time.Time` en `internal/repository/user_repository.go`

#### Error: Missing permission for tenant-admin
```
{"error":"unauthorized: user does not have permission to access this endpoint"}
```
- **Causa**: Faltaba permiso específico `POST /api/v1/documents/track` para tenant-admin
- **Solución**: Agregar permiso en migración `00007_seed_rbac_data.sql`

#### Error: Bcrypt hash mismatch
```
ERROR: invalid credentials (password mismatch)
```
- **Causa**: Los hashes bcrypt en la migración inicial eran incorrectos (no correspondían al password "password")
- **Solución**: Regenerar hashes bcrypt correctos con `bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)`

---

## Backend API Endpoints

### POST /api/v1/documents/track (PHOS-23, PHOS-35, RBAC)
- Permite inyección manual de URLs
- Valida formato de URL
- **RBAC**: Verifica permiso `POST /api/v1/documents/track`
- **Super-admin**: Requiere `tenant_id` en request body
- **Tenant-admin/editor**: Usa `tenant_id` del token JWT
- Método `TrackDocument` en repository con transacción
- Publica directamente en `url-scrape`

### Documents API (PHOS-33, RBAC)
- `GET /api/v1/documents` - Lista documentos con paginación
- **RBAC**: Verifica permiso `GET /api/v1/documents`
- Extrae tenant_id del JWT (aislamiento de tenant)
- **Super-admin**: Puede ver documentos de cualquier tenant (con tenant_id en query param)
- Query params: `limit` (default: 20, max: 100), `offset` (default: 0), `tenant_id` (solo super-admin)
- JOIN entre tenant_documents y global_documents
- Incluye semantic_analysis y social_temperature
- Formato de respuesta: `{"data": [...], "meta": {"total": ..., "limit": ..., "offset": ...}}`

### Concepts API (PHOS-26)
- `GET /api/v1/tenants/{tenant_id}/concepts` - Lista conceptos activos
- `POST /api/v1/tenants/{tenant_id}/concepts` - Crea concepto (body: {concept_term})
- `DELETE /api/v1/tenants/{tenant_id}/concepts/{concept_id}` - Soft delete

### Sources API (PHOS-40)
- `GET /api/v1/tenants/{tenant_id}/sources` - Lista fuentes
- `POST /api/v1/tenants/{tenant_id}/sources` - Crea fuente (body: {name, type, max_links})
- `GET /api/v1/tenants/{tenant_id}/sources/{source_id}` - Obtiene fuente
- `PUT /api/v1/tenants/{tenant_id}/sources/{source_id}` - Actualiza config (body: {config})
- `DELETE /api/v1/tenants/{tenant_id}/sources/{source_id}` - Elimina fuente

---

## Frontend Vue 3 (PHOS-4)

### Stack
- Vue 3 con Composition API (`<script setup lang="ts">`)
- Vite
- TypeScript
- TailwindCSS v4

### Estructura
```
frontend/
├── src/
│   ├── components/layouts/DashboardLayout.vue
│   ├── views/
│   ├── composables/
│   ├── types/
│   └── assets/
```

### DashboardLayout.vue
- Layout base con sidebar de navegación
- Items por defecto: Dashboard, Documents, Sources, Users, Settings

### src/types/index.ts
- Interfaces: Tenant, Role, TenantUser, Source, GlobalDocument, TenantDocument
- Tipo: SentimentScore (number con brand)
- Interfaz: DocumentAnalysis

### Dependencias
- `@tailwindcss/postcss` - Plugin de PostCSS para TailwindCSS

### Testing
- `vitest` - Test runner
- `@vue/test-utils` - Component testing
- `jsdom` - DOM environment

### Tests
- `components/layouts/DashboardLayout.test.ts` - 4 tests passing
- `types/index.test.ts` - 8 tests passing

### PHOS-36: Integración de Auth JWT y Refactor UI
- **Dependencias**: pinia, vue-router, axios
- **Plugin Axios**: `src/plugins/axios.ts` con interceptor para Authorization header
- **Store Pinia**: `src/stores/auth.ts` - Manejo de sesión con localStorage
- **Router**: `src/router/index.ts` - Navigation guard con requiresAuth
- **Login**: `src/views/Login.vue` - Formulario de autenticación
- **Dashboard**: `src/views/Dashboard.vue` - Vista protegida con listado de documentos

### PHOS-37: Rediseño del Dashboard con Navegación y UI de Vuetify
- **Dependencias**: vuetify, @mdi/font, vite-plugin-vuetify
- **Plugin Vuetify**: `src/plugins/vuetify.ts` - Configuración con tema custom
- **DefaultLayout**: `src/components/layouts/DefaultLayout.vue` - Layout con v-navigation-drawer, v-app-bar
- **Dashboard**: Vista con tarjetas estadísticas (Total Documentos, Temperatura Promedio, Conceptos)
- **Documents**: Vista con v-card, v-chip para polarity, v-progress-circular para temperature
- **Concepts**: Vista CRUD para conceptos semilla
- **Navegación**: Menú con iconos MDI (Dashboard, Documentos, Conceptos, Fuentes RSS, Ajustes)

### src/composables/useDocuments.ts
- fetchDocuments(tenantId) - Llama a API
- Estados: loading, error, documents

### src/components/DocumentCard.vue
- Muestra título, badge de fuente, URL truncada
- Diseño tarjeta con Tailwind

### src/components/DocumentList.vue
- Itera documentos
- Estados: loading (skeleton), error, empty, list

### src/views/DashboardView.vue
- Usa composable useDocuments
- Tenant ID de prueba: 00000000-0000-0000-0000-000000000001
- Proxy configurado en Vite para /api → localhost:8080

---

## Pub/Sub DAG Topology (PHOS-18)

### Topics
| Topic | Dead Letter |
|-------|-------------|
| `source-discovery` | `source-discovery-dead-letter` |
| `url-scrape` | `url-scrape-dead-letter` |
| `document-analyze` | `document-analyze-dead-letter` |
| `social-probe` | `social-probe-dead-letter` |
| `climate-aggregate` | `climate-aggregate-dead-letter` |

### Subscriptions
| Subscription | Topic | Type | Push Endpoint | Notas |
|--------------|-------|------|---------------|-------|
| `source-discovery-sub` | `source-discovery` | Push | `worker-discovery` | Worker discovery recibe mensajes HTTP |
| `url-scrape-sub` | `url-scrape` | Push | `worker-scraper` | Worker scraper recibe mensajes HTTP |
| `document-analyze-sub` | `document-analyze` | Push | `worker-semantic` | Worker semantic recibe mensajes HTTP |
| `social-probe-sub` | `social-probe` | Push | `worker-social-probe` | Worker social probe recibe mensajes HTTP |
| `climate-aggregate-sub` | `climate-aggregate` | Push | `worker-climate-aggregate` | Worker climate aggregate recibe mensajes HTTP |

**Configuración común:**
- DeadLetterPolicy: max 5 delivery attempts
- RetryPolicy: 10s minimum, 600s maximum backoff
- **Push endpoints**: Todos los workers usan push subscriptions para integración con Cloud Run

### Setup
```bash
cd backend && go run ./cmd/setup-pubsub
```

---

## MinIO Datalake (PHOS-19)

### Buckets
| Bucket | Purpose |
|--------|---------|
| `raw-html` | HTML scraped by scraper worker |
| `social-payloads` | JSON responses from social media APIs |
| `parquet-files` | Parquet files from RSS fetcher |

### Setup
```bash
cd backend && go run ./cmd/setup-minio
```

### Environment Variables
```
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET_RAW_HTML=raw-html
MINIO_BUCKET_SOCIAL_PAYLOADS=social-payloads
MINIO_BUCKET_PARQUET=parquet-files
```

---

## Tenant de Prueba

- **ID**: `85c5f582-86b1-4217-bd4a-e1b1d0aac195`
- **Nombre**: Muni La Serena
- **Creado**: PHOS-26

## Fuente de Prueba (PHOS-38)

- **Nombre**: Diario El Día
- **URL**: https://www.diarioeldia.cl/rss.xml
- **Intervalo**: 15 minutos
- **ID en DB**: 174ccac3-31d8-4fc6-ba50-fd36be96c1b7

---

## Pipeline Bronze/Silver (v2.0.0+)

### Arquitectura Actualizada
```
backend → discovery → scraper → semantic → (social-probe) → climate-aggregate
```

### Cambios Principales en v2.0.0

#### 1. **Consolidación de Workers**
- **Fetcher deprecado**: Scraper ahora maneja todo el fetching de contenido
- **Parser deprecado**: Patrón Bronze/Silver reemplaza procesamiento por lotes
- **MinIO eliminado**: Usando SDK nativo de GCS con autenticación ADC

#### 2. **Patrón Bronze/Silver Data Lake**
- **Bronze Layer**: HTML crudo almacenado en GCS (`raw-html/{document_id}.html`)
- **Silver Layer**: HTML limpio almacenado en GCS (`clean-html/{document_id}.html`)
- **PostgreSQL**: Solo texto plano (sin límites artificiales)

#### 3. **Flujo de Documentos**
1. **Discovery**: Crea documento en `global_documents` con `source_id` y URL
2. **Scraper**: Procesa URL, almacena Bronze/Silver, extrae texto, actualiza documento
3. **Semantic**: Analiza contenido con Vertex AI, actualiza `semantic_analysis`
4. **Social Probe**: Busca menciones en redes sociales (opcional)
5. **Climate Aggregate**: Calcula temperatura social

#### 4. **Sin Límites Artificiales**
- **Filosofía**: No imponer límites para early adopters
- **Vertex AI**: Soporta 1M tokens (~750k palabras)
- **PostgreSQL**: Columnas TEXT eficientes para documentos grandes
- **GCS**: Sin límites de tamaño para HTML crudo

### Mensajes entre Workers

#### Discovery → Scraper
```json
{
  "document_id": "uuid-from-global_documents",
  "url": "https://example.com/article"
}
```

#### Scraper → Semantic
```json
{
  "document_id": "uuid-from-global_documents",
  "url": "https://example.com/article"
}
```

### Variables de Entorno Actualizadas

#### Scraper Worker
```
DATABASE_URL=postgres://...
GOOGLE_PROJECT_ID=phosboard
GCS_BUCKET=phosboard-documents
```

#### Discovery Worker
```
DATABASE_URL=postgres://...
GOOGLE_PROJECT_ID=phosboard
```

## Cambios Recientes (Abril 2026)

### ✅ **Refactorización Completa v2.0.0+**
1. **Patrón Bronze/Silver Data Lake**: HTML crudo → HTML limpio → texto plano en GCS
2. **Consolidación de Workers**: Fetcher y parser deprecados, scraper unificado
3. **Eficiencia en PostgreSQL**: Guardar solo primeros 15,000 caracteres en DB
4. **Texto completo en GCS**: Vertex AI accede a texto completo desde GCS
5. **Discovery Worker mejorado**: Crea documentos antes de scraping, usa `document_id`
6. **CI/CD tag-only**: Workflows solo ejecutan en tags `v*` (evita errores WIF)

### ✅ **Sistema RBAC Completo (Abril 2026)**
1. **Roles**: super-admin (dueño), tenant-admin (cliente líder), tenant-editor (cuentas asociadas)
2. **Autorización**: Por endpoint + método HTTP (tabla `permissions` con `endpoint` y `method`)
3. **Eliminación de credenciales hardcodeadas**: Todas las credenciales removidas del backend
4. **Middleware RBAC**: `AuthWithAuthorization()` verifica permisos por endpoint
5. **Migración de datos**: `00006_rbac_system.sql` (schema) + `00007_seed_rbac_data.sql` (datos iniciales)
6. **Fix crítico**: Tipo de datos `CreatedAt`/`UpdatedAt` cambiados de `string` a `time.Time`
7. **Permisos faltantes**: Agregado `POST /api/v1/documents/track` para tenant-admin

### ✅ **Versiones Desplegadas**
- **Backend API**: v1.2.5 (RBAC completo, fix tipos timestamp, permisos faltantes)
- **Discovery**: v2.0.1 (crea documentos, publica con `document_id`)
- **Scraper**: v2.0.2 (Bronze/Silver pattern, 15k chars en DB, texto completo en GCS)
- **Semantic**: v2.0.4 (Vertex AI sin truncamiento, lee texto plano de GCS)
- **Climate Aggregate**: v2.0.1 (cálculo temperatura social)
- **Social Probe**: v2.0.1 (mock scraper, opcional)
- **Fetcher**: v2.0.1 (deprecated, reemplazado por scraper)

### ✅ **Pipeline Validado**
```
Backend (RBAC) → Discovery → Scraper (Bronze/Silver) → Semantic → Climate Aggregate
```

## Problemas Pendientes

### ⚠️ **Encoding de Texto en Scraper**
- **Problema**: El texto extraído sigue perdiendo Ñ, tildes y caracteres especiales
- **Ubicación**: `workers/scraper/internal/scraper/collector.go`
- **Estado**: Implementado encoding UTF-8 con `golang.org/x/net/html/charset` pero aún con problemas
- **Solución pendiente**: Investigar mejor detección de encoding o forzar UTF-8 en respuesta HTTP

## Próximos Pasos Sugeridos

1. ✅ **PHOS-40**: CRUD de Fuentes y Control de Límites en Discovery
2. ✅ **Bronze/Silver**: Implementación completa del patrón data lake
3. ✅ **RBAC System**: Sistema completo de control de acceso basado en roles
4. ⏳ **Encoding Fix**: Resolver problemas de Ñ, tildes y caracteres especiales
5. ⏳ **Testing**: Validar pipeline completo con testing end-to-end
6. ⏳ **Monitoreo**: Configurar observabilidad para el pipeline
7. ⏳ **Optimización**: Basada en datos de uso real
8. ⏳ **Logs OpenTelemetry**: Agregar logs estructurados para diagnóstico

---

## Cómo Ejecutar (v2.0.0+)

### Desarrollo Local
```bash
# 1. Levantar servicios dependientes
docker compose up -d postgres minio pubsub

# 2. Setup inicial
cd backend && go run ./cmd/setup-pubsub   # Configura topics/subscriptions
cd backend && go run ./cmd/setup-minio    # Configura buckets MinIO (solo desarrollo)

# 3. Variables de entorno para testing
export DATABASE_URL=postgres://phos_user:phos_password@localhost:5432/phosboard
export GOOGLE_PROJECT_ID=phosboard
export PUBSUB_EMULATOR_HOST=localhost:8085
export GCS_BUCKET=test-bucket  # Para pruebas con emulador

# 4. Ejecutar workers (en terminales separadas)
# Backend API
cd backend && go run ./cmd/api

# Discovery Worker (crea documentos)
cd workers/discovery && go run ./cmd/worker

# Scraper Worker (Bronze/Silver pattern)
cd workers/scraper && go run ./cmd/worker

# Semantic Worker (Vertex AI analysis)
cd workers/semantic && go run ./cmd/worker

# Frontend
cd frontend && npm run dev
```

### Testing del Pipeline
```bash
# 1. Publicar mensaje de prueba a discovery
curl -X POST http://localhost:8085/v1/projects/phosboard/topics/source-discovery:publish \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [{
      "data": "eyJzb3VyY2VfaWQiOiAiMTc0Y2NhYzMtMzFkOC00ZmM2LWJhNTAtZmQzNmJlOTZjMWI3IiwgInVybCI6ICJodHRwczovL3d3dy5kaWFyaW9lbGRpYS5jbC9yc3MueG1sIiwgInRpbWVzdGFtcCI6ICIyMDI0LTAxLTAxVDAwOjAwOjAwWiJ9"
    }]
  }'

# 2. Verificar flujo:
#    - Discovery crea documento en global_documents
#    - Discovery publica a scraper con document_id
#    - Scraper procesa URL, almacena Bronze/Silver
#    - Scraper publica a semantic
#    - Semantic analiza con Vertex AI
```

### Producción (Cloud Run)
```bash
# Los workers se despliegan automáticamente con tags v*
# Ejemplo para semantic worker:
git tag v2.0.3
git push origin v2.0.3

# Para otros workers (discovery, scraper, etc.):
git tag v2.0.1
git push origin v2.0.1

# Verificar servicios en Cloud Run
gcloud run services list --region=us-east1 --project=phosboard

# Health checks
curl https://worker-discovery-544990213867.us-east1.run.app/health
curl https://worker-scraper-544990213867.us-east1.run.app/health
curl https://worker-semantic-544990213867.us-east1.run.app/health
```

### Puertos
- 5432: PostgreSQL
- 9000: MinIO API
- 9001: MinIO Console
- 8085: Pub/Sub Emulator
- 8086: Pub/Sub Dashboard

---

## Git Workflow

### Ramas
```
main         # Producción - solo merges desde dev
dev          # Desarrollo estable - merges desde feature/fix
feature/*    # Nuevas funcionalidades
fix/*        # Correcciones de bugs
```

### Pre-push Hook
El backend tiene un hook de pre-push que corre golangci-lint antes de push:

```bash
# Ubicación: backend/.git/hooks/pre-push
# Require golangci-lint v2.5.0 instalado en ~/go/bin/
```

### Hacer un deploy

```bash
# 1. Asegurarse de estar en main
git checkout main

# 2. Crear tag
git tag v1.0.0

# 3. Push del tag (esto dispara el workflow)
git push origin v1.0.0

# 4. Ver progreso en GitHub Actions
# https://github.com/SamuelAnjel/phosboard-api/actions
```

---

---

## GitHub Actions CI/CD

### Workflow: build-deploy.yml
- **Trigger**: Al crear un tag `v*`
- **Pasos**:
  1. Lint con golangci-lint v2.5.0
  2. Security scan con Trivy
  3. Build imagen Docker
  4. Fetch secrets from GCP Secret Manager
  5. Push a Artifact Registry
  6. Deploy a Cloud Run

### Configuración GCP
- **Project ID**: phosboard
- **Project Number**: 544990213867
- **Artifact Registry**: us-east1-docker.pkg.dev/phosboard/phosboard-images
- **Cloud Run Region**: us-east1

---

## GCP Infrastructure Setup

### 1. Service Account y Workload Identity

#### Crear Service Account
```bash
gcloud iam service-accounts create github-actions-sa \
  --display-name="GitHub Actions Service Account" \
  --project=phosboard
```

#### Crear Workload Identity Pool
```bash
gcloud iam workload-identity-pools create github-pool \
  --project=phosboard \
  --location=global \
  --display-name="GitHub Actions Pool"
```

#### Crear Workload Identity Provider
```bash
gcloud iam workload-identity-pools providers create github-provider \
  --project=phosboard \
  --location=global \
  --workload-identity-pool=github-pool \
  --attribute-mapping="attribute.repository=assertion.repository" \
  --issuer-uri="https://token.actions.githubusercontent.com"
```

#### Agregar principal de GitHub al provider
```bash
gcloud iam workload-identity-pools principals add-iam-policy-binding \
  github-pool \
  --project=phosboard \
  --location=global \
  --principal-set="principalSet:attribute.repository=SamuelAnjel/phosboard-api" \
  --role="roles/iam.workloadIdentityUser"
```

#### Asignar roles a la service account
```bash
# Artifact Registry (escribir imágenes)
gcloud projects add-iam-policy-binding phosboard \
  --member="serviceAccount:github-actions-sa@phosboard.iam.gserviceaccount.com" \
  --role="roles/artifactregistry.writer"

# Cloud Run (desplegar servicios)
gcloud projects add-iam-policy-binding phosboard \
  --member="serviceAccount:github-actions-sa@phosboard.iam.gserviceaccount.com" \
  --role="roles/cloudrun.developer"

# Secret Manager (leer secretos)
gcloud projects add-iam-policy-binding phosboard \
  --member="serviceAccount:github-actions-sa@phosboard.iam.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

---

### 2. Secret Manager (Secrets)

#### Crear secretos
```bash
# DATABASE_URL
echo -n "postgresql://user:password@host:5432/db" | \
  gcloud secrets create phosboard-database-url \
  --data-file=- \
  --project=phosboard

# JWT_SECRET
echo -n "your-secret-jwt-key" | \
  gcloud secrets create phosboard-jwt-secret \
  --data-file=- \
  --project=phosboard
```

#### Versionar secretos (opcional)
```bash
# Actualizar versión
echo -n "new-password" | \
  gcloud secrets versions add phosboard-database-url \
  --data-file=-
```

---

### 3. Artifact Registry

#### Crear repositorio (si no existe)
```bash
gcloud artifacts repositories create phosboard-images \
  --repository-format=docker \
  --location=us-east1 \
  --project=phosboard
```

#### Autenticar Docker
```bash
gcloud auth configure-docker us-east1-docker.pkg.dev
```

---

### 4. Cloud Run

#### Deploy manual (para pruebas)
```bash
gcloud run deploy api-backend \
  --image=us-east1-docker.pkg.dev/phosboard/phosboard-images/api-backend:latest \
  --platform=managed \
  --region=us-east1 \
  --allow-unauthenticated \
  --project=phosboard
```

#### Ver logs
```bash
gcloud logs read --resource-type=cloud_run_revision \
  --filter="resource.serviceName=api-backend" \
  --limit=50
```

---

### 5. Database (Supabase PostgreSQL)

#### Connection String
```
postgresql://postgres.ohrmoiplfblbzstpgpxn:PASSWORD@aws-1-us-east-1.pooler.supabase.com:5432/postgres
```

#### Migraciones
```bash
cd cmd/migrate
DATABASE_URL="postgresql://postgres.ohrmoiplfblbzstpgpxn:PASSWORD@aws-1-us-east-1.pooler.supabase.com:5432/postgres" \
  go run main.go
```

---

### 6. GitHub Actions Workflow

```yaml
name: Build and Deploy to GCP

on:
  push:
    tags:
      - 'v*'

env:
  GCP_PROJECT_ID: phosboard
  GCP_PROJECT_NUMBER: '544990213867'
  WORKLOAD_IDENTITY_POOL: github-pool
  WORKLOAD_IDENTITY_PROVIDER: github-provider
  SERVICE_ACCOUNT: github-actions-sa@phosboard.iam.gserviceaccount.com
  GAR_LOCATION: us-east1
  GAR_REPO: phosboard-images

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - uses: golangci/golangci-lint-action@v7
        with:
          version: 'v2.5.0'

  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'
          format: 'sarif'
          output: 'trivy-results.sarif'
      - name: Fail if critical
        run: |
          if grep -q '"Severity":"CRITICAL"' trivy-results.sarif; then
            echo "Critical vulnerabilities found!"
            exit 1
          fi

  build:
    needs: [lint, security]
    runs-on: ubuntu-latest
    permissions:
      contents: 'read'
      id-token: 'write'

    steps:
      - uses: actions/checkout@v4

      - name: Authenticate to GCP
        uses: google-github-actions/auth@v2
        with:
          workload_identity_provider: 'projects/${{ env.GCP_PROJECT_NUMBER }}/locations/global/workloadIdentityPools/${{ env.WORKLOAD_IDENTITY_POOL }}/providers/${{ env.WORKLOAD_IDENTITY_PROVIDER }}'
          service_account: '${{ env.SERVICE_ACCOUNT }}'

      - name: Fetch secrets
        uses: google-github-actions/get-secretmanager-secrets@v2
        with:
          secrets: |
            DATABASE_URL:projects/${{ env.GCP_PROJECT_NUMBER }}/secrets/phosboard-database-url/versions/latest
            JWT_SECRET:projects/${{ env.GCP_PROJECT_NUMBER }}/secrets/phosboard-jwt-secret/versions/latest
          export_to_environment: true

      - uses: docker/setup-buildx-action@v3

      - name: Configure Docker
        run: |
          gcloud auth configure-docker ${{ env.GAR_LOCATION }}-docker.pkg.dev

      - name: Build and Push
        uses: docker/build-push-action@v6
        with:
          context: .
          push: true
          tags: ${{ env.GAR_LOCATION }}-docker.pkg.dev/${{ env.GCP_PROJECT_ID }}/${{ env.GAR_REPO }}/api-backend:${{ github.ref_name }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

      - name: Deploy to Cloud Run
        uses: google-github-actions/deploy-cloudrun@v2
        with:
          service: api-backend
          image: ${{ env.GAR_LOCATION }}-docker.pkg.dev/${{ env.GCP_PROJECT_ID }}/${{ env.GAR_REPO }}/api-backend:${{ github.ref_name }}
          region: ${{ env.GAR_LOCATION }}
          flags: '--ingress=all'
          env_vars: |
            DATABASE_URL=${{ env.DATABASE_URL }}
            JWT_SECRET=${{ env.JWT_SECRET }}
            GOOGLE_PROJECT_ID=${{ env.GCP_PROJECT_ID }}
```

---

### 7. Cómo hacer un deploy

```bash
# 1. Asegurarse de estar en main
git checkout main

# 2. Crear tag
git tag v1.0.0

# 3. Push del tag (esto dispara el workflow)
git push origin v1.0.0

# 4. Ver progreso en GitHub Actions
# https://github.com/SamuelAnjel/phosboard-api/actions
```

---

### 8. Variables de entorno en producción (v2.0.0+ con RBAC)

| Variable | Fuente | Descripción | Workers que la usan |
|----------|--------|-------------|---------------------|
| **Comunes** | | | |
| DATABASE_URL | Secret Manager | Connection string de PostgreSQL | Todos |
| GOOGLE_PROJECT_ID | Env | ID del proyecto GCP (phosboard) | Todos |
| JWT_SECRET | Secret Manager | Secret para firmar tokens JWT (RBAC) | Backend API |
| **Específicas por Worker** | | | |
| GCS_BUCKET | Env | Bucket de Google Cloud Storage | Scraper, Social Probe, Climate Aggregate |
| GOOGLE_LOCATION | Env | Ubicación de Vertex AI (us-central1) | Semantic |
| **Deprecadas** | | | |
| MINIO_ENDPOINT | Secret Manager | Endpoint de MinIO/S3 (solo desarrollo) | ⚠️ Deprecado |
| MINIO_ACCESS_KEY | Secret Manager | Access key de MinIO (solo desarrollo) | ⚠️ Deprecado |
| MINIO_SECRET_KEY | Secret Manager | Secret key de MinIO (solo desarrollo) | ⚠️ Deprecado |
| **Cloud Run** | | | |
| PORT | Cloud Run | Automáticamente configurado a 8080 | Todos |
| **RBAC** | | | |
| ✅ **Sistema completo**: Sin credenciales hardcodeadas, autenticación por base de datos |

### 9. Secrets necesarios para workers (con RBAC)

```bash
# MINIO_ENDPOINT (deprecated, solo desarrollo)
echo -n "storage.googleapis.com" | \
  gcloud secrets create phosboard-minio-endpoint \
  --data-file=- \
  --project=phosboard

# MINIO_ACCESS_KEY (deprecated, solo desarrollo)
echo -n "minioadmin" | \
  gcloud secrets create phosboard-minio-access-key \
  --data-file=- \
  --project=phosboard

# MINIO_SECRET_KEY (deprecated, solo desarrollo)
echo -n "minioadmin" | \
  gcloud secrets create phosboard-minio-secret-key \
  --data-file=- \
  --project=phosboard

# JWT_SECRET (RBAC - requerido)
echo -n "your-secure-jwt-secret-key-here" | \
  gcloud secrets create phosboard-jwt-secret \
  --data-file=- \
  --project=phosboard
```

### 10. Repositorios de Workers (v2.0.1+)

## Estado Actual del Sistema (Abril 2026)

### ✅ **Sistema RBAC Completamente Implementado**
- **Autenticación**: Login con base de datos (bcrypt), sin credenciales hardcodeadas
- **Autorización**: Permisos por endpoint + método HTTP (tabla `permissions`)
- **Roles**: super-admin, tenant-admin, tenant-editor con permisos específicos
- **Middleware**: `AuthWithAuthorization()` verifica permisos antes de cada request
- **Document handler**: Maneja super-admin (requiere tenant_id) vs tenant users
- **Fix crítico**: Tipo de datos `CreatedAt`/`UpdatedAt` cambiados de `string` a `time.Time`
- **Permisos faltantes**: Agregado `POST /api/v1/documents/track` para tenant-admin

### ✅ **Pipeline de Documentos Funcional**
1. **Backend API (RBAC)**: Autenticación + autorización + track manual de documentos
2. **Discovery Worker**: Crea documentos en `global_documents` antes de scraping
3. **Scraper Worker**: Patrón Bronze/Silver (HTML crudo → HTML limpio → texto plano)
4. **Semantic Worker**: Vertex AI Gemini 1.5 Flash para análisis semántico
5. **Climate Aggregate**: Cálculo de temperatura social basado en menciones

### ✅ **Cloud Run Deployment**
- **Servicio**: `api-backend` en región `us-east1`
- **Revisión actual**: `api-backend-00035-wq8` (ejecutando RBAC)
- **Service account**: `phosboard-api-backend-sa@phosboard.iam.gserviceaccount.com`
- **Permisos**: `roles/pubsub.publisher` agregados para publicación de mensajes

### ✅ **Git Tags Desplegados**
- `v1.2.4`: Fix de tipos timestamp en User struct
- `v1.2.5`: Permiso faltante `POST /api/v1/documents/track` agregado
- Build en progreso para deployment final con RBAC completo

### ⚠️ **Problemas Conocidos**
1. **Serialización JSON**: Algunas respuestas muestran `field: string` en terminal pero JSON real está correcto (ver con `curl -i`)
2. **Pub/Sub permisos**: Error "check topic exists" en algunos workers pero no crítico
3. **Visualización**: Problema de visualización en terminal vs respuesta HTTP real

### 📋 **Próximos Pasos Inmediatos**
1. Agregar logs OpenTelemetry para mejor diagnóstico
2. Verificar/arreglar permisos de Pub/Sub en workers
3. Posible fix para problema de visualización de serialización JSON
4. Testing completo del pipeline RBAC con tenant real

Cada worker tiene su propio repositorio de GitHub con workflow de CI/CD que solo se ejecuta en tags:

| Worker | Repositorio | Versión | Status | Notas |
|--------|-------------|---------|--------|-------|
| **Pipeline Principal** | | | | |
| discovery | `SamuelAnjel/phosboard-worker-discovery` | `v2.0.1` | ✅ Producción | Crea documentos, publica con `document_id` |
| scraper | `SamuelAnjel/phosboard-worker-scraper` | `v2.0.2` | ✅ Producción | Patrón Bronze/Silver, 15k chars en DB, texto completo en GCS |
| semantic | `SamuelAnjel/phosboard-worker-semantic` | `v2.0.4` | ✅ Producción | Vertex AI Gemini analysis, lee texto plano de GCS |
| **Workers Adicionales** | | | | |
| climate_aggregate | `SamuelAnjel/phosboard-worker-climate-aggregate` | `v2.0.1` | ✅ Producción | Calcula temperatura social |
| social_probe | `SamuelAnjel/phosboard-worker-social-probe` | `v2.0.1` | ✅ Producción | Mock scraper para redes sociales |
| **Workers Deprecados** | | | | |
| fetcher | `SamuelAnjel/phosboard-worker-fetcher` | `v2.0.1` | ⚠️ Deprecado | Reemplazado por scraper |
| parser | No implementado | N/A | ❌ Deprecado | Reemplazado por Bronze/Silver |

**Política de CI/CD**:
- **Solo tags generan imágenes**: Todos los workflows tienen condición `if: startsWith(github.ref, 'refs/tags/v')`
- **`main` branch solo para desarrollo**: No genera deployments automáticos
- **Tags `v2.0.1`**: Incluyen fix para workflows tag-only y actualizaciones Bronze/Silver
- **Tag `v2.0.2`**: Scraper con optimización de almacenamiento (15k chars en DB)
- **Tag `v2.0.4`**: Semantic worker lee texto plano de GCS

**Política de versiones (SemVer)**:
- `0.x.x`: Desarrollo inicial, API inestable
- `1.0.0`: Primera versión estable de producción
- `x.y.z`: Incrementos según cambios (MAJOR.MINOR.PATCH)

#### Workflow de climate_aggregate

```yaml
# Trigger: push a main o tags v*
# Jobs: lint (golangci-lint) → security (trivy) → build (docker + deploy)
# Service: worker-climate-aggregate
# Image: us-east1-docker.pkg.dev/phosboard/phosboard-images/worker-climate-aggregate
# Flags: --min-instances=0 --max-instances=3 --service-account=phosboard-runtime-sa
# Env vars: DATABASE_URL, GOOGLE_PROJECT_ID, PUBSUB_EMULATOR_HOST
```

**Nota**: El workflow centralizado `.github/workflows/deploy-workers.yml` del monorepo está **deprecado** en favor de repositorios individuales por worker.

### 11. Runtime Service Account para Workers

Los workers de Cloud Run necesitan una service account para acceder a los recursos de GCP:

```bash
export PROJECT_ID="phosboard"
export SA_NAME="phosboard-runtime-sa"
export SA_EMAIL="${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"

# 1. Crear la Service Account
gcloud iam service-accounts create ${SA_NAME} \
  --display-name="Phosboard Runtime Service Account"

# 2. Darle permisos de administración sobre el bucket de documentos
gcloud storage buckets add-iam-policy-binding gs://phosboard-documents \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/storage.objectAdmin"
```

#### Permisos necesarios para Runtime SA

| Rol | Descripción |
|-----|-------------|
| `roles/artifactregistry.reader` | Leer imágenes de Artifact Registry |
| `roles/pubsub.subscriber` | Consumir mensajes de Pub/Sub |
| `roles/pubsub.publisher` | Publicar mensajes a Pub/Sub |
| `roles/storage.objectAdmin` | Admin sobre bucket GCS (phosboard-documents) |
| `roles/secretmanager.secretAccessor` | Leer secretos de Secret Manager |
| `roles/cloudsql.client` | Conectarse a Cloud SQL |

**Comandos para asignar permisos**:

```bash
export PROJECT_ID="phosboard"
export SA_EMAIL="phosboard-runtime-sa@${PROJECT_ID}.iam.gserviceaccount.com"

# Artifact Registry (leer imágenes)
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/artifactregistry.reader"

# Pub/Sub (subscriber + publisher)
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/pubsub.subscriber"

gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/pubsub.publisher"

# Cloud Storage (admin sobre bucket GCS)
gcloud storage buckets add-iam-policy-binding gs://phosboard-documents \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/storage.objectAdmin"

# Secret Manager (leer secretos)
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/secretmanager.secretAccessor"

# Cloud SQL (conectar a DB)
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/cloudsql.client"
```

---

## Desplegar Workers a Cloud Run

Todos los workers (discovery, scraper, semantic, social_probe, climate_aggregate, fetcher, parser) también se despliegan como servicios de Cloud Run.

### Implementación del Worker Social Probe

**Estado**: ✅ Implementado (repositorio: `SamuelAnjel/phosboard-worker-social-probe`, tag: `v0.1.0`)

**Arquitectura**: Cloud Run Service + Pub/Sub Push (mismo patrón que discovery)

**Funcionalidad**:
- Recibe tareas del topic `social-probe` via Pub/Sub Push
- Procesa búsquedas en redes sociales (mock implementation)
- Guarda resultados en GCS (`mentions/{document_id}/{timestamp}.json`)
- Publica a `climate-aggregate` topic para procesamiento posterior

**Características implementadas**:
1. **Handler HTTP para Pub/Sub Push** (`internal/handler/pubsub.go`)
2. **Publisher con lazy initialization** (sin checks de topic durante startup)
3. **Workflow GitHub Actions** (deploy automático en tags `v*`)
4. **Husky pre-commit hooks** con golangci-lint
5. **Mock scraper** para desarrollo/testing
6. **Health endpoint** en `/health`

**Nota importante**: El worker usa un **MockScraper** que devuelve datos de prueba. Para producción, se necesita implementar un scraper real que consulte APIs de redes sociales.

**Próximos pasos (pendientes)**:
1. Configurar subscription Push en GCP (`social-probe-sub`)
2. Testear pipeline completo (social-probe → climate-aggregate)
3. Implementar scraper real para redes sociales

### Dockerfiles para Workers

Cada worker tiene su propio Dockerfile. Ejemplo para discovery worker:

```dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /worker ./cmd/worker

FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /worker .

CMD ["./worker"]
```

### Servicios de Cloud Run

| Servicio | Descripción | Status | URL |
|----------|-------------|--------|-----|
| api-backend | API REST principal | ✅ | `https://api-backend-544990213867.us-east1.run.app` |
| worker-fetcher | Consume fetcher-tasks, fetch RSS feeds y guarda en GCS (v2.0.0) | ✅ | `https://worker-fetcher-544990213867.us-east1.run.app` |
| worker-scraper | Consume url-scrape, hace scraping de páginas | ✅ | `https://worker-scraper-544990213867.us-east1.run.app` |
| worker-semantic | Consume document-analyze, analiza con Vertex AI (v2.0.0) | ✅ | `https://worker-semantic-544990213867.us-east1.run.app` |
| worker-climate-aggregate | Consume climate-aggregate, calcula temperatura (v2.0.0) | ✅ | `https://worker-climate-aggregate-544990213867.us-east1.run.app` |
| worker-discovery | Consume source-discovery, descubre URLs de RSS | ⏳ | TBD |
| worker-social-probe | Consume social-probe, busca menciones sociales | ⏳ | TBD |

### Notas importantes

1. **Todos los servicios en Cloud Run**: No hay más ejecución local de workers
2. **Secrets compartidos**: Todos los workers leen DATABASE_URL del Secret Manager
3. **Pub/Sub**: Los workers consumen mensajes de los topics de GCP Pub/Sub
4. **Escalabilidad**: Cloud Run escala automáticamente según demanda
5. **Costos**: Solo paga por lo que usa (requests × duración)
---

## Worker: Fetcher (v2.0.0 - Cloud Run Service)

**GitHub Repository**: https://github.com/SamuelAnjel/phosboard-worker-fetcher

**Deployment**: Cloud Run Service (not Jobs)

**Architecture Pattern**: HTTP server receiving Pub/Sub Push messages

### Overview

RSS feed fetcher worker that retrieves content from configured sources and stores it as Parquet files in Google Cloud Storage. This worker was refactored from Pub/Sub Pull subscription to Cloud Run Service with Pub/Sub Push.

### Key Changes from Original Design

1. **Pull → Push Migration**: Converted from Pull subscription to HTTP server
2. **Database Integration**: Fetches source configuration from PostgreSQL instead of hardcoded values
3. **MinIO → GCS**: Migrated from MinIO to Google Cloud Storage with `parquet-files/` prefix
4. **Dynamic Sources**: Removed hardcoded RSS feed URL (`https://www.diarioeldia.cl/rss.xml`)

### Cloud Resources

| Resource | Name | Description |
|----------|------|-------------|
| Cloud Run Service | `worker-fetcher` | HTTP endpoint for Pub/Sub push |
| Pub/Sub Topic | `fetcher-tasks` | Topic for triggering fetches |
| Pub/Sub Subscription | `fetcher-sub` | Push subscription to Cloud Run |
| Service URL | `https://worker-fetcher-544990213867.us-east1.run.app` | Cloud Run endpoint |
| GCS Bucket | `phosboard-documents` | Parquet file storage |
| GCS Prefix | `parquet-files/` | Prefix for all fetcher outputs |

### Configuration

**Environment Variables**:
- `DATABASE_URL` (secret): PostgreSQL connection string
- `GCS_BUCKET` (env var): `phosboard-documents`
- `PORT` (default: `8080`): HTTP server port

**Cloud Run Settings**:
- Min instances: 0 (scales to zero)
- Max instances: 10
- Memory: 512Mi
- CPU: 1
- Timeout: 300s
- Service Account: `phosboard-runtime-sa@phosboard.iam.gserviceaccount.com`

### Database Schema

Queries the `sources` table:

```sql
SELECT id, name, url, fetch_strategy
FROM sources
WHERE id = $1
```

Updates `last_run_at` after successful execution:

```sql
UPDATE sources
SET last_run_at = NOW()
WHERE id = $1
```

### Message Format

Pub/Sub messages must contain:

```json
{
  "source_id": "uuid-of-source-in-database"
}
```

### Flow

1. API backend publishes message to `fetcher-tasks` topic with `source_id`
2. Pub/Sub pushes message to Cloud Run service via HTTP POST to `/`
3. Worker queries database for source configuration (URL, strategy)
4. Worker fetches RSS feed using configured URL
5. Worker writes feed items to Parquet file
6. Worker uploads to GCS: `parquet-files/<timestamp>_<filename>.parquet`
7. Worker updates `last_run_at` timestamp
8. Worker responds HTTP 200 (ACK) or HTTP 500 (retry)

### IAM Permissions

**phosboard-runtime-sa** needs:
- `roles/run.invoker` (on worker-fetcher service)
- `roles/storage.objectAdmin` (on phosboard-documents bucket)
- `roles/secretmanager.secretAccessor`
- `roles/cloudsql.client`

**Pub/Sub service account** needs:
- `roles/iam.serviceAccountTokenCreator` (on phosboard-runtime-sa) - for OIDC tokens

### Deployment

Automated via GitHub Actions on push to `main` or version tags.

**Manual Deployment**:

```bash
# Build image for amd64 (required for Cloud Run)
docker buildx build --platform linux/amd64 \
  -t us-east1-docker.pkg.dev/phosboard/phosboard-images/worker-fetcher:v1.0.0 \
  --push .

# Deploy to Cloud Run
gcloud run deploy worker-fetcher \
  --image=us-east1-docker.pkg.dev/phosboard/phosboard-images/worker-fetcher:v1.0.0 \
  --platform=managed \
  --region=us-east1 \
  --service-account=phosboard-runtime-sa@phosboard.iam.gserviceaccount.com \
  --set-env-vars="GCS_BUCKET=phosboard-documents" \
  --set-secrets="DATABASE_URL=phosboard-database-url:latest" \
  --min-instances=0 \
  --max-instances=10 \
  --memory=512Mi \
  --cpu=1 \
  --timeout=300 \
  --no-allow-unauthenticated
```

**Configure Pub/Sub Subscription**:

```bash
# Create subscription with push endpoint
gcloud pubsub subscriptions create fetcher-sub \
  --topic=fetcher-tasks \
  --push-endpoint=https://worker-fetcher-544990213867.us-east1.run.app/ \
  --push-auth-service-account=phosboard-runtime-sa@phosboard.iam.gserviceaccount.com

# Grant invoker role
gcloud run services add-iam-policy-binding worker-fetcher \
  --member="serviceAccount:phosboard-runtime-sa@phosboard.iam.gserviceaccount.com" \
  --role="roles/run.invoker" \
  --region=us-east1

# Grant token creator role for OIDC
gcloud iam service-accounts add-iam-policy-binding phosboard-runtime-sa@phosboard.iam.gserviceaccount.com \
  --member="serviceAccount:service-544990213867@gcp-sa-pubsub.iam.gserviceaccount.com" \
  --role="roles/iam.serviceAccountTokenCreator"
```

### Testing

```bash
# Publish test message (use a real source_id from database)
gcloud pubsub topics publish fetcher-tasks --message='{"source_id":"YOUR-UUID-HERE"}'

# View logs
gcloud run services logs read worker-fetcher --region=us-east1 --limit=50

# View detailed logs
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=worker-fetcher" --limit=20 --format=json
```

### File Structure

```
workers/fetcher/
├── .github/
│   └── workflows/
│       └── build-deploy.yml          # CI/CD workflow
├── cmd/
│   └── worker/
│       └── main.go                   # HTTP server (port 8080)
├── internal/
│   ├── handler/
│   │   └── pubsub.go                 # HTTP handler for Pub/Sub push
│   ├── repository/
│   │   └── source_repository.go     # Database queries
│   ├── fetcher/
│   │   └── rss_collector.go         # RSS feed parsing (gofeed)
│   ├── storage/
│   │   ├── gcs.go                   # GCS upload (parquet-files/)
│   │   └── parquet_writer.go        # Parquet file writing
│   └── models/
│       └── raw_payload.go            # Data structure
├── Dockerfile                        # Multi-stage build (amd64)
├── go.mod                            # Dependencies (pgx, gofeed, parquet-go)
├── .gitignore
└── README.md
```

### Dependencies

- `cloud.google.com/go/storage` - GCS uploads
- `github.com/jackc/pgx/v5` - PostgreSQL driver
- `github.com/mmcdole/gofeed` - RSS feed parsing
- `github.com/parquet-go/parquet-go` - Parquet file writing
- `github.com/joho/godotenv` - Environment variables

### Monitoring

- **Invocations**: Cloud Run metrics show request count
- **Latency**: Average execution time per request
- **Errors**: 5xx responses indicate failures (triggers Pub/Sub retry)
- **Logs**: JSON structured logs via slog

### Troubleshooting

**403 Forbidden on Push Endpoint**:
- Ensure `service-{PROJECT_NUMBER}@gcp-sa-pubsub.iam.gserviceaccount.com` has `roles/iam.serviceAccountTokenCreator` on runtime SA
- Ensure runtime SA has `roles/run.invoker` on the service

**500 Internal Server Error**:
- Check logs for database connection issues
- Verify `DATABASE_URL` secret exists: `phosboard-database-url`
- Verify source_id is a valid UUID in the database

**No Logs Appearing**:
- Check subscription configuration: `gcloud pubsub subscriptions describe fetcher-sub`
- Verify topic has messages: `gcloud pubsub topics list`
- Check IAM permissions

---

## Worker: Semantic Analyzer (v2.0.0 - Cloud Run Service)

**GitHub Repository**: https://github.com/SamuelAnjel/phosboard-worker-semantic

**Deployment**: Cloud Run Service con Pub/Sub Push

**Arquitectura**: HTTP server que recibe mensajes Push de Pub/Sub para análisis semántico con Vertex AI Gemini

### Overview

Worker de análisis semántico que procesa documentos usando Vertex AI Gemini 1.5 Flash. Refactorizado desde Pull subscription a Cloud Run Service con Pub/Sub Push.

### Key Changes from Original Design

1. **Pull → Push Migration**: Convertido a HTTP server para recibir mensajes Push
2. **MinIO → GCS**: Migrado de MinIO a Google Cloud Storage
3. **Vertex AI Integration**: Análisis real con Gemini 1.5 Flash (PHOS-28)
4. **Cloud Run Service**: Escala automáticamente según demanda

### Cloud Resources

| Resource | Nombre | Descripción |
|----------|--------|-------------|
| Cloud Run Service | `worker-semantic` | Endpoint HTTP para Pub/Sub push |
| Pub/Sub Topic | `document-analyze` | Topic para análisis de documentos |
| Pub/Sub Subscription | `document-analyze-sub` | Push subscription a Cloud Run |
| Service URL | `https://worker-semantic-544990213867.us-east1.run.app` | Cloud Run endpoint |
| GCS Bucket | `phosboard-documents` | Almacenamiento de documentos |
| Vertex AI Location | `us-central1` | Ubicación del modelo Gemini |

### Configuration

**Environment Variables**:
- `DATABASE_URL` (secret): PostgreSQL connection string
- `GOOGLE_PROJECT_ID` (env var): `phosboard`
- `GOOGLE_LOCATION` (env var): `us-central1` (Vertex AI)
- `GCS_BUCKET` (env var): `phosboard-documents`
- `PORT` (default: `8080`): HTTP server port

**Cloud Run Settings**:
- Min instances: 0 (scales to zero)
- Max instances: 3
- Memory: 512Mi
- CPU: 1
- Timeout: 300s
- Service Account: `phosboard-runtime-sa@phosboard.iam.gserviceaccount.com`
- Authentication: `--no-allow-unauthenticated` (solo Pub/Sub)

### Flujo

1. Scraper worker publica mensaje en `document-analyze` topic
2. Pub/Sub push envía HTTP POST a Cloud Run service `/`
3. Worker descarga HTML desde GCS (`raw-html/` bucket)
4. Worker obtiene conceptos activos desde `tenant_concepts`
5. Worker analiza con Vertex AI Gemini 1.5 Flash
6. Worker guarda resultado en `global_documents.semantic_analysis`
7. Si hay search_queries, publica en `social-probe`
8. Worker responde HTTP 200 (ACK) o HTTP 500 (retry)

### Vertex AI Integration (PHOS-28)

- **Modelo**: `gemini-2.5-flash`
- **Ubicación**: `us-central1`
- **Configuración**: ResponseMIMEType = "application/json" para forzar JSON
- **Prompt**: Construido con conceptos activos y content_text completo (sin truncamiento)
- **Error handling**: Si falla conexión a Vertex AI, worker termina con `os.Exit(1)`

### GitHub Actions CI/CD

**Workflow**: `.github/workflows/build-deploy.yml`

**Trigger**: Push a `main` branch o tags `v*`

**Jobs**:
1. **lint**: golangci-lint v2.5.0 (con staticcheck deshabilitado)
2. **security**: Trivy scan
3. **build**: Docker build + push + deploy

**Workload Identity Federation**:
- Service Account: `github-actions-sa@phosboard.iam.gserviceaccount.com`
- Binding: `principalSet://.../attribute.repository/SamuelAnjel/phosboard-worker-semantic`
- **Lección aprendida**: Solo validar por repositorio, sin condiciones de `ref_type`

### IAM Permissions

**github-actions-sa** necesita:
- `roles/iam.serviceAccountTokenCreator`
- `roles/secretmanager.secretAccessor`
- `roles/artifactregistry.writer`
- `roles/run.developer`
- `roles/iam.workloadIdentityUser` (con binding específico por repositorio)

**phosboard-runtime-sa** necesita:
- `roles/run.invoker` (para Pub/Sub push)
- `roles/storage.objectAdmin` (en bucket phosboard-documents)
- `roles/secretmanager.secretAccessor`
- `roles/cloudsql.client`
- `roles/aiplatform.user` (para Vertex AI)

### Deployment

**Manual Deployment**:
```bash
# Build image for amd64
docker buildx build --platform linux/amd64 \
  -t us-east1-docker.pkg.dev/phosboard/phosboard-images/worker-semantic:v1.0.0 \
  --push .

# Deploy to Cloud Run
gcloud run deploy worker-semantic \
  --image=us-east1-docker.pkg.dev/phosboard/phosboard-images/worker-semantic:v1.0.0 \
  --platform=managed \
  --region=us-east1 \
  --service-account=phosboard-runtime-sa@phosboard.iam.gserviceaccount.com \
  --set-env-vars="GOOGLE_PROJECT_ID=phosboard,GOOGLE_LOCATION=us-central1,GCS_BUCKET=phosboard-documents" \
  --set-secrets="DATABASE_URL=phosboard-database-url:latest" \
  --min-instances=0 \
  --max-instances=3 \
  --memory=512Mi \
  --cpu=1 \
  --timeout=300 \
  --no-allow-unauthenticated
```

**Configure Pub/Sub Subscription**:
```bash
# Create subscription with push endpoint
gcloud pubsub subscriptions create document-analyze-sub \
  --topic=document-analyze \
  --push-endpoint=https://worker-semantic-544990213867.us-east1.run.app/ \
  --push-auth-service-account=phosboard-runtime-sa@phosboard.iam.gserviceaccount.com

# Grant invoker role
gcloud run services add-iam-policy-binding worker-semantic \
  --member="serviceAccount:phosboard-runtime-sa@phosboard.iam.gserviceaccount.com" \
  --role="roles/run.invoker" \
  --region=us-east1
```

### Workload Identity Federation - Lecciones Aprendidas

**Problema**: Error `Permission 'iam.serviceAccounts.getAccessToken' denied`

**Causa**: Condición `assertion.ref_type=="branch"` en binding de Workload Identity

**Solución**: Solo validar por repositorio sin condiciones de `ref_type`

**Binding correcto**:
```
principalSet://iam.googleapis.com/projects/544990213867/locations/global/workloadIdentityPools/github-pool/attribute.repository/SamuelAnjel/phosboard-worker-semantic
```

**Explicación**:
- GitHub Actions ejecuta workflows en:
  - **Branches**: `ref_type = "branch"`
  - **Tags**: `ref_type = "tag"`
- Si el workflow solo se ejecuta en tags (`tags: - 'v*'`) pero el binding requiere `ref_type=="branch"`, falla
- **Solución**: Remover condición de `ref_type`, solo validar repositorio

**Comando para agregar binding**:
```bash
gcloud iam service-accounts add-iam-policy-binding \
  github-actions-sa@phosboard.iam.gserviceaccount.com \
  --project=phosboard \
  --role="roles/iam.workloadIdentityUser" \
  --member="principalSet://iam.googleapis.com/projects/544990213867/locations/global/workloadIdentityPools/github-pool/attribute.repository/SamuelAnjel/phosboard-worker-semantic"
```

### File Structure

```
workers/semantic/
├── .github/
│   └── workflows/
│       └── build-deploy.yml          # CI/CD workflow
├── cmd/
│   └── worker/
│       └── main.go                   # HTTP server (port 8080)
├── internal/
│   ├── handler/
│   │   └── pubsub.go                 # HTTP handler for Pub/Sub push
│   ├── analyzer/
│   │   └── analyzer.go               # Vertex AI Gemini integration
│   ├── publisher/
│   │   └── publisher.go              # Publishes to social-probe
│   ├── repository/
│   │   └── document_repository.go    # Database operations
│   └── storage/
│       └── gcs.go                    # GCS download (raw-html/)
├── Dockerfile                        # Multi-stage build (amd64)
├── go.mod                            # Dependencies (pgx, vertexai)
├── .golangci.yml                     # Lint config (disables staticcheck)
└── README.md                         # Deployment instructions
```

### Dependencies

- `cloud.google.com/go/vertexai` - Vertex AI Gemini
- `cloud.google.com/go/storage` - GCS downloads
- `github.com/jackc/pgx/v5` - PostgreSQL driver
- `github.com/joho/godotenv` - Environment variables
- `github.com/jackc/puddle/v2` - Connection pooling

### Lint Configuration

`.golangci.yml` deshabilita `staticcheck` para evitar warnings SA1019 sobre `cloud.google.com/go/pubsub` deprecado (requerido por Vertex AI v1.50.1).

### Testing

```bash
# Publish test message
gcloud pubsub topics publish document-analyze --message='{"document_id":"uuid-here"}'

# View logs
gcloud run services logs read worker-semantic --region=us-east1 --limit=50

# Check service status
gcloud run services describe worker-semantic --region=us-east1
```

### Troubleshooting

**Vertex AI Connection Error**:
- Verify `GOOGLE_LOCATION=us-central1`
- Check `phosboard-runtime-sa` has `roles/aiplatform.user`
- Verify Vertex AI API is enabled

**GCS Access Error**:
- Check `phosboard-runtime-sa` has `roles/storage.objectAdmin` on bucket
- Verify bucket exists: `phosboard-documents`

**Database Connection Error**:
- Verify `DATABASE_URL` secret exists and is accessible
- Check `phosboard-runtime-sa` has `roles/cloudsql.client`

**Workload Identity Error**:
- Verify binding exists for repository
- Check no restrictive `ref_type` conditions
- Ensure `github-actions-sa` has `roles/iam.serviceAccountTokenCreator`

**Error específico**: `"The given credential is rejected by the attribute condition"`
- **Causa 1**: El repositorio no está incluido en la `attributeCondition` del provider
- **Causa 2**: No existe binding de Workload Identity para el repositorio específico

**Solución completa**:
1. **Actualizar provider condition** (incluir todos los repositorios):
   ```bash
   gcloud iam workload-identity-pools providers update-oidc github-provider \
     --project=phosboard \
     --location=global \
     --workload-identity-pool=github-pool \
     --attribute-condition="assertion.repository in ['SamuelAnjel/phosboard-api','SamuelAnjel/phosboard-worker-climate-aggregate','SamuelAnjel/phosboard-worker-fetcher','SamuelAnjel/phosboard-worker-scraper','SamuelAnjel/phosboard-worker-semantic','SamuelAnjel/phosboard-worker-discovery']" \
     --attribute-mapping="google.subject=assertion.sub,attribute.actor=assertion.actor,attribute.repository=assertion.repository" \
     --issuer-uri="https://token.actions.githubusercontent.com"
   ```

2. **Agregar binding específico**:
   ```bash
   gcloud iam service-accounts add-iam-policy-binding \
     github-actions-sa@phosboard.iam.gserviceaccount.com \
     --project=phosboard \
     --role="roles/iam.workloadIdentityUser" \
     --member="principalSet://iam.googleapis.com/projects/544990213867/locations/global/workloadIdentityPools/github-pool/attribute.repository/SamuelAnjel/<repo-name>"
    ```

### Cloud Run Service Errors

**Error**: `"rpc error: code = PermissionDenied desc = User not authorized to perform this action. permission:pubsub.topics.get"`

**Causa**: El worker está intentando verificar/crear topics de Pub/Sub durante la inicialización, pero en Cloud Run con Pub/Sub Push esto no es necesario.

**Solución**: Usar **lazy initialization** en el publisher:
1. No verificar si topics existen (`topic.Exists()`)
2. No crear topics (`client.CreateTopic()`)
3. Inicializar cliente Pub/Sub solo cuando sea necesario para publicar
4. Asumir que los topics ya existen y tenemos permisos para publicar

**Código del publisher corregido**:
```go
// Lazy initialization - don't check topics during startup
func (p *Publisher) init(ctx context.Context) error {
    // Initialize client only when first publish is needed
    // Don't call topic.Exists() or client.CreateTopic()
}
```

---

## Workload Identity Federation - Best Practices

### Lecciones Aprendidas

#### 1. Validación Simple por Repositorio
**Problema**: Condiciones complejas como `assertion.ref_type=="branch"` causan errores cuando los workflows se ejecutan en tags.

**Solución**: Solo validar por repositorio:
```
principalSet://iam.googleapis.com/projects/544990213867/locations/global/workloadIdentityPools/github-pool/attribute.repository/SamuelAnjel/<repo-name>
```

**Ventajas**:
- Funciona para branches y tags
- Más simple de mantener
- Igual de seguro (solo tu repositorio puede acceder)

#### 2. Cada Repositorio Necesita su Propio Binding
**Comando para agregar binding**:
```bash
gcloud iam service-accounts add-iam-policy-binding \
  github-actions-sa@phosboard.iam.gserviceaccount.com \
  --project=phosboard \
  --role="roles/iam.workloadIdentityUser" \
  --member="principalSet://iam.googleapis.com/projects/544990213867/locations/global/workloadIdentityPools/github-pool/attribute.repository/SamuelAnjel/<repo-name>"
```

#### 3. Permisos Requeridos para GitHub Actions SA
La service account `github-actions-sa@phosboard.iam.gserviceaccount.com` necesita:
- `roles/iam.serviceAccountTokenCreator` (CRÍTICO para `getAccessToken`)
- `roles/secretmanager.secretAccessor` (para leer secretos)
- `roles/artifactregistry.writer` (para push de imágenes)
- `roles/run.developer` (para deploy a Cloud Run)

#### 4. Error Común: `Permission 'iam.serviceAccounts.getAccessToken' denied`
**Causas**:
1. Falta rol `roles/iam.serviceAccountTokenCreator`
2. Binding de Workload Identity no existe para el repositorio
3. Condición restrictiva (ej: `ref_type=="branch"` cuando workflow corre en tags)

**Solución**:
1. Verificar roles: `gcloud projects get-iam-policy phosboard --filter="bindings.members:github-actions-sa"`
2. Verificar binding: `gcloud iam service-accounts get-iam-policy github-actions-sa@phosboard.iam.gserviceaccount.com`
3. Agregar binding simple sin condiciones

#### 5. Configuración de GitHub Actions Workflow
**Requisitos mínimos**:
```yaml
permissions:
  contents: 'read'
  id-token: 'write'  # CRÍTICO para Workload Identity

steps:
  - name: Authenticate to GCP
    uses: google-github-actions/auth@v2
    with:
      workload_identity_provider: 'projects/${{ env.GCP_PROJECT_NUMBER }}/locations/global/workloadIdentityPools/${{ env.WORKLOAD_IDENTITY_POOL }}/providers/${{ env.WORKLOAD_IDENTITY_PROVIDER }}'
      service_account: '${{ env.SERVICE_ACCOUNT }}'
```

#### 6. Testing Local
Para probar permisos localmente:
```bash
# Intentar impersonar la service account (debería fallar sin Workload Identity)
gcloud auth print-access-token --impersonate-service-account=github-actions-sa@phosboard.iam.gserviceaccount.com

# Verificar secretos (con cuenta personal que tenga permisos)
gcloud secrets versions access latest --secret="phosboard-database-url" --project=phosboard
```

#### 7. Monitoreo
- Ver logs de GitHub Actions para errores de autenticación
- Verificar que el binding existe para cada nuevo repositorio
- Mantener lista actualizada de repositorios en PHOSBOARD_CONTEXT.md

### Resumen
Workload Identity Federation es poderoso pero requiere configuración precisa. La mejor práctica es:
1. **Binding simple por repositorio** sin condiciones complejas
2. **Todos los permisos necesarios** en la service account
3. **Configuración correcta** en el workflow de GitHub Actions
4. **Documentación actualizada** de cada repositorio y sus bindings

---

## Política de Versionado (Semantic Versioning)

### Estrategia de Versionado

**Semantic Versioning (SemVer)**: `MAJOR.MINOR.PATCH`

#### Reglas:
1. **Versión 0.x.x**: Desarrollo inicial
   - API inestable, breaking changes permitidos
   - Ejemplo: `v0.1.0`, `v0.2.0`, `v0.3.0`
   - Usado para primeros releases funcionales

2. **Versión 1.0.0**: Producción estable
   - API estable, sin breaking changes
   - Documentación completa
   - Tests exhaustivos
   - Ready for production

3. **Incrementos**:
   - **PATCH** (`x.y.z+1`): Bug fixes, sin cambios de API
   - **MINOR** (`x.y+1.0`): Nuevas features, backward compatible
   - **MAJOR** (`x+1.0.0`): Breaking changes, nueva API

### Estado Actual de Workers

| Worker | Versión | Estado | Notas |
|--------|---------|--------|-------|
| **climate_aggregate** | `v2.0.0` | ✅ Producción | Migrado a Cloud Run + Pub/Sub Push |
| **fetcher** | `v2.0.0` | ✅ Producción | Migrado a Cloud Run + Pub/Sub Push |
| **scraper** | `v1.0.0` | ✅ Producción | Cloud Run Service estable |
| **semantic** | `v1.0.8` | ✅ Producción | Cloud Run Service estable |
| **discovery** | `v0.1.4` | ⏳ Publisher fixed (lazy init), deploy pending | Primera versión Cloud Run |
| **social_probe** | TBD | ⏳ Por implementar | |
| **parser** | TBD | ⏳ Por implementar | |

### Workflow de CI/CD por Versión

#### Para versiones `0.x.x`:
```yaml
# .github/workflows/build-deploy.yml
on:
  push:
    branches:
      - main    # Solo lint + security
    tags:
      - 'v0.*'  # Deploy a Cloud Run
```

#### Para versiones `1.x.x`+:
```yaml
on:
  push:
    tags:
      - 'v*'    # Solo tags para producción
```

### Creación de Tags

```bash
# Desarrollo (0.x.x)
git tag v0.1.0
git tag v0.2.0
git tag v0.3.0

# Producción (1.0.0+)
git tag v1.0.0
git tag v1.1.0
git tag v2.0.0

# Push tags
git push origin --tags
```

### Migración de 0.x.x a 1.0.0

**Requisitos para 1.0.0**:
1. ✅ Tests completos y passing
2. ✅ Documentación actualizada
3. ✅ API estable (no breaking changes planeados)
4. ✅ Monitoreo y logging implementados
5. ✅ Performance aceptable en producción

**Proceso**:
1. Última versión 0.x.x: `v0.9.0`
2. Release candidate: `v1.0.0-rc.1`
3. Testing en staging
4. Release final: `v1.0.0`

### Best Practices

1. **Siempre usar SemVer**: Claridad sobre el estado del software
2. **Versionar independientemente**: Cada worker tiene su propio versionado
3. **Documentar breaking changes**: En CHANGELOG.md
4. **Automate tag creation**: Usar GitHub Actions para releases
5. **Maintain CHANGELOG**: Registro de cambios por versión

### Ejemplo de CHANGELOG.md

```markdown
# Changelog

## [v0.1.0] - 2024-01-01
### Added
- Initial Cloud Run Service implementation
- Pub/Sub Push handler
- GitHub Actions CI/CD

## [v0.2.0] - 2024-01-15
### Changed
- Improved error handling
- Added retry logic
- Updated dependencies

## [v1.0.0] - 2024-02-01
### Added
- Production-ready features
- Comprehensive monitoring
- Performance optimizations
```

Esta política asegura claridad, estabilidad y profesionalismo en el desarrollo del proyecto.

---

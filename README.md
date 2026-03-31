# Phosboard API

Backend API for Phosboard, a document management system with RBAC (Role-Based Access Control).

## Overview

Phosboard is a document management system that:
- Ingests documents from RSS feeds via a worker pipeline
- Stores documents in PostgreSQL with MinIO for large payloads
- Analyzes documents using Vertex AI Gemini
- Provides REST APIs with JWT authentication

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Discovery  │────▶│   Scraper   │────▶│  Semantic   │────▶│   Climate   │
│   Worker    │     │   Worker    │     │   Worker    │     │  Aggregate  │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
       │                   │                   │                   │
       ▼                   ▼                   ▼                   ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                         Pub/Sub (GCP)                                   │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  PostgreSQL  │  MinIO (S3)  │  Cloud Run (API)                         │
└─────────────────────────────────────────────────────────────────────────┘
```

## Technology Stack

- **Language**: Go 1.25
- **Framework**: Gin
- **Database**: PostgreSQL (pgx/v5)
- **Storage**: MinIO (S3-compatible)
- **AI**: Google Vertex AI Gemini
- **Messaging**: Google Cloud Pub/Sub
- **Auth**: JWT

## Database Schema

### Tables

| Table | Description |
|-------|-------------|
| `tenants` | Multi-tenant organizations |
| `roles` | Role definitions per tenant |
| `tenant_users` | User-role associations |
| `sources` | RSS feed sources with config |
| `global_documents` | Master document records |
| `tenant_documents` | Tenant-document associations |
| `discovery_tasks` | URL discovery queue |
| `tenant_concepts` | Seed concepts per tenant |
| `social_mentions` | Social media mentions (partitioned) |
| `document_temperatures` | Social temperature scores |

## API Endpoints

### Authentication

```bash
# Login
POST /api/auth/login
{
  "email": "admin@phosboard.cl",
  "password": "password123"
}

# Response
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Documents (Protected)

```bash
# List documents with pagination
GET /api/v1/documents?limit=20&offset=0
Authorization: Bearer <token>

# Response
{
  "data": [...],
  "meta": {"total": 100, "limit": 20, "offset": 0}
}

# Track URL manually
POST /api/v1/documents/track
{
  "url": "https://example.com/article",
  "source_type": "manual",
  "priority": 5
}
```

### Concepts (Protected)

```bash
# List concepts
GET /api/v1/tenants/{tenant_id}/concepts

# Create concept
POST /api/v1/tenants/{tenant_id}/concepts
{
  "concept_term": "climate change"
}

# Delete concept
DELETE /api/v1/tenants/{tenant_id}/concepts/{concept_id}
```

### Sources (Protected)

```bash
# List sources
GET /api/v1/tenants/{tenant_id}/sources

# Create source
POST /api/v1/tenants/{tenant_id}/sources
{
  "name": "Tech News",
  "type": "rss",
  "max_links": 20
}

# Get source
GET /api/v1/tenants/{tenant_id}/sources/{source_id}

# Update source config
PUT /api/v1/tenants/{tenant_id}/sources/{source_id}
{
  "config": {"max_links": 50}
}

# Delete source
DELETE /api/v1/tenants/{tenant_id}/sources/{source_id}
```

### Health

```bash
GET /health
```

## Worker Pipeline

### 1. Discovery Worker
- Consumes from `source-discovery` topic
- Parses RSS feeds
- Discovers URLs
- Respects `max_links` config per source (default: 20)
- Publishes to `url-scrape`

### 2. Scraper Worker
- Consumes from `url-scrape` topic
- Fetches HTML content
- Cleans HTML (removes scripts, ads, nav)
- Extracts plain text
- Stores HTML in MinIO (`raw-html` bucket)
- Stores text in PostgreSQL
- Publishes to `document-analyze`

### 3. Semantic Worker
- Consumes from `document-analyze` topic
- Downloads HTML from MinIO
- Analyzes with Vertex AI Gemini
- Extracts entities and sentiment
- Publishes to `social-probe` if queries found

### 4. Social Probe Worker
- Consumes from `social-probe` topic
- Mocks social media search (for demo)
- Stores mentions in MinIO (`social-payloads` bucket)
- Publishes to `climate-aggregate`

### 5. Climate Aggregate Worker
- Consumes from `climate-aggregate` topic
- Calculates social temperature
- Updates `global_documents.social_temperature`

## Pub/Sub Topics

| Topic | Dead Letter |
|-------|-------------|
| `source-discovery` | `source-discovery-dead-letter` |
| `url-scrape` | `url-scrape-dead-letter` |
| `document-analyze` | `document-analyze-dead-letter` |
| `social-probe` | `social-probe-dead-letter` |
| `climate-aggregate` | `climate-aggregate-dead-letter` |

## Configuration

### Environment Variables

```bash
# Database
DATABASE_URL=postgres://user:pass@host:5432/phosboard

# GCP
GOOGLE_PROJECT_ID=phosboard
GOOGLE_LOCATION=us-central1
PUBSUB_EMULATOR_HOST=localhost:8085

# MinIO
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin

# API
JWT_SECRET=your-secret-key
DISPATCHER_INTERVAL_SECONDS=900
```

## Running Locally

### Prerequisites
- PostgreSQL 16+
- MinIO
- Pub/Sub Emulator (optional)

### Setup

```bash
# Install dependencies
go mod download

# Run migrations
psql $DATABASE_URL -f data/migrations/00001_initial_schema.sql
# ... run other migrations

# Setup Pub/Sub (optional)
go run ./cmd/setup-pubsub

# Setup MinIO buckets (optional)
go run ./cmd/setup-minio

# Run API
go run ./cmd/api
```

### Workers

```bash
# Discovery Worker
cd workers/discovery && go run ./cmd/worker

# Scraper Worker
cd workers/scraper && go run ./cmd/worker

# Semantic Worker
cd workers/semantic && go run ./cmd/worker

# Social Probe Worker
cd workers/social_probe && go run ./cmd/worker

# Climate Aggregate Worker
cd workers/climate_aggregate && go run ./cmd/worker
```

## Docker

```bash
# Build
docker build -t phosboard-api .

# Run
docker run -p 8080:8080 \
  -e DATABASE_URL=$DATABASE_URL \
  -e JWT_SECRET=$JWT_SECRET \
  phosboard-api
```

## Git Workflow

```
main         # Production
dev          # Development stable
feature/*    # New features
fix/*        # Bug fixes
```

## CI/CD

GitHub Actions workflow triggers on tag push (`v*`):
1. Lint with golangci-lint
2. Security scan with Trivy
3. Build Docker image
4. Push to Artifact Registry
5. Deploy to Cloud Run

## Project Structure

```
backend/
├── cmd/
│   ├── api/           # Main API server
│   ├── setup-pubsub/  # Pub/Sub topology setup
│   └── setup-minio/   # MinIO buckets setup
├── internal/
│   ├── config/        # Configuration loading
│   ├── db/            # Database connection
│   ├── models/        # Data models
│   ├── repository/    # Data access layer
│   ├── handler/       # HTTP handlers
│   ├── dispatcher/    # Scheduled task dispatcher
│   ├── publisher/     # Pub/Sub publisher
│   ├── auth/          # JWT authentication
│   └── http/middleware/ # HTTP middleware
├── .github/workflows/ # CI/CD
├── Dockerfile
└── go.mod
```

## License

MIT

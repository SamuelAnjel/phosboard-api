# Social Probe Worker

Cloud Run service that processes social media monitoring tasks via Pub/Sub Push.

## Architecture

- **Cloud Run Service**: HTTP server running on port 8080
- **Pub/Sub Push**: Receives messages from `social-probe` topic
- **Processing**: Scrapes social media mentions for given search queries
- **Storage**: Saves results to Google Cloud Storage (GCS)
- **Publisher**: Publishes processed data to `climate-aggregate` topic

## Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `GOOGLE_PROJECT_ID` | Google Cloud Project ID | Yes | `phosboard` |
| `GCS_BUCKET_NAME` | GCS bucket for storing mentions | Yes | `phosboard-documents` |
| `PORT` | HTTP server port (Cloud Run sets this) | No | `8080` |
| `PUBSUB_EMULATOR_HOST` | For local development | No | - |

## Local Development

1. Copy environment file:
   ```bash
   cp .env.example .env
   ```

2. Run locally:
   ```bash
   go run ./cmd/worker
   ```

3. Test with curl:
   ```bash
   curl -X POST http://localhost:8080/health
   ```

## Deployment

### Manual Deployment

```bash
# Build and push image
gcloud builds submit --tag us-east1-docker.pkg.dev/phosboard/phosboard/worker-social-probe

# Deploy to Cloud Run
gcloud run deploy worker-social-probe \
  --image=us-east1-docker.pkg.dev/phosboard/phosboard/worker-social-probe \
  --region=us-east1 \
  --platform=managed \
  --allow-unauthenticated \
  --ingress=all \
  --service-account=phosboard-runtime-sa@phosboard.iam.gserviceaccount.com \
  --set-env-vars="GOOGLE_PROJECT_ID=phosboard,GCS_BUCKET_NAME=phosboard-documents"
```

### CI/CD Deployment (GitHub Actions)

The worker is automatically deployed when tags are pushed:

```bash
# Create and push a tag
git tag v0.1.0
git push origin v0.1.0
```

## Pub/Sub Configuration

### Create Topics and Subscriptions

```bash
# Create topics
gcloud pubsub topics create social-probe --project=phosboard
gcloud pubsub topics create social-probe-dead-letter --project=phosboard
gcloud pubsub topics create climate-aggregate --project=phosboard

# Create Push subscription for social-probe
gcloud pubsub subscriptions create social-probe-sub \
  --topic=social-probe \
  --push-endpoint=https://worker-social-probe-544990213867.us-east1.run.app/ \
  --push-auth-service-account=phosboard-runtime-sa@phosboard.iam.gserviceaccount.com \
  --dead-letter-topic=social-probe-dead-letter \
  --ack-deadline=10
```

### Test Message

```bash
gcloud pubsub topics publish social-probe --message='{
  "document_id": "test-doc-123",
  "search_queries": ["climate change", "global warming"]
}'
```

## Message Format

### Input (social-probe topic)
```json
{
  "document_id": "uuid-of-document",
  "search_queries": ["query 1", "query 2", "query 3"]
}
```

### Output (climate-aggregate topic)
```json
{
  "document_id": "uuid-of-document",
  "gcs_mentions_key": "mentions/{document_id}/{timestamp}.json",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## GCS Output Format

The worker saves mentions to GCS in the following location:
```
mentions/{document_id}/{timestamp}.json
```

Example content:
```json
{
  "document_id": "uuid-of-document",
  "queries": ["climate change", "global warming"],
  "mentions": [
    {
      "text": "Mock mention about climate change",
      "author": "user_a_handle",
      "date": "2024-01-01T00:00:00Z",
      "platform": "twitter",
      "engagement_score": 100
    }
  ],
  "scraped_at": "2024-01-01T00:00:00Z"
}
```

## Development Notes

- Uses lazy initialization for Pub/Sub publisher (no topic checks during startup)
- Mock scraper implementation for development
- Health endpoint at `/health`
- Pre-commit hooks with golangci-lint
- Semantic versioning (tags trigger deployment)
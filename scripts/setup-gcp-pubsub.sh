#!/bin/bash

set -e

PROJECT_ID="${GOOGLE_PROJECT_ID:-phosboard}"

echo "Enabling Pub/Sub API for project: $PROJECT_ID..."
gcloud services enable pubsub.googleapis.com --project="$PROJECT_ID"

echo "Creating Pub/Sub topics and subscriptions for project: $PROJECT_ID"

# Create topics (ignore if exists)
echo "Creating topics..."
gcloud pubsub topics create source-discovery --project="$PROJECT_ID" || true
gcloud pubsub topics create source-discovery-dead-letter --project="$PROJECT_ID" || true
gcloud pubsub topics create url-scrape --project="$PROJECT_ID" || true
gcloud pubsub topics create url-scrape-dead-letter --project="$PROJECT_ID" || true
gcloud pubsub topics create document-analyze --project="$PROJECT_ID" || true
gcloud pubsub topics create document-analyze-dead-letter --project="$PROJECT_ID" || true
gcloud pubsub topics create social-probe --project="$PROJECT_ID" || true
gcloud pubsub topics create social-probe-dead-letter --project="$PROJECT_ID" || true
gcloud pubsub topics create climate-aggregate --project="$PROJECT_ID" || true
gcloud pubsub topics create climate-aggregate-dead-letter --project="$PROJECT_ID" || true

echo "Creating subscriptions..."

# Create subscriptions with Dead Letter and Retry Policy (ignore if exists)
gcloud pubsub subscriptions create source-discovery-sub \
  --project="$PROJECT_ID" \
  --topic=source-discovery \
  --dead-letter-topic=source-discovery-dead-letter \
  --max-delivery-attempts=5 \
  --min-retry-delay=10s \
  --max-retry-delay=600s || true

# Delete existing subscription if it exists (to recreate as push)
gcloud pubsub subscriptions delete url-scrape-sub --project="$PROJECT_ID" --quiet 2>/dev/null || true

# Get Cloud Run service URL
SCRAPER_SERVICE_URL=$(gcloud run services describe worker-scraper --region=us-east1 --project="$PROJECT_ID" --format='value(status.url)' 2>/dev/null || echo "")

if [ -n "$SCRAPER_SERVICE_URL" ]; then
  echo "Found scraper service URL: $SCRAPER_SERVICE_URL"
  gcloud pubsub subscriptions create url-scrape-sub \
    --project="$PROJECT_ID" \
    --topic=url-scrape \
    --push-endpoint="${SCRAPER_SERVICE_URL}/" \
    --push-auth-service-account=phosboard-runtime-sa@phosboard.iam.gserviceaccount.com \
    --dead-letter-topic=url-scrape-dead-letter \
    --max-delivery-attempts=5 \
    --min-retry-delay=10s \
    --max-retry-delay=600s || true
else
  echo "Warning: Scraper service not found, creating pull subscription"
  gcloud pubsub subscriptions create url-scrape-sub \
    --project="$PROJECT_ID" \
    --topic=url-scrape \
    --dead-letter-topic=url-scrape-dead-letter \
    --max-delivery-attempts=5 \
    --min-retry-delay=10s \
    --max-retry-delay=600s || true
fi

gcloud pubsub subscriptions create document-analyze-sub \
  --project="$PROJECT_ID" \
  --topic=document-analyze \
  --dead-letter-topic=document-analyze-dead-letter \
  --max-delivery-attempts=5 \
  --min-retry-delay=10s \
  --max-retry-delay=600s || true

gcloud pubsub subscriptions create social-probe-sub \
  --project="$PROJECT_ID" \
  --topic=social-probe \
  --dead-letter-topic=social-probe-dead-letter \
  --max-delivery-attempts=5 \
  --min-retry-delay=10s \
  --max-retry-delay=600s || true

gcloud pubsub subscriptions create climate-aggregate-sub \
  --project="$PROJECT_ID" \
  --topic=climate-aggregate \
  --dead-letter-topic=climate-aggregate-dead-letter \
  --max-delivery-attempts=5 \
  --min-retry-delay=10s \
  --max-retry-delay=600s || true

echo "Done! Pub/Sub topology created successfully."

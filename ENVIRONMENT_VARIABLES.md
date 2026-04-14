# Environment Variables for Phosboard v2.0.0+

## Social Media Monitoring (Crisis Track) Configuration

### Apify Integration
Required for real social media monitoring:

```
APIFY_API_TOKEN=your_apify_api_token_here
APIFY_ACTOR_ID=apify/actor-twitter-scraper  # Example actor ID
```

### Backend API Configuration
```
DATABASE_URL=postgresql://user:password@host:port/database
PORT=8080  # Default: 8080
```

### Worker Configuration
#### Social Probe Worker
```
GCS_BUCKET_NAME=phosboard-documents
GOOGLE_PROJECT_ID=phosboard
PUBSUB_EMULATOR_HOST=localhost:8085  # Optional, for local development
```

#### Climate Aggregate Worker
```
DATABASE_URL=postgresql://user:password@host:port/database
GCS_BUCKET_NAME=phosboard-documents
```

## Deployment Notes

1. **Apify Setup**:
   - Sign up at https://apify.com
   - Create API token in Account Settings
   - Choose Twitter/X scraper actor (e.g., `apify/actor-twitter-scraper`)
   - Set `APIFY_ACTOR_ID` to the actor ID

2. **Database Migration**:
   Run the new migration for social tracks:
   ```bash
   psql $DATABASE_URL -f data/migrations/00010_social_tracks.sql
   psql $DATABASE_URL -f data/migrations/00011_social_tracks_temperature.sql
   ```

3. **Backend API**:
   Build and deploy:
   ```bash
   cd backend
   go build -o api ./cmd/api
   ./api
   ```

4. **Workers**:
   Each worker has its own deployment. Update environment variables in:
   - Cloud Run (production)
   - Docker containers
   - Local .env files
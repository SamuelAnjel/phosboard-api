#!/bin/bash

set -e

# Get database URL from environment or parameter
DATABASE_URL="${DATABASE_URL:-$1}"
if [ -z "$DATABASE_URL" ]; then
    echo "Usage: DATABASE_URL=postgres://... ./reprocess_tasks.sh"
    echo "Or: ./reprocess_tasks.sh postgres://..."
    exit 1
fi

PROJECT_ID="${GOOGLE_PROJECT_ID:-phosboard}"

echo "Reprocessing pending discovery tasks for project: $PROJECT_ID"

# Create a temporary Python script
cat > /tmp/reprocess_tasks.py << 'EOF'
import os
import json
import psycopg2
from google.cloud import pubsub_v1

def main():
    db_url = os.environ.get('DATABASE_URL')
    project_id = os.environ.get('GOOGLE_PROJECT_ID', 'phosboard')
    
    if not db_url:
        print("DATABASE_URL environment variable is required")
        return
    
    # Connect to database
    conn = psycopg2.connect(db_url)
    cur = conn.cursor()
    
    # Get pending discovery tasks with their document IDs
    cur.execute("""
        SELECT dt.url, gd.id as document_id
        FROM discovery_tasks dt
        LEFT JOIN global_documents gd ON dt.url = gd.url
        WHERE dt.status = 'pending'
        AND gd.id IS NOT NULL
        ORDER BY dt.created_at
        LIMIT 50
    """)
    
    tasks = cur.fetchall()
    print(f"Found {len(tasks)} pending tasks to reprocess")
    
    if not tasks:
        print("No pending tasks found")
        return
    
    # Initialize Pub/Sub client
    publisher = pubsub_v1.PublisherClient()
    topic_path = publisher.topic_path(project_id, 'url-scrape')
    
    # Publish each task
    success_count = 0
    for url, document_id in tasks:
        task_data = {
            'document_id': document_id,
            'url': url
        }
        
        data = json.dumps(task_data).encode('utf-8')
        
        try:
            future = publisher.publish(topic_path, data=data)
            future.result()  # Wait for publish to complete
            success_count += 1
            print(f"Published: {document_id} -> {url}")
        except Exception as e:
            print(f"Failed to publish task {url}: {e}")
    
    print(f"\nSuccessfully published {success_count}/{len(tasks)} tasks")
    
    cur.close()
    conn.close()

if __name__ == '__main__':
    main()
EOF

# Check if Python and required packages are available
if ! command -v python3 &> /dev/null; then
    echo "Python3 is required but not installed"
    exit 1
fi

# Install required Python packages if needed
echo "Checking Python dependencies..."
python3 -c "import psycopg2, google.cloud.pubsub_v1" 2>/dev/null || {
    echo "Installing required Python packages..."
    pip3 install psycopg2-binary google-cloud-pubsub
}

# Run the Python script
DATABASE_URL="$DATABASE_URL" GOOGLE_PROJECT_ID="$PROJECT_ID" python3 /tmp/reprocess_tasks.py

# Clean up
rm -f /tmp/reprocess_tasks.py

echo "Done!"
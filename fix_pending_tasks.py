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
    
    conn = psycopg2.connect(db_url)
    cur = conn.cursor()
    
    # Get the source ID for Diario El Día
    cur.execute("SELECT id FROM sources WHERE name = 'Diario El Día'")
    source_row = cur.fetchone()
    if not source_row:
        print("Source 'Diario El Día' not found")
        return
    
    source_id = source_row[0]
    print(f"Source ID: {source_id}")
    
    # Get pending discovery tasks
    cur.execute("""
        SELECT url
        FROM discovery_tasks
        WHERE status = 'pending'
        AND source_type = 'web-crawl'
        ORDER BY created_at
        LIMIT 50
    """)
    
    pending_urls = [row[0] for row in cur.fetchall()]
    print(f"Found {len(pending_urls)} pending URLs")
    
    if not pending_urls:
        print("No pending URLs found")
        return
    
    # Initialize Pub/Sub client
    publisher = pubsub_v1.PublisherClient()
    topic_path = publisher.topic_path(project_id, 'url-scrape')
    
    success_count = 0
    error_count = 0
    
    for url in pending_urls:
        try:
            # Check if document already exists
            cur.execute("SELECT id FROM global_documents WHERE url = %s", (url,))
            existing = cur.fetchone()
            
            if existing:
                document_id = existing[0]
                print(f"Document already exists: {document_id} for {url[:80]}...")
            else:
                # Create document
                cur.execute("""
                    INSERT INTO global_documents (source_id, url, content_text, created_at, updated_at)
                    VALUES (%s, %s, '', NOW(), NOW())
                    RETURNING id
                """, (source_id, url))
                
                document_id = cur.fetchone()[0]
                conn.commit()
                print(f"Created document: {document_id} for {url[:80]}...")
            
            # Publish to Pub/Sub
            task_data = {
                'document_id': document_id,
                'url': url
            }
            
            data = json.dumps(task_data).encode('utf-8')
            future = publisher.publish(topic_path, data=data)
            future.result()  # Wait for publish to complete
            
            # Optionally update task status (or leave as pending for scraper to update)
            # cur.execute("UPDATE discovery_tasks SET status = 'processing' WHERE url = %s", (url,))
            # conn.commit()
            
            success_count += 1
            print(f"  Published to url-scrape topic")
            
        except psycopg2.IntegrityError as e:
            print(f"Integrity error for {url[:80]}...: {e}")
            conn.rollback()
            error_count += 1
        except Exception as e:
            print(f"Error processing {url[:80]}...: {e}")
            conn.rollback()
            error_count += 1
    
    print(f"\nSummary:")
    print(f"  Successfully processed: {success_count}")
    print(f"  Errors: {error_count}")
    print(f"  Total: {len(pending_urls)}")
    
    cur.close()
    conn.close()

if __name__ == '__main__':
    main()
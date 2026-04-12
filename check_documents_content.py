import os
import psycopg2
import json

def main():
    db_url = os.environ.get('DATABASE_URL')
    if not db_url:
        print("DATABASE_URL environment variable is required")
        return
    
    conn = psycopg2.connect(db_url)
    cur = conn.cursor()
    
    # Get documents from Diario El Día source
    print("=== Documents from Diario El Día ===")
    cur.execute("""
        SELECT gd.id, gd.url, gd.title, 
               LENGTH(gd.content_text) as content_length,
               gd.created_at,
               dt.status as task_status
        FROM global_documents gd
        LEFT JOIN sources s ON gd.source_id = s.id
        LEFT JOIN discovery_tasks dt ON gd.url = dt.url
        WHERE s.name = 'Diario El Día'
        ORDER BY gd.created_at DESC
        LIMIT 10
    """)
    
    documents = cur.fetchall()
    print(f"Found {len(documents)} documents")
    print("\n" + "="*80)
    
    for i, (doc_id, url, title, content_len, created_at, task_status) in enumerate(documents):
        print(f"\n[{i+1}] Document ID: {doc_id}")
        print(f"    URL: {url}")
        print(f"    Title: {title[:100] if title else 'No title'}")
        print(f"    Content length: {content_len} characters")
        print(f"    Created: {created_at}")
        print(f"    Task status: {task_status}")
        
        # Get first 500 chars of content
        if content_len > 0:
            cur2 = conn.cursor()
            cur2.execute("SELECT content_text FROM global_documents WHERE id = %s", (doc_id,))
            content = cur2.fetchone()[0]
            cur2.close()
            
            print(f"\n    Content preview (first 500 chars):")
            print("    " + "-"*70)
            if content:
                preview = content[:500].replace('\n', '\n    ')
                print(f"    {preview}")
                if len(content) > 500:
                    print(f"    ... [truncated, total: {len(content)} chars]")
            else:
                print("    No content")
            print("    " + "-"*70)
        
        print("\n" + "="*80)
    
    # Check raw payload structure
    print("\n\n=== Checking raw payload structure ===")
    cur.execute("""
        SELECT raw_payload->>'title' as payload_title,
               jsonb_typeof(raw_payload) as payload_type,
               COUNT(*) as count
        FROM global_documents
        WHERE source_id = (SELECT id FROM sources WHERE name = 'Diario El Día')
        GROUP BY raw_payload->>'title', jsonb_typeof(raw_payload)
        ORDER BY count DESC
        LIMIT 5
    """)
    
    payload_stats = cur.fetchall()
    print("Raw payload statistics:")
    for payload_title, payload_type, count in payload_stats:
        print(f"  Type: {payload_type}, Title in payload: {payload_title}, Count: {count}")
    
    # Check if raw_payload has useful fields
    print("\n=== Sample raw payload keys ===")
    cur.execute("""
        SELECT jsonb_object_keys(raw_payload) as key
        FROM global_documents
        WHERE source_id = (SELECT id FROM sources WHERE name = 'Diario El Día')
        AND raw_payload IS NOT NULL
        LIMIT 1
    """)
    
    sample_keys = cur.fetchall()
    if sample_keys:
        print("Keys in raw_payload:")
        for (key,) in sample_keys:
            print(f"  - {key}")
    
    # Check content quality metrics
    print("\n=== Content quality metrics ===")
    cur.execute("""
        SELECT 
            COUNT(*) as total_docs,
            AVG(LENGTH(content_text)) as avg_content_length,
            MIN(LENGTH(content_text)) as min_content_length,
            MAX(LENGTH(content_text)) as max_content_length,
            COUNT(CASE WHEN LENGTH(content_text) < 100 THEN 1 END) as short_docs,
            COUNT(CASE WHEN title IS NULL OR title = '' THEN 1 END) as missing_titles
        FROM global_documents
        WHERE source_id = (SELECT id FROM sources WHERE name = 'Diario El Día')
    """)
    
    total, avg_len, min_len, max_len, short_docs, missing_titles = cur.fetchone()
    print(f"Total documents: {total}")
    print(f"Average content length: {avg_len:.0f} chars")
    print(f"Min content length: {min_len} chars")
    print(f"Max content length: {max_len} chars")
    print(f"Short documents (<100 chars): {short_docs}")
    print(f"Documents missing titles: {missing_titles}")
    
    # Show examples of very short content
    if short_docs > 0:
        print(f"\n=== Examples of short documents (<100 chars) ===")
        cur.execute("""
            SELECT url, title, LENGTH(content_text) as len
            FROM global_documents
            WHERE source_id = (SELECT id FROM sources WHERE name = 'Diario El Día')
            AND LENGTH(content_text) < 100
            ORDER BY LENGTH(content_text)
            LIMIT 3
        """)
        
        short_examples = cur.fetchall()
        for url, title, length in short_examples:
            print(f"  URL: {url}")
            print(f"  Title: {title}")
            print(f"  Length: {length} chars")
            print()
    
    cur.close()
    conn.close()

if __name__ == '__main__':
    main()
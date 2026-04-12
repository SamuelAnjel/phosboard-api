import os
import psycopg2

def main():
    db_url = os.environ.get('DATABASE_URL')
    if not db_url:
        print("DATABASE_URL environment variable is required")
        return
    
    conn = psycopg2.connect(db_url)
    cur = conn.cursor()
    
    # Check for duplicate URLs in global_documents
    print("=== Checking for duplicate URLs in global_documents ===")
    cur.execute("""
        SELECT url, COUNT(*) as count
        FROM global_documents
        GROUP BY url
        HAVING COUNT(*) > 1
        ORDER BY count DESC
        LIMIT 10
    """)
    duplicates = cur.fetchall()
    if duplicates:
        print(f"Found {len(duplicates)} duplicate URLs:")
        for url, count in duplicates:
            print(f"  {url[:80]}...: {count} duplicates")
    else:
        print("No duplicate URLs found")
    
    # Check total documents count
    print("\n=== Global Documents Count ===")
    cur.execute("SELECT COUNT(*) FROM global_documents")
    total = cur.fetchone()[0]
    print(f"Total documents: {total}")
    
    # Check if any of the pending task URLs already exist in global_documents
    print("\n=== Checking if pending task URLs already exist ===")
    cur.execute("""
        SELECT COUNT(DISTINCT dt.url)
        FROM discovery_tasks dt
        INNER JOIN global_documents gd ON dt.url = gd.url
        WHERE dt.status = 'pending'
    """)
    existing = cur.fetchone()[0]
    print(f"Pending tasks with existing documents: {existing}")
    
    # Show a few examples
    cur.execute("""
        SELECT dt.url, gd.id, gd.created_at
        FROM discovery_tasks dt
        INNER JOIN global_documents gd ON dt.url = gd.url
        WHERE dt.status = 'pending'
        ORDER BY gd.created_at DESC
        LIMIT 5
    """)
    examples = cur.fetchall()
    if examples:
        print("Examples:")
        for url, doc_id, created_at in examples:
            print(f"  {created_at} | {doc_id} | {url[:80]}...")
    
    # Check the source for diarioeldia
    print("\n=== Sources table ===")
    cur.execute("SELECT id, name, type, url FROM sources WHERE url LIKE '%diarioeldia%' OR name LIKE '%Día%'")
    sources = cur.fetchall()
    for source_id, name, source_type, url in sources:
        print(f"ID: {source_id}")
        print(f"  Name: {name}")
        print(f"  Type: {source_type}")
        print(f"  URL: {url}")
        
        # Check config
        cur2 = conn.cursor()
        cur2.execute("SELECT config FROM sources WHERE id = %s", (source_id,))
        config = cur2.fetchone()[0]
        print(f"  Config: {config}")
        cur2.close()
    
    cur.close()
    conn.close()

if __name__ == '__main__':
    main()
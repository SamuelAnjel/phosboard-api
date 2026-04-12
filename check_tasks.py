import os
import psycopg2

def main():
    db_url = os.environ.get('DATABASE_URL')
    if not db_url:
        print("DATABASE_URL environment variable is required")
        return
    
    conn = psycopg2.connect(db_url)
    cur = conn.cursor()
    
    # Check discovery_tasks table
    print("=== Discovery Tasks Status ===")
    cur.execute("""
        SELECT status, COUNT(*) as count
        FROM discovery_tasks
        GROUP BY status
        ORDER BY status
    """)
    for status, count in cur.fetchall():
        print(f"{status}: {count}")
    
    print("\n=== Sample Pending Tasks ===")
    cur.execute("""
        SELECT url, created_at, source_type
        FROM discovery_tasks
        WHERE status = 'pending'
        ORDER BY created_at DESC
        LIMIT 10
    """)
    for url, created_at, source_type in cur.fetchall():
        print(f"{created_at} | {source_type} | {url[:80]}...")
    
    print("\n=== Check if documents exist for pending tasks ===")
    cur.execute("""
        SELECT 
            COUNT(DISTINCT dt.url) as total_pending,
            COUNT(DISTINCT gd.url) as with_documents,
            COUNT(DISTINCT CASE WHEN gd.id IS NULL THEN dt.url END) as without_documents
        FROM discovery_tasks dt
        LEFT JOIN global_documents gd ON dt.url = gd.url
        WHERE dt.status = 'pending'
    """)
    total, with_docs, without_docs = cur.fetchone()
    print(f"Total pending: {total}")
    print(f"With documents: {with_docs}")
    print(f"Without documents: {without_docs}")
    
    print("\n=== Sample tasks without documents ===")
    cur.execute("""
        SELECT dt.url, dt.created_at
        FROM discovery_tasks dt
        LEFT JOIN global_documents gd ON dt.url = gd.url
        WHERE dt.status = 'pending'
        AND gd.id IS NULL
        ORDER BY dt.created_at DESC
        LIMIT 5
    """)
    for url, created_at in cur.fetchall():
        print(f"{created_at} | {url[:80]}...")
    
    cur.close()
    conn.close()

if __name__ == '__main__':
    main()
import os
import psycopg2

def test_create_document():
    db_url = os.environ.get('DATABASE_URL')
    if not db_url:
        print("DATABASE_URL environment variable is required")
        return
    
    conn = psycopg2.connect(db_url)
    cur = conn.cursor()
    
    # Get the source ID for Diario El Día
    cur.execute("SELECT id FROM sources WHERE name = 'Diario El Día'")
    source_id = cur.fetchone()[0]
    print(f"Source ID: {source_id}")
    
    # Test URL
    test_url = "https://www.diarioeldia.cl/noticias/2026/04/09/134164-murio-michael-j-fox-public"
    
    # Try to create a document
    try:
        cur.execute("""
            INSERT INTO global_documents (source_id, url, content_text, created_at, updated_at)
            VALUES (%s, %s, '', NOW(), NOW())
            RETURNING id
        """, (source_id, test_url))
        
        document_id = cur.fetchone()[0]
        conn.commit()
        print(f"Successfully created document: {document_id}")
        
    except psycopg2.IntegrityError as e:
        print(f"IntegrityError (likely duplicate): {e}")
        # Check if URL already exists
        cur.execute("SELECT id FROM global_documents WHERE url = %s", (test_url,))
        existing = cur.fetchone()
        if existing:
            print(f"URL already exists with document ID: {existing[0]}")
    except Exception as e:
        print(f"Error: {e}")
        conn.rollback()
    
    # Try with a different URL that definitely doesn't exist
    test_url2 = "https://www.diarioeldia.cl/test-unique-url-" + os.urandom(4).hex()
    try:
        cur.execute("""
            INSERT INTO global_documents (source_id, url, content_text, created_at, updated_at)
            VALUES (%s, %s, '', NOW(), NOW())
            RETURNING id
        """, (source_id, test_url2))
        
        document_id2 = cur.fetchone()[0]
        conn.commit()
        print(f"\nSuccessfully created document with unique URL: {document_id2}")
        
        # Clean up
        cur.execute("DELETE FROM global_documents WHERE id = %s", (document_id2,))
        conn.commit()
        print(f"Cleaned up test document: {document_id2}")
        
    except Exception as e:
        print(f"\nError with unique URL: {e}")
        conn.rollback()
    
    cur.close()
    conn.close()

if __name__ == '__main__':
    test_create_document()
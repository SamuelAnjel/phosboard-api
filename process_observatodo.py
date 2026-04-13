#!/usr/bin/env python3
"""
Script para procesar manualmente URLs de El Observatodo
que están atascadas en discovery_tasks.
"""
import os
import sys
import psycopg2
import json
import subprocess
from datetime import datetime

def get_db_connection():
    """Obtener conexión a la base de datos"""
    db_url = os.environ.get('DATABASE_URL')
    if not db_url:
        # Intentar obtener de gcloud secrets
        try:
            result = subprocess.run(
                ['gcloud', 'secrets', 'versions', 'access', 'latest', '--secret=phosboard-database-url'],
                capture_output=True,
                text=True,
                check=True
            )
            db_url = result.stdout.strip()
        except subprocess.CalledProcessError as e:
            print(f"Error getting database URL: {e}")
            sys.exit(1)
    
    return psycopg2.connect(db_url)

def process_observatodo_urls():
    """Procesar URLs de El Observatodo"""
    conn = get_db_connection()
    cur = conn.cursor()
    
    # Source ID de El Observatodo
    source_id = '4062fcbb-51dd-48db-b977-aed17dbe6ab2'
    
    # Obtener URLs pendientes de El Observatodo
    cur.execute('''
        SELECT id, url 
        FROM discovery_tasks 
        WHERE url LIKE 'https://www.elobservatodo.cl/%'
        AND status = 'pending'
        ORDER BY created_at
        LIMIT 20  -- Procesar solo 20 para empezar
    ''')
    
    tasks = cur.fetchall()
    print(f"Found {len(tasks)} pending tasks for El Observatodo")
    
    processed = 0
    for task_id, url in tasks:
        try:
            # Verificar si la URL ya tiene documento
            cur.execute('''
                SELECT id FROM global_documents 
                WHERE url = %s AND source_id = %s
            ''', (url, source_id))
            
            if cur.fetchone():
                print(f"Document already exists for: {url[:80]}...")
                # Marcar task como completed
                cur.execute('''
                    UPDATE discovery_tasks 
                    SET status = 'completed', completed_at = NOW()
                    WHERE id = %s
                ''', (task_id,))
                continue
            
            # Crear documento
            cur.execute('''
                INSERT INTO global_documents (source_id, url, content_text, created_at, updated_at)
                VALUES (%s, %s, '', NOW(), NOW())
                RETURNING id
            ''', (source_id, url))
            
            document_id = cur.fetchone()[0]
            
            # Publicar mensaje a url-scrape topic
            message_data = json.dumps({
                "document_id": document_id,
                "url": url
            })
            
            # Codificar en base64 para Pub/Sub
            import base64
            encoded_data = base64.b64encode(message_data.encode()).decode()
            
            # Publicar usando gcloud
            cmd = [
                'gcloud', 'pubsub', 'topics', 'publish', 'url-scrape',
                '--message', message_data
            ]
            
            result = subprocess.run(cmd, capture_output=True, text=True)
            if result.returncode == 0:
                print(f"Published: {url[:80]}... (doc: {document_id[:8]}...)")
                
                # Marcar task como completed
                cur.execute('''
                    UPDATE discovery_tasks 
                    SET status = 'completed', completed_at = NOW()
                    WHERE id = %s
                ''', (task_id,))
                
                processed += 1
            else:
                print(f"Failed to publish: {result.stderr}")
                
        except Exception as e:
            print(f"Error processing URL {url[:80]}...: {e}")
            conn.rollback()
            continue
    
    conn.commit()
    cur.close()
    conn.close()
    
    print(f"\nProcessed {processed} URLs successfully")

if __name__ == "__main__":
    process_observatodo_urls()
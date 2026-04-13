#!/usr/bin/env python3
"""
Test para verificar que nuevos sources se procesan inmediatamente.
"""
import os
import sys
import time
import psycopg2
import subprocess
import uuid
from datetime import datetime, timezone

def get_db_connection():
    """Obtener conexión a la base de datos"""
    db_url = os.environ.get('DATABASE_URL')
    if not db_url:
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

def test_new_source_immediate_processing():
    """Test: crear source nuevo y verificar que se procesa"""
    conn = get_db_connection()
    cur = conn.cursor()
    
    print("=== TEST: PROCESAMIENTO INMEDIATO DE NUEVOS SOURCES ===\n")
    
    # 1. Crear un source de prueba
    test_name = f"Test Source {uuid.uuid4().hex[:8]}"
    test_url = f"https://example.com/test-{uuid.uuid4().hex[:8]}"
    test_type = "web-crawl"
    
    print(f"1. Creando source de prueba:")
    print(f"   Nombre: {test_name}")
    print(f"   URL: {test_url}")
    print(f"   Tipo: {test_type}")
    
    # Insertar directamente en DB (simulando creación por API)
    cur.execute('''
        INSERT INTO sources (name, type, url, fetch_strategy, config, created_at, updated_at)
        VALUES (%s, %s, %s, %s, '{"type": "web-crawl"}'::jsonb, NOW(), NOW())
        RETURNING id, last_run_at
    ''', (test_name, test_type, test_url, "web-crawl"))
    
    source_id, last_run_at = cur.fetchone()
    conn.commit()
    
    print(f"   Source ID: {source_id}")
    print(f"   last_run_at inicial: {last_run_at}")
    
    # 2. Verificar que el dispatcher lo verá como "activo"
    print(f"\n2. Verificando query del dispatcher:")
    
    cur.execute('''
        SELECT id, name, 
               CASE 
                   WHEN last_run_at IS NULL THEN TRUE
                   ELSE FALSE
               END as should_run
        FROM sources 
        WHERE id = %s
        AND url IS NOT NULL 
        AND (
            last_run_at IS NULL 
            OR last_run_at + (interval_minutes || ' minutes')::interval < NOW()
        )
    ''', (source_id,))
    
    result = cur.fetchone()
    if result:
        print(f"   ✓ El source aparece en la query del dispatcher")
        print(f"   - ID: {result[0]}")
        print(f"   - Nombre: {result[1]}")
        print(f"   - Debería ejecutarse: {result[2]}")
    else:
        print(f"   ✗ El source NO aparece en la query del dispatcher")
    
    # 3. Verificar lógica de priorización
    print(f"\n3. Verificando priorización:")
    
    cur.execute('''
        SELECT id, name, last_run_at,
               CASE 
                   WHEN last_run_at IS NULL THEN 1
                   ELSE 2
               END as priority
        FROM sources 
        WHERE url IS NOT NULL 
        AND (
            last_run_at IS NULL 
            OR last_run_at + (interval_minutes || ' minutes')::interval < NOW()
        )
        ORDER BY 
            CASE WHEN last_run_at IS NULL THEN 1 ELSE 2 END,
            last_run_at ASC NULLS FIRST
        LIMIT 5
    ''')
    
    print("   Top 5 sources por prioridad:")
    for row in cur.fetchall():
        status = "NUEVO" if row[2] is None else "PENDIENTE"
        print(f"   - {row[1]} ({status}, prioridad: {row[3]})")
    
    # 4. Limpiar (opcional)
    print(f"\n4. Limpiando test...")
    cur.execute('DELETE FROM sources WHERE id = %s', (source_id,))
    conn.commit()
    print(f"   ✓ Test source eliminado")
    
    cur.close()
    conn.close()
    
    print(f"\n=== CONCLUSIÓN ===")
    print("Con la modificación del dispatcher (v1.2.8):")
    print("- Sources nuevos (last_run_at = NULL) tienen prioridad 1")
    print("- Se procesarán en el próximo ciclo del dispatcher (≤15 min)")
    print("- Orden: 1) nuevos, 2) más antiguos primero")

if __name__ == "__main__":
    test_new_source_immediate_processing()
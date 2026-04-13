#!/bin/bash
# Script para ejecutar migraciones RBAC en Supabase

set -e  # Exit on error

DATABASE_URL="postgresql://postgres.ohrmoiplfblbzstpgpxn:&2s9d-3cXALSPtd@aws-1-us-east-1.pooler.supabase.com:5432/postgres"

echo "=== Ejecutando migraciones RBAC en Supabase ==="
echo "Database: $(echo $DATABASE_URL | sed 's/:.*@/:****@/')"

# Función para ejecutar SQL
execute_sql() {
    local sql_file=$1
    echo "Ejecutando: $sql_file"
    psql "$DATABASE_URL" -f "$sql_file" -v ON_ERROR_STOP=1
    echo "✅ $sql_file completado"
}

# Verificar que psql está instalado
if ! command -v psql &> /dev/null; then
    echo "Error: psql no está instalado. Instala PostgreSQL client."
    exit 1
fi

# Ejecutar migraciones en orden
echo ""
echo "1. Ejecutando migración RBAC schema (00006_rbac_system.sql)..."
execute_sql "data/migrations/00006_rbac_system.sql"

echo ""
echo "2. Ejecutando seed data RBAC (00007_seed_rbac_data.sql)..."
execute_sql "data/migrations/00007_seed_rbac_data.sql"

echo ""
echo "=== Migraciones completadas exitosamente ==="
echo ""
echo "Datos creados:"
echo "- Tablas: users, permissions, user_roles, role_permissions"
echo "- Super-admin: admin@phosboard.cl / admin123"
echo "- Tenant-admin: tenant.admin@example.com / admin123"
echo "- Tenant por defecto: 85c5f582-86b1-4217-bd4a-e1b1d0aac195"
echo "- Permisos: Configurados por endpoint + método HTTP"
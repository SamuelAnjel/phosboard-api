-- Fix invalid UUIDs in RBAC seed data
-- Run after 00007_seed_rbac_data.sql

-- Temporarily disable foreign key constraints
ALTER TABLE user_roles DROP CONSTRAINT IF EXISTS user_roles_user_id_fkey;
ALTER TABLE user_roles DROP CONSTRAINT IF EXISTS user_roles_role_id_fkey;
ALTER TABLE user_roles DROP CONSTRAINT IF EXISTS user_roles_tenant_id_fkey;

-- 1. Generate valid UUIDs for users
DO $$
DECLARE
    new_admin_uuid UUID := uuid_generate_v4();
    new_tenant_admin_uuid UUID := uuid_generate_v4();
    old_admin_uuid UUID := 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid;
    old_tenant_admin_uuid UUID := 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::uuid;
BEGIN
    -- Update admin user with valid UUID
    UPDATE users 
    SET id = new_admin_uuid 
    WHERE id = old_admin_uuid;
    
    -- Update user_roles for admin
    UPDATE user_roles 
    SET user_id = new_admin_uuid 
    WHERE user_id = old_admin_uuid;
    
    -- Update tenant admin user with valid UUID
    UPDATE users 
    SET id = new_tenant_admin_uuid 
    WHERE id = old_tenant_admin_uuid;
    
    -- Update user_roles for tenant admin
    UPDATE user_roles 
    SET user_id = new_tenant_admin_uuid 
    WHERE user_id = old_tenant_admin_uuid;
    
    RAISE NOTICE 'Fixed UUIDs: admin=%, tenant_admin=%', new_admin_uuid, new_tenant_admin_uuid;
END $$;

-- Re-enable foreign key constraints
ALTER TABLE user_roles 
ADD CONSTRAINT user_roles_user_id_fkey 
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE user_roles 
ADD CONSTRAINT user_roles_role_id_fkey 
FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE;

ALTER TABLE user_roles 
ADD CONSTRAINT user_roles_tenant_id_fkey 
FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

-- 2. Verify the fixes
SELECT 
    'users' as table_name,
    COUNT(*) as total,
    COUNT(*) as valid_uuids -- All should be valid now
FROM users
UNION ALL
SELECT 
    'user_roles' as table_name,
    COUNT(*) as total,
    COUNT(*) as valid_uuids -- All should be valid now
FROM user_roles;
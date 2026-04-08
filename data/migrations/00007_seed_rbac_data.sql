-- Seed RBAC data: Super Admin, roles, and permissions
-- Password for super admin: 'admin123' (bcrypt hash)

-- 1. Create default tenant (if not exists)
INSERT INTO tenants (id, name, created_at, updated_at)
VALUES (
    '85c5f582-86b1-4217-bd4a-e1b1d0aac195',
    'Default System Tenant',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- 2. Create or update roles
INSERT INTO roles (id, name, description, created_at, updated_at) VALUES
    ('11111111-1111-1111-1111-111111111111', 'super-admin', 'System super administrator with full access', NOW(), NOW()),
    ('22222222-2222-2222-2222-222222222222', 'tenant-admin', 'Tenant administrator with full tenant access', NOW(), NOW()),
    ('33333333-3333-3333-3333-333333333333', 'tenant-editor', 'Tenant editor with limited access', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    updated_at = NOW();

-- 3. Create super admin user
-- Password: 'password' (bcrypt hash - $2a$10$bMPWlUqSeT.VdRH6dC1Ga.pUh.J8gM0UYTIAJm01BWM7pAya.IOAq)
INSERT INTO users (id, email, password_hash, full_name, is_active, created_at, updated_at) VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 
     'admin@phosboard.cl', 
     '$2a$10$bMPWlUqSeT.VdRH6dC1Ga.pUh.J8gM0UYTIAJm01BWM7pAya.IOAq', -- password
     'System Administrator',
     TRUE,
     NOW(),
     NOW())
ON CONFLICT (email) DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    full_name = EXCLUDED.full_name,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- 4. Assign super-admin role to admin user (global role, no tenant)
INSERT INTO user_roles (user_id, role_id, tenant_id, created_at) VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '11111111-1111-1111-1111-111111111111', NULL, NOW())
ON CONFLICT (user_id, role_id, tenant_id) DO NOTHING;

-- 5. Define permissions for our API endpoints
-- Note: Using wildcard * for method means all methods
INSERT INTO permissions (id, endpoint, method, description, created_at) VALUES
    -- Super Admin permissions (global access)
    ('11111111-1111-1111-1111-111111111101', '/api/v1/*', '*', 'Full access to all API v1 endpoints', NOW()),
    
    -- Tenant Admin permissions (tenant-scoped)
    ('11111111-1111-1111-1111-111111111102', '/api/v1/tenants/{tenant_id}/*', '*', 'Full access to tenant-specific endpoints', NOW()),
    ('11111111-1111-1111-1111-111111111103', '/api/v1/documents', '*', 'Full access to documents', NOW()),
    ('11111111-1111-1111-1111-111111111104', '/api/v1/concepts', '*', 'Full access to concepts', NOW()),
    ('11111111-1111-1111-1111-111111111105', '/api/v1/sources', '*', 'Full access to sources', NOW()),
    
    -- Tenant Editor permissions (limited)
    ('11111111-1111-1111-1111-111111111106', '/api/v1/documents', 'GET', 'Read access to documents', NOW()),
    ('11111111-1111-1111-1111-111111111107', '/api/v1/concepts', 'GET', 'Read access to concepts', NOW())
ON CONFLICT (endpoint, method) DO UPDATE SET
    description = EXCLUDED.description,
    created_at = NOW();

-- 6. Assign permissions to roles
-- Super Admin gets all permissions
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT 
    '11111111-1111-1111-1111-111111111111', 
    id, 
    NOW()
FROM permissions
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Tenant Admin gets tenant-specific permissions
INSERT INTO role_permissions (role_id, permission_id, created_at) VALUES
    ('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111102', NOW()),
    ('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111103', NOW()),
    ('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111104', NOW()),
    ('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111105', NOW())
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Tenant Editor gets read-only permissions
INSERT INTO role_permissions (role_id, permission_id, created_at) VALUES
    ('33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111106', NOW()),
    ('33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111107', NOW())
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 7. Create a test tenant admin user for demonstration
-- Password: 'password' (same bcrypt hash)
INSERT INTO users (id, email, password_hash, full_name, is_active, created_at, updated_at) VALUES
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 
     'tenant.admin@example.com', 
     '$2a$10$bMPWlUqSeT.VdRH6dC1Ga.pUh.J8gM0UYTIAJm01BWM7pAya.IOAq', -- password
     'Tenant Administrator',
     TRUE,
     NOW(),
     NOW())
ON CONFLICT (email) DO NOTHING;

-- Assign tenant-admin role to test user for default tenant
INSERT INTO user_roles (user_id, role_id, tenant_id, created_at) VALUES
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 
     '22222222-2222-2222-2222-222222222222', 
     '85c5f582-86b1-4217-bd4a-e1b1d0aac195', 
     NOW())
ON CONFLICT (user_id, role_id, tenant_id) DO NOTHING;
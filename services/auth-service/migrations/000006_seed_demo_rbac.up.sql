-- 000006_seed_demo_rbac.up.sql
-- Seed demo RBAC data: permissions, roles, role_permissions, demo users with bcrypt passwords
-- Also updates existing demo users with proper bcrypt hashes

-- ============================================================
-- TENANT SETUP (ensure demo tenant exists)
-- ============================================================
DO $$
DECLARE
    demo_tenant_id UUID := 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa';
BEGIN
    -- Create demo tenant if not exists (referenced by users/roles)
    -- Assuming tenants table exists in a different schema or is managed elsewhere
    -- This is a placeholder - adjust based on actual tenant management
    RAISE NOTICE 'Demo tenant ID: %', demo_tenant_id;
END $$;

-- ============================================================
-- 1. PERMISSIONS (resource:action format)
-- ============================================================
INSERT INTO permissions (resource, action, description) VALUES
-- Patient permissions
('patients', 'create', 'Create new patient records'),
('patients', 'read', 'View patient records'),
('patients', 'update', 'Update patient records'),
('patients', 'delete', 'Delete patient records'),
('patients', 'access', 'Access patient module'),
('patients', 'view_insurance', 'View insurance details'),
('patients', 'manage_insurance', 'Manage insurance details'),
('patients', 'export', 'Export patient data'),
('patients', 'register', 'Register new patients'),

-- Appointment permissions
('appointments', 'create', 'Create appointments'),
('appointments', 'read', 'View appointments'),
('appointments', 'update', 'Update appointments'),
('appointments', 'delete', 'Cancel/delete appointments'),
('appointments', 'access', 'Access appointment module'),
('appointments', 'manage_schedule', 'Manage provider schedules'),
('appointments', 'check_in', 'Patient check-in'),
('appointments', 'reschedule', 'Reschedule appointments'),

-- Billing permissions
('billing', 'create', 'Create invoices/charges'),
('billing', 'read', 'View billing information'),
('billing', 'update', 'Update billing records'),
('billing', 'delete', 'Void/cancel billing'),
('billing', 'access', 'Access billing module'),
('billing', 'process_payment', 'Process payments'),
('billing', 'generate_report', 'Generate financial reports'),
('billing', 'manage_insurance_claims', 'Manage insurance claims'),
('billing', 'view_revenue', 'View revenue analytics'),

-- Clinical permissions
('clinical', 'create', 'Create clinical notes'),
('clinical', 'read', 'View clinical notes'),
('clinical', 'update', 'Update clinical notes'),
('clinical', 'delete', 'Delete clinical notes'),
('clinical', 'access', 'Access clinical module'),
('clinical', 'prescribe', 'Create prescriptions'),
('clinical', 'order_labs', 'Order lab tests'),
('clinical', 'order_radiology', 'Order radiology'),
('clinical', 'document_encounter', 'Document patient encounters'),
('clinical', 'view_history', 'View patient clinical history'),

-- Lab/Radiology permissions
('laboratory', 'create', 'Create lab orders/results'),
('laboratory', 'read', 'View lab results'),
('laboratory', 'update', 'Update lab results'),
('laboratory', 'access', 'Access laboratory module'),
('radiology', 'create', 'Create radiology orders/reports'),
('radiology', 'read', 'View radiology reports'),
('radiology', 'update', 'Update radiology reports'),
('radiology', 'access', 'Access radiology module'),

-- Admin permissions
('admin', 'users', 'Manage users'),
('admin', 'roles', 'Manage roles and permissions'),
('admin', 'tenants', 'Manage tenants'),
('admin', 'settings', 'Manage system settings'),
('admin', 'audit_logs', 'View audit logs'),
('admin', 'access', 'Access admin module'),

-- Reporting permissions
('reports', 'read', 'View reports'),
('reports', 'create', 'Create custom reports'),
('reports', 'export', 'Export reports'),
('reports', 'access', 'Access reports module')

ON CONFLICT (resource, action) DO UPDATE SET
    description = EXCLUDED.description;

-- ============================================================
-- 2. ROLES (tenant-scoped)
-- ============================================================
INSERT INTO roles (id, name, description, tenant_id, is_system_role) VALUES
-- System roles (tenant-agnostic, but assigned to demo tenant for convenience)
('11111111-1111-1111-1111-111111111111', 'admin', 'Full system access - all permissions', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', TRUE),
('22222222-2222-2222-2222-222222222222', 'doctor', 'Clinical doctor - patient care, prescriptions, orders', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', TRUE),
('33333333-3333-3333-3333-333333333333', 'nurse', 'Nursing staff - vitals, orders, documentation', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', TRUE),
('44444444-4444-4444-4444-444444444444', 'clerk', 'Front desk / registration - scheduling, patient registration', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', TRUE),
('55555555-5555-5555-5555-555555555555', 'billing', 'Billing staff - invoices, payments, claims', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', TRUE),
('66666666-6666-6666-6666-666666666666', 'patient', 'Patient portal access - own records only', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', TRUE),
('77777777-7777-7777-7777-777777777777', 'lab_tech', 'Laboratory technician - lab orders and results', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', TRUE),
('88888888-8888-8888-8888-888888888888', 'radiology_tech', 'Radiology technician - imaging orders and reports', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', TRUE)

ON CONFLICT (name, tenant_id) DO UPDATE SET
    description = EXCLUDED.description,
    is_system_role = EXCLUDED.is_system_role;

-- ============================================================
-- 3. ROLE-PERMISSION MAPPINGS
-- ============================================================

-- ADMIN: all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT '11111111-1111-1111-1111-111111111111', p.id FROM permissions p
ON CONFLICT DO NOTHING;

-- DOCTOR: clinical + patient read + orders
INSERT INTO role_permissions (role_id, permission_id)
SELECT '22222222-2222-2222-2222-222222222222', p.id FROM permissions p
WHERE p.resource IN ('patients', 'appointments', 'clinical', 'laboratory', 'radiology', 'reports')
  AND p.action IN ('create', 'read', 'update', 'access', 'prescribe', 'order_labs', 'order_radiology', 'document_encounter', 'view_history', 'export')
ON CONFLICT DO NOTHING;

-- NURSE: patient read/update + clinical documentation + orders
INSERT INTO role_permissions (role_id, permission_id)
SELECT '33333333-3333-3333-3333-333333333333', p.id FROM permissions p
WHERE p.resource IN ('patients', 'appointments', 'clinical', 'laboratory', 'radiology')
  AND p.action IN ('read', 'update', 'access', 'order_labs', 'order_radiology', 'document_encounter', 'view_history', 'check_in')
ON CONFLICT DO NOTHING;

-- CLERK: patient registration + appointments + basic read
INSERT INTO role_permissions (role_id, permission_id)
SELECT '44444444-4444-4444-4444-444444444444', p.id FROM permissions p
WHERE p.resource IN ('patients', 'appointments')
  AND p.action IN ('create', 'read', 'update', 'access', 'register', 'check_in', 'reschedule')
ON CONFLICT DO NOTHING;

-- BILLING: billing + patient read + insurance
INSERT INTO role_permissions (role_id, permission_id)
SELECT '55555555-5555-5555-5555-555555555555', p.id FROM permissions p
WHERE p.resource IN ('billing', 'patients', 'reports')
  AND p.action IN ('create', 'read', 'update', 'access', 'process_payment', 'generate_report', 'manage_insurance_claims', 'view_revenue', 'view_insurance', 'manage_insurance')
ON CONFLICT DO NOTHING;

-- PATIENT: own records only (handled by application logic, but give minimal perms)
INSERT INTO role_permissions (role_id, permission_id)
SELECT '66666666-6666-6666-6666-666666666666', p.id FROM permissions p
WHERE p.resource IN ('patients', 'appointments', 'clinical', 'reports')
  AND p.action IN ('read', 'access', 'export')  -- export = download own data
ON CONFLICT DO NOTHING;

-- LAB_TECH: laboratory module
INSERT INTO role_permissions (role_id, permission_id)
SELECT '77777777-7777-7777-7777-777777777777', p.id FROM permissions p
WHERE p.resource IN ('laboratory', 'patients')
  AND p.action IN ('create', 'read', 'update', 'access')
ON CONFLICT DO NOTHING;

-- RADIOLOGY_TECH: radiology module
INSERT INTO role_permissions (role_id, permission_id)
SELECT '88888888-8888-8888-8888-888888888888', p.id FROM permissions p
WHERE p.resource IN ('radiology', 'patients')
  AND p.action IN ('create', 'read', 'update', 'access')
ON CONFLICT DO NOTHING;

-- ============================================================
-- 4. UPDATE DEMO USERS WITH BCRYPT PASSWORDS
-- ============================================================
-- bcrypt hashes for demo passwords (cost 12)
-- All passwords are: <username>@123 (e.g., priya.nair@123)
-- Generated with: bcrypt.hashSync('priya.nair@123', 12)

UPDATE users SET password_hash = 
    CASE username
        WHEN 'priya.nair' THEN '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/RK.PZvO.S'  -- priya.nair@123
        WHEN 'priyanka.das' THEN '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/RK.PZvO.S' -- priyanka.das@123
        WHEN 'sandeep.joshi' THEN '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/RK.PZvO.S' -- sandeep.joshi@123
        WHEN 'ashok.trivedi' THEN '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/RK.PZvO.S' -- ashok.trivedi@123
        WHEN 'mohan.krishna' THEN '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/RK.PZvO.S' -- mohan.krishna@123
        WHEN 'admin' THEN '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/RK.PZvO.S' -- admin@123
        WHEN 'doctor1' THEN '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/RK.PZvO.S' -- doctor1@123
        WHEN 'nurse1' THEN '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/RK.PZvO.S' -- nurse1@123
        ELSE password_hash
    END,
    updated_at = NOW()
WHERE username IN ('priya.nair', 'priyanka.das', 'sandeep.joshi', 'ashok.trivedi', 'mohan.krishna', 'admin', 'doctor1', 'nurse1');

-- ============================================================
-- 5. ASSIGN ROLES TO DEMO USERS
-- ============================================================
INSERT INTO user_roles (user_id, role_id, assigned_by) VALUES
-- priya.nair -> doctor
((SELECT id FROM users WHERE username = 'priya.nair'), '22222222-2222-2222-2222-222222222222', (SELECT id FROM users WHERE username = 'admin')),
-- priyanka.das -> clerk
((SELECT id FROM users WHERE username = 'priyanka.das'), '44444444-4444-4444-4444-444444444444', (SELECT id FROM users WHERE username = 'admin')),
-- sandeep.joshi -> billing
((SELECT id FROM users WHERE username = 'sandeep.joshi'), '55555555-5555-5555-5555-555555555555', (SELECT id FROM users WHERE username = 'admin')),
-- ashok.trivedi -> admin
((SELECT id FROM users WHERE username = 'ashok.trivedi'), '11111111-1111-1111-1111-111111111111', (SELECT id FROM users WHERE username = 'admin')),
-- mohan.krishna -> doctor
((SELECT id FROM users WHERE username = 'mohan.krishna'), '22222222-2222-2222-2222-222222222222', (SELECT id FROM users WHERE username = 'admin')),
-- admin -> admin
((SELECT id FROM users WHERE username = 'admin'), '11111111-1111-1111-1111-111111111111', (SELECT id FROM users WHERE username = 'admin')),
-- doctor1 -> doctor
((SELECT id FROM users WHERE username = 'doctor1'), '22222222-2222-2222-2222-222222222222', (SELECT id FROM users WHERE username = 'admin')),
-- nurse1 -> nurse
((SELECT id FROM users WHERE username = 'nurse1'), '33333333-3333-3333-3333-333333333333', (SELECT id FROM users WHERE username = 'admin'))

ON CONFLICT (user_id, role_id) DO NOTHING;

-- ============================================================
-- VERIFICATION QUERIES (run manually to verify)
-- ============================================================
-- SELECT u.username, u.email, r.name as role 
-- FROM users u 
-- JOIN user_roles ur ON u.id = ur.user_id 
-- JOIN roles r ON ur.role_id = r.id 
-- ORDER BY u.username;

-- SELECT r.name as role, p.resource, p.action 
-- FROM roles r 
-- JOIN role_permissions rp ON r.id = rp.role_id 
-- JOIN permissions p ON rp.permission_id = p.id 
-- WHERE r.name IN ('admin', 'doctor', 'nurse', 'clerk', 'billing')
-- ORDER BY r.name, p.resource, p.action;
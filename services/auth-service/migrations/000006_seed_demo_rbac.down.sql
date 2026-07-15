-- 000006_seed_demo_rbac.down.sql
-- Remove seed data (reverse order)

-- Remove user-role assignments
DELETE FROM user_roles 
WHERE user_id IN (SELECT id FROM users WHERE username IN (
    'priya.nair', 'priyanka.das', 'sandeep.joshi', 'ashok.trivedi', 'mohan.krishna',
    'admin', 'doctor1', 'nurse1'
));

-- Remove role-permissions
DELETE FROM role_permissions 
WHERE role_id IN (
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    '33333333-3333-3333-3333-333333333333',
    '44444444-4444-4444-4444-444444444444',
    '55555555-5555-5555-5555-555555555555',
    '66666666-6666-6666-6666-666666666666',
    '77777777-7777-7777-7777-777777777777',
    '88888888-8888-8888-8888-888888888888'
);

-- Remove demo roles
DELETE FROM roles 
WHERE id IN (
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    '33333333-3333-3333-3333-333333333333',
    '44444444-4444-4444-4444-444444444444',
    '55555555-5555-5555-5555-555555555555',
    '66666666-6666-6666-6666-666666666666',
    '77777777-7777-7777-7777-777777777777',
    '88888888-8888-8888-8888-888888888888'
);

-- Remove permissions
DELETE FROM permissions 
WHERE resource IN ('patients', 'appointments', 'billing', 'clinical', 'laboratory', 'radiology', 'admin', 'reports');

-- Note: Don't revert user password hashes - they're one-way
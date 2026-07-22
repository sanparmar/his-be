-- Doctor permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT '88888888-8888-8888-8888-888888888888', id FROM permissions WHERE name IN (
  'patient:read', 'encounter:read', 'encounter:write', 'encounter:close',
  'clinical:read', 'clinical:write', 'clinical:sign', 'order:read', 'order:write',
  'medication:read', 'medication:administer', 'diagnosis:write', 'procedure:write',
  'allergy:write', 'vitals:read', 'vitals:write', 'lab:order', 'lab:result:read',
  'rad:order', 'rad:result:read', 'ipd:read', 'ipd:write', 'opd:read', 'opd:write',
  'ed:read', 'ed:write', 'ed:disposition', 'appointment:read', 'appointment:schedule',
  'billing:read', 'report:run'
);

-- Senior Doctor (inherits doctor + additional)
INSERT INTO role_permissions (role_id, permission_id)
SELECT '99999999-9999-9999-9999-999999999999', id FROM permissions WHERE name IN (
  'patient:read', 'encounter:read', 'encounter:write', 'encounter:close',
  'clinical:read', 'clinical:write', 'clinical:sign', 'order:read', 'order:write', 'order:approve',
  'medication:read', 'medication:administer', 'diagnosis:write', 'procedure:write',
  'allergy:write', 'vitals:read', 'vitals:write', 'lab:order', 'lab:result:read', 'lab:result:verify',
  'rad:order', 'rad:result:read', 'rad:result:verify', 'ipd:read', 'ipd:write', 'opd:read', 'opd:write',
  'ed:read', 'ed:write', 'ed:disposition', 'appointment:read', 'appointment:schedule',
  'billing:read', 'report:run'
);

-- Nurse permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'dddddddd-dddd-dddd-dddd-dddddddddddd', id FROM permissions WHERE name IN (
  'patient:read', 'encounter:read', 'clinical:read', 'clinical:write',
  'order:read', 'medication:read', 'medication:administer', 'vitals:read', 'vitals:write',
  'lab:result:read', 'rad:result:read', 'ipd:read', 'ipd:write',
  'bed:assign', 'specimen:collect'
);

-- Senior Nurse (inherits nurse + additional)
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', id FROM permissions WHERE name IN (
  'patient:read', 'encounter:read', 'encounter:write', 'clinical:read', 'clinical:write',
  'order:read', 'medication:read', 'medication:administer', 'vitals:read', 'vitals:write',
  'lab:result:read', 'rad:result:read', 'ipd:read', 'ipd:write', 'bed:assign', 'bed:transfer',
  'specimen:collect', 'appointment:read', 'appointment:schedule'
);

-- Front Desk permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT '66666666-7777-8888-9999-000000000000', id FROM permissions WHERE name IN (
  'patient:read', 'patient:write', 'encounter:read', 'encounter:write',
  'appointment:read', 'appointment:schedule', 'appointment:reschedule', 'appointment:cancel',
  'checkin:process', 'billing:read', 'coverage:verify'
);

-- Billing Officer permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT '77777777-8888-9999-0000-111111111111', id FROM permissions WHERE name IN (
  'billing:read', 'billing:write', 'billing:post', 'invoice:generate',
  'payment:post', 'claim:create', 'report:run'
);

-- Pharmacist permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT '11111111-2222-3333-4444-555555555555', id FROM permissions WHERE name IN (
  'pharmacy:read', 'pharmacy:verify', 'pharmacy:dispense', 'interaction:check',
  'medication:read', 'medication:dispense', 'stock:manage', 'controlled:dispense'
);

-- Lab Technician permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT '22222222-3333-4444-5555-666666666666', id FROM permissions WHERE name IN (
  'lab:result:read', 'lab:result:verify', 'specimen:collect'
);

-- Radiology Technician permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT '33333333-4444-5555-6666-777777777777', id FROM permissions WHERE name IN (
  'rad:result:read', 'rad:result:verify'
);

-- Patient permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'bbbbbbbb-cccc-dddd-eeee-ffffffffffff', id FROM permissions WHERE name IN (
  'patient:read', 'appointment:read', 'appointment:schedule', 'appointment:reschedule', 'appointment:cancel'
);

-- Super Admin gets all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT '11111111-1111-1111-1111-111111111111', id FROM permissions;
-- Assign roles to seeded users (from 000002_seed_mock_data.up.sql)
-- admin -> super_admin
INSERT INTO user_roles (user_id, role_id, tenant_id, assigned_by, assigned_at)
VALUES (
  '11111111-1111-1111-1111-111111111111',
  '11111111-1111-1111-1111-111111111111',
  'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
  '11111111-1111-1111-1111-111111111111',
  NOW()
);

-- doctor1 -> doctor
INSERT INTO user_roles (user_id, role_id, tenant_id, assigned_by, assigned_at)
VALUES (
  '22222222-2222-2222-2222-222222222222',
  '88888888-8888-8888-8888-888888888888',
  'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
  '11111111-1111-1111-1111-111111111111',
  NOW()
);

-- nurse1 -> nurse
INSERT INTO user_roles (user_id, role_id, tenant_id, assigned_by, assigned_at)
VALUES (
  '33333333-3333-3333-3333-333333333333',
  'dddddddd-dddd-dddd-dddd-dddddddddddd',
  'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
  '11111111-1111-1111-1111-111111111111',
  NOW()
);
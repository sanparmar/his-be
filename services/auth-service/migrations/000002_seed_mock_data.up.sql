BEGIN;

-- Mock users for local/dev testing.
INSERT INTO users (
    id,
    username,
    email,
    password_hash`,
    tenant_id,
    organization_id,
    hospital_id,
    created_at,
    updated_at
) VALUES
(
    '11111111-1111-1111-1111-111111111111',
    'admin',
    'admin@example.com',
    '$2a$12$uMseXOXn1LDKyy0n3r8/LuAdD/qIC4/2z7fijg3M1qGIYlYYYal66',
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    'cccccccc-cccc-cccc-cccc-cccccccccccc',
    NOW(),
    NOW()
),
(
    '22222222-2222-2222-2222-222222222222',
    'doctor1',
    'doctor1@example.com',
    '$2a$12$mockedhashfordoctor1234567890abcdef',
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    'cccccccc-cccc-cccc-cccc-cccccccccccc',
    NOW(),
    NOW()
),
(
    '33333333-3333-3333-3333-333333333333',
    'nurse1',
    'nurse1@example.com',
    '$2a$12$mockedhashfornurse1234567890abcdefg',
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    'cccccccc-cccc-cccc-cccc-cccccccccccc',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO UPDATE
SET
    username = EXCLUDED.username,
    email = EXCLUDED.email,
    password_hash = EXCLUDED.password_hash,
    tenant_id = EXCLUDED.tenant_id,
    organization_id = EXCLUDED.organization_id,
    hospital_id = EXCLUDED.hospital_id,
    updated_at = NOW();

-- Mock sessions linked to seeded users.
INSERT INTO sessions (
    id,
    user_id,
    token,
    expires_at,
    ip_address,
    user_agent,
    created_at
) VALUES
(
    '44444444-4444-4444-4444-444444444444',
    '11111111-1111-1111-1111-111111111111',
    'seed-refresh-token-admin',
    NOW() + INTERVAL '7 days',
    '127.0.0.1',
    'seed-script',
    NOW()
),
(
    '55555555-5555-5555-5555-555555555555',
    '22222222-2222-2222-2222-222222222222',
    'seed-refresh-token-doctor1',
    NOW() + INTERVAL '7 days',
    '127.0.0.1',
    'seed-script',
    NOW()
)
ON CONFLICT (id) DO UPDATE
SET
    user_id = EXCLUDED.user_id,
    token = EXCLUDED.token,
    expires_at = EXCLUDED.expires_at,
    ip_address = EXCLUDED.ip_address,
    user_agent = EXCLUDED.user_agent;

COMMIT;
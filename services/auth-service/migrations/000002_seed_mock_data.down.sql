BEGIN;

DELETE FROM sessions
WHERE id IN (
    '44444444-4444-4444-4444-444444444444',
    '55555555-5555-5555-5555-555555555555'
)
OR token IN (
    'seed-refresh-token-admin',
    'seed-refresh-token-doctor1'
);

DELETE FROM users
WHERE id IN (
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    '33333333-3333-3333-3333-333333333333'
)
OR username IN ('admin', 'doctor1', 'nurse1')
OR email IN ('admin@example.com', 'doctor1@example.com', 'nurse1@example.com');

COMMIT;
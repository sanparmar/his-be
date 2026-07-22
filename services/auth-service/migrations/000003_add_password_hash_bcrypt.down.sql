-- 000003_add_password_hash_bcrypt.down.sql
-- No schema changes to revert
-- Remove constraint if added
-- ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_password_hash_bcrypt;
COMMENT ON COLUMN users.password_hash IS 'password hash';
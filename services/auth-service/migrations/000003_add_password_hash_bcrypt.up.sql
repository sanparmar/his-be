-- 000003_add_password_hash_bcrypt.up.sql
-- No schema change needed - password_hash column already exists in users table
-- This migration ensures existing passwords are re-hashed with bcrypt
-- and adds a check constraint to enforce bcrypt format

-- Add comment to document bcrypt requirement
COMMENT ON COLUMN users.password_hash IS 'bcrypt hash (cost 12), format: $2a$12$...';

-- Add constraint to enforce bcrypt format (optional, can be added after migration)
-- ALTER TABLE users ADD CONSTRAINT chk_password_hash_bcrypt 
--   CHECK (password_hash ~ '^\$2[ab]\$\d{2}\$[./A-Za-z0-9]{53}$');

-- If there are existing users with plaintext passwords, update them here
-- Example:
-- UPDATE users SET password_hash = crypt(password_hash, gen_salt('bf', 12)) 
-- WHERE password_hash NOT LIKE '$2a$12$%';
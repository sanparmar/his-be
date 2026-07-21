-- 000005_add_mfa_tables.up.sql
-- MFA credentials table for TOTP and WebAuthn

CREATE TABLE IF NOT EXISTS mfa_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('totp', 'webauthn', 'backup_codes')),
    -- TOTP fields
    secret TEXT,                          -- base32 encoded secret for TOTP
    -- WebAuthn fields
    credential_id TEXT,                   -- WebAuthn credential ID (base64url)
    public_key TEXT,                      -- COSE public key (base64)
    sign_count BIGINT,                    -- WebAuthn signature counter
    -- Common fields
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    backup_codes TEXT[],                  -- array of hashed backup codes (bcrypt)
    device_name TEXT,                     -- user-friendly name (e.g., "Authenticator App", "YubiKey")
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Ensure one enabled TOTP per user
    UNIQUE(user_id, type) WHERE (enabled AND type = 'totp')
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_mfa_credentials_user ON mfa_credentials(user_id);
CREATE INDEX IF NOT EXISTS idx_mfa_credentials_type ON mfa_credentials(type);
CREATE INDEX IF NOT EXISTS idx_mfa_credentials_credential_id ON mfa_credentials(credential_id);

-- Comments
COMMENT ON TABLE mfa_credentials IS 'Multi-factor authentication credentials (TOTP, WebAuthn, Backup Codes)';
COMMENT ON COLUMN mfa_credentials.type IS 'Type: totp, webauthn, or backup_codes';
COMMENT ON COLUMN mfa_credentials.backup_codes IS 'Array of bcrypt-hashed backup codes';
COMMENT ON COLUMN mfa_credentials.secret IS 'TOTP secret in base32 encoding';
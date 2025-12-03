-- +goose Up
-- Auth credentials table
CREATE TABLE IF NOT EXISTS auth_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Auth sessions table
CREATE TABLE IF NOT EXISTS auth_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE,
    revoked_at TIMESTAMP WITH TIME ZONE,
    user_agent TEXT,
    ip_address VARCHAR(45)
);

-- Indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_auth_credentials_username ON auth_credentials(username) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_auth_credentials_email ON auth_credentials(email) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_auth_credentials_user_id ON auth_credentials(user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_auth_sessions_token ON auth_sessions(token) WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_auth_sessions_user_id ON auth_sessions(user_id) WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_auth_sessions_expires_at ON auth_sessions(expires_at);

-- +goose Down
DROP TABLE auth_sessions IF EXISTS;
DROP TABLE auth_credentials IF EXISTS;
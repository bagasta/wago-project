CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_name VARCHAR(100) NOT NULL,
    webhook_url TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'disconnected',
    qr_code TEXT,
    phone_number VARCHAR(20),
    device_info JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_connected TIMESTAMP,
    uptime_seconds BIGINT DEFAULT 0,
    CONSTRAINT unique_session_name UNIQUE(user_id, session_name),
    CONSTRAINT valid_status CHECK (status IN ('qr', 'connected', 'disconnected'))
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);
CREATE INDEX IF NOT EXISTS idx_sessions_created_at ON sessions(created_at DESC);

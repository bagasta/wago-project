CREATE TABLE IF NOT EXISTS analytics (
    id BIGSERIAL PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    message_id VARCHAR(100),
    from_number VARCHAR(50),
    message_type VARCHAR(20),
    is_group BOOLEAN DEFAULT false,
    is_mention BOOLEAN DEFAULT false,
    webhook_sent BOOLEAN DEFAULT false,
    webhook_success BOOLEAN DEFAULT false,
    webhook_response_time_ms INTEGER,
    webhook_status_code INTEGER,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_message_type CHECK (message_type IN ('text', 'image', 'document', 'audio', 'video', 'sticker', 'location', 'contact'))
);

CREATE INDEX IF NOT EXISTS idx_analytics_session_id ON analytics(session_id);
CREATE INDEX IF NOT EXISTS idx_analytics_created_at ON analytics(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_analytics_session_created ON analytics(session_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_analytics_webhook_sent ON analytics(webhook_sent, created_at);
CREATE INDEX IF NOT EXISTS idx_analytics_is_group ON analytics(is_group, created_at);

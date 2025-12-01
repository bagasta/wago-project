CREATE TABLE IF NOT EXISTS messages_log (
    id BIGSERIAL PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    direction VARCHAR(10) NOT NULL,
    from_number VARCHAR(50),
    to_number VARCHAR(50),
    message_type VARCHAR(20),
    content TEXT,
    media_url TEXT,
    group_id VARCHAR(100),
    group_name VARCHAR(200),
    is_group BOOLEAN DEFAULT false,
    quoted_message_id VARCHAR(100),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_direction CHECK (direction IN ('incoming', 'outgoing'))
);

CREATE INDEX IF NOT EXISTS idx_messages_session_timestamp ON messages_log(session_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_messages_from_number ON messages_log(from_number);
CREATE INDEX IF NOT EXISTS idx_messages_group_id ON messages_log(group_id) WHERE is_group = true;

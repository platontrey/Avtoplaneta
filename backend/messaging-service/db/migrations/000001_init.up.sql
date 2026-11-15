-- 000001_init.up.sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS conversations (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_message_at TIMESTAMP WITH TIME ZONE NOT NULL,
    title VARCHAR(255) NOT NULL DEFAULT '',
    last_message TEXT NOT NULL DEFAULT '',
    participants BIGINT[] NOT NULL
);

CREATE TABLE IF NOT EXISTS messages (
    id BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL,
    sender_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    message_type VARCHAR(50) NOT NULL DEFAULT 'text',
    voice_url TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    read_by BIGINT[] NOT NULL DEFAULT '{}',
    attachments TEXT[] NOT NULL DEFAULT '{}',
    search_vector tsvector
) WITH (fillfactor = 85);

CREATE TABLE IF NOT EXISTS mentions (
    id BIGSERIAL PRIMARY KEY,
    message_id BIGINT NOT NULL,
    resource_id BIGINT NOT NULL,
    type VARCHAR(50) NOT NULL,
    text TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS reactions (
    id BIGSERIAL PRIMARY KEY,
    message_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    emoji VARCHAR(50) NOT NULL
);

CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    related_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    read BOOLEAN NOT NULL DEFAULT FALSE,
    type VARCHAR(50) NOT NULL,
    message TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS user_status (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT UNIQUE NOT NULL,
    last_seen TIMESTAMP WITH TIME ZONE NOT NULL,
    is_online BOOLEAN NOT NULL DEFAULT FALSE
) WITH (fillfactor = 85);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_conversations_participants ON conversations USING GIN (participants);
CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages (conversation_id);
CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages (sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_read_by ON messages USING GIN (read_by);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_conv_created ON messages (conversation_id, id DESC);
CREATE INDEX IF NOT EXISTS idx_messages_search ON messages USING GIN (search_vector);
CREATE INDEX IF NOT EXISTS idx_messages_content_trgm ON messages USING GIN (content gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications (user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_user_status_user_id ON user_status (user_id);
CREATE INDEX IF NOT EXISTS idx_reactions_message_id ON reactions (message_id);
CREATE INDEX IF NOT EXISTS idx_mentions_message_id ON mentions (message_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_reactions_unique ON reactions (message_id, user_id, emoji);

-- Alter statements for migration if tables exist
ALTER TABLE messages ADD COLUMN IF NOT EXISTS search_vector tsvector;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS message_type VARCHAR(50) DEFAULT 'text';
ALTER TABLE messages ADD COLUMN IF NOT EXISTS voice_url TEXT;

UPDATE messages SET search_vector = to_tsvector('russian', content) WHERE search_vector IS NULL;

-- Trigger to automatically update search_vector
CREATE OR REPLACE FUNCTION messages_search_vector_trigger() RETURNS trigger AS $$
BEGIN
  new.search_vector := to_tsvector('russian', COALESCE(new.content, ''));
  RETURN new;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS tsvectorupdate ON messages;
CREATE TRIGGER tsvectorupdate BEFORE INSERT OR UPDATE ON messages
FOR EACH ROW EXECUTE FUNCTION messages_search_vector_trigger();

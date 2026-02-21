-- Index on expires_at for efficient cleanup queries (DELETE WHERE expires_at < NOW()).
-- token_hash already has an implicit B-tree index from the UNIQUE constraint.
-- idx_refresh_tokens_user_id and idx_refresh_tokens_active were created in 000014.
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

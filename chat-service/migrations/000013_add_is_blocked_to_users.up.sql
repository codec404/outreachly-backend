ALTER TABLE users
    ADD COLUMN is_blocked BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX idx_users_is_blocked ON users (is_blocked) WHERE is_blocked = TRUE;

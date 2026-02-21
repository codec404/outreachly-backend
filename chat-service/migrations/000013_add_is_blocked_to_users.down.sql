DROP INDEX IF EXISTS idx_users_is_blocked;

ALTER TABLE users
    DROP COLUMN IF EXISTS is_blocked;

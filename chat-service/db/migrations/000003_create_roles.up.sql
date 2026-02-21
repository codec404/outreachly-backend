CREATE TABLE roles (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(50)  NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Seed default roles.
INSERT INTO roles (name, description) VALUES
    ('admin', 'Full system access'),
    ('user',  'Standard user access');

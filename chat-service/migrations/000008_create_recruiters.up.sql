CREATE TABLE recruiters (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name         VARCHAR(255) NOT NULL,
    company      VARCHAR(255),
    email        VARCHAR(255) NOT NULL,
    linkedin_url TEXT,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    -- A user cannot store the same recruiter email twice.
    CONSTRAINT recruiters_user_email_unique UNIQUE (user_id, email),

    CONSTRAINT recruiters_email_format CHECK (email ~* '^[^@\s]+@[^@\s]+\.[^@\s]+$')
);

CREATE INDEX idx_recruiters_user_id ON recruiters (user_id);
CREATE INDEX idx_recruiters_email   ON recruiters (email);

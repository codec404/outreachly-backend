CREATE TABLE templates (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    subject         VARCHAR(500) NOT NULL,
    body            TEXT,
    s3_url          TEXT,
    is_ai_generated BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    -- Exactly one of body or s3_url must be set.
    CONSTRAINT template_has_content CHECK (
        (body IS NOT NULL AND s3_url IS NULL) OR
        (body IS NULL     AND s3_url IS NOT NULL)
    )
);

CREATE INDEX idx_templates_user_id        ON templates (user_id);
CREATE INDEX idx_templates_is_ai_generated ON templates (is_ai_generated);

CREATE TRIGGER set_templates_updated_at
    BEFORE UPDATE ON templates
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

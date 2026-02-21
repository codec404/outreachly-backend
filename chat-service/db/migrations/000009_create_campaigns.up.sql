CREATE TYPE campaign_status AS ENUM (
    'draft',
    'scheduled',
    'running',
    'completed',
    'failed'
);

CREATE TABLE campaigns (
    id               UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID            NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    -- Allow templates to be deleted without destroying campaign history.
    template_id      UUID            REFERENCES templates(id) ON DELETE SET NULL,
    name             VARCHAR(255)    NOT NULL,
    status           campaign_status NOT NULL DEFAULT 'draft',
    total_recipients INTEGER         NOT NULL DEFAULT 0 CHECK (total_recipients >= 0),
    scheduled_at     TIMESTAMPTZ,
    created_at       TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ     NOT NULL DEFAULT NOW(),

    CONSTRAINT campaigns_scheduled_requires_time CHECK (
        status != 'scheduled' OR scheduled_at IS NOT NULL
    )
);

CREATE INDEX idx_campaigns_user_id      ON campaigns (user_id);
CREATE INDEX idx_campaigns_status       ON campaigns (status);
-- Partial index: only future scheduled campaigns need fast lookup.
CREATE INDEX idx_campaigns_scheduled_at ON campaigns (scheduled_at)
    WHERE scheduled_at IS NOT NULL AND status = 'scheduled';

CREATE TRIGGER set_campaigns_updated_at
    BEFORE UPDATE ON campaigns
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

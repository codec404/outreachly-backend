CREATE TYPE email_event_type AS ENUM (
    'delivered',
    'opened',
    'clicked',
    'bounced',
    'complaint',
    'unsubscribed'
);

CREATE TABLE email_events (
    id                 UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_target_id UUID             NOT NULL REFERENCES campaign_targets(id) ON DELETE CASCADE,
    event_type         email_event_type NOT NULL,
    -- Raw SES/webhook payload stored as JSONB for flexible querying.
    metadata           JSONB,
    created_at         TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_email_events_campaign_target_id ON email_events (campaign_target_id);
CREATE INDEX idx_email_events_event_type         ON email_events (event_type);
-- Time-range queries for analytics dashboards.
CREATE INDEX idx_email_events_created_at         ON email_events (created_at DESC);
-- GIN index for querying inside the metadata JSONB payload.
CREATE INDEX idx_email_events_metadata           ON email_events USING GIN (metadata)
    WHERE metadata IS NOT NULL;

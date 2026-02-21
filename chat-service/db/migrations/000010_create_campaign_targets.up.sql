CREATE TYPE target_status AS ENUM (
    'pending',
    'sent',
    'failed',
    'bounced'
);

CREATE TABLE campaign_targets (
    id            UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id   UUID          NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    recruiter_id  UUID          NOT NULL REFERENCES recruiters(id) ON DELETE CASCADE,
    status        target_status NOT NULL DEFAULT 'pending',
    error_message TEXT,
    sent_at       TIMESTAMPTZ,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    -- A recruiter can only appear once per campaign.
    CONSTRAINT campaign_targets_unique UNIQUE (campaign_id, recruiter_id),

    CONSTRAINT campaign_targets_sent_at_requires_sent CHECK (
        status != 'sent' OR sent_at IS NOT NULL
    )
);

CREATE INDEX idx_campaign_targets_campaign_id  ON campaign_targets (campaign_id);
CREATE INDEX idx_campaign_targets_recruiter_id ON campaign_targets (recruiter_id);
CREATE INDEX idx_campaign_targets_status       ON campaign_targets (status);

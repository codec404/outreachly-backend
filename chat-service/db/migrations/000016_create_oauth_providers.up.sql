CREATE TABLE oauth_providers (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider         VARCHAR(64)  NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    email            VARCHAR(255) NOT NULL,
    name             VARCHAR(255) NOT NULL DEFAULT '',
    avatar_url       TEXT         NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    -- One provider account can only be linked to one local user.
    CONSTRAINT uq_oauth_provider_user UNIQUE (provider, provider_user_id)
);

CREATE INDEX idx_oauth_providers_user_id ON oauth_providers(user_id);
CREATE INDEX idx_oauth_providers_lookup  ON oauth_providers(provider, provider_user_id);

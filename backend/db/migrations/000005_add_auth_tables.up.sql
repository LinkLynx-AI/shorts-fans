CREATE TABLE app.auth_identities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES app.users (id),
    provider TEXT NOT NULL,
    provider_subject TEXT NOT NULL,
    email_normalized TEXT,
    verified_at TIMESTAMPTZ,
    last_authenticated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (provider, provider_subject),
    CHECK (length(btrim(provider)) > 0),
    CHECK (length(btrim(provider_subject)) > 0),
    CHECK (
        email_normalized IS NULL
        OR (
            email_normalized = lower(btrim(email_normalized))
            AND length(btrim(email_normalized)) > 0
        )
    )
);

CREATE INDEX idx_auth_identities_user_id
    ON app.auth_identities (user_id, created_at DESC, id DESC);

CREATE INDEX idx_auth_identities_email_normalized
    ON app.auth_identities (email_normalized)
    WHERE email_normalized IS NOT NULL;

CREATE TABLE app.auth_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES app.users (id),
    active_mode TEXT NOT NULL DEFAULT 'fan' CHECK (
        active_mode IN ('fan', 'creator')
    ),
    session_token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (length(btrim(session_token_hash)) > 0)
);

CREATE INDEX idx_auth_sessions_user_id
    ON app.auth_sessions (user_id, created_at DESC, id DESC);

CREATE INDEX idx_auth_sessions_active_lookup
    ON app.auth_sessions (expires_at, id)
    WHERE revoked_at IS NULL;

CREATE TABLE app.auth_login_challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider TEXT NOT NULL,
    provider_subject TEXT NOT NULL,
    email_normalized TEXT,
    challenge_token_hash TEXT NOT NULL,
    purpose TEXT NOT NULL DEFAULT 'login' CHECK (
        purpose IN ('login')
    ),
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    attempt_count INTEGER NOT NULL DEFAULT 0 CHECK (
        attempt_count >= 0
    ),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (length(btrim(provider)) > 0),
    CHECK (length(btrim(provider_subject)) > 0),
    CHECK (length(btrim(challenge_token_hash)) > 0),
    CHECK (
        email_normalized IS NULL
        OR (
            email_normalized = lower(btrim(email_normalized))
            AND length(btrim(email_normalized)) > 0
        )
    )
);

CREATE INDEX idx_auth_login_challenges_subject_created_at
    ON app.auth_login_challenges (provider, provider_subject, created_at DESC, id DESC);

CREATE INDEX idx_auth_login_challenges_pending_expiration
    ON app.auth_login_challenges (expires_at, id)
    WHERE consumed_at IS NULL;

ALTER TABLE app.auth_sessions
    ADD COLUMN recent_authenticated_at TIMESTAMPTZ;

UPDATE app.auth_sessions
SET recent_authenticated_at = created_at
WHERE recent_authenticated_at IS NULL;

ALTER TABLE app.auth_sessions
    ALTER COLUMN recent_authenticated_at SET DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE app.auth_sessions
    ALTER COLUMN recent_authenticated_at SET NOT NULL;

CREATE UNIQUE INDEX idx_auth_identities_cognito_email_normalized
    ON app.auth_identities (email_normalized)
    WHERE provider = 'cognito' AND email_normalized IS NOT NULL;

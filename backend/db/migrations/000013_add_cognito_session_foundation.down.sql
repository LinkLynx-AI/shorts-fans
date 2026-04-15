DROP INDEX IF EXISTS app.idx_auth_identities_cognito_email_normalized;

ALTER TABLE app.auth_sessions
    DROP COLUMN IF EXISTS recent_authenticated_at;

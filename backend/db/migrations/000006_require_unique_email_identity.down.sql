DROP INDEX IF EXISTS app.idx_auth_identities_email_normalized;

ALTER TABLE app.auth_identities
    DROP CONSTRAINT IF EXISTS auth_identities_email_provider_requires_email_check;

CREATE INDEX idx_auth_identities_email_normalized
    ON app.auth_identities (email_normalized)
    WHERE email_normalized IS NOT NULL;

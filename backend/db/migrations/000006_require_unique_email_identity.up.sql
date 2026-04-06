DROP INDEX IF EXISTS app.idx_auth_identities_email_normalized;

ALTER TABLE app.auth_identities
    ADD CONSTRAINT auth_identities_email_provider_requires_email_check CHECK (
        provider <> 'email'
        OR email_normalized IS NOT NULL
    );

CREATE UNIQUE INDEX idx_auth_identities_email_normalized
    ON app.auth_identities (email_normalized)
    WHERE provider = 'email'
        AND email_normalized IS NOT NULL;

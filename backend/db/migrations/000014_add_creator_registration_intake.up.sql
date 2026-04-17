CREATE TABLE app.creator_registration_intakes (
    user_id UUID PRIMARY KEY REFERENCES app.creator_capabilities (user_id) ON DELETE CASCADE,
    legal_name TEXT NOT NULL DEFAULT '',
    birth_date DATE,
    payout_recipient_type TEXT CHECK (
        payout_recipient_type IS NULL
        OR payout_recipient_type IN ('self', 'business')
    ),
    payout_recipient_name TEXT NOT NULL DEFAULT '',
    declares_no_prohibited_category BOOLEAN NOT NULL DEFAULT FALSE,
    accepts_consent_responsibility BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE app.creator_registration_evidences (
    user_id UUID NOT NULL REFERENCES app.creator_capabilities (user_id) ON DELETE CASCADE,
    kind TEXT NOT NULL CHECK (
        kind IN ('government_id', 'payout_proof')
    ),
    file_name TEXT NOT NULL,
    mime_type TEXT NOT NULL CHECK (
        mime_type IN (
            'image/jpeg',
            'image/png',
            'image/webp',
            'application/pdf'
        )
    ),
    file_size_bytes BIGINT NOT NULL CHECK (
        file_size_bytes > 0
        AND file_size_bytes <= 10485760
    ),
    storage_bucket TEXT NOT NULL,
    storage_key TEXT NOT NULL,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, kind)
);

CREATE INDEX idx_creator_registration_evidences_user_id
    ON app.creator_registration_evidences (user_id);

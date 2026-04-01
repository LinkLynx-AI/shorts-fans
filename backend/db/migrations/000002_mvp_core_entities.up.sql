CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE app.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    auth_provider_user_ref TEXT UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE app.creator_capabilities (
    user_id UUID PRIMARY KEY REFERENCES app.users (id),
    state TEXT NOT NULL CHECK (
        state IN ('draft', 'submitted', 'approved', 'rejected', 'suspended')
    ),
    rejection_reason_code TEXT,
    is_resubmit_eligible BOOLEAN NOT NULL DEFAULT FALSE,
    is_support_review_required BOOLEAN NOT NULL DEFAULT FALSE,
    self_serve_resubmit_count INTEGER NOT NULL DEFAULT 0 CHECK (
        self_serve_resubmit_count BETWEEN 0 AND 2
    ),
    kyc_provider_case_ref TEXT UNIQUE,
    payout_provider_account_ref TEXT UNIQUE,
    submitted_at TIMESTAMPTZ,
    approved_at TIMESTAMPTZ,
    rejected_at TIMESTAMPTZ,
    suspended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_creator_capabilities_state
    ON app.creator_capabilities (state);

CREATE TABLE app.creator_profile_drafts (
    user_id UUID PRIMARY KEY REFERENCES app.creator_capabilities (user_id),
    display_name TEXT,
    avatar_url TEXT,
    bio TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE app.creator_profiles (
    user_id UUID PRIMARY KEY REFERENCES app.creator_capabilities (user_id),
    display_name TEXT NOT NULL CHECK (length(btrim(display_name)) > 0),
    avatar_url TEXT,
    bio TEXT NOT NULL DEFAULT '',
    published_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE app.media_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_user_id UUID NOT NULL REFERENCES app.creator_capabilities (user_id),
    processing_state TEXT NOT NULL CHECK (
        processing_state IN ('uploaded', 'ready', 'failed')
    ),
    storage_provider TEXT NOT NULL,
    storage_bucket TEXT NOT NULL,
    storage_key TEXT NOT NULL,
    playback_url TEXT,
    mime_type TEXT NOT NULL,
    duration_ms BIGINT CHECK (duration_ms IS NULL OR duration_ms > 0),
    external_upload_ref TEXT UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (id, creator_user_id),
    CHECK (
        (
            processing_state = 'ready'
            AND playback_url IS NOT NULL
            AND duration_ms IS NOT NULL
        )
        OR (
            processing_state IN ('uploaded', 'failed')
            AND playback_url IS NULL
        )
    )
);

CREATE INDEX idx_media_assets_creator_user_id
    ON app.media_assets (creator_user_id);

CREATE TABLE app.mains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_user_id UUID NOT NULL REFERENCES app.creator_capabilities (user_id),
    media_asset_id UUID NOT NULL UNIQUE,
    state TEXT NOT NULL CHECK (
        state IN (
            'draft',
            'pending_review',
            'approved_for_unlock',
            'revision_requested',
            'rejected',
            'suspended',
            'removed'
        )
    ),
    review_reason_code TEXT,
    post_report_state TEXT CHECK (
        post_report_state IS NULL
        OR post_report_state IN (
            'under_review',
            'temporarily_limited',
            'removed',
            'restored'
        )
    ),
    price_minor BIGINT CHECK (price_minor IS NULL OR price_minor >= 0),
    currency_code TEXT CHECK (
        currency_code IS NULL
        OR currency_code ~ '^[A-Z]{3}$'
    ),
    ownership_confirmed BOOLEAN NOT NULL DEFAULT FALSE,
    consent_confirmed BOOLEAN NOT NULL DEFAULT FALSE,
    approved_for_unlock_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (id, creator_user_id),
    FOREIGN KEY (media_asset_id, creator_user_id)
        REFERENCES app.media_assets (id, creator_user_id),
    CHECK (
        (price_minor IS NULL AND currency_code IS NULL)
        OR (price_minor IS NOT NULL AND currency_code IS NOT NULL)
    )
);

CREATE INDEX idx_mains_creator_user_id
    ON app.mains (creator_user_id);

CREATE INDEX idx_mains_state
    ON app.mains (state);

CREATE TABLE app.shorts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_user_id UUID NOT NULL REFERENCES app.creator_capabilities (user_id),
    canonical_main_id UUID NOT NULL,
    media_asset_id UUID NOT NULL UNIQUE,
    state TEXT NOT NULL CHECK (
        state IN (
            'draft',
            'pending_review',
            'approved_for_publish',
            'revision_requested',
            'rejected',
            'removed'
        )
    ),
    review_reason_code TEXT,
    post_report_state TEXT CHECK (
        post_report_state IS NULL
        OR post_report_state IN (
            'under_review',
            'temporarily_limited',
            'removed',
            'restored'
        )
    ),
    approved_for_publish_at TIMESTAMPTZ,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (canonical_main_id, creator_user_id)
        REFERENCES app.mains (id, creator_user_id),
    FOREIGN KEY (media_asset_id, creator_user_id)
        REFERENCES app.media_assets (id, creator_user_id)
);

CREATE INDEX idx_shorts_creator_user_id
    ON app.shorts (creator_user_id);

CREATE INDEX idx_shorts_canonical_main_id
    ON app.shorts (canonical_main_id);

CREATE INDEX idx_shorts_state
    ON app.shorts (state);

CREATE TABLE app.main_unlocks (
    user_id UUID NOT NULL REFERENCES app.users (id),
    main_id UUID NOT NULL REFERENCES app.mains (id),
    payment_provider_purchase_ref TEXT UNIQUE,
    purchased_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, main_id)
);

CREATE INDEX idx_main_unlocks_user_id
    ON app.main_unlocks (user_id);

CREATE INDEX idx_main_unlocks_main_id
    ON app.main_unlocks (main_id);

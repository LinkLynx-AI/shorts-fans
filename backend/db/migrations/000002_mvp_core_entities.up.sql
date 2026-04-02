CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE app.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (state <> 'submitted' OR submitted_at IS NOT NULL),
    CHECK (state <> 'approved' OR approved_at IS NOT NULL),
    CHECK (state <> 'rejected' OR rejected_at IS NOT NULL),
    CHECK (state <> 'suspended' OR suspended_at IS NOT NULL),
    CHECK (approved_at IS NULL OR state IN ('approved', 'suspended')),
    CHECK (rejected_at IS NULL OR state = 'rejected'),
    CHECK (suspended_at IS NULL OR state = 'suspended')
);

CREATE INDEX idx_creator_capabilities_state
    ON app.creator_capabilities (state);

CREATE TABLE app.creator_profile_drafts (
    user_id UUID PRIMARY KEY REFERENCES app.creator_capabilities (user_id),
    display_name TEXT,
    avatar_url TEXT,
    bio TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (
        display_name IS NULL
        OR length(btrim(display_name)) > 0
    )
);

CREATE TABLE app.creator_profiles (
    user_id UUID PRIMARY KEY REFERENCES app.creator_capabilities (user_id),
    display_name TEXT NOT NULL,
    avatar_url TEXT,
    bio TEXT NOT NULL DEFAULT '',
    published_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (
        display_name IS NOT NULL
        AND length(btrim(display_name)) > 0
    )
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
    ),
    CHECK (
        approved_for_unlock_at IS NULL
        OR state IN ('approved_for_unlock', 'suspended', 'removed')
    ),
    CHECK (
        state <> 'approved_for_unlock'
        OR (
            ownership_confirmed
            AND consent_confirmed
            AND approved_for_unlock_at IS NOT NULL
        )
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
        REFERENCES app.media_assets (id, creator_user_id),
    CHECK (
        approved_for_publish_at IS NULL
        OR state IN ('approved_for_publish', 'removed')
    ),
    CHECK (
        state <> 'approved_for_publish'
        OR approved_for_publish_at IS NOT NULL
    ),
    CHECK (
        published_at IS NULL
        OR state IN ('approved_for_publish', 'removed')
    ),
    CHECK (
        published_at IS NULL
        OR approved_for_publish_at IS NOT NULL
    ),
    CHECK (
        published_at IS NULL
        OR published_at >= approved_for_publish_at
    )
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

CREATE TABLE app.creator_follows (
    user_id UUID NOT NULL REFERENCES app.users (id),
    creator_user_id UUID NOT NULL REFERENCES app.creator_profiles (user_id),
    followed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, creator_user_id),
    CHECK (user_id <> creator_user_id)
);

CREATE INDEX idx_creator_follows_creator_user_id
    ON app.creator_follows (creator_user_id);

CREATE TABLE app.pinned_shorts (
    user_id UUID NOT NULL REFERENCES app.users (id),
    short_id UUID NOT NULL REFERENCES app.shorts (id),
    pinned_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, short_id)
);

CREATE INDEX idx_pinned_shorts_short_id
    ON app.pinned_shorts (short_id);

CREATE TABLE app.main_playback_progress (
    user_id UUID NOT NULL,
    main_id UUID NOT NULL,
    last_position_ms BIGINT NOT NULL DEFAULT 0 CHECK (last_position_ms >= 0),
    last_played_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, main_id),
    FOREIGN KEY (user_id, main_id)
        REFERENCES app.main_unlocks (user_id, main_id),
    CHECK (completed_at IS NULL OR completed_at >= created_at)
);

CREATE INDEX idx_main_playback_progress_user_last_played_at
    ON app.main_playback_progress (user_id, last_played_at DESC);

CREATE FUNCTION app.assert_creator_capability_state(
    p_user_id UUID,
    p_allowed_states TEXT[],
    p_context TEXT
) RETURNS VOID
LANGUAGE plpgsql
AS $$
DECLARE
    current_state TEXT;
BEGIN
    SELECT state
    INTO current_state
    FROM app.creator_capabilities
    WHERE user_id = p_user_id;

    IF current_state IS NULL THEN
        RAISE EXCEPTION '% requires creator capability for user %', p_context, p_user_id;
    END IF;

    IF current_state <> ALL (p_allowed_states) THEN
        RAISE EXCEPTION '% requires creator capability state in %, got % for user %',
            p_context,
            p_allowed_states,
            current_state,
            p_user_id;
    END IF;
END;
$$;

CREATE FUNCTION app.enforce_public_creator_profile_requires_approved_capability()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    PERFORM app.assert_creator_capability_state(
        NEW.user_id,
        ARRAY['approved'],
        'creator_profiles'
    );

    RETURN NEW;
END;
$$;

CREATE FUNCTION app.enforce_creator_content_requires_approved_capability()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    PERFORM app.assert_creator_capability_state(
        NEW.creator_user_id,
        ARRAY['approved'],
        TG_TABLE_SCHEMA || '.' || TG_TABLE_NAME
    );

    RETURN NEW;
END;
$$;

CREATE FUNCTION app.enforce_follow_target_requires_approved_capability()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    PERFORM app.assert_creator_capability_state(
        NEW.creator_user_id,
        ARRAY['approved'],
        'creator_follows'
    );

    RETURN NEW;
END;
$$;

CREATE FUNCTION app.enforce_unlockable_main_purchase()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    main_owner_user_id UUID;
    main_state TEXT;
    main_post_report_state TEXT;
    creator_state TEXT;
BEGIN
    SELECT
        m.creator_user_id,
        m.state,
        m.post_report_state,
        c.state
    INTO
        main_owner_user_id,
        main_state,
        main_post_report_state,
        creator_state
    FROM app.mains AS m
    JOIN app.creator_capabilities AS c
        ON c.user_id = m.creator_user_id
    WHERE m.id = NEW.main_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'main_unlocks requires an existing main, got %', NEW.main_id;
    END IF;

    IF NEW.user_id = main_owner_user_id THEN
        RAISE EXCEPTION 'main_unlocks must not encode creator ownership for main %', NEW.main_id;
    END IF;

    IF creator_state <> 'approved' THEN
        RAISE EXCEPTION 'main_unlocks requires approved creator capability for main %', NEW.main_id;
    END IF;

    IF main_state <> 'approved_for_unlock' THEN
        RAISE EXCEPTION 'main_unlocks requires main state approved_for_unlock, got % for main %',
            main_state,
            NEW.main_id;
    END IF;

    IF main_post_report_state IN ('temporarily_limited', 'removed') THEN
        RAISE EXCEPTION 'main_unlocks requires a playable main, got post_report_state % for main %',
            main_post_report_state,
            NEW.main_id;
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_creator_profiles_require_approved_capability
    BEFORE INSERT OR UPDATE OF user_id
    ON app.creator_profiles
    FOR EACH ROW
    EXECUTE FUNCTION app.enforce_public_creator_profile_requires_approved_capability();

CREATE TRIGGER trg_media_assets_require_approved_capability
    BEFORE INSERT OR UPDATE OF creator_user_id
    ON app.media_assets
    FOR EACH ROW
    EXECUTE FUNCTION app.enforce_creator_content_requires_approved_capability();

CREATE TRIGGER trg_mains_require_approved_capability
    BEFORE INSERT OR UPDATE OF creator_user_id
    ON app.mains
    FOR EACH ROW
    EXECUTE FUNCTION app.enforce_creator_content_requires_approved_capability();

CREATE TRIGGER trg_shorts_require_approved_capability
    BEFORE INSERT OR UPDATE OF creator_user_id
    ON app.shorts
    FOR EACH ROW
    EXECUTE FUNCTION app.enforce_creator_content_requires_approved_capability();

CREATE TRIGGER trg_main_unlocks_require_unlockable_main
    BEFORE INSERT OR UPDATE OF user_id, main_id
    ON app.main_unlocks
    FOR EACH ROW
    EXECUTE FUNCTION app.enforce_unlockable_main_purchase();

CREATE TRIGGER trg_creator_follows_require_approved_capability
    BEFORE INSERT OR UPDATE OF creator_user_id
    ON app.creator_follows
    FOR EACH ROW
    EXECUTE FUNCTION app.enforce_follow_target_requires_approved_capability();

CREATE VIEW app.public_creator_profiles AS
SELECT
    p.user_id,
    p.display_name,
    p.avatar_url,
    p.bio,
    p.published_at,
    p.created_at,
    p.updated_at
FROM app.creator_profiles AS p
JOIN app.creator_capabilities AS c
    ON c.user_id = p.user_id
WHERE c.state = 'approved';

CREATE VIEW app.unlockable_mains AS
SELECT
    m.id,
    m.creator_user_id,
    m.media_asset_id,
    m.state,
    m.review_reason_code,
    m.post_report_state,
    m.price_minor,
    m.currency_code,
    m.ownership_confirmed,
    m.consent_confirmed,
    m.approved_for_unlock_at,
    m.created_at,
    m.updated_at
FROM app.mains AS m
JOIN app.creator_capabilities AS c
    ON c.user_id = m.creator_user_id
WHERE c.state = 'approved'
    AND m.state = 'approved_for_unlock'
    AND (m.post_report_state IS NULL OR m.post_report_state NOT IN ('temporarily_limited', 'removed'));

CREATE VIEW app.public_shorts AS
SELECT
    s.id,
    s.creator_user_id,
    s.canonical_main_id,
    s.media_asset_id,
    s.state,
    s.review_reason_code,
    s.post_report_state,
    s.approved_for_publish_at,
    s.published_at,
    s.created_at,
    s.updated_at
FROM app.shorts AS s
JOIN app.unlockable_mains AS m
    ON m.id = s.canonical_main_id
JOIN app.creator_capabilities AS c
    ON c.user_id = s.creator_user_id
WHERE c.state = 'approved'
    AND s.state = 'approved_for_publish'
    AND s.published_at IS NOT NULL
    AND (s.post_report_state IS NULL OR s.post_report_state NOT IN ('temporarily_limited', 'removed'));

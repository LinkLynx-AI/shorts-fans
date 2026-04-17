ALTER TABLE app.shorts
    ADD CONSTRAINT recommendation_shorts_identity_unique
    UNIQUE (id, creator_user_id, canonical_main_id);

CREATE TABLE app.recommendation_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    viewer_user_id UUID NOT NULL REFERENCES app.users (id) ON DELETE CASCADE,
    event_kind TEXT NOT NULL CHECK (
        event_kind IN (
            'impression',
            'view_start',
            'view_completion',
            'rewatch_loop',
            'profile_click',
            'main_click',
            'unlock_conversion'
        )
    ),
    creator_user_id UUID REFERENCES app.creator_capabilities (user_id) ON DELETE CASCADE,
    canonical_main_id UUID REFERENCES app.mains (id) ON DELETE CASCADE,
    short_id UUID REFERENCES app.shorts (id) ON DELETE CASCADE,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    idempotency_key TEXT NOT NULL CHECK (length(btrim(idempotency_key)) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (
        creator_user_id IS NOT NULL
        OR canonical_main_id IS NOT NULL
        OR short_id IS NOT NULL
    ),
    CHECK (
        short_id IS NULL
        OR canonical_main_id IS NOT NULL
    ),
    CHECK (
        canonical_main_id IS NULL
        OR creator_user_id IS NOT NULL
    ),
    CHECK (
        short_id IS NULL
        OR creator_user_id IS NOT NULL
    ),
    CONSTRAINT recommendation_events_payload_shape_check CHECK (
        (
            event_kind IN ('impression', 'view_start', 'view_completion', 'rewatch_loop')
            AND creator_user_id IS NOT NULL
            AND canonical_main_id IS NOT NULL
            AND short_id IS NOT NULL
        )
        OR (
            event_kind = 'profile_click'
            AND creator_user_id IS NOT NULL
            AND canonical_main_id IS NULL
            AND short_id IS NULL
        )
        OR (
            event_kind IN ('main_click', 'unlock_conversion')
            AND creator_user_id IS NOT NULL
            AND canonical_main_id IS NOT NULL
        )
    ),
    CONSTRAINT recommendation_events_main_identity_fkey
        FOREIGN KEY (canonical_main_id, creator_user_id)
        REFERENCES app.mains (id, creator_user_id),
    CONSTRAINT recommendation_events_short_identity_fkey
        FOREIGN KEY (short_id, creator_user_id, canonical_main_id)
        REFERENCES app.shorts (id, creator_user_id, canonical_main_id)
);

CREATE UNIQUE INDEX recommendation_events_viewer_idempotency_key_idx
    ON app.recommendation_events (viewer_user_id, idempotency_key);

CREATE INDEX recommendation_events_viewer_occurred_at_idx
    ON app.recommendation_events (viewer_user_id, occurred_at DESC, id DESC);

CREATE INDEX recommendation_events_creator_user_id_idx
    ON app.recommendation_events (creator_user_id);

CREATE INDEX recommendation_events_main_identity_idx
    ON app.recommendation_events (canonical_main_id, creator_user_id);

CREATE INDEX recommendation_events_short_identity_idx
    ON app.recommendation_events (short_id, creator_user_id, canonical_main_id);

CREATE TABLE app.recommendation_viewer_short_features (
    viewer_user_id UUID NOT NULL REFERENCES app.users (id) ON DELETE CASCADE,
    short_id UUID NOT NULL REFERENCES app.shorts (id) ON DELETE CASCADE,
    creator_user_id UUID NOT NULL REFERENCES app.creator_capabilities (user_id) ON DELETE CASCADE,
    canonical_main_id UUID NOT NULL REFERENCES app.mains (id) ON DELETE CASCADE,
    impression_count BIGINT NOT NULL DEFAULT 0 CHECK (impression_count >= 0),
    last_impression_at TIMESTAMPTZ,
    view_start_count BIGINT NOT NULL DEFAULT 0 CHECK (view_start_count >= 0),
    last_view_start_at TIMESTAMPTZ,
    view_completion_count BIGINT NOT NULL DEFAULT 0 CHECK (view_completion_count >= 0),
    last_view_completion_at TIMESTAMPTZ,
    rewatch_loop_count BIGINT NOT NULL DEFAULT 0 CHECK (rewatch_loop_count >= 0),
    last_rewatch_loop_at TIMESTAMPTZ,
    main_click_count BIGINT NOT NULL DEFAULT 0 CHECK (main_click_count >= 0),
    last_main_click_at TIMESTAMPTZ,
    unlock_conversion_count BIGINT NOT NULL DEFAULT 0 CHECK (unlock_conversion_count >= 0),
    last_unlock_conversion_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (viewer_user_id, short_id),
    CONSTRAINT recommendation_viewer_short_features_short_identity_fkey
        FOREIGN KEY (short_id, creator_user_id, canonical_main_id)
        REFERENCES app.shorts (id, creator_user_id, canonical_main_id)
);

CREATE INDEX recommendation_viewer_short_features_creator_user_id_idx
    ON app.recommendation_viewer_short_features (creator_user_id);

CREATE INDEX recommendation_viewer_short_features_main_identity_idx
    ON app.recommendation_viewer_short_features (canonical_main_id, creator_user_id);

CREATE INDEX recommendation_viewer_short_features_short_identity_idx
    ON app.recommendation_viewer_short_features (short_id, creator_user_id, canonical_main_id);

CREATE TABLE app.recommendation_viewer_creator_features (
    viewer_user_id UUID NOT NULL REFERENCES app.users (id) ON DELETE CASCADE,
    creator_user_id UUID NOT NULL REFERENCES app.creator_capabilities (user_id) ON DELETE CASCADE,
    impression_count BIGINT NOT NULL DEFAULT 0 CHECK (impression_count >= 0),
    last_impression_at TIMESTAMPTZ,
    view_start_count BIGINT NOT NULL DEFAULT 0 CHECK (view_start_count >= 0),
    last_view_start_at TIMESTAMPTZ,
    view_completion_count BIGINT NOT NULL DEFAULT 0 CHECK (view_completion_count >= 0),
    last_view_completion_at TIMESTAMPTZ,
    rewatch_loop_count BIGINT NOT NULL DEFAULT 0 CHECK (rewatch_loop_count >= 0),
    last_rewatch_loop_at TIMESTAMPTZ,
    profile_click_count BIGINT NOT NULL DEFAULT 0 CHECK (profile_click_count >= 0),
    last_profile_click_at TIMESTAMPTZ,
    main_click_count BIGINT NOT NULL DEFAULT 0 CHECK (main_click_count >= 0),
    last_main_click_at TIMESTAMPTZ,
    unlock_conversion_count BIGINT NOT NULL DEFAULT 0 CHECK (unlock_conversion_count >= 0),
    last_unlock_conversion_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (viewer_user_id, creator_user_id)
);

CREATE INDEX recommendation_viewer_creator_features_creator_user_id_idx
    ON app.recommendation_viewer_creator_features (creator_user_id);

CREATE TABLE app.recommendation_viewer_main_features (
    viewer_user_id UUID NOT NULL REFERENCES app.users (id) ON DELETE CASCADE,
    canonical_main_id UUID NOT NULL REFERENCES app.mains (id) ON DELETE CASCADE,
    creator_user_id UUID NOT NULL REFERENCES app.creator_capabilities (user_id) ON DELETE CASCADE,
    impression_count BIGINT NOT NULL DEFAULT 0 CHECK (impression_count >= 0),
    last_impression_at TIMESTAMPTZ,
    view_start_count BIGINT NOT NULL DEFAULT 0 CHECK (view_start_count >= 0),
    last_view_start_at TIMESTAMPTZ,
    view_completion_count BIGINT NOT NULL DEFAULT 0 CHECK (view_completion_count >= 0),
    last_view_completion_at TIMESTAMPTZ,
    rewatch_loop_count BIGINT NOT NULL DEFAULT 0 CHECK (rewatch_loop_count >= 0),
    last_rewatch_loop_at TIMESTAMPTZ,
    main_click_count BIGINT NOT NULL DEFAULT 0 CHECK (main_click_count >= 0),
    last_main_click_at TIMESTAMPTZ,
    unlock_conversion_count BIGINT NOT NULL DEFAULT 0 CHECK (unlock_conversion_count >= 0),
    last_unlock_conversion_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (viewer_user_id, canonical_main_id),
    CONSTRAINT recommendation_viewer_main_features_main_identity_fkey
        FOREIGN KEY (canonical_main_id, creator_user_id)
        REFERENCES app.mains (id, creator_user_id)
);

CREATE INDEX recommendation_viewer_main_features_creator_user_id_idx
    ON app.recommendation_viewer_main_features (creator_user_id);

CREATE INDEX recommendation_viewer_main_features_main_identity_idx
    ON app.recommendation_viewer_main_features (canonical_main_id, creator_user_id);

CREATE TABLE app.recommendation_short_global_features (
    short_id UUID PRIMARY KEY REFERENCES app.shorts (id) ON DELETE CASCADE,
    creator_user_id UUID NOT NULL REFERENCES app.creator_capabilities (user_id) ON DELETE CASCADE,
    canonical_main_id UUID NOT NULL REFERENCES app.mains (id) ON DELETE CASCADE,
    impression_count BIGINT NOT NULL DEFAULT 0 CHECK (impression_count >= 0),
    last_impression_at TIMESTAMPTZ,
    view_start_count BIGINT NOT NULL DEFAULT 0 CHECK (view_start_count >= 0),
    last_view_start_at TIMESTAMPTZ,
    view_completion_count BIGINT NOT NULL DEFAULT 0 CHECK (view_completion_count >= 0),
    last_view_completion_at TIMESTAMPTZ,
    rewatch_loop_count BIGINT NOT NULL DEFAULT 0 CHECK (rewatch_loop_count >= 0),
    last_rewatch_loop_at TIMESTAMPTZ,
    main_click_count BIGINT NOT NULL DEFAULT 0 CHECK (main_click_count >= 0),
    last_main_click_at TIMESTAMPTZ,
    unlock_conversion_count BIGINT NOT NULL DEFAULT 0 CHECK (unlock_conversion_count >= 0),
    last_unlock_conversion_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT recommendation_short_global_features_short_identity_fkey
        FOREIGN KEY (short_id, creator_user_id, canonical_main_id)
        REFERENCES app.shorts (id, creator_user_id, canonical_main_id)
);

CREATE INDEX recommendation_short_global_features_creator_user_id_idx
    ON app.recommendation_short_global_features (creator_user_id);

CREATE INDEX recommendation_short_global_features_main_identity_idx
    ON app.recommendation_short_global_features (canonical_main_id, creator_user_id);

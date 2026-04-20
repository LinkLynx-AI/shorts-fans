ALTER TABLE app.mains
    ADD COLUMN review_decision_source TEXT,
    ADD COLUMN review_decisioned_at TIMESTAMPTZ;

ALTER TABLE app.mains
    ADD CONSTRAINT mains_review_decision_source_check CHECK (
        review_decision_source IS NULL
        OR review_decision_source IN ('manual', 'auto', 'manual_override')
    ),
    ADD CONSTRAINT mains_review_decisioned_at_pair_check CHECK (
        (review_decision_source IS NULL AND review_decisioned_at IS NULL)
        OR (review_decision_source IS NOT NULL AND review_decisioned_at IS NOT NULL)
    );

ALTER TABLE app.shorts
    ADD COLUMN review_decision_source TEXT,
    ADD COLUMN review_decisioned_at TIMESTAMPTZ;

ALTER TABLE app.shorts
    ADD CONSTRAINT shorts_review_decision_source_check CHECK (
        review_decision_source IS NULL
        OR review_decision_source IN ('manual', 'auto', 'manual_override')
    ),
    ADD CONSTRAINT shorts_review_decisioned_at_pair_check CHECK (
        (review_decision_source IS NULL AND review_decisioned_at IS NULL)
        OR (review_decision_source IS NOT NULL AND review_decisioned_at IS NOT NULL)
    );

DROP VIEW IF EXISTS app.public_shorts;
DROP VIEW IF EXISTS app.unlockable_mains;

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
JOIN app.media_assets AS main_asset
    ON main_asset.id = m.media_asset_id
WHERE c.state = 'approved'
    AND main_asset.processing_state = 'ready'
    AND m.state = 'approved_for_unlock'
    AND m.ownership_confirmed = TRUE
    AND m.consent_confirmed = TRUE
    AND (m.post_report_state IS NULL OR m.post_report_state NOT IN ('temporarily_limited', 'removed'));

CREATE VIEW app.public_shorts AS
SELECT
    s.id,
    s.creator_user_id,
    s.canonical_main_id,
    s.media_asset_id,
    s.caption,
    s.state,
    s.review_reason_code,
    s.post_report_state,
    s.approved_for_publish_at,
    COALESCE(
        s.published_at,
        GREATEST(s.approved_for_publish_at, m.approved_for_unlock_at)
    ) AS published_at,
    s.created_at,
    s.updated_at
FROM app.shorts AS s
JOIN app.unlockable_mains AS m
    ON m.id = s.canonical_main_id
JOIN app.creator_capabilities AS c
    ON c.user_id = s.creator_user_id
JOIN app.media_assets AS short_asset
    ON short_asset.id = s.media_asset_id
WHERE c.state = 'approved'
    AND short_asset.processing_state = 'ready'
    AND s.state = 'approved_for_publish'
    AND COALESCE(
        s.published_at,
        GREATEST(s.approved_for_publish_at, m.approved_for_unlock_at)
    ) IS NOT NULL
    AND (s.post_report_state IS NULL OR s.post_report_state NOT IN ('temporarily_limited', 'removed'));

CREATE UNIQUE INDEX uq_submission_review_intakes_id_canonical_main_id
    ON app.submission_review_intakes (id, canonical_main_id);

CREATE TABLE app.submission_review_main_decisions (
    submission_review_intake_id UUID NOT NULL
        REFERENCES app.submission_review_intakes (id)
        ON DELETE CASCADE,
    main_id UUID NOT NULL,
    target_state TEXT NOT NULL CHECK (
        target_state IN (
            'approved_for_unlock',
            'revision_requested',
            'rejected'
        )
    ),
    reason_code TEXT,
    decision_source TEXT NOT NULL CHECK (
        decision_source IN ('manual', 'auto', 'manual_override')
    ),
    decisioned_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (submission_review_intake_id, main_id),
    FOREIGN KEY (submission_review_intake_id, main_id)
        REFERENCES app.submission_review_intakes (id, canonical_main_id)
        ON DELETE CASCADE,
    CHECK (
        (target_state = 'approved_for_unlock' AND reason_code IS NULL)
        OR (target_state IN ('revision_requested', 'rejected') AND reason_code IS NOT NULL)
    )
);

CREATE INDEX idx_submission_review_main_decisions_main_id
    ON app.submission_review_main_decisions (main_id, decisioned_at DESC);

CREATE TABLE app.submission_review_short_decisions (
    submission_review_intake_id UUID NOT NULL
        REFERENCES app.submission_review_intakes (id)
        ON DELETE CASCADE,
    short_id UUID NOT NULL,
    target_state TEXT NOT NULL CHECK (
        target_state IN (
            'approved_for_publish',
            'revision_requested',
            'rejected'
        )
    ),
    reason_code TEXT,
    decision_source TEXT NOT NULL CHECK (
        decision_source IN ('manual', 'auto', 'manual_override')
    ),
    decisioned_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (submission_review_intake_id, short_id),
    FOREIGN KEY (submission_review_intake_id, short_id)
        REFERENCES app.submission_review_intake_shorts (submission_review_intake_id, short_id)
        ON DELETE CASCADE,
    CHECK (
        (target_state = 'approved_for_publish' AND reason_code IS NULL)
        OR (target_state IN ('revision_requested', 'rejected') AND reason_code IS NOT NULL)
    )
);

CREATE INDEX idx_submission_review_short_decisions_short_id
    ON app.submission_review_short_decisions (short_id, decisioned_at DESC);

CREATE TABLE app.submission_review_intakes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    canonical_main_id UUID NOT NULL,
    creator_user_id UUID NOT NULL REFERENCES app.creator_capabilities (user_id),
    status TEXT NOT NULL CHECK (
        status IN (
            'pending_review',
            'decision_applied'
        )
    ),
    submit_kind TEXT NOT NULL CHECK (
        submit_kind IN (
            'initial_submit',
            'resubmit'
        )
    ),
    previous_intake_id UUID REFERENCES app.submission_review_intakes (id),
    main_media_asset_id UUID NOT NULL REFERENCES app.media_assets (id),
    main_price_minor BIGINT NOT NULL CHECK (main_price_minor > 0),
    ownership_confirmed BOOLEAN NOT NULL,
    consent_confirmed BOOLEAN NOT NULL,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (canonical_main_id, creator_user_id)
        REFERENCES app.mains (id, creator_user_id),
    CHECK (
        (submit_kind = 'initial_submit' AND previous_intake_id IS NULL)
        OR (submit_kind = 'resubmit' AND previous_intake_id IS NOT NULL)
    )
);

CREATE UNIQUE INDEX uq_submission_review_intakes_pending_main
    ON app.submission_review_intakes (canonical_main_id)
    WHERE status = 'pending_review';

CREATE UNIQUE INDEX uq_submission_review_intakes_previous_intake_id
    ON app.submission_review_intakes (previous_intake_id)
    WHERE previous_intake_id IS NOT NULL;

CREATE INDEX idx_submission_review_intakes_canonical_main_id
    ON app.submission_review_intakes (canonical_main_id, submitted_at DESC, id DESC);

CREATE INDEX idx_submission_review_intakes_creator_user_id
    ON app.submission_review_intakes (creator_user_id, submitted_at DESC, id DESC);

CREATE TABLE app.submission_review_intake_shorts (
    submission_review_intake_id UUID NOT NULL
        REFERENCES app.submission_review_intakes (id)
        ON DELETE CASCADE,
    short_id UUID NOT NULL
        REFERENCES app.shorts (id),
    media_asset_id UUID NOT NULL REFERENCES app.media_assets (id),
    caption TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (submission_review_intake_id, short_id)
);

CREATE INDEX idx_submission_review_intake_shorts_short_id
    ON app.submission_review_intake_shorts (short_id);

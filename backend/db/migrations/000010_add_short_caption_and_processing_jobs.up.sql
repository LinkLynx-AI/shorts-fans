ALTER TABLE app.shorts
    ADD COLUMN caption TEXT;

CREATE TABLE app.media_processing_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_user_id UUID NOT NULL REFERENCES app.creator_capabilities (user_id),
    media_asset_id UUID NOT NULL UNIQUE,
    asset_role TEXT NOT NULL CHECK (asset_role IN ('main', 'short')),
    status TEXT NOT NULL CHECK (status IN ('queued', 'processing', 'succeeded', 'failed')),
    attempt_count INTEGER NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    last_error_code TEXT,
    last_error_message TEXT,
    queued_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    failed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (media_asset_id, creator_user_id)
        REFERENCES app.media_assets (id, creator_user_id)
);

CREATE INDEX idx_media_processing_jobs_status
    ON app.media_processing_jobs (status, queued_at, id);

CREATE INDEX idx_media_processing_jobs_creator_user_id
    ON app.media_processing_jobs (creator_user_id);

CREATE OR REPLACE VIEW app.public_shorts AS
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

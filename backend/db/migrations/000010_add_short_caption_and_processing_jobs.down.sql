CREATE OR REPLACE VIEW app.public_shorts AS
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

DROP INDEX IF EXISTS idx_media_processing_jobs_creator_user_id;
DROP INDEX IF EXISTS idx_media_processing_jobs_status;
DROP TABLE IF EXISTS app.media_processing_jobs;

ALTER TABLE app.shorts
    DROP COLUMN IF EXISTS caption;

DROP INDEX IF EXISTS idx_submission_review_short_decisions_short_id;
DROP TABLE IF EXISTS app.submission_review_short_decisions;

DROP INDEX IF EXISTS idx_submission_review_main_decisions_main_id;
DROP TABLE IF EXISTS app.submission_review_main_decisions;

DROP INDEX IF EXISTS uq_submission_review_intakes_id_canonical_main_id;

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
WHERE c.state = 'approved'
    AND m.state = 'approved_for_unlock'
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

ALTER TABLE app.shorts
    DROP CONSTRAINT IF EXISTS shorts_review_decisioned_at_pair_check,
    DROP CONSTRAINT IF EXISTS shorts_review_decision_source_check,
    DROP COLUMN IF EXISTS review_decisioned_at,
    DROP COLUMN IF EXISTS review_decision_source;

ALTER TABLE app.mains
    DROP CONSTRAINT IF EXISTS mains_review_decisioned_at_pair_check,
    DROP CONSTRAINT IF EXISTS mains_review_decision_source_check,
    DROP COLUMN IF EXISTS review_decisioned_at,
    DROP COLUMN IF EXISTS review_decision_source;

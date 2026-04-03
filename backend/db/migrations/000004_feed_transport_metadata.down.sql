DROP VIEW app.public_shorts;
DROP VIEW app.public_creator_profiles;

ALTER TABLE app.shorts
    DROP COLUMN caption,
    DROP COLUMN title;

ALTER TABLE app.creator_profiles
    DROP CONSTRAINT creator_profiles_public_fields_check,
    DROP CONSTRAINT creator_profiles_handle_trimmed_check,
    ADD CONSTRAINT creator_profiles_public_fields_check
        CHECK (
            published_at IS NULL
            OR display_name IS NOT NULL
        ),
    DROP COLUMN handle;

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
WHERE c.state = 'approved'
    AND p.published_at IS NOT NULL;

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

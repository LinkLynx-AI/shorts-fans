ALTER TABLE app.creator_profiles
    ADD COLUMN handle TEXT;

ALTER TABLE app.creator_profiles
    DROP CONSTRAINT creator_profiles_public_fields_check,
    ADD CONSTRAINT creator_profiles_handle_trimmed_check
        CHECK (
            handle IS NULL
            OR length(btrim(handle)) > 0
        ),
    ADD CONSTRAINT creator_profiles_public_fields_check
        CHECK (
            published_at IS NULL
            OR (
                display_name IS NOT NULL
                AND handle IS NOT NULL
            )
        );

ALTER TABLE app.shorts
    ADD COLUMN title TEXT NOT NULL DEFAULT '',
    ADD COLUMN caption TEXT NOT NULL DEFAULT '';

DROP VIEW app.public_shorts;
DROP VIEW app.public_creator_profiles;

CREATE VIEW app.public_creator_profiles AS
SELECT
    p.user_id,
    p.display_name,
    p.handle,
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
    s.title,
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

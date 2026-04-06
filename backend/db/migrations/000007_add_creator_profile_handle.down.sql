DROP VIEW IF EXISTS app.public_creator_profiles;

DROP INDEX IF EXISTS app.creator_profiles_public_display_name_trgm_idx;
DROP INDEX IF EXISTS app.creator_profiles_public_published_handle_idx;
DROP INDEX IF EXISTS app.creator_profiles_handle_unique_idx;

ALTER TABLE app.creator_profiles
    DROP CONSTRAINT IF EXISTS creator_profiles_handle_format_check,
    DROP CONSTRAINT IF EXISTS creator_profiles_handle_lowercase_check,
    DROP CONSTRAINT IF EXISTS creator_profiles_handle_trimmed_check,
    DROP CONSTRAINT IF EXISTS creator_profiles_public_fields_check;

ALTER TABLE app.creator_profiles
    ADD CONSTRAINT creator_profiles_public_fields_check
        CHECK (
            published_at IS NULL
            OR display_name IS NOT NULL
        );

ALTER TABLE app.creator_profiles
    DROP COLUMN handle;

CREATE OR REPLACE VIEW app.public_creator_profiles AS
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

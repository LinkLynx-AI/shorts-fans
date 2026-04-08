DROP VIEW IF EXISTS app.public_creator_profiles;

DROP INDEX IF EXISTS app.creator_profiles_public_published_handle_idx;
DROP INDEX IF EXISTS app.creator_profiles_handle_unique_idx;

ALTER TABLE app.creator_profiles
    DROP CONSTRAINT IF EXISTS creator_profiles_handle_format_check,
    DROP CONSTRAINT IF EXISTS creator_profiles_handle_lowercase_check,
    DROP CONSTRAINT IF EXISTS creator_profiles_handle_trimmed_check;

ALTER TABLE app.creator_profiles
    ALTER COLUMN handle DROP NOT NULL;

ALTER TABLE app.creator_profiles
    ADD CONSTRAINT creator_profiles_handle_trimmed_check
        CHECK (
            handle IS NULL
            OR handle = btrim(handle)
        ),
    ADD CONSTRAINT creator_profiles_handle_lowercase_check
        CHECK (
            handle IS NULL
            OR handle = lower(handle)
        ),
    ADD CONSTRAINT creator_profiles_handle_format_check
        CHECK (
            handle IS NULL
            OR handle ~ '^[a-z0-9._]+$'
        );

CREATE UNIQUE INDEX creator_profiles_handle_unique_idx
    ON app.creator_profiles (handle)
    WHERE handle IS NOT NULL;

CREATE INDEX creator_profiles_public_published_handle_idx
    ON app.creator_profiles (published_at DESC, handle ASC)
    WHERE published_at IS NOT NULL;

CREATE OR REPLACE VIEW app.public_creator_profiles AS
SELECT
    p.user_id,
    p.display_name,
    p.avatar_url,
    p.bio,
    p.published_at,
    p.created_at,
    p.updated_at,
    p.handle
FROM app.creator_profiles AS p
JOIN app.creator_capabilities AS c
    ON c.user_id = p.user_id
WHERE c.state = 'approved'
    AND p.published_at IS NOT NULL;

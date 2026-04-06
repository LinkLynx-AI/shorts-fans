CREATE EXTENSION IF NOT EXISTS pg_trgm;

ALTER TABLE app.creator_profiles
    ADD COLUMN handle TEXT;

UPDATE app.creator_profiles AS p
SET handle = generated.generated_handle
FROM (
    WITH prepared AS (
        SELECT
            user_id,
            created_at,
            COALESCE(
                NULLIF(regexp_replace(lower(display_name), '[^a-z0-9]+', '', 'g'), ''),
                'creator'
            ) AS base_handle
        FROM app.creator_profiles
        WHERE published_at IS NOT NULL
    ),
    ranked AS (
        SELECT
            user_id,
            CASE
                WHEN row_number() OVER (PARTITION BY base_handle ORDER BY created_at, user_id) = 1 THEN base_handle
                ELSE base_handle || '_' || substr(replace(user_id::text, '-', ''), 1, 8)
            END AS generated_handle
        FROM prepared
    )
    SELECT
        user_id,
        generated_handle
    FROM ranked
) AS generated
WHERE p.user_id = generated.user_id;

ALTER TABLE app.creator_profiles
    DROP CONSTRAINT IF EXISTS creator_profiles_public_fields_check;

ALTER TABLE app.creator_profiles
    ADD CONSTRAINT creator_profiles_public_fields_check
        CHECK (
            published_at IS NULL
            OR (
                display_name IS NOT NULL
                AND handle IS NOT NULL
            )
        ),
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

CREATE INDEX creator_profiles_public_display_name_trgm_idx
    ON app.creator_profiles
    USING gin (display_name gin_trgm_ops)
    WHERE published_at IS NOT NULL
        AND display_name IS NOT NULL;

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

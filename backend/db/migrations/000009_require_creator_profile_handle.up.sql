DO $$
DECLARE
    profile_row RECORD;
    base_handle TEXT;
    candidate_handle TEXT;
    collision_suffix INTEGER;
BEGIN
    FOR profile_row IN
        SELECT
            user_id,
            display_name
        FROM app.creator_profiles
        WHERE handle IS NULL
        ORDER BY created_at, user_id
    LOOP
        base_handle := COALESCE(
            NULLIF(
                regexp_replace(lower(COALESCE(profile_row.display_name, '')), '[^a-z0-9]+', '', 'g'),
                ''
            ),
            'creator'
        );
        candidate_handle := base_handle;
        collision_suffix := 0;

        WHILE EXISTS (
            SELECT 1
            FROM app.creator_profiles
            WHERE handle = candidate_handle
                AND user_id <> profile_row.user_id
        ) LOOP
            collision_suffix := collision_suffix + 1;
            candidate_handle := base_handle
                || '_'
                || replace(profile_row.user_id::text, '-', '');

            IF collision_suffix > 1 THEN
                candidate_handle := candidate_handle || '_' || collision_suffix::text;
            END IF;
        END LOOP;

        UPDATE app.creator_profiles
        SET handle = candidate_handle
        WHERE user_id = profile_row.user_id;
    END LOOP;
END;
$$;

DROP VIEW IF EXISTS app.public_creator_profiles;

DROP INDEX IF EXISTS app.creator_profiles_public_published_handle_idx;
DROP INDEX IF EXISTS app.creator_profiles_handle_unique_idx;

ALTER TABLE app.creator_profiles
    DROP CONSTRAINT IF EXISTS creator_profiles_handle_format_check,
    DROP CONSTRAINT IF EXISTS creator_profiles_handle_lowercase_check,
    DROP CONSTRAINT IF EXISTS creator_profiles_handle_trimmed_check;

ALTER TABLE app.creator_profiles
    ALTER COLUMN handle SET NOT NULL;

ALTER TABLE app.creator_profiles
    ADD CONSTRAINT creator_profiles_handle_trimmed_check
        CHECK (handle = btrim(handle)),
    ADD CONSTRAINT creator_profiles_handle_lowercase_check
        CHECK (handle = lower(handle)),
    ADD CONSTRAINT creator_profiles_handle_format_check
        CHECK (handle ~ '^[a-z0-9._]+$');

CREATE UNIQUE INDEX creator_profiles_handle_unique_idx
    ON app.creator_profiles (handle);

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

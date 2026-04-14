CREATE TABLE app.user_profiles (
    user_id UUID PRIMARY KEY REFERENCES app.users (id),
    display_name TEXT NOT NULL,
    handle TEXT NOT NULL,
    avatar_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT user_profiles_display_name_trimmed_check
        CHECK (length(btrim(display_name)) > 0),
    CONSTRAINT user_profiles_handle_trimmed_check
        CHECK (handle = btrim(handle)),
    CONSTRAINT user_profiles_handle_lowercase_check
        CHECK (handle = lower(handle)),
    CONSTRAINT user_profiles_handle_format_check
        CHECK (handle ~ '^[a-z0-9._]+$')
);

CREATE UNIQUE INDEX user_profiles_handle_unique_idx
    ON app.user_profiles (handle);

INSERT INTO app.user_profiles (
    user_id,
    display_name,
    handle,
    avatar_url,
    created_at,
    updated_at
)
SELECT
    p.user_id,
    COALESCE(
        NULLIF(btrim(p.display_name), ''),
        'User ' || substr(replace(p.user_id::text, '-', ''), 1, 8)
    ) AS display_name,
    p.handle,
    p.avatar_url,
    p.created_at,
    p.updated_at
FROM app.creator_profiles AS p
ON CONFLICT (user_id) DO NOTHING;

DO $$
DECLARE
    profile_row RECORD;
    base_handle TEXT;
    candidate_handle TEXT;
    collision_suffix BIGINT;
BEGIN
    FOR profile_row IN
        SELECT u.id
        FROM app.users AS u
        LEFT JOIN app.user_profiles AS p
            ON p.user_id = u.id
        WHERE p.user_id IS NULL
        ORDER BY u.created_at, u.id
    LOOP
        base_handle := 'user_' || replace(profile_row.id::text, '-', '');
        candidate_handle := base_handle;
        collision_suffix := 0;

        WHILE EXISTS (
            SELECT 1
            FROM app.user_profiles
            WHERE handle = candidate_handle
        ) LOOP
            collision_suffix := collision_suffix + 1;
            candidate_handle := base_handle || '_' || collision_suffix::text;
        END LOOP;

        INSERT INTO app.user_profiles (
            user_id,
            display_name,
            handle,
            avatar_url
        ) VALUES (
            profile_row.id,
            'User ' || substr(replace(profile_row.id::text, '-', ''), 1, 8),
            candidate_handle,
            NULL
        );
    END LOOP;
END $$;

DROP VIEW IF EXISTS app.public_creator_profiles;

DROP TRIGGER IF EXISTS trg_creator_profiles_require_approved_capability
    ON app.creator_profiles;

CREATE TABLE app.creator_profile_drafts (
    user_id UUID PRIMARY KEY REFERENCES app.creator_capabilities (user_id),
    display_name TEXT,
    avatar_url TEXT,
    bio TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (
        display_name IS NULL
        OR length(btrim(display_name)) > 0
    )
);

INSERT INTO app.creator_profile_drafts (
    user_id,
    display_name,
    avatar_url,
    bio,
    created_at,
    updated_at
)
SELECT
    p.user_id,
    p.display_name,
    p.avatar_url,
    p.bio,
    p.created_at,
    p.updated_at
FROM app.creator_profiles AS p
WHERE p.published_at IS NULL;

DELETE FROM app.creator_profiles
WHERE published_at IS NULL;

DO $$
DECLARE
    check_constraint_name TEXT;
BEGIN
    FOR check_constraint_name IN
        SELECT con.conname
        FROM pg_constraint AS con
        JOIN pg_class AS rel
            ON rel.oid = con.conrelid
        JOIN pg_namespace AS n
            ON n.oid = rel.relnamespace
        WHERE n.nspname = 'app'
            AND rel.relname = 'creator_profiles'
            AND con.contype = 'c'
    LOOP
        EXECUTE format(
            'ALTER TABLE app.creator_profiles DROP CONSTRAINT %I',
            check_constraint_name
        );
    END LOOP;
END;
$$;

ALTER TABLE app.creator_profiles
    ALTER COLUMN display_name SET NOT NULL,
    ALTER COLUMN published_at SET DEFAULT CURRENT_TIMESTAMP,
    ALTER COLUMN published_at SET NOT NULL,
    ADD CONSTRAINT creator_profiles_display_name_required_check
        CHECK (
            display_name IS NOT NULL
            AND length(btrim(display_name)) > 0
        );

CREATE OR REPLACE FUNCTION app.enforce_public_creator_profile_requires_approved_capability()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    PERFORM app.assert_creator_capability_state(
        NEW.user_id,
        ARRAY['approved'],
        'creator_profiles'
    );

    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION app.enforce_follow_target_requires_approved_capability()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    PERFORM app.assert_creator_capability_state(
        NEW.creator_user_id,
        ARRAY['approved'],
        'creator_follows'
    );

    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_creator_profiles_require_approved_capability
    BEFORE INSERT OR UPDATE OF user_id
    ON app.creator_profiles
    FOR EACH ROW
    EXECUTE FUNCTION app.enforce_public_creator_profile_requires_approved_capability();

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
WHERE c.state = 'approved';

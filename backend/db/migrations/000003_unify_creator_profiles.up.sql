DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM app.creator_profile_drafts AS d
        JOIN app.creator_profiles AS p
            ON p.user_id = d.user_id
    ) THEN
        RAISE EXCEPTION
            'creator_profile_drafts and creator_profiles overlap; aborting unified profile migration';
    END IF;
END;
$$;

DROP TRIGGER IF EXISTS trg_creator_profiles_require_approved_capability
    ON app.creator_profiles;

ALTER TABLE app.creator_profiles
    ALTER COLUMN display_name DROP NOT NULL,
    ALTER COLUMN published_at DROP NOT NULL,
    ALTER COLUMN published_at DROP DEFAULT;

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
    ADD CONSTRAINT creator_profiles_display_name_trimmed_check
        CHECK (
            display_name IS NULL
            OR length(btrim(display_name)) > 0
        ),
    ADD CONSTRAINT creator_profiles_public_fields_check
        CHECK (
            published_at IS NULL
            OR display_name IS NOT NULL
        );

INSERT INTO app.creator_profiles (
    user_id,
    display_name,
    avatar_url,
    bio,
    published_at,
    created_at,
    updated_at
)
SELECT
    d.user_id,
    d.display_name,
    d.avatar_url,
    d.bio,
    NULL,
    d.created_at,
    d.updated_at
FROM app.creator_profile_drafts AS d;

DROP TABLE app.creator_profile_drafts;

CREATE OR REPLACE FUNCTION app.enforce_public_creator_profile_requires_approved_capability()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'UPDATE'
        AND OLD.published_at IS NOT NULL
        AND NEW.published_at IS NULL THEN
        RAISE EXCEPTION
            'creator_profiles published_at cannot be cleared once set for user %',
            NEW.user_id;
    END IF;

    IF NEW.published_at IS NULL THEN
        RETURN NEW;
    END IF;

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
DECLARE
    creator_published_at TIMESTAMPTZ;
BEGIN
    PERFORM app.assert_creator_capability_state(
        NEW.creator_user_id,
        ARRAY['approved'],
        'creator_follows'
    );

    SELECT published_at
    INTO creator_published_at
    FROM app.creator_profiles
    WHERE user_id = NEW.creator_user_id;

    IF creator_published_at IS NULL THEN
        RAISE EXCEPTION
            'creator_follows requires a public creator profile for user %',
            NEW.creator_user_id;
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_creator_profiles_require_approved_capability
    BEFORE INSERT OR UPDATE OF user_id, published_at
    ON app.creator_profiles
    FOR EACH ROW
    EXECUTE FUNCTION app.enforce_public_creator_profile_requires_approved_capability();

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

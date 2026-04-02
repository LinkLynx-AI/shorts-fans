DROP VIEW IF EXISTS app.public_shorts;
DROP VIEW IF EXISTS app.unlockable_mains;
DROP VIEW IF EXISTS app.public_creator_profiles;

DROP TRIGGER IF EXISTS trg_main_unlocks_require_unlockable_main ON app.main_unlocks;
DROP TRIGGER IF EXISTS trg_creator_follows_require_approved_capability ON app.creator_follows;
DROP TRIGGER IF EXISTS trg_shorts_require_approved_capability ON app.shorts;
DROP TRIGGER IF EXISTS trg_mains_require_approved_capability ON app.mains;
DROP TRIGGER IF EXISTS trg_media_assets_require_approved_capability ON app.media_assets;
DROP TRIGGER IF EXISTS trg_creator_profiles_require_approved_capability ON app.creator_profiles;

DROP FUNCTION IF EXISTS app.enforce_unlockable_main_purchase();
DROP FUNCTION IF EXISTS app.enforce_follow_target_requires_approved_capability();
DROP FUNCTION IF EXISTS app.enforce_creator_content_requires_approved_capability();
DROP FUNCTION IF EXISTS app.enforce_public_creator_profile_requires_approved_capability();
DROP FUNCTION IF EXISTS app.assert_creator_capability_state(UUID, TEXT[], TEXT);

DROP TABLE IF EXISTS app.main_playback_progress;
DROP TABLE IF EXISTS app.pinned_shorts;
DROP TABLE IF EXISTS app.creator_follows;
DROP TABLE IF EXISTS app.main_unlocks;
DROP TABLE IF EXISTS app.shorts;
DROP TABLE IF EXISTS app.mains;
DROP TABLE IF EXISTS app.media_assets;
DROP TABLE IF EXISTS app.creator_profiles;
DROP TABLE IF EXISTS app.creator_profile_drafts;
DROP TABLE IF EXISTS app.creator_capabilities;
DROP TABLE IF EXISTS app.users;

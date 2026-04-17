DROP TABLE IF EXISTS app.recommendation_short_global_features;
DROP TABLE IF EXISTS app.recommendation_viewer_main_features;
DROP TABLE IF EXISTS app.recommendation_viewer_creator_features;
DROP TABLE IF EXISTS app.recommendation_viewer_short_features;
DROP INDEX IF EXISTS app.recommendation_events_viewer_occurred_at_idx;
DROP INDEX IF EXISTS app.recommendation_events_viewer_idempotency_key_idx;
DROP TABLE IF EXISTS app.recommendation_events;
ALTER TABLE app.shorts
    DROP CONSTRAINT IF EXISTS recommendation_shorts_identity_unique;

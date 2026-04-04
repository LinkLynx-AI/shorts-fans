ALTER TABLE app.media_assets
    DROP CONSTRAINT IF EXISTS media_assets_check,
    DROP CONSTRAINT IF EXISTS media_assets_processing_state_check;

ALTER TABLE app.media_assets
    ADD CONSTRAINT media_assets_processing_state_check CHECK (
        processing_state IN ('uploaded', 'processing', 'ready', 'failed')
    ),
    ADD CONSTRAINT media_assets_check CHECK (
        (
            processing_state = 'ready'
            AND playback_url IS NOT NULL
            AND duration_ms IS NOT NULL
        )
        OR (
            processing_state IN ('uploaded', 'processing', 'failed')
            AND playback_url IS NULL
        )
    );

CREATE INDEX idx_media_processing_jobs_review_submit_retry
    ON app.media_processing_jobs (completed_at, updated_at, id)
    WHERE status = 'succeeded'
        AND last_error_code = 'review_submit_failed';

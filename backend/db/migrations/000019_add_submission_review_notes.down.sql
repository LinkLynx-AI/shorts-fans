ALTER TABLE app.submission_review_short_decisions
    DROP COLUMN IF EXISTS review_note;

ALTER TABLE app.submission_review_main_decisions
    DROP COLUMN IF EXISTS review_note;

ALTER TABLE app.shorts
    DROP COLUMN IF EXISTS review_note;

ALTER TABLE app.mains
    DROP COLUMN IF EXISTS review_note;

ALTER TABLE app.mains
    ADD COLUMN review_note TEXT;

ALTER TABLE app.shorts
    ADD COLUMN review_note TEXT;

ALTER TABLE app.submission_review_main_decisions
    ADD COLUMN review_note TEXT;

ALTER TABLE app.submission_review_short_decisions
    ADD COLUMN review_note TEXT;

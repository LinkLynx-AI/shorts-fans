CREATE TABLE app.short_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    short_id UUID NOT NULL REFERENCES app.shorts (id),
    author_user_id UUID NOT NULL REFERENCES app.users (id),
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (length(btrim(body)) BETWEEN 1 AND 500)
);

CREATE INDEX idx_short_comments_short_created
    ON app.short_comments (short_id, created_at DESC, id DESC);

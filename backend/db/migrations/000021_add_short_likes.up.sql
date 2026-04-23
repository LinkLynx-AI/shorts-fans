CREATE TABLE app.short_likes (
    user_id UUID NOT NULL REFERENCES app.users (id),
    short_id UUID NOT NULL REFERENCES app.shorts (id),
    liked_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, short_id)
);

CREATE INDEX idx_short_likes_short_id
    ON app.short_likes (short_id);

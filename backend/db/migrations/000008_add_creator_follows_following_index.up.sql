CREATE INDEX idx_creator_follows_user_followed_at_creator_user_id
    ON app.creator_follows (user_id, followed_at DESC, creator_user_id ASC);

CREATE TABLE IF NOT EXISTS user_blocks (
    blocker_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blocked_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT user_blocks_are_different CHECK (blocker_id <> blocked_user_id),
    CONSTRAINT user_blocks_pair_unique UNIQUE (blocker_id, blocked_user_id)
);

CREATE INDEX IF NOT EXISTS user_blocks_blocked_user_id_idx ON user_blocks(blocked_user_id);
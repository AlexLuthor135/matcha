CREATE TABLE IF NOT EXISTS user_reports (
    reporter_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reported_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT user_reports_users_are_different CHECK (reporter_id <> reported_user_id),
    CONSTRAINT user_reports_pair_unique UNIQUE (reporter_id, reported_user_id)
);

CREATE INDEX IF NOT EXISTS user_reports_reported_user_id_idx ON user_reports (reported_user_id);
-- +goose Up
CREATE TABLE sessions (
    id text PRIMARY KEY,
    user_id text NOT NULL,
    display_name text NOT NULL,
    csrf_token text NOT NULL,
    created_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL
);
CREATE INDEX sessions_expires_at_idx ON sessions (expires_at);

-- +goose Down
DROP TABLE sessions;

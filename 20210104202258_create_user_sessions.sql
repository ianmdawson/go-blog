-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE user_sessions (
    session_key PRIMARY KEY,
    id UUID,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
)

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE user_sessions;

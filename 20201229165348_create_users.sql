-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE users (
		id UUID PRIMARY KEY,
		role VARCHAR,
		username VARCHAR,
    password TEXT,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
	);
CREATE UNIQUE INDEX user_username_unique_idx ON users (username);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE users;

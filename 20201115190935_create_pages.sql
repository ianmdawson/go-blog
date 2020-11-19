-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE pages (
		id UUID PRIMARY KEY,
		title VARCHAR,
		body TEXT,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
	);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE pages;
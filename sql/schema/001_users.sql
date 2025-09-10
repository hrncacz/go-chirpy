-- +goose Up
CREATE TABLE users (
	id UUID PRIMARY KEY,
	create_at TIMESTAMP,
	updated_at TIMESTAMP,
	email TEXT
);

-- +goose Down
DROP TABLE users;

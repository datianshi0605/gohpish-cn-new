-- +goose Up
ALTER TABLE campaigns ADD COLUMN long_term boolean NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE campaigns DROP COLUMN long_term;

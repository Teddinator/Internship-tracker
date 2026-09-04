-- +goose Up

DELETE FROM idempotency_keys;

ALTER TABLE idempotency_keys
ADD COLUMN request_hash TEXT NOT NULL;

-- +goose Down

ALTER TABLE idempotency_keys
DROP COLUMN request_hash;

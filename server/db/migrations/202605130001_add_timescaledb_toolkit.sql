-- +goose Up
CREATE EXTENSION IF NOT EXISTS timescaledb_toolkit;

-- +goose Down
DROP EXTENSION IF EXISTS timescaledb_toolkit;

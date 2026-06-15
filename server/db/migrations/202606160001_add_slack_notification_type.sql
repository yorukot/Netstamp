-- +goose Up
ALTER TYPE notification_type ADD VALUE IF NOT EXISTS 'slack';

-- +goose Down
-- PostgreSQL enum values cannot be safely removed without rebuilding dependent columns.

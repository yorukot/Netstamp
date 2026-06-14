-- +goose Up
ALTER TYPE notification_channel_type ADD VALUE IF NOT EXISTS 'discord';
ALTER TYPE notification_channel_type ADD VALUE IF NOT EXISTS 'telegram';

-- +goose Down
-- PostgreSQL enum values cannot be safely removed without rebuilding dependent columns.

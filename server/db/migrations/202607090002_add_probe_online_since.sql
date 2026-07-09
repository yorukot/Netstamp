-- +goose Up
ALTER TABLE probe_statuses
    ADD COLUMN online_since timestamptz;

UPDATE probe_statuses
SET online_since = last_seen_at
WHERE status = 'online'
  AND last_seen_at IS NOT NULL
  AND last_seen_at >= now() - interval '35 seconds';

-- +goose Down
ALTER TABLE probe_statuses
    DROP COLUMN IF EXISTS online_since;

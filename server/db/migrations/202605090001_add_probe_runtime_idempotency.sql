-- +goose Up
CREATE UNIQUE INDEX uq_ping_results_project_probe_check_started_at
    ON ping_results (project_id, probe_id, check_id, started_at);

-- +goose Down
DROP INDEX IF EXISTS uq_ping_results_project_probe_check_started_at;

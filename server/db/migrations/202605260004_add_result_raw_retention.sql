-- +goose NO TRANSACTION
-- +goose Up
SELECT add_retention_policy('ping_results', INTERVAL '3 days', if_not_exists => TRUE);
SELECT add_retention_policy('tcp_results', INTERVAL '3 days', if_not_exists => TRUE);
SELECT add_retention_policy('traceroute_results', INTERVAL '3 days', if_not_exists => TRUE);

-- +goose Down
SELECT remove_retention_policy('ping_results', if_exists => TRUE);
SELECT remove_retention_policy('tcp_results', if_exists => TRUE);
SELECT remove_retention_policy('traceroute_results', if_exists => TRUE);

-- +goose NO TRANSACTION
-- +goose Up
CREATE MATERIALIZED VIEW tcp_result_rollups_1m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 minute', started_at) AS bucket,
    probe_id,
    check_id,
    count(*)::bigint AS result_count,
    sum(CASE WHEN status = 'successful' THEN 1 ELSE 0 END)::bigint AS successful_count,
    sum(CASE WHEN status = 'timeout' THEN 1 ELSE 0 END)::bigint AS timeout_count,
    sum(CASE WHEN status = 'error' THEN 1 ELSE 0 END)::bigint AS error_count,
    coalesce(sum(connect_duration_ms), 0)::double precision AS connect_duration_sum_ms,
    count(connect_duration_ms)::bigint AS connect_duration_count,
    min(connect_duration_ms)::double precision AS connect_duration_min_ms,
    max(connect_duration_ms)::double precision AS connect_duration_max_ms
FROM tcp_results
GROUP BY bucket, probe_id, check_id
WITH DATA;

SELECT add_continuous_aggregate_policy('tcp_result_rollups_1m', start_offset => INTERVAL '3 days', end_offset => INTERVAL '2 minutes', schedule_interval => INTERVAL '1 minute', if_not_exists => TRUE);

-- +goose Down
SELECT remove_continuous_aggregate_policy('tcp_result_rollups_1m', if_exists => TRUE);
DROP MATERIALIZED VIEW IF EXISTS tcp_result_rollups_1m;

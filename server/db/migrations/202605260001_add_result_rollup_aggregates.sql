-- +goose NO TRANSACTION
-- +goose Up
CREATE MATERIALIZED VIEW ping_results_1m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 minute', started_at) AS bucket,
    probe_id,
    check_id,
    count(*)::bigint AS result_count,
    sum(CASE WHEN status = 'successful' THEN 1 ELSE 0 END)::bigint AS successful_count,
    sum(CASE WHEN status = 'timeout' THEN 1 ELSE 0 END)::bigint AS timeout_count,
    sum(CASE WHEN status = 'error' THEN 1 ELSE 0 END)::bigint AS error_count,
    sum(duration_ms)::double precision AS duration_sum_ms,
    count(duration_ms)::bigint AS duration_count,
    min(duration_ms)::double precision AS duration_min_ms,
    max(duration_ms)::double precision AS duration_max_ms,
    sum(sent_count)::bigint AS sent_count,
    sum(received_count)::bigint AS received_count,
    sum(loss_percent)::double precision AS loss_sum_percent,
    count(loss_percent)::bigint AS loss_count,
    min(rtt_min_ms)::double precision AS rtt_min_ms,
    sum(rtt_avg_ms)::double precision AS rtt_avg_sum_ms,
    count(rtt_avg_ms)::bigint AS rtt_avg_count,
    max(rtt_max_ms)::double precision AS rtt_max_ms,
    sum(rtt_stddev_ms)::double precision AS rtt_stddev_sum_ms,
    count(rtt_stddev_ms)::bigint AS rtt_stddev_count
FROM ping_results
GROUP BY bucket, probe_id, check_id
WITH DATA;

CREATE MATERIALIZED VIEW ping_results_10m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('10 minutes', started_at) AS bucket,
    probe_id,
    check_id,
    count(*)::bigint AS result_count,
    sum(CASE WHEN status = 'successful' THEN 1 ELSE 0 END)::bigint AS successful_count,
    sum(CASE WHEN status = 'timeout' THEN 1 ELSE 0 END)::bigint AS timeout_count,
    sum(CASE WHEN status = 'error' THEN 1 ELSE 0 END)::bigint AS error_count,
    sum(duration_ms)::double precision AS duration_sum_ms,
    count(duration_ms)::bigint AS duration_count,
    min(duration_ms)::double precision AS duration_min_ms,
    max(duration_ms)::double precision AS duration_max_ms,
    sum(sent_count)::bigint AS sent_count,
    sum(received_count)::bigint AS received_count,
    sum(loss_percent)::double precision AS loss_sum_percent,
    count(loss_percent)::bigint AS loss_count,
    min(rtt_min_ms)::double precision AS rtt_min_ms,
    sum(rtt_avg_ms)::double precision AS rtt_avg_sum_ms,
    count(rtt_avg_ms)::bigint AS rtt_avg_count,
    max(rtt_max_ms)::double precision AS rtt_max_ms,
    sum(rtt_stddev_ms)::double precision AS rtt_stddev_sum_ms,
    count(rtt_stddev_ms)::bigint AS rtt_stddev_count
FROM ping_results
GROUP BY bucket, probe_id, check_id
WITH DATA;

CREATE MATERIALIZED VIEW ping_results_15m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('15 minutes', started_at) AS bucket,
    probe_id,
    check_id,
    count(*)::bigint AS result_count,
    sum(CASE WHEN status = 'successful' THEN 1 ELSE 0 END)::bigint AS successful_count,
    sum(CASE WHEN status = 'timeout' THEN 1 ELSE 0 END)::bigint AS timeout_count,
    sum(CASE WHEN status = 'error' THEN 1 ELSE 0 END)::bigint AS error_count,
    sum(duration_ms)::double precision AS duration_sum_ms,
    count(duration_ms)::bigint AS duration_count,
    min(duration_ms)::double precision AS duration_min_ms,
    max(duration_ms)::double precision AS duration_max_ms,
    sum(sent_count)::bigint AS sent_count,
    sum(received_count)::bigint AS received_count,
    sum(loss_percent)::double precision AS loss_sum_percent,
    count(loss_percent)::bigint AS loss_count,
    min(rtt_min_ms)::double precision AS rtt_min_ms,
    sum(rtt_avg_ms)::double precision AS rtt_avg_sum_ms,
    count(rtt_avg_ms)::bigint AS rtt_avg_count,
    max(rtt_max_ms)::double precision AS rtt_max_ms,
    sum(rtt_stddev_ms)::double precision AS rtt_stddev_sum_ms,
    count(rtt_stddev_ms)::bigint AS rtt_stddev_count
FROM ping_results
GROUP BY bucket, probe_id, check_id
WITH DATA;

CREATE MATERIALIZED VIEW ping_results_30m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('30 minutes', started_at) AS bucket,
    probe_id,
    check_id,
    count(*)::bigint AS result_count,
    sum(CASE WHEN status = 'successful' THEN 1 ELSE 0 END)::bigint AS successful_count,
    sum(CASE WHEN status = 'timeout' THEN 1 ELSE 0 END)::bigint AS timeout_count,
    sum(CASE WHEN status = 'error' THEN 1 ELSE 0 END)::bigint AS error_count,
    sum(duration_ms)::double precision AS duration_sum_ms,
    count(duration_ms)::bigint AS duration_count,
    min(duration_ms)::double precision AS duration_min_ms,
    max(duration_ms)::double precision AS duration_max_ms,
    sum(sent_count)::bigint AS sent_count,
    sum(received_count)::bigint AS received_count,
    sum(loss_percent)::double precision AS loss_sum_percent,
    count(loss_percent)::bigint AS loss_count,
    min(rtt_min_ms)::double precision AS rtt_min_ms,
    sum(rtt_avg_ms)::double precision AS rtt_avg_sum_ms,
    count(rtt_avg_ms)::bigint AS rtt_avg_count,
    max(rtt_max_ms)::double precision AS rtt_max_ms,
    sum(rtt_stddev_ms)::double precision AS rtt_stddev_sum_ms,
    count(rtt_stddev_ms)::bigint AS rtt_stddev_count
FROM ping_results
GROUP BY bucket, probe_id, check_id
WITH DATA;

CREATE MATERIALIZED VIEW ping_results_1h
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', started_at) AS bucket,
    probe_id,
    check_id,
    count(*)::bigint AS result_count,
    sum(CASE WHEN status = 'successful' THEN 1 ELSE 0 END)::bigint AS successful_count,
    sum(CASE WHEN status = 'timeout' THEN 1 ELSE 0 END)::bigint AS timeout_count,
    sum(CASE WHEN status = 'error' THEN 1 ELSE 0 END)::bigint AS error_count,
    sum(duration_ms)::double precision AS duration_sum_ms,
    count(duration_ms)::bigint AS duration_count,
    min(duration_ms)::double precision AS duration_min_ms,
    max(duration_ms)::double precision AS duration_max_ms,
    sum(sent_count)::bigint AS sent_count,
    sum(received_count)::bigint AS received_count,
    sum(loss_percent)::double precision AS loss_sum_percent,
    count(loss_percent)::bigint AS loss_count,
    min(rtt_min_ms)::double precision AS rtt_min_ms,
    sum(rtt_avg_ms)::double precision AS rtt_avg_sum_ms,
    count(rtt_avg_ms)::bigint AS rtt_avg_count,
    max(rtt_max_ms)::double precision AS rtt_max_ms,
    sum(rtt_stddev_ms)::double precision AS rtt_stddev_sum_ms,
    count(rtt_stddev_ms)::bigint AS rtt_stddev_count
FROM ping_results
GROUP BY bucket, probe_id, check_id
WITH DATA;

CREATE MATERIALIZED VIEW tcp_results_1m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 minute', started_at) AS bucket,
    probe_id,
    check_id,
    count(*)::bigint AS result_count,
    sum(CASE WHEN status = 'successful' THEN 1 ELSE 0 END)::bigint AS successful_count,
    sum(CASE WHEN status = 'timeout' THEN 1 ELSE 0 END)::bigint AS timeout_count,
    sum(CASE WHEN status = 'error' THEN 1 ELSE 0 END)::bigint AS error_count,
    sum(duration_ms)::double precision AS duration_sum_ms,
    count(duration_ms)::bigint AS duration_count,
    min(duration_ms)::double precision AS duration_min_ms,
    max(duration_ms)::double precision AS duration_max_ms,
    sum(connect_duration_ms)::double precision AS connect_duration_sum_ms,
    count(connect_duration_ms)::bigint AS connect_duration_count,
    min(connect_duration_ms)::double precision AS connect_duration_min_ms,
    max(connect_duration_ms)::double precision AS connect_duration_max_ms
FROM tcp_results
GROUP BY bucket, probe_id, check_id
WITH DATA;

CREATE MATERIALIZED VIEW tcp_results_10m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('10 minutes', started_at) AS bucket,
    probe_id,
    check_id,
    count(*)::bigint AS result_count,
    sum(CASE WHEN status = 'successful' THEN 1 ELSE 0 END)::bigint AS successful_count,
    sum(CASE WHEN status = 'timeout' THEN 1 ELSE 0 END)::bigint AS timeout_count,
    sum(CASE WHEN status = 'error' THEN 1 ELSE 0 END)::bigint AS error_count,
    sum(duration_ms)::double precision AS duration_sum_ms,
    count(duration_ms)::bigint AS duration_count,
    min(duration_ms)::double precision AS duration_min_ms,
    max(duration_ms)::double precision AS duration_max_ms,
    sum(connect_duration_ms)::double precision AS connect_duration_sum_ms,
    count(connect_duration_ms)::bigint AS connect_duration_count,
    min(connect_duration_ms)::double precision AS connect_duration_min_ms,
    max(connect_duration_ms)::double precision AS connect_duration_max_ms
FROM tcp_results
GROUP BY bucket, probe_id, check_id
WITH DATA;

CREATE MATERIALIZED VIEW tcp_results_15m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('15 minutes', started_at) AS bucket,
    probe_id,
    check_id,
    count(*)::bigint AS result_count,
    sum(CASE WHEN status = 'successful' THEN 1 ELSE 0 END)::bigint AS successful_count,
    sum(CASE WHEN status = 'timeout' THEN 1 ELSE 0 END)::bigint AS timeout_count,
    sum(CASE WHEN status = 'error' THEN 1 ELSE 0 END)::bigint AS error_count,
    sum(duration_ms)::double precision AS duration_sum_ms,
    count(duration_ms)::bigint AS duration_count,
    min(duration_ms)::double precision AS duration_min_ms,
    max(duration_ms)::double precision AS duration_max_ms,
    sum(connect_duration_ms)::double precision AS connect_duration_sum_ms,
    count(connect_duration_ms)::bigint AS connect_duration_count,
    min(connect_duration_ms)::double precision AS connect_duration_min_ms,
    max(connect_duration_ms)::double precision AS connect_duration_max_ms
FROM tcp_results
GROUP BY bucket, probe_id, check_id
WITH DATA;

CREATE MATERIALIZED VIEW tcp_results_30m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('30 minutes', started_at) AS bucket,
    probe_id,
    check_id,
    count(*)::bigint AS result_count,
    sum(CASE WHEN status = 'successful' THEN 1 ELSE 0 END)::bigint AS successful_count,
    sum(CASE WHEN status = 'timeout' THEN 1 ELSE 0 END)::bigint AS timeout_count,
    sum(CASE WHEN status = 'error' THEN 1 ELSE 0 END)::bigint AS error_count,
    sum(duration_ms)::double precision AS duration_sum_ms,
    count(duration_ms)::bigint AS duration_count,
    min(duration_ms)::double precision AS duration_min_ms,
    max(duration_ms)::double precision AS duration_max_ms,
    sum(connect_duration_ms)::double precision AS connect_duration_sum_ms,
    count(connect_duration_ms)::bigint AS connect_duration_count,
    min(connect_duration_ms)::double precision AS connect_duration_min_ms,
    max(connect_duration_ms)::double precision AS connect_duration_max_ms
FROM tcp_results
GROUP BY bucket, probe_id, check_id
WITH DATA;

CREATE MATERIALIZED VIEW tcp_results_1h
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', started_at) AS bucket,
    probe_id,
    check_id,
    count(*)::bigint AS result_count,
    sum(CASE WHEN status = 'successful' THEN 1 ELSE 0 END)::bigint AS successful_count,
    sum(CASE WHEN status = 'timeout' THEN 1 ELSE 0 END)::bigint AS timeout_count,
    sum(CASE WHEN status = 'error' THEN 1 ELSE 0 END)::bigint AS error_count,
    sum(duration_ms)::double precision AS duration_sum_ms,
    count(duration_ms)::bigint AS duration_count,
    min(duration_ms)::double precision AS duration_min_ms,
    max(duration_ms)::double precision AS duration_max_ms,
    sum(connect_duration_ms)::double precision AS connect_duration_sum_ms,
    count(connect_duration_ms)::bigint AS connect_duration_count,
    min(connect_duration_ms)::double precision AS connect_duration_min_ms,
    max(connect_duration_ms)::double precision AS connect_duration_max_ms
FROM tcp_results
GROUP BY bucket, probe_id, check_id
WITH DATA;

CREATE MATERIALIZED VIEW traceroute_results_1m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 minute', started_at) AS bucket,
    probe_id,
    check_id,
    count(*)::bigint AS result_count,
    sum(CASE WHEN status = 'successful' THEN 1 ELSE 0 END)::bigint AS successful_count,
    sum(CASE WHEN status = 'timeout' THEN 1 ELSE 0 END)::bigint AS timeout_count,
    sum(CASE WHEN status = 'error' THEN 1 ELSE 0 END)::bigint AS error_count,
    sum(CASE WHEN status = 'partial' THEN 1 ELSE 0 END)::bigint AS partial_count,
    sum(duration_ms)::double precision AS duration_sum_ms,
    count(duration_ms)::bigint AS duration_count,
    min(duration_ms)::double precision AS duration_min_ms,
    max(duration_ms)::double precision AS duration_max_ms,
    sum(CASE WHEN destination_reached THEN 1 ELSE 0 END)::bigint AS destination_reached_count,
    sum(hop_count)::bigint AS hop_count_sum,
    count(hop_count)::bigint AS hop_count_count,
    min(hop_count)::bigint AS hop_count_min,
    max(hop_count)::bigint AS hop_count_max
FROM traceroute_results
GROUP BY bucket, probe_id, check_id
WITH DATA;

CREATE MATERIALIZED VIEW traceroute_results_10m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('10 minutes', started_at) AS bucket,
    probe_id,
    check_id,
    count(*)::bigint AS result_count,
    sum(CASE WHEN status = 'successful' THEN 1 ELSE 0 END)::bigint AS successful_count,
    sum(CASE WHEN status = 'timeout' THEN 1 ELSE 0 END)::bigint AS timeout_count,
    sum(CASE WHEN status = 'error' THEN 1 ELSE 0 END)::bigint AS error_count,
    sum(CASE WHEN status = 'partial' THEN 1 ELSE 0 END)::bigint AS partial_count,
    sum(duration_ms)::double precision AS duration_sum_ms,
    count(duration_ms)::bigint AS duration_count,
    min(duration_ms)::double precision AS duration_min_ms,
    max(duration_ms)::double precision AS duration_max_ms,
    sum(CASE WHEN destination_reached THEN 1 ELSE 0 END)::bigint AS destination_reached_count,
    sum(hop_count)::bigint AS hop_count_sum,
    count(hop_count)::bigint AS hop_count_count,
    min(hop_count)::bigint AS hop_count_min,
    max(hop_count)::bigint AS hop_count_max
FROM traceroute_results
GROUP BY bucket, probe_id, check_id
WITH DATA;

CREATE MATERIALIZED VIEW traceroute_results_15m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('15 minutes', started_at) AS bucket,
    probe_id,
    check_id,
    count(*)::bigint AS result_count,
    sum(CASE WHEN status = 'successful' THEN 1 ELSE 0 END)::bigint AS successful_count,
    sum(CASE WHEN status = 'timeout' THEN 1 ELSE 0 END)::bigint AS timeout_count,
    sum(CASE WHEN status = 'error' THEN 1 ELSE 0 END)::bigint AS error_count,
    sum(CASE WHEN status = 'partial' THEN 1 ELSE 0 END)::bigint AS partial_count,
    sum(duration_ms)::double precision AS duration_sum_ms,
    count(duration_ms)::bigint AS duration_count,
    min(duration_ms)::double precision AS duration_min_ms,
    max(duration_ms)::double precision AS duration_max_ms,
    sum(CASE WHEN destination_reached THEN 1 ELSE 0 END)::bigint AS destination_reached_count,
    sum(hop_count)::bigint AS hop_count_sum,
    count(hop_count)::bigint AS hop_count_count,
    min(hop_count)::bigint AS hop_count_min,
    max(hop_count)::bigint AS hop_count_max
FROM traceroute_results
GROUP BY bucket, probe_id, check_id
WITH DATA;

CREATE MATERIALIZED VIEW traceroute_results_30m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('30 minutes', started_at) AS bucket,
    probe_id,
    check_id,
    count(*)::bigint AS result_count,
    sum(CASE WHEN status = 'successful' THEN 1 ELSE 0 END)::bigint AS successful_count,
    sum(CASE WHEN status = 'timeout' THEN 1 ELSE 0 END)::bigint AS timeout_count,
    sum(CASE WHEN status = 'error' THEN 1 ELSE 0 END)::bigint AS error_count,
    sum(CASE WHEN status = 'partial' THEN 1 ELSE 0 END)::bigint AS partial_count,
    sum(duration_ms)::double precision AS duration_sum_ms,
    count(duration_ms)::bigint AS duration_count,
    min(duration_ms)::double precision AS duration_min_ms,
    max(duration_ms)::double precision AS duration_max_ms,
    sum(CASE WHEN destination_reached THEN 1 ELSE 0 END)::bigint AS destination_reached_count,
    sum(hop_count)::bigint AS hop_count_sum,
    count(hop_count)::bigint AS hop_count_count,
    min(hop_count)::bigint AS hop_count_min,
    max(hop_count)::bigint AS hop_count_max
FROM traceroute_results
GROUP BY bucket, probe_id, check_id
WITH DATA;

CREATE MATERIALIZED VIEW traceroute_results_1h
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', started_at) AS bucket,
    probe_id,
    check_id,
    count(*)::bigint AS result_count,
    sum(CASE WHEN status = 'successful' THEN 1 ELSE 0 END)::bigint AS successful_count,
    sum(CASE WHEN status = 'timeout' THEN 1 ELSE 0 END)::bigint AS timeout_count,
    sum(CASE WHEN status = 'error' THEN 1 ELSE 0 END)::bigint AS error_count,
    sum(CASE WHEN status = 'partial' THEN 1 ELSE 0 END)::bigint AS partial_count,
    sum(duration_ms)::double precision AS duration_sum_ms,
    count(duration_ms)::bigint AS duration_count,
    min(duration_ms)::double precision AS duration_min_ms,
    max(duration_ms)::double precision AS duration_max_ms,
    sum(CASE WHEN destination_reached THEN 1 ELSE 0 END)::bigint AS destination_reached_count,
    sum(hop_count)::bigint AS hop_count_sum,
    count(hop_count)::bigint AS hop_count_count,
    min(hop_count)::bigint AS hop_count_min,
    max(hop_count)::bigint AS hop_count_max
FROM traceroute_results
GROUP BY bucket, probe_id, check_id
WITH DATA;


-- Refresh only the raw retention window; refreshing older windows after raw chunks expire would erase backfilled aggregates.
SELECT add_continuous_aggregate_policy('ping_results_1m', start_offset => INTERVAL '3 days', end_offset => INTERVAL '2 minutes', schedule_interval => INTERVAL '1 minute', if_not_exists => TRUE);
SELECT add_continuous_aggregate_policy('ping_results_10m', start_offset => INTERVAL '3 days', end_offset => INTERVAL '10 minutes', schedule_interval => INTERVAL '10 minutes', if_not_exists => TRUE);
SELECT add_continuous_aggregate_policy('ping_results_15m', start_offset => INTERVAL '3 days', end_offset => INTERVAL '15 minutes', schedule_interval => INTERVAL '15 minutes', if_not_exists => TRUE);
SELECT add_continuous_aggregate_policy('ping_results_30m', start_offset => INTERVAL '3 days', end_offset => INTERVAL '30 minutes', schedule_interval => INTERVAL '30 minutes', if_not_exists => TRUE);
SELECT add_continuous_aggregate_policy('ping_results_1h', start_offset => INTERVAL '3 days', end_offset => INTERVAL '1 hour', schedule_interval => INTERVAL '1 hour', if_not_exists => TRUE);
SELECT add_continuous_aggregate_policy('tcp_results_1m', start_offset => INTERVAL '3 days', end_offset => INTERVAL '2 minutes', schedule_interval => INTERVAL '1 minute', if_not_exists => TRUE);
SELECT add_continuous_aggregate_policy('tcp_results_10m', start_offset => INTERVAL '3 days', end_offset => INTERVAL '10 minutes', schedule_interval => INTERVAL '10 minutes', if_not_exists => TRUE);
SELECT add_continuous_aggregate_policy('tcp_results_15m', start_offset => INTERVAL '3 days', end_offset => INTERVAL '15 minutes', schedule_interval => INTERVAL '15 minutes', if_not_exists => TRUE);
SELECT add_continuous_aggregate_policy('tcp_results_30m', start_offset => INTERVAL '3 days', end_offset => INTERVAL '30 minutes', schedule_interval => INTERVAL '30 minutes', if_not_exists => TRUE);
SELECT add_continuous_aggregate_policy('tcp_results_1h', start_offset => INTERVAL '3 days', end_offset => INTERVAL '1 hour', schedule_interval => INTERVAL '1 hour', if_not_exists => TRUE);
SELECT add_continuous_aggregate_policy('traceroute_results_1m', start_offset => INTERVAL '3 days', end_offset => INTERVAL '2 minutes', schedule_interval => INTERVAL '1 minute', if_not_exists => TRUE);
SELECT add_continuous_aggregate_policy('traceroute_results_10m', start_offset => INTERVAL '3 days', end_offset => INTERVAL '10 minutes', schedule_interval => INTERVAL '10 minutes', if_not_exists => TRUE);
SELECT add_continuous_aggregate_policy('traceroute_results_15m', start_offset => INTERVAL '3 days', end_offset => INTERVAL '15 minutes', schedule_interval => INTERVAL '15 minutes', if_not_exists => TRUE);
SELECT add_continuous_aggregate_policy('traceroute_results_30m', start_offset => INTERVAL '3 days', end_offset => INTERVAL '30 minutes', schedule_interval => INTERVAL '30 minutes', if_not_exists => TRUE);
SELECT add_continuous_aggregate_policy('traceroute_results_1h', start_offset => INTERVAL '3 days', end_offset => INTERVAL '1 hour', schedule_interval => INTERVAL '1 hour', if_not_exists => TRUE);
SELECT add_retention_policy('ping_results_1m', INTERVAL '30 days', if_not_exists => TRUE);
SELECT add_retention_policy('ping_results_10m', INTERVAL '90 days', if_not_exists => TRUE);
SELECT add_retention_policy('ping_results_15m', INTERVAL '90 days', if_not_exists => TRUE);
SELECT add_retention_policy('ping_results_30m', INTERVAL '180 days', if_not_exists => TRUE);
SELECT add_retention_policy('tcp_results_1m', INTERVAL '30 days', if_not_exists => TRUE);
SELECT add_retention_policy('tcp_results_10m', INTERVAL '90 days', if_not_exists => TRUE);
SELECT add_retention_policy('tcp_results_15m', INTERVAL '90 days', if_not_exists => TRUE);
SELECT add_retention_policy('tcp_results_30m', INTERVAL '180 days', if_not_exists => TRUE);
SELECT add_retention_policy('traceroute_results_1m', INTERVAL '30 days', if_not_exists => TRUE);
SELECT add_retention_policy('traceroute_results_10m', INTERVAL '90 days', if_not_exists => TRUE);
SELECT add_retention_policy('traceroute_results_15m', INTERVAL '90 days', if_not_exists => TRUE);
SELECT add_retention_policy('traceroute_results_30m', INTERVAL '180 days', if_not_exists => TRUE);

-- +goose Down
SELECT remove_retention_policy('ping_results_1m', if_exists => TRUE);
SELECT remove_retention_policy('ping_results_10m', if_exists => TRUE);
SELECT remove_retention_policy('ping_results_15m', if_exists => TRUE);
SELECT remove_retention_policy('ping_results_30m', if_exists => TRUE);
SELECT remove_retention_policy('tcp_results_1m', if_exists => TRUE);
SELECT remove_retention_policy('tcp_results_10m', if_exists => TRUE);
SELECT remove_retention_policy('tcp_results_15m', if_exists => TRUE);
SELECT remove_retention_policy('tcp_results_30m', if_exists => TRUE);
SELECT remove_retention_policy('traceroute_results_1m', if_exists => TRUE);
SELECT remove_retention_policy('traceroute_results_10m', if_exists => TRUE);
SELECT remove_retention_policy('traceroute_results_15m', if_exists => TRUE);
SELECT remove_retention_policy('traceroute_results_30m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('ping_results_1m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('ping_results_10m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('ping_results_15m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('ping_results_30m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('ping_results_1h', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('tcp_results_1m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('tcp_results_10m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('tcp_results_15m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('tcp_results_30m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('tcp_results_1h', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('traceroute_results_1m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('traceroute_results_10m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('traceroute_results_15m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('traceroute_results_30m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('traceroute_results_1h', if_exists => TRUE);
DROP MATERIALIZED VIEW IF EXISTS traceroute_results_1h;
DROP MATERIALIZED VIEW IF EXISTS traceroute_results_30m;
DROP MATERIALIZED VIEW IF EXISTS traceroute_results_15m;
DROP MATERIALIZED VIEW IF EXISTS traceroute_results_10m;
DROP MATERIALIZED VIEW IF EXISTS traceroute_results_1m;
DROP MATERIALIZED VIEW IF EXISTS tcp_results_1h;
DROP MATERIALIZED VIEW IF EXISTS tcp_results_30m;
DROP MATERIALIZED VIEW IF EXISTS tcp_results_15m;
DROP MATERIALIZED VIEW IF EXISTS tcp_results_10m;
DROP MATERIALIZED VIEW IF EXISTS tcp_results_1m;
DROP MATERIALIZED VIEW IF EXISTS ping_results_1h;
DROP MATERIALIZED VIEW IF EXISTS ping_results_30m;
DROP MATERIALIZED VIEW IF EXISTS ping_results_15m;
DROP MATERIALIZED VIEW IF EXISTS ping_results_10m;
DROP MATERIALIZED VIEW IF EXISTS ping_results_1m;

-- +goose NO TRANSACTION
-- +goose Up
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
SELECT remove_retention_policy('ping_rtt_sample_density_1m', if_exists => TRUE);
SELECT remove_retention_policy('ping_rtt_sample_density_10m', if_exists => TRUE);
SELECT remove_retention_policy('ping_rtt_sample_density_15m', if_exists => TRUE);
SELECT remove_retention_policy('ping_rtt_sample_density_30m', if_exists => TRUE);
SELECT remove_retention_policy('traceroute_topology_nodes_1m', if_exists => TRUE);
SELECT remove_retention_policy('traceroute_topology_nodes_10m', if_exists => TRUE);
SELECT remove_retention_policy('traceroute_topology_nodes_15m', if_exists => TRUE);
SELECT remove_retention_policy('traceroute_topology_nodes_30m', if_exists => TRUE);
SELECT remove_retention_policy('traceroute_topology_edges_1m', if_exists => TRUE);
SELECT remove_retention_policy('traceroute_topology_edges_10m', if_exists => TRUE);
SELECT remove_retention_policy('traceroute_topology_edges_15m', if_exists => TRUE);
SELECT remove_retention_policy('traceroute_topology_edges_30m', if_exists => TRUE);
SELECT remove_retention_policy('traceroute_hop_observations', if_exists => TRUE);
SELECT remove_retention_policy('traceroute_edge_observations', if_exists => TRUE);

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
SELECT remove_continuous_aggregate_policy('ping_rtt_sample_density_1m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('ping_rtt_sample_density_10m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('ping_rtt_sample_density_15m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('ping_rtt_sample_density_30m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('ping_rtt_sample_density_1h', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('traceroute_topology_nodes_1m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('traceroute_topology_nodes_10m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('traceroute_topology_nodes_15m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('traceroute_topology_nodes_30m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('traceroute_topology_nodes_1h', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('traceroute_topology_edges_1m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('traceroute_topology_edges_10m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('traceroute_topology_edges_15m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('traceroute_topology_edges_30m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('traceroute_topology_edges_1h', if_exists => TRUE);

DROP MATERIALIZED VIEW IF EXISTS traceroute_topology_edges_1h;
DROP MATERIALIZED VIEW IF EXISTS traceroute_topology_edges_30m;
DROP MATERIALIZED VIEW IF EXISTS traceroute_topology_edges_15m;
DROP MATERIALIZED VIEW IF EXISTS traceroute_topology_edges_10m;
DROP MATERIALIZED VIEW IF EXISTS traceroute_topology_edges_1m;
DROP MATERIALIZED VIEW IF EXISTS traceroute_topology_nodes_1h;
DROP MATERIALIZED VIEW IF EXISTS traceroute_topology_nodes_30m;
DROP MATERIALIZED VIEW IF EXISTS traceroute_topology_nodes_15m;
DROP MATERIALIZED VIEW IF EXISTS traceroute_topology_nodes_10m;
DROP MATERIALIZED VIEW IF EXISTS traceroute_topology_nodes_1m;
DROP MATERIALIZED VIEW IF EXISTS ping_rtt_sample_density_1h;
DROP MATERIALIZED VIEW IF EXISTS ping_rtt_sample_density_30m;
DROP MATERIALIZED VIEW IF EXISTS ping_rtt_sample_density_15m;
DROP MATERIALIZED VIEW IF EXISTS ping_rtt_sample_density_10m;
DROP MATERIALIZED VIEW IF EXISTS ping_rtt_sample_density_1m;
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

DROP TRIGGER IF EXISTS record_traceroute_topology_observations ON traceroute_result_hops;
DROP FUNCTION IF EXISTS record_traceroute_topology_observations();
DROP TABLE IF EXISTS traceroute_edge_observations;
DROP TABLE IF EXISTS traceroute_hop_observations;

CREATE MATERIALIZED VIEW ping_result_rollups_1m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 minute', started_at) AS bucket,
    probe_id,
    check_id,
    count(*)::bigint AS result_count,
    sum(CASE WHEN status = 'successful' THEN 1 ELSE 0 END)::bigint AS successful_count,
    sum(sent_count)::bigint AS sent_count,
    sum(received_count)::bigint AS received_count,
    coalesce(sum(rtt_avg_ms), 0)::double precision AS rtt_avg_sum_ms,
    count(rtt_avg_ms)::bigint AS rtt_avg_count,
    min(rtt_min_ms)::double precision AS rtt_min_ms,
    max(rtt_max_ms)::double precision AS rtt_max_ms
FROM ping_results
GROUP BY bucket, probe_id, check_id
WITH DATA;

SELECT add_continuous_aggregate_policy('ping_result_rollups_1m', start_offset => INTERVAL '3 days', end_offset => INTERVAL '2 minutes', schedule_interval => INTERVAL '1 minute', if_not_exists => TRUE);

-- +goose Down
SELECT remove_continuous_aggregate_policy('ping_result_rollups_1m', if_exists => TRUE);
DROP MATERIALIZED VIEW IF EXISTS ping_result_rollups_1m;

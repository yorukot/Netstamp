-- +goose NO TRANSACTION
-- +goose Up
CREATE TABLE traceroute_hop_observations (
    probe_id bigint NOT NULL,
    check_id bigint NOT NULL,
    started_at timestamptz NOT NULL,
    hop_index integer NOT NULL,
    address inet,
    hostname text,
    sent_count integer NOT NULL,
    received_count integer NOT NULL,
    loss_percent double precision NOT NULL,
    rtt_avg_ms double precision,
    PRIMARY KEY (probe_id, check_id, started_at, hop_index),
    CONSTRAINT traceroute_hop_observations_hop_index_positive CHECK (hop_index > 0),
    CONSTRAINT traceroute_hop_observations_sent_count_non_negative CHECK (sent_count >= 0),
    CONSTRAINT traceroute_hop_observations_received_count_range CHECK (received_count >= 0 AND received_count <= sent_count),
    CONSTRAINT traceroute_hop_observations_loss_percent_range CHECK (loss_percent >= 0 AND loss_percent <= 100),
    CONSTRAINT traceroute_hop_observations_rtt_avg_ms_non_negative CHECK (rtt_avg_ms IS NULL OR rtt_avg_ms >= 0),
    CONSTRAINT traceroute_hop_observations_hostname_not_empty CHECK (hostname IS NULL OR length(btrim(hostname)) > 0)
);

CREATE TABLE traceroute_edge_observations (
    probe_id bigint NOT NULL,
    check_id bigint NOT NULL,
    started_at timestamptz NOT NULL,
    source_hop_index integer NOT NULL,
    target_hop_index integer NOT NULL,
    source_address inet,
    target_address inet,
    source_hostname text,
    target_hostname text,
    source_loss_percent double precision NOT NULL,
    target_loss_percent double precision NOT NULL,
    source_rtt_avg_ms double precision,
    target_rtt_avg_ms double precision,
    PRIMARY KEY (probe_id, check_id, started_at, source_hop_index, target_hop_index),
    CONSTRAINT traceroute_edge_observations_source_hop_index_positive CHECK (source_hop_index > 0),
    CONSTRAINT traceroute_edge_observations_target_hop_index_next CHECK (target_hop_index = source_hop_index + 1),
    CONSTRAINT traceroute_edge_observations_source_loss_percent_range CHECK (source_loss_percent >= 0 AND source_loss_percent <= 100),
    CONSTRAINT traceroute_edge_observations_target_loss_percent_range CHECK (target_loss_percent >= 0 AND target_loss_percent <= 100),
    CONSTRAINT traceroute_edge_observations_source_rtt_avg_ms_non_negative CHECK (source_rtt_avg_ms IS NULL OR source_rtt_avg_ms >= 0),
    CONSTRAINT traceroute_edge_observations_target_rtt_avg_ms_non_negative CHECK (target_rtt_avg_ms IS NULL OR target_rtt_avg_ms >= 0),
    CONSTRAINT traceroute_edge_observations_source_hostname_not_empty CHECK (source_hostname IS NULL OR length(btrim(source_hostname)) > 0),
    CONSTRAINT traceroute_edge_observations_target_hostname_not_empty CHECK (target_hostname IS NULL OR length(btrim(target_hostname)) > 0)
);

SELECT create_hypertable('traceroute_hop_observations', 'started_at', if_not_exists => TRUE);
SELECT create_hypertable('traceroute_edge_observations', 'started_at', if_not_exists => TRUE);

CREATE INDEX idx_traceroute_hop_observations_probe_check_started_at
    ON traceroute_hop_observations (probe_id, check_id, started_at DESC);
CREATE INDEX idx_traceroute_edge_observations_probe_check_started_at
    ON traceroute_edge_observations (probe_id, check_id, started_at DESC);

INSERT INTO traceroute_hop_observations (
    probe_id,
    check_id,
    started_at,
    hop_index,
    address,
    hostname,
    sent_count,
    received_count,
    loss_percent,
    rtt_avg_ms
)
SELECT
    probe_id,
    check_id,
    started_at,
    hop_index,
    address,
    hostname,
    sent_count,
    received_count,
    loss_percent,
    rtt_avg_ms
FROM traceroute_result_hops
ON CONFLICT (probe_id, check_id, started_at, hop_index) DO NOTHING;

INSERT INTO traceroute_edge_observations (
    probe_id,
    check_id,
    started_at,
    source_hop_index,
    target_hop_index,
    source_address,
    target_address,
    source_hostname,
    target_hostname,
    source_loss_percent,
    target_loss_percent,
    source_rtt_avg_ms,
    target_rtt_avg_ms
)
SELECT
    source.probe_id,
    source.check_id,
    source.started_at,
    source.hop_index AS source_hop_index,
    target.hop_index AS target_hop_index,
    source.address AS source_address,
    target.address AS target_address,
    source.hostname AS source_hostname,
    target.hostname AS target_hostname,
    source.loss_percent AS source_loss_percent,
    target.loss_percent AS target_loss_percent,
    source.rtt_avg_ms AS source_rtt_avg_ms,
    target.rtt_avg_ms AS target_rtt_avg_ms
FROM traceroute_result_hops AS source
JOIN traceroute_result_hops AS target
    ON target.probe_id = source.probe_id
    AND target.check_id = source.check_id
    AND target.started_at = source.started_at
    AND target.hop_index = source.hop_index + 1
ON CONFLICT (probe_id, check_id, started_at, source_hop_index, target_hop_index) DO NOTHING;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION record_traceroute_topology_observations()
RETURNS trigger AS $$
BEGIN
    INSERT INTO traceroute_hop_observations (
        probe_id,
        check_id,
        started_at,
        hop_index,
        address,
        hostname,
        sent_count,
        received_count,
        loss_percent,
        rtt_avg_ms
    )
    VALUES (
        NEW.probe_id,
        NEW.check_id,
        NEW.started_at,
        NEW.hop_index,
        NEW.address,
        NEW.hostname,
        NEW.sent_count,
        NEW.received_count,
        NEW.loss_percent,
        NEW.rtt_avg_ms
    )
    ON CONFLICT (probe_id, check_id, started_at, hop_index) DO NOTHING;

    INSERT INTO traceroute_edge_observations (
        probe_id,
        check_id,
        started_at,
        source_hop_index,
        target_hop_index,
        source_address,
        target_address,
        source_hostname,
        target_hostname,
        source_loss_percent,
        target_loss_percent,
        source_rtt_avg_ms,
        target_rtt_avg_ms
    )
    SELECT
        previous.probe_id,
        previous.check_id,
        previous.started_at,
        previous.hop_index AS source_hop_index,
        NEW.hop_index AS target_hop_index,
        previous.address AS source_address,
        NEW.address AS target_address,
        previous.hostname AS source_hostname,
        NEW.hostname AS target_hostname,
        previous.loss_percent AS source_loss_percent,
        NEW.loss_percent AS target_loss_percent,
        previous.rtt_avg_ms AS source_rtt_avg_ms,
        NEW.rtt_avg_ms AS target_rtt_avg_ms
    FROM traceroute_result_hops AS previous
    WHERE previous.probe_id = NEW.probe_id
        AND previous.check_id = NEW.check_id
        AND previous.started_at = NEW.started_at
        AND previous.hop_index = NEW.hop_index - 1
    ON CONFLICT (probe_id, check_id, started_at, source_hop_index, target_hop_index) DO NOTHING;

    INSERT INTO traceroute_edge_observations (
        probe_id,
        check_id,
        started_at,
        source_hop_index,
        target_hop_index,
        source_address,
        target_address,
        source_hostname,
        target_hostname,
        source_loss_percent,
        target_loss_percent,
        source_rtt_avg_ms,
        target_rtt_avg_ms
    )
    SELECT
        NEW.probe_id,
        NEW.check_id,
        NEW.started_at,
        NEW.hop_index AS source_hop_index,
        next.hop_index AS target_hop_index,
        NEW.address AS source_address,
        next.address AS target_address,
        NEW.hostname AS source_hostname,
        next.hostname AS target_hostname,
        NEW.loss_percent AS source_loss_percent,
        next.loss_percent AS target_loss_percent,
        NEW.rtt_avg_ms AS source_rtt_avg_ms,
        next.rtt_avg_ms AS target_rtt_avg_ms
    FROM traceroute_result_hops AS next
    WHERE next.probe_id = NEW.probe_id
        AND next.check_id = NEW.check_id
        AND next.started_at = NEW.started_at
        AND next.hop_index = NEW.hop_index + 1
    ON CONFLICT (probe_id, check_id, started_at, source_hop_index, target_hop_index) DO NOTHING;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER record_traceroute_topology_observations
    AFTER INSERT ON traceroute_result_hops
    FOR EACH ROW
    EXECUTE FUNCTION record_traceroute_topology_observations();

CREATE MATERIALIZED VIEW traceroute_topology_nodes_1m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 minute', started_at) AS bucket,
    probe_id,
    check_id,
    hop_index,
    address,
    max(hostname) AS hostname,
    count(*)::bigint AS seen_count,
    sum(sent_count)::bigint AS sent_count,
    sum(received_count)::bigint AS received_count,
    sum(loss_percent)::double precision AS loss_sum_percent,
    count(loss_percent)::bigint AS loss_count,
    sum(rtt_avg_ms)::double precision AS rtt_avg_sum_ms,
    count(rtt_avg_ms)::bigint AS rtt_avg_count
FROM traceroute_hop_observations
GROUP BY bucket, probe_id, check_id, hop_index, address
WITH DATA;

CREATE MATERIALIZED VIEW traceroute_topology_nodes_10m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('10 minutes', started_at) AS bucket,
    probe_id,
    check_id,
    hop_index,
    address,
    max(hostname) AS hostname,
    count(*)::bigint AS seen_count,
    sum(sent_count)::bigint AS sent_count,
    sum(received_count)::bigint AS received_count,
    sum(loss_percent)::double precision AS loss_sum_percent,
    count(loss_percent)::bigint AS loss_count,
    sum(rtt_avg_ms)::double precision AS rtt_avg_sum_ms,
    count(rtt_avg_ms)::bigint AS rtt_avg_count
FROM traceroute_hop_observations
GROUP BY bucket, probe_id, check_id, hop_index, address
WITH DATA;

CREATE MATERIALIZED VIEW traceroute_topology_nodes_15m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('15 minutes', started_at) AS bucket,
    probe_id,
    check_id,
    hop_index,
    address,
    max(hostname) AS hostname,
    count(*)::bigint AS seen_count,
    sum(sent_count)::bigint AS sent_count,
    sum(received_count)::bigint AS received_count,
    sum(loss_percent)::double precision AS loss_sum_percent,
    count(loss_percent)::bigint AS loss_count,
    sum(rtt_avg_ms)::double precision AS rtt_avg_sum_ms,
    count(rtt_avg_ms)::bigint AS rtt_avg_count
FROM traceroute_hop_observations
GROUP BY bucket, probe_id, check_id, hop_index, address
WITH DATA;

CREATE MATERIALIZED VIEW traceroute_topology_nodes_30m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('30 minutes', started_at) AS bucket,
    probe_id,
    check_id,
    hop_index,
    address,
    max(hostname) AS hostname,
    count(*)::bigint AS seen_count,
    sum(sent_count)::bigint AS sent_count,
    sum(received_count)::bigint AS received_count,
    sum(loss_percent)::double precision AS loss_sum_percent,
    count(loss_percent)::bigint AS loss_count,
    sum(rtt_avg_ms)::double precision AS rtt_avg_sum_ms,
    count(rtt_avg_ms)::bigint AS rtt_avg_count
FROM traceroute_hop_observations
GROUP BY bucket, probe_id, check_id, hop_index, address
WITH DATA;

CREATE MATERIALIZED VIEW traceroute_topology_nodes_1h
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', started_at) AS bucket,
    probe_id,
    check_id,
    hop_index,
    address,
    max(hostname) AS hostname,
    count(*)::bigint AS seen_count,
    sum(sent_count)::bigint AS sent_count,
    sum(received_count)::bigint AS received_count,
    sum(loss_percent)::double precision AS loss_sum_percent,
    count(loss_percent)::bigint AS loss_count,
    sum(rtt_avg_ms)::double precision AS rtt_avg_sum_ms,
    count(rtt_avg_ms)::bigint AS rtt_avg_count
FROM traceroute_hop_observations
GROUP BY bucket, probe_id, check_id, hop_index, address
WITH DATA;

CREATE MATERIALIZED VIEW traceroute_topology_edges_1m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 minute', started_at) AS bucket,
    probe_id,
    check_id,
    source_hop_index,
    target_hop_index,
    source_address,
    target_address,
    max(source_hostname) AS source_hostname,
    max(target_hostname) AS target_hostname,
    count(*)::bigint AS seen_count,
    sum(source_loss_percent)::double precision AS source_loss_sum_percent,
    count(source_loss_percent)::bigint AS source_loss_count,
    sum(target_loss_percent)::double precision AS target_loss_sum_percent,
    count(target_loss_percent)::bigint AS target_loss_count,
    sum(source_rtt_avg_ms)::double precision AS source_rtt_avg_sum_ms,
    count(source_rtt_avg_ms)::bigint AS source_rtt_avg_count,
    sum(target_rtt_avg_ms)::double precision AS target_rtt_avg_sum_ms,
    count(target_rtt_avg_ms)::bigint AS target_rtt_avg_count
FROM traceroute_edge_observations
GROUP BY bucket, probe_id, check_id, source_hop_index, target_hop_index, source_address, target_address
WITH DATA;

CREATE MATERIALIZED VIEW traceroute_topology_edges_10m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('10 minutes', started_at) AS bucket,
    probe_id,
    check_id,
    source_hop_index,
    target_hop_index,
    source_address,
    target_address,
    max(source_hostname) AS source_hostname,
    max(target_hostname) AS target_hostname,
    count(*)::bigint AS seen_count,
    sum(source_loss_percent)::double precision AS source_loss_sum_percent,
    count(source_loss_percent)::bigint AS source_loss_count,
    sum(target_loss_percent)::double precision AS target_loss_sum_percent,
    count(target_loss_percent)::bigint AS target_loss_count,
    sum(source_rtt_avg_ms)::double precision AS source_rtt_avg_sum_ms,
    count(source_rtt_avg_ms)::bigint AS source_rtt_avg_count,
    sum(target_rtt_avg_ms)::double precision AS target_rtt_avg_sum_ms,
    count(target_rtt_avg_ms)::bigint AS target_rtt_avg_count
FROM traceroute_edge_observations
GROUP BY bucket, probe_id, check_id, source_hop_index, target_hop_index, source_address, target_address
WITH DATA;

CREATE MATERIALIZED VIEW traceroute_topology_edges_15m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('15 minutes', started_at) AS bucket,
    probe_id,
    check_id,
    source_hop_index,
    target_hop_index,
    source_address,
    target_address,
    max(source_hostname) AS source_hostname,
    max(target_hostname) AS target_hostname,
    count(*)::bigint AS seen_count,
    sum(source_loss_percent)::double precision AS source_loss_sum_percent,
    count(source_loss_percent)::bigint AS source_loss_count,
    sum(target_loss_percent)::double precision AS target_loss_sum_percent,
    count(target_loss_percent)::bigint AS target_loss_count,
    sum(source_rtt_avg_ms)::double precision AS source_rtt_avg_sum_ms,
    count(source_rtt_avg_ms)::bigint AS source_rtt_avg_count,
    sum(target_rtt_avg_ms)::double precision AS target_rtt_avg_sum_ms,
    count(target_rtt_avg_ms)::bigint AS target_rtt_avg_count
FROM traceroute_edge_observations
GROUP BY bucket, probe_id, check_id, source_hop_index, target_hop_index, source_address, target_address
WITH DATA;

CREATE MATERIALIZED VIEW traceroute_topology_edges_30m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('30 minutes', started_at) AS bucket,
    probe_id,
    check_id,
    source_hop_index,
    target_hop_index,
    source_address,
    target_address,
    max(source_hostname) AS source_hostname,
    max(target_hostname) AS target_hostname,
    count(*)::bigint AS seen_count,
    sum(source_loss_percent)::double precision AS source_loss_sum_percent,
    count(source_loss_percent)::bigint AS source_loss_count,
    sum(target_loss_percent)::double precision AS target_loss_sum_percent,
    count(target_loss_percent)::bigint AS target_loss_count,
    sum(source_rtt_avg_ms)::double precision AS source_rtt_avg_sum_ms,
    count(source_rtt_avg_ms)::bigint AS source_rtt_avg_count,
    sum(target_rtt_avg_ms)::double precision AS target_rtt_avg_sum_ms,
    count(target_rtt_avg_ms)::bigint AS target_rtt_avg_count
FROM traceroute_edge_observations
GROUP BY bucket, probe_id, check_id, source_hop_index, target_hop_index, source_address, target_address
WITH DATA;

CREATE MATERIALIZED VIEW traceroute_topology_edges_1h
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', started_at) AS bucket,
    probe_id,
    check_id,
    source_hop_index,
    target_hop_index,
    source_address,
    target_address,
    max(source_hostname) AS source_hostname,
    max(target_hostname) AS target_hostname,
    count(*)::bigint AS seen_count,
    sum(source_loss_percent)::double precision AS source_loss_sum_percent,
    count(source_loss_percent)::bigint AS source_loss_count,
    sum(target_loss_percent)::double precision AS target_loss_sum_percent,
    count(target_loss_percent)::bigint AS target_loss_count,
    sum(source_rtt_avg_ms)::double precision AS source_rtt_avg_sum_ms,
    count(source_rtt_avg_ms)::bigint AS source_rtt_avg_count,
    sum(target_rtt_avg_ms)::double precision AS target_rtt_avg_sum_ms,
    count(target_rtt_avg_ms)::bigint AS target_rtt_avg_count
FROM traceroute_edge_observations
GROUP BY bucket, probe_id, check_id, source_hop_index, target_hop_index, source_address, target_address
WITH DATA;


-- Refresh only the raw retention window; refreshing older windows after raw chunks expire would erase backfilled aggregates.
SELECT add_continuous_aggregate_policy('traceroute_topology_nodes_1m', start_offset => INTERVAL '3 days', end_offset => INTERVAL '2 minutes', schedule_interval => INTERVAL '1 minute', if_not_exists => TRUE);
SELECT add_continuous_aggregate_policy('traceroute_topology_nodes_10m', start_offset => INTERVAL '3 days', end_offset => INTERVAL '10 minutes', schedule_interval => INTERVAL '10 minutes', if_not_exists => TRUE);
SELECT add_continuous_aggregate_policy('traceroute_topology_nodes_15m', start_offset => INTERVAL '3 days', end_offset => INTERVAL '15 minutes', schedule_interval => INTERVAL '15 minutes', if_not_exists => TRUE);
SELECT add_continuous_aggregate_policy('traceroute_topology_nodes_30m', start_offset => INTERVAL '3 days', end_offset => INTERVAL '30 minutes', schedule_interval => INTERVAL '30 minutes', if_not_exists => TRUE);
SELECT add_continuous_aggregate_policy('traceroute_topology_nodes_1h', start_offset => INTERVAL '3 days', end_offset => INTERVAL '1 hour', schedule_interval => INTERVAL '1 hour', if_not_exists => TRUE);
SELECT add_continuous_aggregate_policy('traceroute_topology_edges_1m', start_offset => INTERVAL '3 days', end_offset => INTERVAL '2 minutes', schedule_interval => INTERVAL '1 minute', if_not_exists => TRUE);
SELECT add_continuous_aggregate_policy('traceroute_topology_edges_10m', start_offset => INTERVAL '3 days', end_offset => INTERVAL '10 minutes', schedule_interval => INTERVAL '10 minutes', if_not_exists => TRUE);
SELECT add_continuous_aggregate_policy('traceroute_topology_edges_15m', start_offset => INTERVAL '3 days', end_offset => INTERVAL '15 minutes', schedule_interval => INTERVAL '15 minutes', if_not_exists => TRUE);
SELECT add_continuous_aggregate_policy('traceroute_topology_edges_30m', start_offset => INTERVAL '3 days', end_offset => INTERVAL '30 minutes', schedule_interval => INTERVAL '30 minutes', if_not_exists => TRUE);
SELECT add_continuous_aggregate_policy('traceroute_topology_edges_1h', start_offset => INTERVAL '3 days', end_offset => INTERVAL '1 hour', schedule_interval => INTERVAL '1 hour', if_not_exists => TRUE);
SELECT add_retention_policy('traceroute_topology_nodes_1m', INTERVAL '30 days', if_not_exists => TRUE);
SELECT add_retention_policy('traceroute_topology_nodes_10m', INTERVAL '90 days', if_not_exists => TRUE);
SELECT add_retention_policy('traceroute_topology_nodes_15m', INTERVAL '90 days', if_not_exists => TRUE);
SELECT add_retention_policy('traceroute_topology_nodes_30m', INTERVAL '180 days', if_not_exists => TRUE);
SELECT add_retention_policy('traceroute_topology_edges_1m', INTERVAL '30 days', if_not_exists => TRUE);
SELECT add_retention_policy('traceroute_topology_edges_10m', INTERVAL '90 days', if_not_exists => TRUE);
SELECT add_retention_policy('traceroute_topology_edges_15m', INTERVAL '90 days', if_not_exists => TRUE);
SELECT add_retention_policy('traceroute_topology_edges_30m', INTERVAL '180 days', if_not_exists => TRUE);
SELECT add_retention_policy('traceroute_hop_observations', INTERVAL '3 days', if_not_exists => TRUE);
SELECT add_retention_policy('traceroute_edge_observations', INTERVAL '3 days', if_not_exists => TRUE);

-- +goose Down
SELECT remove_retention_policy('traceroute_hop_observations', if_exists => TRUE);
SELECT remove_retention_policy('traceroute_edge_observations', if_exists => TRUE);
SELECT remove_retention_policy('traceroute_topology_nodes_1m', if_exists => TRUE);
SELECT remove_retention_policy('traceroute_topology_nodes_10m', if_exists => TRUE);
SELECT remove_retention_policy('traceroute_topology_nodes_15m', if_exists => TRUE);
SELECT remove_retention_policy('traceroute_topology_nodes_30m', if_exists => TRUE);
SELECT remove_retention_policy('traceroute_topology_edges_1m', if_exists => TRUE);
SELECT remove_retention_policy('traceroute_topology_edges_10m', if_exists => TRUE);
SELECT remove_retention_policy('traceroute_topology_edges_15m', if_exists => TRUE);
SELECT remove_retention_policy('traceroute_topology_edges_30m', if_exists => TRUE);
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
DROP TRIGGER IF EXISTS record_traceroute_topology_observations ON traceroute_result_hops;
DROP FUNCTION IF EXISTS record_traceroute_topology_observations();
DROP TABLE IF EXISTS traceroute_edge_observations;
DROP TABLE IF EXISTS traceroute_hop_observations;

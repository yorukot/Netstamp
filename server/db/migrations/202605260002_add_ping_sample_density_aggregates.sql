-- +goose NO TRANSACTION
-- +goose Up
CREATE TABLE ping_rtt_sample_observations (
    probe_id bigint NOT NULL,
    check_id bigint NOT NULL,
    started_at timestamptz NOT NULL,
    sample_index integer NOT NULL,
    rtt_sample_ms double precision NOT NULL,
    PRIMARY KEY (probe_id, check_id, started_at, sample_index),
    CONSTRAINT ping_rtt_sample_observations_sample_index_positive CHECK (sample_index > 0),
    CONSTRAINT ping_rtt_sample_observations_rtt_sample_ms_non_negative CHECK (rtt_sample_ms >= 0)
);

SELECT create_hypertable('ping_rtt_sample_observations', 'started_at', if_not_exists => TRUE);

CREATE INDEX idx_ping_rtt_sample_observations_probe_check_started_at
    ON ping_rtt_sample_observations (probe_id, check_id, started_at DESC);

INSERT INTO ping_rtt_sample_observations (
    probe_id,
    check_id,
    started_at,
    sample_index,
    rtt_sample_ms
)
SELECT
    ping_results.probe_id,
    ping_results.check_id,
    ping_results.started_at,
    sample.ordinality::integer AS sample_index,
    sample.value::double precision AS rtt_sample_ms
FROM ping_results
CROSS JOIN LATERAL unnest(ping_results.rtt_samples_ms) WITH ORDINALITY AS sample(value, ordinality)
ON CONFLICT (probe_id, check_id, started_at, sample_index) DO NOTHING;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION record_ping_rtt_sample_observations()
RETURNS trigger AS $$
BEGIN
    INSERT INTO ping_rtt_sample_observations (
        probe_id,
        check_id,
        started_at,
        sample_index,
        rtt_sample_ms
    )
    SELECT
        NEW.probe_id,
        NEW.check_id,
        NEW.started_at,
        sample.ordinality::integer,
        sample.value::double precision
    FROM unnest(NEW.rtt_samples_ms) WITH ORDINALITY AS sample(value, ordinality)
    ON CONFLICT (probe_id, check_id, started_at, sample_index) DO NOTHING;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER record_ping_rtt_sample_observations
    AFTER INSERT ON ping_results
    FOR EACH ROW
    EXECUTE FUNCTION record_ping_rtt_sample_observations();

CREATE MATERIALIZED VIEW ping_rtt_sample_density_1m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 minute', started_at) AS bucket,
    probe_id,
    check_id,
    (floor(rtt_sample_ms / 10.0) * 10.0)::double precision AS rtt_bucket_start_ms,
    ((floor(rtt_sample_ms / 10.0) + 1.0) * 10.0)::double precision AS rtt_bucket_end_ms,
    count(*)::bigint AS sample_count
FROM ping_rtt_sample_observations
GROUP BY bucket, probe_id, check_id, rtt_bucket_start_ms, rtt_bucket_end_ms
WITH DATA;

CREATE MATERIALIZED VIEW ping_rtt_sample_density_10m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('10 minutes', started_at) AS bucket,
    probe_id,
    check_id,
    (floor(rtt_sample_ms / 10.0) * 10.0)::double precision AS rtt_bucket_start_ms,
    ((floor(rtt_sample_ms / 10.0) + 1.0) * 10.0)::double precision AS rtt_bucket_end_ms,
    count(*)::bigint AS sample_count
FROM ping_rtt_sample_observations
GROUP BY bucket, probe_id, check_id, rtt_bucket_start_ms, rtt_bucket_end_ms
WITH DATA;

CREATE MATERIALIZED VIEW ping_rtt_sample_density_15m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('15 minutes', started_at) AS bucket,
    probe_id,
    check_id,
    (floor(rtt_sample_ms / 10.0) * 10.0)::double precision AS rtt_bucket_start_ms,
    ((floor(rtt_sample_ms / 10.0) + 1.0) * 10.0)::double precision AS rtt_bucket_end_ms,
    count(*)::bigint AS sample_count
FROM ping_rtt_sample_observations
GROUP BY bucket, probe_id, check_id, rtt_bucket_start_ms, rtt_bucket_end_ms
WITH DATA;

CREATE MATERIALIZED VIEW ping_rtt_sample_density_30m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('30 minutes', started_at) AS bucket,
    probe_id,
    check_id,
    (floor(rtt_sample_ms / 10.0) * 10.0)::double precision AS rtt_bucket_start_ms,
    ((floor(rtt_sample_ms / 10.0) + 1.0) * 10.0)::double precision AS rtt_bucket_end_ms,
    count(*)::bigint AS sample_count
FROM ping_rtt_sample_observations
GROUP BY bucket, probe_id, check_id, rtt_bucket_start_ms, rtt_bucket_end_ms
WITH DATA;

CREATE MATERIALIZED VIEW ping_rtt_sample_density_1h
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', started_at) AS bucket,
    probe_id,
    check_id,
    (floor(rtt_sample_ms / 10.0) * 10.0)::double precision AS rtt_bucket_start_ms,
    ((floor(rtt_sample_ms / 10.0) + 1.0) * 10.0)::double precision AS rtt_bucket_end_ms,
    count(*)::bigint AS sample_count
FROM ping_rtt_sample_observations
GROUP BY bucket, probe_id, check_id, rtt_bucket_start_ms, rtt_bucket_end_ms
WITH DATA;


-- Refresh only the raw retention window; refreshing older windows after raw chunks expire would erase backfilled aggregates.
SELECT add_continuous_aggregate_policy('ping_rtt_sample_density_1m', start_offset => INTERVAL '3 days', end_offset => INTERVAL '2 minutes', schedule_interval => INTERVAL '1 minute', if_not_exists => TRUE);
SELECT add_continuous_aggregate_policy('ping_rtt_sample_density_10m', start_offset => INTERVAL '3 days', end_offset => INTERVAL '10 minutes', schedule_interval => INTERVAL '10 minutes', if_not_exists => TRUE);
SELECT add_continuous_aggregate_policy('ping_rtt_sample_density_15m', start_offset => INTERVAL '3 days', end_offset => INTERVAL '15 minutes', schedule_interval => INTERVAL '15 minutes', if_not_exists => TRUE);
SELECT add_continuous_aggregate_policy('ping_rtt_sample_density_30m', start_offset => INTERVAL '3 days', end_offset => INTERVAL '30 minutes', schedule_interval => INTERVAL '30 minutes', if_not_exists => TRUE);
SELECT add_continuous_aggregate_policy('ping_rtt_sample_density_1h', start_offset => INTERVAL '3 days', end_offset => INTERVAL '1 hour', schedule_interval => INTERVAL '1 hour', if_not_exists => TRUE);
SELECT add_retention_policy('ping_rtt_sample_density_1m', INTERVAL '30 days', if_not_exists => TRUE);
SELECT add_retention_policy('ping_rtt_sample_density_10m', INTERVAL '90 days', if_not_exists => TRUE);
SELECT add_retention_policy('ping_rtt_sample_density_15m', INTERVAL '90 days', if_not_exists => TRUE);
SELECT add_retention_policy('ping_rtt_sample_density_30m', INTERVAL '180 days', if_not_exists => TRUE);
SELECT add_retention_policy('ping_rtt_sample_observations', INTERVAL '3 days', if_not_exists => TRUE);

-- +goose Down
SELECT remove_retention_policy('ping_rtt_sample_observations', if_exists => TRUE);
SELECT remove_retention_policy('ping_rtt_sample_density_1m', if_exists => TRUE);
SELECT remove_retention_policy('ping_rtt_sample_density_10m', if_exists => TRUE);
SELECT remove_retention_policy('ping_rtt_sample_density_15m', if_exists => TRUE);
SELECT remove_retention_policy('ping_rtt_sample_density_30m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('ping_rtt_sample_density_1m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('ping_rtt_sample_density_10m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('ping_rtt_sample_density_15m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('ping_rtt_sample_density_30m', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('ping_rtt_sample_density_1h', if_exists => TRUE);
DROP MATERIALIZED VIEW IF EXISTS ping_rtt_sample_density_1h;
DROP MATERIALIZED VIEW IF EXISTS ping_rtt_sample_density_30m;
DROP MATERIALIZED VIEW IF EXISTS ping_rtt_sample_density_15m;
DROP MATERIALIZED VIEW IF EXISTS ping_rtt_sample_density_10m;
DROP MATERIALIZED VIEW IF EXISTS ping_rtt_sample_density_1m;
DROP TRIGGER IF EXISTS record_ping_rtt_sample_observations ON ping_results;
DROP FUNCTION IF EXISTS record_ping_rtt_sample_observations();
DROP TABLE IF EXISTS ping_rtt_sample_observations;

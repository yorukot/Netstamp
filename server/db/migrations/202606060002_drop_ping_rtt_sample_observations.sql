-- +goose NO TRANSACTION
-- +goose Up
SELECT remove_retention_policy('ping_rtt_sample_observations', if_exists => TRUE);

DROP TRIGGER IF EXISTS record_ping_rtt_sample_observations ON ping_results;
DROP FUNCTION IF EXISTS record_ping_rtt_sample_observations();
DROP TABLE IF EXISTS ping_rtt_sample_observations;

-- +goose Down
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

SELECT create_hypertable(
    'ping_rtt_sample_observations',
    'started_at',
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

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

SELECT add_retention_policy('ping_rtt_sample_observations', INTERVAL '3 days', if_not_exists => TRUE);

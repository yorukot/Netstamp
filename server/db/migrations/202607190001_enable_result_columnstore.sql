-- +goose NO TRANSACTION
-- +goose Up

-- Columnstore support for continuous aggregates requires TimescaleDB 2.20.0.
-- Require the latest 2.20 patch release so fresh and upgraded installations use
-- the fixes shipped before Netstamp starts managing these policies.
-- +goose StatementBegin
DO $$
DECLARE
    installed_version text;
    version_parts text[];
BEGIN
    SELECT extversion
    INTO installed_version
    FROM pg_extension
    WHERE extname = 'timescaledb';

    version_parts := regexp_match(installed_version, '^([0-9]+)\.([0-9]+)\.([0-9]+)');

    IF version_parts IS NULL OR
       (version_parts[1]::integer, version_parts[2]::integer, version_parts[3]::integer) < (2, 20, 3) THEN
        RAISE EXCEPTION
            'Netstamp columnstore policies require TimescaleDB 2.20.3 or newer; installed version is %',
            coalesce(installed_version, 'unknown')
            USING HINT = 'Upgrade the TimescaleDB image, run ALTER EXTENSION timescaledb UPDATE, then retry the migration.';
    END IF;
END;
$$;
-- +goose StatementEnd

-- Keep recent, write-heavy raw chunks in rowstore. Closed one-day chunks become
-- eligible for columnstore conversion after one day and are retained for three.
ALTER TABLE ping_results SET (
    timescaledb.enable_columnstore = true,
    timescaledb.segmentby = 'probe_id, check_id',
    timescaledb.orderby = 'started_at DESC'
);

ALTER TABLE tcp_results SET (
    timescaledb.enable_columnstore = true,
    timescaledb.segmentby = 'probe_id, check_id',
    timescaledb.orderby = 'started_at DESC'
);

ALTER TABLE http_results SET (
    timescaledb.enable_columnstore = true,
    timescaledb.segmentby = 'probe_id, check_id',
    timescaledb.orderby = 'started_at DESC'
);

ALTER TABLE traceroute_results SET (
    timescaledb.enable_columnstore = true,
    timescaledb.segmentby = 'probe_id, check_id',
    timescaledb.orderby = 'started_at DESC'
);

ALTER TABLE traceroute_result_hops SET (
    timescaledb.enable_columnstore = true,
    timescaledb.segmentby = 'probe_id, check_id',
    timescaledb.orderby = 'started_at DESC, hop_index ASC'
);

-- Long-lived sampled and rollup data stays mutable beyond each aggregate's
-- three-day refresh window, then moves to columnstore after seven days.
ALTER TABLE traceroute_sampled_runs_1m SET (
    timescaledb.enable_columnstore = true,
    timescaledb.segmentby = 'probe_id, check_id',
    timescaledb.orderby = 'bucket DESC'
);

ALTER MATERIALIZED VIEW ping_result_rollups_1m SET (
    timescaledb.enable_columnstore = true,
    timescaledb.segmentby = 'probe_id, check_id',
    timescaledb.orderby = 'bucket DESC'
);

ALTER MATERIALIZED VIEW tcp_result_rollups_1m SET (
    timescaledb.enable_columnstore = true,
    timescaledb.segmentby = 'probe_id, check_id',
    timescaledb.orderby = 'bucket DESC'
);

ALTER MATERIALIZED VIEW http_result_rollups_1m SET (
    timescaledb.enable_columnstore = true,
    timescaledb.segmentby = 'probe_id, check_id',
    timescaledb.orderby = 'bucket DESC'
);

-- Stagger the first runs so an existing installation does not rewrite every
-- eligible hypertable at once immediately after deployment.
CALL add_columnstore_policy(
    'ping_results',
    after => INTERVAL '1 day',
    schedule_interval => INTERVAL '6 hours',
    initial_start => now() + INTERVAL '10 minutes',
    if_not_exists => true
);

CALL add_columnstore_policy(
    'tcp_results',
    after => INTERVAL '1 day',
    schedule_interval => INTERVAL '6 hours',
    initial_start => now() + INTERVAL '20 minutes',
    if_not_exists => true
);

CALL add_columnstore_policy(
    'http_results',
    after => INTERVAL '1 day',
    schedule_interval => INTERVAL '6 hours',
    initial_start => now() + INTERVAL '30 minutes',
    if_not_exists => true
);

CALL add_columnstore_policy(
    'traceroute_results',
    after => INTERVAL '1 day',
    schedule_interval => INTERVAL '6 hours',
    initial_start => now() + INTERVAL '40 minutes',
    if_not_exists => true
);

CALL add_columnstore_policy(
    'traceroute_result_hops',
    after => INTERVAL '1 day',
    schedule_interval => INTERVAL '6 hours',
    initial_start => now() + INTERVAL '50 minutes',
    if_not_exists => true
);

CALL add_columnstore_policy(
    'traceroute_sampled_runs_1m',
    after => INTERVAL '7 days',
    schedule_interval => INTERVAL '12 hours',
    initial_start => now() + INTERVAL '60 minutes',
    if_not_exists => true
);

CALL add_columnstore_policy(
    'ping_result_rollups_1m',
    after => INTERVAL '7 days',
    schedule_interval => INTERVAL '12 hours',
    initial_start => now() + INTERVAL '70 minutes',
    if_not_exists => true
);

CALL add_columnstore_policy(
    'tcp_result_rollups_1m',
    after => INTERVAL '7 days',
    schedule_interval => INTERVAL '12 hours',
    initial_start => now() + INTERVAL '80 minutes',
    if_not_exists => true
);

CALL add_columnstore_policy(
    'http_result_rollups_1m',
    after => INTERVAL '7 days',
    schedule_interval => INTERVAL '12 hours',
    initial_start => now() + INTERVAL '90 minutes',
    if_not_exists => true
);

-- +goose Down

-- Stop future conversion without forcing a potentially large rowstore rewrite.
-- Existing columnstore chunks remain queryable; restore the pre-upgrade backup
-- when a complete physical rollback is required.
CALL remove_columnstore_policy('http_result_rollups_1m', if_exists => true);
CALL remove_columnstore_policy('tcp_result_rollups_1m', if_exists => true);
CALL remove_columnstore_policy('ping_result_rollups_1m', if_exists => true);
CALL remove_columnstore_policy('traceroute_sampled_runs_1m', if_exists => true);
CALL remove_columnstore_policy('traceroute_result_hops', if_exists => true);
CALL remove_columnstore_policy('traceroute_results', if_exists => true);
CALL remove_columnstore_policy('http_results', if_exists => true);
CALL remove_columnstore_policy('tcp_results', if_exists => true);
CALL remove_columnstore_policy('ping_results', if_exists => true);

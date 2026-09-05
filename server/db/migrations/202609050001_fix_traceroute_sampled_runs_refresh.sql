-- +goose Up

-- Redefine the traceroute sampled-run refresh procedure with two fixes:
--
-- 1. The hop join is bounded to the refresh window. Without a time predicate
--    on traceroute_result_hops, TimescaleDB cannot exclude chunks and the
--    planner hashes the entire hop hypertable (including decompressing
--    columnstore chunks) on every one-minute run, spilling gigabytes of temp
--    files each time.
-- 2. The upsert guard uses a strict comparison so unchanged rows are not
--    rewritten on every run.
--
-- Job 1063 (add_job in the v0.1.0 baseline) calls this procedure by name, so
-- the replacement takes effect on the next scheduled run.

-- +goose StatementBegin
CREATE OR REPLACE PROCEDURE refresh_traceroute_sampled_runs_1m(
    job_id integer DEFAULT NULL,
    config jsonb DEFAULT '{}'::jsonb
)
LANGUAGE plpgsql
AS $$
DECLARE
    refresh_lookback interval;
    refresh_lag interval;
    window_from timestamptz;
    window_to timestamptz;
BEGIN
    refresh_lookback := COALESCE((config ->> 'lookback')::interval, INTERVAL '10 minutes');
    refresh_lag := COALESCE((config ->> 'refresh_lag')::interval, INTERVAL '1 minute');
    window_to := COALESCE((config ->> 'to')::timestamptz, now() - refresh_lag);
    window_from := COALESCE((config ->> 'from')::timestamptz, window_to - refresh_lookback);

    IF window_from >= window_to THEN
        RAISE EXCEPTION 'refresh window start must be before end: from %, to %', window_from, window_to;
    END IF;

    INSERT INTO traceroute_sampled_runs_1m (
        bucket,
        probe_id,
        check_id,
        sampled_started_at,
        finished_at,
        duration_ms,
        status,
        resolved_ip,
        ip_family,
        destination_reached,
        hop_count,
        path_signature,
        hops
    )
    WITH candidate_runs AS (
        SELECT
            time_bucket(INTERVAL '1 minute', traceroute_results.started_at) AS bucket,
            traceroute_results.probe_id,
            traceroute_results.check_id,
            traceroute_results.started_at,
            traceroute_results.finished_at,
            traceroute_results.duration_ms,
            traceroute_results.status,
            traceroute_results.resolved_ip,
            traceroute_results.ip_family,
            traceroute_results.destination_reached,
            traceroute_results.hop_count
        FROM traceroute_results
        WHERE traceroute_results.started_at >= window_from
            AND traceroute_results.started_at < window_to
    ),
    sampled_runs AS (
        SELECT DISTINCT ON (candidate_runs.probe_id, candidate_runs.check_id, candidate_runs.bucket)
            candidate_runs.bucket,
            candidate_runs.probe_id,
            candidate_runs.check_id,
            candidate_runs.started_at,
            candidate_runs.finished_at,
            candidate_runs.duration_ms,
            candidate_runs.status,
            candidate_runs.resolved_ip,
            candidate_runs.ip_family,
            candidate_runs.destination_reached,
            candidate_runs.hop_count
        FROM candidate_runs
        ORDER BY
            candidate_runs.probe_id,
            candidate_runs.check_id,
            candidate_runs.bucket,
            candidate_runs.started_at DESC
    ),
    sampled_hops AS (
        SELECT
            sampled_runs.bucket,
            sampled_runs.probe_id,
            sampled_runs.check_id,
            sampled_runs.started_at,
            sampled_runs.finished_at,
            sampled_runs.duration_ms,
            sampled_runs.status,
            sampled_runs.resolved_ip,
            sampled_runs.ip_family,
            sampled_runs.destination_reached,
            sampled_runs.hop_count,
            string_agg(
                COALESCE(
                    traceroute_result_hops.address::text,
                    traceroute_result_hops.hostname,
                    traceroute_result_hops.error_code,
                    'unknown:' || traceroute_result_hops.hop_index::text
                ),
                '>' ORDER BY traceroute_result_hops.hop_index
            ) FILTER (WHERE traceroute_result_hops.hop_index IS NOT NULL) AS path_signature,
            COALESCE(
                jsonb_agg(
                    jsonb_build_object(
                        'hopIndex', traceroute_result_hops.hop_index,
                        'address', traceroute_result_hops.address::text,
                        'hostname', traceroute_result_hops.hostname,
                        'sentCount', traceroute_result_hops.sent_count,
                        'receivedCount', traceroute_result_hops.received_count,
                        'lossPercent', traceroute_result_hops.loss_percent,
                        'rttMinMs', traceroute_result_hops.rtt_min_ms,
                        'rttAvgMs', traceroute_result_hops.rtt_avg_ms,
                        'rttMedianMs', traceroute_result_hops.rtt_median_ms,
                        'rttMaxMs', traceroute_result_hops.rtt_max_ms,
                        'rttStddevMs', traceroute_result_hops.rtt_stddev_ms,
                        'rttSamplesMs', traceroute_result_hops.rtt_samples_ms,
                        'errorCode', traceroute_result_hops.error_code,
                        'errorMessage', traceroute_result_hops.error_message
                    )
                    ORDER BY traceroute_result_hops.hop_index
                ) FILTER (WHERE traceroute_result_hops.hop_index IS NOT NULL),
                '[]'::jsonb
            ) AS hops
        FROM sampled_runs
        LEFT JOIN traceroute_result_hops
            ON traceroute_result_hops.probe_id = sampled_runs.probe_id
            AND traceroute_result_hops.check_id = sampled_runs.check_id
            AND traceroute_result_hops.started_at = sampled_runs.started_at
            AND traceroute_result_hops.started_at >= window_from
            AND traceroute_result_hops.started_at < window_to
        GROUP BY
            sampled_runs.bucket,
            sampled_runs.probe_id,
            sampled_runs.check_id,
            sampled_runs.started_at,
            sampled_runs.finished_at,
            sampled_runs.duration_ms,
            sampled_runs.status,
            sampled_runs.resolved_ip,
            sampled_runs.ip_family,
            sampled_runs.destination_reached,
            sampled_runs.hop_count
    )
    SELECT
        sampled_hops.bucket,
        sampled_hops.probe_id,
        sampled_hops.check_id,
        sampled_hops.started_at,
        sampled_hops.finished_at,
        sampled_hops.duration_ms,
        sampled_hops.status,
        sampled_hops.resolved_ip,
        sampled_hops.ip_family,
        sampled_hops.destination_reached,
        sampled_hops.hop_count,
        sampled_hops.path_signature,
        sampled_hops.hops
    FROM sampled_hops
    ON CONFLICT (probe_id, check_id, bucket) DO UPDATE
    SET
        sampled_started_at = EXCLUDED.sampled_started_at,
        finished_at = EXCLUDED.finished_at,
        duration_ms = EXCLUDED.duration_ms,
        status = EXCLUDED.status,
        resolved_ip = EXCLUDED.resolved_ip,
        ip_family = EXCLUDED.ip_family,
        destination_reached = EXCLUDED.destination_reached,
        hop_count = EXCLUDED.hop_count,
        path_signature = EXCLUDED.path_signature,
        hops = EXCLUDED.hops
    WHERE traceroute_sampled_runs_1m.sampled_started_at < EXCLUDED.sampled_started_at;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    RAISE EXCEPTION 'traceroute sampled-run refresh fix is irreversible; the previous procedure body is retained only in the v0.1.0 baseline';
END;
$$;
-- +goose StatementEnd

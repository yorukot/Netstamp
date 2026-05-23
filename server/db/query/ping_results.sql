-- name: CreatePingResult :exec
INSERT INTO ping_results (
    probe_id,
    check_id,
    started_at,
    finished_at,
    duration_ms,
    status,
    sent_count,
    received_count,
    loss_percent,
    rtt_min_ms,
    rtt_avg_ms,
    rtt_median_ms,
    rtt_max_ms,
    rtt_stddev_ms,
    rtt_samples_ms,
    resolved_ip,
    ip_family,
    error_code,
    error_message
)
VALUES (
    sqlc.arg(probe_storage_id),
    sqlc.arg(check_storage_id),
    sqlc.arg(started_at),
    sqlc.arg(finished_at),
    sqlc.arg(duration_ms),
    sqlc.arg(status),
    sqlc.arg(sent_count),
    sqlc.arg(received_count),
    sqlc.arg(loss_percent),
    sqlc.narg(rtt_min_ms),
    sqlc.narg(rtt_avg_ms),
    sqlc.narg(rtt_median_ms),
    sqlc.narg(rtt_max_ms),
    sqlc.narg(rtt_stddev_ms),
    sqlc.arg(rtt_samples_ms)::double precision[],
    sqlc.narg(resolved_ip),
    sqlc.arg(ip_family),
    sqlc.narg(error_code),
    sqlc.narg(error_message)
)
ON CONFLICT (probe_id, check_id, started_at) DO NOTHING;

-- name: ResolvePingSeriesStorageIDs :one
SELECT probes.internal_id AS probe_storage_id,
       checks.internal_id AS check_storage_id
FROM probes
JOIN checks
    ON checks.project_id = probes.project_id
    AND checks.id = sqlc.arg(check_id)
    AND checks.deleted_at IS NULL
WHERE probes.project_id = sqlc.arg(project_id)
  AND probes.id = sqlc.arg(probe_id)
  AND probes.deleted_at IS NULL;

-- name: CountPingResultSeriesPoints :one
SELECT count(*)::bigint
FROM ping_results
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to)
    AND (sqlc.arg(metric)::text != 'rttAvgMs' OR rtt_avg_ms IS NOT NULL);

-- name: CountPingInsightPoints :one
SELECT count(*)::bigint
FROM ping_results
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to);

-- name: ListPingResultRawSeries :many
SELECT
    (extract(epoch FROM started_at) * 1000)::bigint AS bucket_ms,
    CASE sqlc.arg(metric)::text
        WHEN 'rttAvgMs' THEN rtt_avg_ms
        WHEN 'lossPercent' THEN loss_percent
        WHEN 'successRate' THEN CASE WHEN status = 'successful' THEN 100.0 ELSE 0.0 END
    END::double precision AS value
FROM ping_results
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to)
    AND (sqlc.arg(metric)::text != 'rttAvgMs' OR rtt_avg_ms IS NOT NULL)
ORDER BY started_at ASC;

-- name: ListPingResultBucketSeries :many
SELECT
    (extract(epoch FROM time_bucket(
        (ceil(extract(epoch FROM (sqlc.arg(started_at_to)::timestamptz - sqlc.arg(started_at_from)::timestamptz)) / sqlc.arg(max_data_points)::double precision)::bigint * interval '1 second'),
        started_at,
        sqlc.arg(started_at_from)::timestamptz
    )) * 1000)::bigint AS bucket_ms,
    CASE sqlc.arg(metric)::text
        WHEN 'rttAvgMs' THEN avg(rtt_avg_ms)
        WHEN 'lossPercent' THEN avg(loss_percent)
        WHEN 'successRate' THEN 100.0 * count(*) FILTER (WHERE status = 'successful') / NULLIF(count(*), 0)
    END::double precision AS value
FROM ping_results
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to)
    AND (sqlc.arg(metric)::text != 'rttAvgMs' OR rtt_avg_ms IS NOT NULL)
GROUP BY bucket_ms
ORDER BY bucket_ms ASC;

-- name: ListPingInsightRawRows :many
SELECT
    (extract(epoch FROM started_at) * 1000)::bigint AS bucket_ms,
    1::bigint AS result_count,
    duration_ms::double precision AS duration_avg_ms,
    rtt_min_ms,
    rtt_avg_ms,
    rtt_median_ms,
    rtt_max_ms,
    rtt_stddev_ms,
    loss_percent,
    CASE WHEN status = 'successful' THEN 100.0 ELSE 0.0 END::double precision AS success_rate,
    sent_count::bigint AS sent_count,
    received_count::bigint AS received_count,
    CASE WHEN status = 'timeout' THEN 1 ELSE 0 END::bigint AS timeout_count,
    CASE WHEN status = 'error' THEN 1 ELSE 0 END::bigint AS error_count
FROM ping_results
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to)
ORDER BY started_at ASC;

-- name: ListPingInsightBucketRows :many
SELECT
    (extract(epoch FROM time_bucket(
        (ceil(extract(epoch FROM (sqlc.arg(started_at_to)::timestamptz - sqlc.arg(started_at_from)::timestamptz)) / sqlc.arg(max_data_points)::double precision)::bigint * interval '1 second'),
        started_at,
        sqlc.arg(started_at_from)::timestamptz
    )) * 1000)::bigint AS bucket_ms,
    count(*)::bigint AS result_count,
    count(rtt_avg_ms)::bigint AS rtt_value_count,
    coalesce(avg(duration_ms), 0)::double precision AS duration_avg_ms,
    coalesce(min(rtt_min_ms), 0)::double precision AS rtt_min_ms,
    coalesce(avg(rtt_avg_ms), 0)::double precision AS rtt_avg_ms,
    coalesce((percentile_cont(0.5) WITHIN GROUP (ORDER BY rtt_median_ms) FILTER (WHERE rtt_median_ms IS NOT NULL)), 0)::double precision AS rtt_median_ms,
    coalesce(max(rtt_max_ms), 0)::double precision AS rtt_max_ms,
    coalesce(avg(rtt_stddev_ms), 0)::double precision AS rtt_stddev_ms,
    coalesce(avg(loss_percent), 0)::double precision AS loss_percent,
    coalesce((100.0 * count(*) FILTER (WHERE status = 'successful') / NULLIF(count(*), 0)), 0)::double precision AS success_rate,
    coalesce(sum(sent_count), 0)::bigint AS sent_count,
    coalesce(sum(received_count), 0)::bigint AS received_count,
    count(*) FILTER (WHERE status = 'timeout')::bigint AS timeout_count,
    count(*) FILTER (WHERE status = 'error')::bigint AS error_count
FROM ping_results
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to)
GROUP BY bucket_ms
ORDER BY bucket_ms ASC;

-- name: ListPingInsightRawSampleDensity :many
WITH samples AS (
    SELECT
        (extract(epoch FROM started_at) * 1000)::bigint AS bucket_ms,
        sample.value::double precision AS rtt_sample_ms
    FROM ping_results
    CROSS JOIN LATERAL unnest(rtt_samples_ms) AS sample(value)
    WHERE probe_id = sqlc.arg(probe_storage_id)
        AND check_id = sqlc.arg(check_storage_id)
        AND started_at >= sqlc.arg(started_at_from)
        AND started_at < sqlc.arg(started_at_to)
),
bounds AS (
    SELECT greatest(1.0, ceil(coalesce(max(rtt_sample_ms), 0) / 40.0))::double precision AS latency_bucket_ms
    FROM samples
)
SELECT
    bucket_ms,
    (floor(rtt_sample_ms / bounds.latency_bucket_ms) * bounds.latency_bucket_ms)::double precision AS rtt_bucket_start_ms,
    ((floor(rtt_sample_ms / bounds.latency_bucket_ms) + 1) * bounds.latency_bucket_ms)::double precision AS rtt_bucket_end_ms,
    count(*)::bigint AS sample_count
FROM samples
CROSS JOIN bounds
GROUP BY bucket_ms, rtt_bucket_start_ms, rtt_bucket_end_ms
ORDER BY bucket_ms ASC, rtt_bucket_start_ms ASC;

-- name: ListPingInsightBucketSampleDensity :many
WITH samples AS (
    SELECT
        (extract(epoch FROM time_bucket(
            (ceil(extract(epoch FROM (sqlc.arg(started_at_to)::timestamptz - sqlc.arg(started_at_from)::timestamptz)) / sqlc.arg(max_data_points)::double precision)::bigint * interval '1 second'),
            started_at,
            sqlc.arg(started_at_from)::timestamptz
        )) * 1000)::bigint AS bucket_ms,
        sample.value::double precision AS rtt_sample_ms
    FROM ping_results
    CROSS JOIN LATERAL unnest(rtt_samples_ms) AS sample(value)
    WHERE probe_id = sqlc.arg(probe_storage_id)
        AND check_id = sqlc.arg(check_storage_id)
        AND started_at >= sqlc.arg(started_at_from)
        AND started_at < sqlc.arg(started_at_to)
),
bounds AS (
    SELECT greatest(1.0, ceil(coalesce(max(rtt_sample_ms), 0) / 40.0))::double precision AS latency_bucket_ms
    FROM samples
)
SELECT
    bucket_ms,
    (floor(rtt_sample_ms / bounds.latency_bucket_ms) * bounds.latency_bucket_ms)::double precision AS rtt_bucket_start_ms,
    ((floor(rtt_sample_ms / bounds.latency_bucket_ms) + 1) * bounds.latency_bucket_ms)::double precision AS rtt_bucket_end_ms,
    count(*)::bigint AS sample_count
FROM samples
CROSS JOIN bounds
GROUP BY bucket_ms, rtt_bucket_start_ms, rtt_bucket_end_ms
ORDER BY bucket_ms ASC, rtt_bucket_start_ms ASC;

-- name: GetPingInsightSummary :one
WITH filtered AS (
    SELECT *
    FROM ping_results
    WHERE probe_id = sqlc.arg(probe_storage_id)
        AND check_id = sqlc.arg(check_storage_id)
        AND started_at >= sqlc.arg(started_at_from)
        AND started_at < sqlc.arg(started_at_to)
),
samples AS (
    SELECT sample.value::double precision AS rtt_sample_ms
    FROM filtered
    CROSS JOIN LATERAL unnest(rtt_samples_ms) AS sample(value)
),
latest AS (
    SELECT *
    FROM filtered
    ORDER BY started_at DESC
    LIMIT 1
)
SELECT
    count(*)::bigint AS total_results,
    count(rtt_avg_ms)::bigint AS rtt_value_count,
    (SELECT count(*)::bigint FROM samples) AS sample_count,
    count(*) FILTER (WHERE status = 'successful')::bigint AS successful_count,
    count(*) FILTER (WHERE status = 'timeout')::bigint AS timeout_count,
    count(*) FILTER (WHERE status = 'error')::bigint AS error_count,
    coalesce(sum(sent_count), 0)::bigint AS sent_count,
    coalesce(sum(received_count), 0)::bigint AS received_count,
    coalesce(avg(loss_percent), 0)::double precision AS avg_loss_percent,
    coalesce(avg(rtt_avg_ms), 0)::double precision AS avg_rtt_ms,
    coalesce((percentile_cont(0.5) WITHIN GROUP (ORDER BY rtt_median_ms) FILTER (WHERE rtt_median_ms IS NOT NULL)), 0)::double precision AS median_rtt_ms,
    coalesce(max(rtt_max_ms), 0)::double precision AS max_rtt_ms,
    coalesce((SELECT percentile_cont(0.95) WITHIN GROUP (ORDER BY rtt_sample_ms) FROM samples), 0)::double precision AS p95_rtt_ms,
    coalesce((SELECT percentile_cont(0.99) WITHIN GROUP (ORDER BY rtt_sample_ms) FROM samples), 0)::double precision AS p99_rtt_ms,
    coalesce((SELECT status::text FROM latest), '')::text AS latest_status,
    coalesce((SELECT (extract(epoch FROM started_at) * 1000)::bigint FROM latest), 0)::bigint AS latest_started_at_ms,
    (SELECT rtt_avg_ms FROM latest) AS latest_rtt_avg_ms,
    coalesce((SELECT loss_percent FROM latest), 0)::double precision AS latest_loss_percent,
    (SELECT resolved_ip FROM latest) AS latest_resolved_ip
FROM filtered;

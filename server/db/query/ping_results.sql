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
    AND started_at < sqlc.arg(started_at_to);

-- name: CountPingResultRollupSeriesPoints :one
SELECT coalesce(sum(result_count), 0)::bigint
FROM ping_result_rollups_1m
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND bucket >= sqlc.arg(started_at_from)::timestamptz
    AND bucket < sqlc.arg(started_at_to)::timestamptz;

-- name: ListPingLatencyAvgRawSeries :many
SELECT
    (extract(epoch FROM started_at) * 1000)::bigint AS bucket_ms,
    coalesce(rtt_avg_ms, 0)::double precision AS value
FROM ping_results
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to)
    AND rtt_avg_ms IS NOT NULL
ORDER BY started_at ASC;

-- name: ListPingLatencyMinRawSeries :many
SELECT
    (extract(epoch FROM started_at) * 1000)::bigint AS bucket_ms,
    coalesce(rtt_min_ms, 0)::double precision AS value
FROM ping_results
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to)
    AND rtt_min_ms IS NOT NULL
ORDER BY started_at ASC;

-- name: ListPingLatencyMaxRawSeries :many
SELECT
    (extract(epoch FROM started_at) * 1000)::bigint AS bucket_ms,
    coalesce(rtt_max_ms, 0)::double precision AS value
FROM ping_results
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to)
    AND rtt_max_ms IS NOT NULL
ORDER BY started_at ASC;

-- name: ListPingLossPercentRawSeries :many
SELECT
    (extract(epoch FROM started_at) * 1000)::bigint AS bucket_ms,
    loss_percent::double precision AS value
FROM ping_results
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to)
ORDER BY started_at ASC;

-- name: ListPingLatencyAvgBucketSeries :many
WITH settings AS (
    SELECT (
        greatest(
            1,
            ceil(extract(epoch FROM (sqlc.arg(started_at_to)::timestamptz - sqlc.arg(started_at_from)::timestamptz)) / sqlc.arg(max_data_points)::double precision)::bigint
        ) * interval '1 second'
    ) AS bucket_width
)
SELECT
    (extract(epoch FROM time_bucket(settings.bucket_width, started_at, sqlc.arg(started_at_from)::timestamptz)) * 1000)::bigint AS bucket_ms,
    coalesce(avg(rtt_avg_ms), 0)::double precision AS value
FROM ping_results
CROSS JOIN settings
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to)
GROUP BY bucket_ms
HAVING count(rtt_avg_ms) > 0
ORDER BY bucket_ms ASC;

-- name: ListPingLatencyMinBucketSeries :many
WITH settings AS (
    SELECT (
        greatest(
            1,
            ceil(extract(epoch FROM (sqlc.arg(started_at_to)::timestamptz - sqlc.arg(started_at_from)::timestamptz)) / sqlc.arg(max_data_points)::double precision)::bigint
        ) * interval '1 second'
    ) AS bucket_width
)
SELECT
    (extract(epoch FROM time_bucket(settings.bucket_width, started_at, sqlc.arg(started_at_from)::timestamptz)) * 1000)::bigint AS bucket_ms,
    coalesce(min(rtt_min_ms), 0)::double precision AS value
FROM ping_results
CROSS JOIN settings
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to)
GROUP BY bucket_ms
HAVING count(rtt_min_ms) > 0
ORDER BY bucket_ms ASC;

-- name: ListPingLatencyMaxBucketSeries :many
WITH settings AS (
    SELECT (
        greatest(
            1,
            ceil(extract(epoch FROM (sqlc.arg(started_at_to)::timestamptz - sqlc.arg(started_at_from)::timestamptz)) / sqlc.arg(max_data_points)::double precision)::bigint
        ) * interval '1 second'
    ) AS bucket_width
)
SELECT
    (extract(epoch FROM time_bucket(settings.bucket_width, started_at, sqlc.arg(started_at_from)::timestamptz)) * 1000)::bigint AS bucket_ms,
    coalesce(max(rtt_max_ms), 0)::double precision AS value
FROM ping_results
CROSS JOIN settings
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to)
GROUP BY bucket_ms
HAVING count(rtt_max_ms) > 0
ORDER BY bucket_ms ASC;

-- name: ListPingLossPercentBucketSeries :many
WITH settings AS (
    SELECT (
        greatest(
            1,
            ceil(extract(epoch FROM (sqlc.arg(started_at_to)::timestamptz - sqlc.arg(started_at_from)::timestamptz)) / sqlc.arg(max_data_points)::double precision)::bigint
        ) * interval '1 second'
    ) AS bucket_width
)
SELECT
    (extract(epoch FROM time_bucket(settings.bucket_width, started_at, sqlc.arg(started_at_from)::timestamptz)) * 1000)::bigint AS bucket_ms,
    coalesce(
        100.0 * (sum(sent_count) - sum(received_count)) / NULLIF(sum(sent_count), 0),
        avg(loss_percent),
        0
    )::double precision AS value
FROM ping_results
CROSS JOIN settings
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to)
GROUP BY bucket_ms
ORDER BY bucket_ms ASC;

-- name: ListPingLatencyAvgRollupSeries :many
WITH settings AS (
    SELECT (
        greatest(
            60,
            ceil(extract(epoch FROM (sqlc.arg(started_at_to)::timestamptz - sqlc.arg(started_at_from)::timestamptz)) / sqlc.arg(max_data_points)::double precision)::bigint
        ) * interval '1 second'
    ) AS bucket_width
),
bucketed AS (
    SELECT
        time_bucket(settings.bucket_width, bucket, sqlc.arg(started_at_from)::timestamptz) AS query_bucket,
        rtt_avg_sum_ms,
        rtt_avg_count
    FROM ping_result_rollups_1m
    CROSS JOIN settings
    WHERE probe_id = sqlc.arg(probe_storage_id)
        AND check_id = sqlc.arg(check_storage_id)
        AND bucket >= sqlc.arg(started_at_from)
        AND bucket < sqlc.arg(started_at_to)
)
SELECT
    (extract(epoch FROM query_bucket) * 1000)::bigint AS bucket_ms,
    coalesce(sum(rtt_avg_sum_ms) / NULLIF(sum(rtt_avg_count), 0), 0)::double precision AS value
FROM bucketed
GROUP BY query_bucket
HAVING sum(rtt_avg_count) > 0
ORDER BY query_bucket ASC;

-- name: ListPingLatencyMinRollupSeries :many
WITH settings AS (
    SELECT (
        greatest(
            60,
            ceil(extract(epoch FROM (sqlc.arg(started_at_to)::timestamptz - sqlc.arg(started_at_from)::timestamptz)) / sqlc.arg(max_data_points)::double precision)::bigint
        ) * interval '1 second'
    ) AS bucket_width
),
bucketed AS (
    SELECT
        time_bucket(settings.bucket_width, bucket, sqlc.arg(started_at_from)::timestamptz) AS query_bucket,
        rtt_avg_count,
        rtt_min_ms
    FROM ping_result_rollups_1m
    CROSS JOIN settings
    WHERE probe_id = sqlc.arg(probe_storage_id)
        AND check_id = sqlc.arg(check_storage_id)
        AND bucket >= sqlc.arg(started_at_from)
        AND bucket < sqlc.arg(started_at_to)
)
SELECT
    (extract(epoch FROM query_bucket) * 1000)::bigint AS bucket_ms,
    coalesce(min(rtt_min_ms), 0)::double precision AS value
FROM bucketed
GROUP BY query_bucket
HAVING sum(rtt_avg_count) > 0
ORDER BY query_bucket ASC;

-- name: ListPingLatencyMaxRollupSeries :many
WITH settings AS (
    SELECT (
        greatest(
            60,
            ceil(extract(epoch FROM (sqlc.arg(started_at_to)::timestamptz - sqlc.arg(started_at_from)::timestamptz)) / sqlc.arg(max_data_points)::double precision)::bigint
        ) * interval '1 second'
    ) AS bucket_width
),
bucketed AS (
    SELECT
        time_bucket(settings.bucket_width, bucket, sqlc.arg(started_at_from)::timestamptz) AS query_bucket,
        rtt_avg_count,
        rtt_max_ms
    FROM ping_result_rollups_1m
    CROSS JOIN settings
    WHERE probe_id = sqlc.arg(probe_storage_id)
        AND check_id = sqlc.arg(check_storage_id)
        AND bucket >= sqlc.arg(started_at_from)
        AND bucket < sqlc.arg(started_at_to)
)
SELECT
    (extract(epoch FROM query_bucket) * 1000)::bigint AS bucket_ms,
    coalesce(max(rtt_max_ms), 0)::double precision AS value
FROM bucketed
GROUP BY query_bucket
HAVING sum(rtt_avg_count) > 0
ORDER BY query_bucket ASC;

-- name: ListPingLossPercentRollupSeries :many
WITH settings AS (
    SELECT (
        greatest(
            60,
            ceil(extract(epoch FROM (sqlc.arg(started_at_to)::timestamptz - sqlc.arg(started_at_from)::timestamptz)) / sqlc.arg(max_data_points)::double precision)::bigint
        ) * interval '1 second'
    ) AS bucket_width
),
bucketed AS (
    SELECT
        time_bucket(settings.bucket_width, bucket, sqlc.arg(started_at_from)::timestamptz) AS query_bucket,
        sent_count,
        received_count
    FROM ping_result_rollups_1m
    CROSS JOIN settings
    WHERE probe_id = sqlc.arg(probe_storage_id)
        AND check_id = sqlc.arg(check_storage_id)
        AND bucket >= sqlc.arg(started_at_from)
        AND bucket < sqlc.arg(started_at_to)
)
SELECT
    (extract(epoch FROM query_bucket) * 1000)::bigint AS bucket_ms,
    coalesce(
        100.0 * (sum(sent_count) - sum(received_count)) / NULLIF(sum(sent_count), 0),
        0
    )::double precision AS value
FROM bucketed
GROUP BY query_bucket
ORDER BY query_bucket ASC;

-- name: GetPingInsightSummary :one
SELECT
    count(*)::bigint AS total_results,
    count(rtt_avg_ms)::bigint AS rtt_value_count,
    coalesce(sum(received_count), 0)::bigint AS samples,
    coalesce(avg(rtt_avg_ms), 0)::double precision AS average_rtt_ms,
    coalesce(max(rtt_max_ms), 0)::double precision AS max_rtt_ms,
    coalesce(
        100.0 * (sum(sent_count) - sum(received_count)) / NULLIF(sum(sent_count), 0),
        avg(loss_percent),
        0
    )::double precision AS loss_percent,
    coalesce(
        100.0 * count(*) FILTER (WHERE status = 'successful') / NULLIF(count(*), 0),
        0
    )::double precision AS success_rate
FROM ping_results
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to);

-- name: GetPingInsightRollupSummary :one
SELECT
    coalesce(sum(result_count), 0)::bigint AS total_results,
    coalesce(sum(rtt_avg_count), 0)::bigint AS rtt_value_count,
    coalesce(sum(received_count), 0)::bigint AS samples,
    coalesce(sum(rtt_avg_sum_ms) / NULLIF(sum(rtt_avg_count), 0), 0)::double precision AS average_rtt_ms,
    coalesce(max(rtt_max_ms), 0)::double precision AS max_rtt_ms,
    coalesce(
        100.0 * (sum(sent_count) - sum(received_count)) / NULLIF(sum(sent_count), 0),
        0
    )::double precision AS loss_percent,
    coalesce(
        100.0 * sum(successful_count) / NULLIF(sum(result_count), 0),
        0
    )::double precision AS success_rate
FROM ping_result_rollups_1m
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND bucket >= sqlc.arg(started_at_from)::timestamptz
    AND bucket < sqlc.arg(started_at_to)::timestamptz;

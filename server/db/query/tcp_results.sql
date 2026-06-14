-- name: CreateTCPResult :one
INSERT INTO tcp_results (
    probe_id,
    check_id,
    started_at,
    finished_at,
    duration_ms,
    status,
    connect_duration_ms,
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
    sqlc.narg(connect_duration_ms),
    sqlc.narg(resolved_ip),
    sqlc.arg(ip_family),
    sqlc.narg(error_code),
    sqlc.narg(error_message)
)
ON CONFLICT (probe_id, check_id, started_at) DO NOTHING
RETURNING true::boolean AS inserted;

-- name: ResolveTCPInsightStorageIDs :one
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

-- name: CountTCPResultSeriesPoints :one
SELECT count(*)::bigint
FROM tcp_results
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to);

-- name: ListTCPConnectAvgRawSeries :many
SELECT
    (extract(epoch FROM started_at) * 1000)::bigint AS bucket_ms,
    coalesce(connect_duration_ms, 0)::double precision AS value
FROM tcp_results
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to)
    AND connect_duration_ms IS NOT NULL
ORDER BY started_at ASC;

-- name: ListTCPConnectMinRawSeries :many
SELECT
    (extract(epoch FROM started_at) * 1000)::bigint AS bucket_ms,
    coalesce(connect_duration_ms, 0)::double precision AS value
FROM tcp_results
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to)
    AND connect_duration_ms IS NOT NULL
ORDER BY started_at ASC;

-- name: ListTCPConnectMaxRawSeries :many
SELECT
    (extract(epoch FROM started_at) * 1000)::bigint AS bucket_ms,
    coalesce(connect_duration_ms, 0)::double precision AS value
FROM tcp_results
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to)
    AND connect_duration_ms IS NOT NULL
ORDER BY started_at ASC;

-- name: ListTCPFailurePercentRawSeries :many
SELECT
    (extract(epoch FROM started_at) * 1000)::bigint AS bucket_ms,
    CASE WHEN status IN ('timeout', 'error') THEN 100.0 ELSE 0.0 END::double precision AS value
FROM tcp_results
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to)
ORDER BY started_at ASC;

-- name: ListTCPConnectAvgBucketSeries :many
WITH settings AS (
    SELECT greatest(
        ceil(extract(epoch FROM (sqlc.arg(started_at_to)::timestamptz - sqlc.arg(started_at_from)::timestamptz)) / sqlc.arg(max_data_points)::double precision)::bigint,
        1
    ) * interval '1 second' AS bucket_width
)
SELECT
    (extract(epoch FROM time_bucket(settings.bucket_width, started_at, sqlc.arg(started_at_from)::timestamptz)) * 1000)::bigint AS bucket_ms,
    coalesce(avg(connect_duration_ms), 0)::double precision AS value
FROM tcp_results, settings
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to)
    AND connect_duration_ms IS NOT NULL
GROUP BY bucket_ms
ORDER BY bucket_ms ASC;

-- name: ListTCPConnectMinBucketSeries :many
WITH settings AS (
    SELECT greatest(
        ceil(extract(epoch FROM (sqlc.arg(started_at_to)::timestamptz - sqlc.arg(started_at_from)::timestamptz)) / sqlc.arg(max_data_points)::double precision)::bigint,
        1
    ) * interval '1 second' AS bucket_width
)
SELECT
    (extract(epoch FROM time_bucket(settings.bucket_width, started_at, sqlc.arg(started_at_from)::timestamptz)) * 1000)::bigint AS bucket_ms,
    coalesce(min(connect_duration_ms), 0)::double precision AS value
FROM tcp_results, settings
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to)
    AND connect_duration_ms IS NOT NULL
GROUP BY bucket_ms
ORDER BY bucket_ms ASC;

-- name: ListTCPConnectMaxBucketSeries :many
WITH settings AS (
    SELECT greatest(
        ceil(extract(epoch FROM (sqlc.arg(started_at_to)::timestamptz - sqlc.arg(started_at_from)::timestamptz)) / sqlc.arg(max_data_points)::double precision)::bigint,
        1
    ) * interval '1 second' AS bucket_width
)
SELECT
    (extract(epoch FROM time_bucket(settings.bucket_width, started_at, sqlc.arg(started_at_from)::timestamptz)) * 1000)::bigint AS bucket_ms,
    coalesce(max(connect_duration_ms), 0)::double precision AS value
FROM tcp_results, settings
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to)
    AND connect_duration_ms IS NOT NULL
GROUP BY bucket_ms
ORDER BY bucket_ms ASC;

-- name: ListTCPFailurePercentBucketSeries :many
WITH settings AS (
    SELECT greatest(
        ceil(extract(epoch FROM (sqlc.arg(started_at_to)::timestamptz - sqlc.arg(started_at_from)::timestamptz)) / sqlc.arg(max_data_points)::double precision)::bigint,
        1
    ) * interval '1 second' AS bucket_width
)
SELECT
    (extract(epoch FROM time_bucket(settings.bucket_width, started_at, sqlc.arg(started_at_from)::timestamptz)) * 1000)::bigint AS bucket_ms,
    coalesce(
        100.0 * count(*) FILTER (WHERE status IN ('timeout', 'error')) / NULLIF(count(*), 0),
        0
    )::double precision AS value
FROM tcp_results, settings
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to)
GROUP BY bucket_ms
ORDER BY bucket_ms ASC;

-- name: ListTCPConnectAvgRollupSeries :many
WITH settings AS (
    SELECT greatest(
        ceil(extract(epoch FROM (sqlc.arg(started_at_to)::timestamptz - sqlc.arg(started_at_from)::timestamptz)) / sqlc.arg(max_data_points)::double precision)::bigint,
        60
    ) * interval '1 second' AS bucket_width
),
bucketed AS (
    SELECT
        time_bucket(settings.bucket_width, bucket, sqlc.arg(started_at_from)::timestamptz) AS query_bucket,
        connect_duration_sum_ms,
        connect_duration_count
    FROM tcp_result_rollups_1m, settings
    WHERE probe_id = sqlc.arg(probe_storage_id)
        AND check_id = sqlc.arg(check_storage_id)
        AND bucket >= sqlc.arg(started_at_from)
        AND bucket < sqlc.arg(started_at_to)
)
SELECT
    (extract(epoch FROM query_bucket) * 1000)::bigint AS bucket_ms,
    coalesce(sum(connect_duration_sum_ms) / NULLIF(sum(connect_duration_count), 0), 0)::double precision AS value
FROM bucketed
GROUP BY query_bucket
HAVING sum(connect_duration_count) > 0
ORDER BY query_bucket ASC;

-- name: ListTCPConnectMinRollupSeries :many
WITH settings AS (
    SELECT greatest(
        ceil(extract(epoch FROM (sqlc.arg(started_at_to)::timestamptz - sqlc.arg(started_at_from)::timestamptz)) / sqlc.arg(max_data_points)::double precision)::bigint,
        60
    ) * interval '1 second' AS bucket_width
),
bucketed AS (
    SELECT
        time_bucket(settings.bucket_width, bucket, sqlc.arg(started_at_from)::timestamptz) AS query_bucket,
        connect_duration_min_ms,
        connect_duration_count
    FROM tcp_result_rollups_1m, settings
    WHERE probe_id = sqlc.arg(probe_storage_id)
        AND check_id = sqlc.arg(check_storage_id)
        AND bucket >= sqlc.arg(started_at_from)
        AND bucket < sqlc.arg(started_at_to)
)
SELECT
    (extract(epoch FROM query_bucket) * 1000)::bigint AS bucket_ms,
    coalesce(min(connect_duration_min_ms), 0)::double precision AS value
FROM bucketed
GROUP BY query_bucket
HAVING sum(connect_duration_count) > 0
ORDER BY query_bucket ASC;

-- name: ListTCPConnectMaxRollupSeries :many
WITH settings AS (
    SELECT greatest(
        ceil(extract(epoch FROM (sqlc.arg(started_at_to)::timestamptz - sqlc.arg(started_at_from)::timestamptz)) / sqlc.arg(max_data_points)::double precision)::bigint,
        60
    ) * interval '1 second' AS bucket_width
),
bucketed AS (
    SELECT
        time_bucket(settings.bucket_width, bucket, sqlc.arg(started_at_from)::timestamptz) AS query_bucket,
        connect_duration_max_ms,
        connect_duration_count
    FROM tcp_result_rollups_1m, settings
    WHERE probe_id = sqlc.arg(probe_storage_id)
        AND check_id = sqlc.arg(check_storage_id)
        AND bucket >= sqlc.arg(started_at_from)
        AND bucket < sqlc.arg(started_at_to)
)
SELECT
    (extract(epoch FROM query_bucket) * 1000)::bigint AS bucket_ms,
    coalesce(max(connect_duration_max_ms), 0)::double precision AS value
FROM bucketed
GROUP BY query_bucket
HAVING sum(connect_duration_count) > 0
ORDER BY query_bucket ASC;

-- name: ListTCPFailurePercentRollupSeries :many
WITH settings AS (
    SELECT greatest(
        ceil(extract(epoch FROM (sqlc.arg(started_at_to)::timestamptz - sqlc.arg(started_at_from)::timestamptz)) / sqlc.arg(max_data_points)::double precision)::bigint,
        60
    ) * interval '1 second' AS bucket_width
),
bucketed AS (
    SELECT
        time_bucket(settings.bucket_width, bucket, sqlc.arg(started_at_from)::timestamptz) AS query_bucket,
        result_count,
        timeout_count,
        error_count
    FROM tcp_result_rollups_1m, settings
    WHERE probe_id = sqlc.arg(probe_storage_id)
        AND check_id = sqlc.arg(check_storage_id)
        AND bucket >= sqlc.arg(started_at_from)
        AND bucket < sqlc.arg(started_at_to)
)
SELECT
    (extract(epoch FROM query_bucket) * 1000)::bigint AS bucket_ms,
    coalesce(
        100.0 * (sum(timeout_count) + sum(error_count)) / NULLIF(sum(result_count), 0),
        0
    )::double precision AS value
FROM bucketed
GROUP BY query_bucket
ORDER BY query_bucket ASC;

-- name: GetTCPInsightSummary :one
SELECT
    count(*)::bigint AS total_results,
    count(connect_duration_ms)::bigint AS connect_value_count,
    count(*)::bigint AS samples,
    coalesce(avg(connect_duration_ms), 0)::double precision AS average_connect_ms,
    coalesce(max(connect_duration_ms), 0)::double precision AS max_connect_ms,
    coalesce(
        100.0 * count(*) FILTER (WHERE status IN ('timeout', 'error')) / NULLIF(count(*), 0),
        0
    )::double precision AS failure_percent,
    coalesce(
        100.0 * count(*) FILTER (WHERE status = 'successful') / NULLIF(count(*), 0),
        0
    )::double precision AS success_rate
FROM tcp_results
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to);

-- name: GetTCPInsightRollupSummary :one
SELECT
    coalesce(sum(result_count), 0)::bigint AS total_results,
    coalesce(sum(connect_duration_count), 0)::bigint AS connect_value_count,
    coalesce(sum(result_count), 0)::bigint AS samples,
    coalesce(sum(connect_duration_sum_ms) / NULLIF(sum(connect_duration_count), 0), 0)::double precision AS average_connect_ms,
    coalesce(max(connect_duration_max_ms), 0)::double precision AS max_connect_ms,
    coalesce(
        100.0 * (sum(timeout_count) + sum(error_count)) / NULLIF(sum(result_count), 0),
        0
    )::double precision AS failure_percent,
    coalesce(
        100.0 * sum(successful_count) / NULLIF(sum(result_count), 0),
        0
    )::double precision AS success_rate
FROM tcp_result_rollups_1m
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND bucket >= sqlc.arg(started_at_from)::timestamptz
    AND bucket < sqlc.arg(started_at_to)::timestamptz;

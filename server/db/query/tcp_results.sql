-- name: CreateTCPResult :exec
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
ON CONFLICT (probe_id, check_id, started_at) DO NOTHING;

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

-- name: CountTCPInsightPoints :one
SELECT count(*)::bigint
FROM tcp_results
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to);

-- name: ListTCPInsightRawRows :many
SELECT
    (extract(epoch FROM started_at) * 1000)::bigint AS bucket_ms,
    1::bigint AS result_count,
    duration_ms::double precision AS duration_avg_ms,
    connect_duration_ms AS connect_min_ms,
    connect_duration_ms AS connect_avg_ms,
    connect_duration_ms AS connect_median_ms,
    connect_duration_ms AS connect_max_ms,
    NULL::double precision AS connect_stddev_ms,
    CASE WHEN status = 'successful' THEN 100.0 ELSE 0.0 END::double precision AS success_rate,
    CASE WHEN status = 'timeout' THEN 1 ELSE 0 END::bigint AS timeout_count,
    CASE WHEN status = 'error' THEN 1 ELSE 0 END::bigint AS error_count
FROM tcp_results
WHERE probe_id = sqlc.arg(probe_storage_id)
    AND check_id = sqlc.arg(check_storage_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to)
ORDER BY started_at ASC;

-- name: ListTCPInsightBucketRows :many
WITH bucketed AS (
    SELECT
        (extract(epoch FROM time_bucket(
            (ceil(extract(epoch FROM (sqlc.arg(started_at_to)::timestamptz - sqlc.arg(started_at_from)::timestamptz)) / sqlc.arg(max_data_points)::double precision)::bigint * interval '1 second'),
            started_at,
            sqlc.arg(started_at_from)::timestamptz
        )) * 1000)::bigint AS bucket_ms,
        duration_ms,
        status,
        connect_duration_ms
    FROM tcp_results
    WHERE probe_id = sqlc.arg(probe_storage_id)
        AND check_id = sqlc.arg(check_storage_id)
        AND started_at >= sqlc.arg(started_at_from)
        AND started_at < sqlc.arg(started_at_to)
)
SELECT
    bucket_ms,
    count(*)::bigint AS result_count,
    count(connect_duration_ms)::bigint AS connect_value_count,
    coalesce(avg(duration_ms), 0)::double precision AS duration_avg_ms,
    coalesce(min(connect_duration_ms), 0)::double precision AS connect_min_ms,
    coalesce(avg(connect_duration_ms), 0)::double precision AS connect_avg_ms,
    coalesce(percentile_cont(0.5) WITHIN GROUP (ORDER BY connect_duration_ms) FILTER (WHERE connect_duration_ms IS NOT NULL), 0)::double precision AS connect_median_ms,
    coalesce(max(connect_duration_ms), 0)::double precision AS connect_max_ms,
    coalesce(stddev_pop(connect_duration_ms), 0)::double precision AS connect_stddev_ms,
    coalesce((100.0 * count(*) FILTER (WHERE status = 'successful') / NULLIF(count(*), 0)), 0)::double precision AS success_rate,
    count(*) FILTER (WHERE status = 'timeout')::bigint AS timeout_count,
    count(*) FILTER (WHERE status = 'error')::bigint AS error_count
FROM bucketed
GROUP BY bucket_ms
ORDER BY bucket_ms ASC;

-- name: GetTCPInsightSummary :one
WITH filtered AS (
    SELECT *
    FROM tcp_results
    WHERE probe_id = sqlc.arg(probe_storage_id)
        AND check_id = sqlc.arg(check_storage_id)
        AND started_at >= sqlc.arg(started_at_from)
        AND started_at < sqlc.arg(started_at_to)
),
latest AS (
    SELECT *
    FROM filtered
    ORDER BY started_at DESC
    LIMIT 1
)
SELECT
    count(*)::bigint AS total_results,
    count(connect_duration_ms)::bigint AS connect_value_count,
    count(*) FILTER (WHERE status = 'successful')::bigint AS successful_count,
    count(*) FILTER (WHERE status = 'timeout')::bigint AS timeout_count,
    count(*) FILTER (WHERE status = 'error')::bigint AS error_count,
    coalesce(avg(connect_duration_ms), 0)::double precision AS avg_connect_ms,
    coalesce(percentile_cont(0.5) WITHIN GROUP (ORDER BY connect_duration_ms) FILTER (WHERE connect_duration_ms IS NOT NULL), 0)::double precision AS median_connect_ms,
    coalesce(max(connect_duration_ms), 0)::double precision AS max_connect_ms,
    coalesce(percentile_cont(0.95) WITHIN GROUP (ORDER BY connect_duration_ms) FILTER (WHERE connect_duration_ms IS NOT NULL), 0)::double precision AS p95_connect_ms,
    coalesce(percentile_cont(0.99) WITHIN GROUP (ORDER BY connect_duration_ms) FILTER (WHERE connect_duration_ms IS NOT NULL), 0)::double precision AS p99_connect_ms,
    coalesce((SELECT status::text FROM latest), '')::text AS latest_status,
    coalesce((SELECT (extract(epoch FROM started_at) * 1000)::bigint FROM latest), 0)::bigint AS latest_started_at_ms,
    (SELECT connect_duration_ms FROM latest) AS latest_connect_ms,
    (SELECT resolved_ip FROM latest) AS latest_resolved_ip
FROM filtered;

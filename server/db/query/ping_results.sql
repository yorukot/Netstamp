-- name: CreatePingResult :exec
INSERT INTO ping_results (
    project_id,
    check_id,
    probe_id,
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
    raw,
    error_code,
    error_message
)
VALUES (
    sqlc.arg(project_id),
    sqlc.arg(check_id),
    sqlc.arg(probe_id),
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
    sqlc.arg(raw)::jsonb,
    sqlc.narg(error_code),
    sqlc.narg(error_message)
)
ON CONFLICT (project_id, probe_id, check_id, started_at) DO NOTHING;

-- name: CountPingResultSeriesPoints :one
SELECT count(*)::bigint
FROM ping_results
WHERE project_id = sqlc.arg(project_id)
    AND probe_id = sqlc.arg(probe_id)
    AND check_id = sqlc.arg(check_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to)
    AND (sqlc.arg(metric)::text != 'rttAvgMs' OR rtt_avg_ms IS NOT NULL);

-- name: ListPingResultRawSeries :many
SELECT
    (extract(epoch FROM started_at) * 1000)::bigint AS bucket_ms,
    CASE sqlc.arg(metric)::text
        WHEN 'rttAvgMs' THEN rtt_avg_ms
        WHEN 'lossPercent' THEN loss_percent
        WHEN 'successRate' THEN CASE WHEN status = 'successful' THEN 100.0 ELSE 0.0 END
    END::double precision AS value
FROM ping_results
WHERE project_id = sqlc.arg(project_id)
    AND probe_id = sqlc.arg(probe_id)
    AND check_id = sqlc.arg(check_id)
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
WHERE project_id = sqlc.arg(project_id)
    AND probe_id = sqlc.arg(probe_id)
    AND check_id = sqlc.arg(check_id)
    AND started_at >= sqlc.arg(started_at_from)
    AND started_at < sqlc.arg(started_at_to)
    AND (sqlc.arg(metric)::text != 'rttAvgMs' OR rtt_avg_ms IS NOT NULL)
GROUP BY bucket_ms
ORDER BY bucket_ms ASC;

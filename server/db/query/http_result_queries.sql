-- name: ResolveHTTPInsightStorageIDs :one
SELECT probes.internal_id AS probe_storage_id, checks.internal_id AS check_storage_id
FROM probes
JOIN checks ON checks.project_id = probes.project_id
    AND checks.id = sqlc.arg(check_id) AND checks.deleted_at IS NULL
WHERE probes.project_id = sqlc.arg(project_id)
  AND probes.id = sqlc.arg(probe_id) AND probes.deleted_at IS NULL;

-- name: ListLatestHTTPResults :many
SELECT probes.id AS probe_id,
       checks.id AS check_id,
       latest.started_at,
       latest.finished_at,
       latest.duration_ms,
       latest.status,
       latest.dns_duration_ms,
       latest.connect_duration_ms,
       latest.tls_duration_ms,
       latest.ttfb_duration_ms,
       latest.resolved_ip,
       latest.ip_family,
       latest.status_code,
       latest.final_url,
       latest.redirect_count,
       latest.response_bytes,
       latest.response_truncated,
       latest.body_matched,
       latest.tls_version,
       latest.tls_cipher_suite,
       latest.certificate_not_before,
       latest.certificate_not_after,
       latest.error_code,
       latest.error_message
FROM probe_check_assignments
JOIN probes
    ON probes.project_id = probe_check_assignments.project_id
    AND probes.id = probe_check_assignments.probe_id
JOIN checks
    ON checks.project_id = probe_check_assignments.project_id
    AND checks.id = probe_check_assignments.check_id
    AND checks.check_type = 'http'
JOIN LATERAL (
    SELECT http_results.started_at,
           http_results.finished_at,
           http_results.duration_ms,
           http_results.status,
           http_results.dns_duration_ms,
           http_results.connect_duration_ms,
           http_results.tls_duration_ms,
           http_results.ttfb_duration_ms,
           http_results.resolved_ip,
           http_results.ip_family,
           http_results.status_code,
           http_results.final_url,
           http_results.redirect_count,
           http_results.response_bytes,
           http_results.response_truncated,
           http_results.body_matched,
           http_results.tls_version,
           http_results.tls_cipher_suite,
           http_results.certificate_not_before,
           http_results.certificate_not_after,
           http_results.error_code,
           http_results.error_message
    FROM http_results
    WHERE http_results.probe_id = probes.internal_id
      AND http_results.check_id = checks.internal_id
    ORDER BY http_results.started_at DESC
    LIMIT 1
) latest ON TRUE
WHERE probe_check_assignments.project_id = sqlc.arg(project_id)
  AND probe_check_assignments.deleted_at IS NULL
  AND probes.deleted_at IS NULL
  AND checks.deleted_at IS NULL
  AND (
      sqlc.narg(probe_id)::uuid IS NULL
      OR probes.id = sqlc.narg(probe_id)::uuid
  )
  AND (
      sqlc.narg(check_id)::uuid IS NULL
      OR checks.id = sqlc.narg(check_id)::uuid
  )
ORDER BY latest.started_at DESC, probes.id ASC, checks.id ASC;

-- name: CountHTTPResultSeriesPoints :one
SELECT count(*)::bigint FROM http_results
WHERE probe_id = sqlc.arg(probe_storage_id) AND check_id = sqlc.arg(check_storage_id)
  AND started_at >= sqlc.arg(started_at_from) AND started_at < sqlc.arg(started_at_to);

-- name: ListHTTPMetricRawSeries :many
SELECT (extract(epoch FROM started_at) * 1000)::bigint AS bucket_ms,
       CASE sqlc.arg(metric)::text
           WHEN 'dns_avg' THEN dns_duration_ms
           WHEN 'connect_avg' THEN connect_duration_ms
           WHEN 'tls_avg' THEN tls_duration_ms
           WHEN 'ttfb_avg' THEN ttfb_duration_ms
           WHEN 'total_avg' THEN duration_ms::double precision
           WHEN 'failure_percent' THEN CASE WHEN status IN ('timeout', 'error') THEN 100.0 ELSE 0.0 END
       END::double precision AS value
FROM http_results
WHERE probe_id = sqlc.arg(probe_storage_id) AND check_id = sqlc.arg(check_storage_id)
  AND started_at >= sqlc.arg(started_at_from) AND started_at < sqlc.arg(started_at_to)
  AND CASE sqlc.arg(metric)::text
      WHEN 'dns_avg' THEN dns_duration_ms IS NOT NULL
      WHEN 'connect_avg' THEN connect_duration_ms IS NOT NULL
      WHEN 'tls_avg' THEN tls_duration_ms IS NOT NULL
      WHEN 'ttfb_avg' THEN ttfb_duration_ms IS NOT NULL
      ELSE true END
ORDER BY started_at ASC;

-- name: ListHTTPMetricBucketSeries :many
WITH settings AS (
    SELECT greatest(ceil(extract(epoch FROM (sqlc.arg(started_at_to)::timestamptz - sqlc.arg(started_at_from)::timestamptz)) / sqlc.arg(max_data_points)::double precision)::bigint, 1) * interval '1 second' AS width
)
SELECT (extract(epoch FROM time_bucket(settings.width, started_at, sqlc.arg(started_at_from)::timestamptz)) * 1000)::bigint AS bucket_ms,
       CASE sqlc.arg(metric)::text
           WHEN 'dns_avg' THEN avg(dns_duration_ms)
           WHEN 'connect_avg' THEN avg(connect_duration_ms)
           WHEN 'tls_avg' THEN avg(tls_duration_ms)
           WHEN 'ttfb_avg' THEN avg(ttfb_duration_ms)
           WHEN 'total_avg' THEN avg(duration_ms)::double precision
           WHEN 'failure_percent' THEN 100.0 * count(*) FILTER (WHERE status IN ('timeout', 'error')) / NULLIF(count(*), 0)
       END::double precision AS value
FROM http_results, settings
WHERE probe_id = sqlc.arg(probe_storage_id) AND check_id = sqlc.arg(check_storage_id)
  AND started_at >= sqlc.arg(started_at_from) AND started_at < sqlc.arg(started_at_to)
GROUP BY bucket_ms
HAVING CASE sqlc.arg(metric)::text
    WHEN 'dns_avg' THEN count(dns_duration_ms) > 0
    WHEN 'connect_avg' THEN count(connect_duration_ms) > 0
    WHEN 'tls_avg' THEN count(tls_duration_ms) > 0
    WHEN 'ttfb_avg' THEN count(ttfb_duration_ms) > 0
    ELSE count(*) > 0 END
ORDER BY bucket_ms ASC;

-- name: ListHTTPMetricRollupSeries :many
WITH settings AS (
    SELECT greatest(ceil(extract(epoch FROM (sqlc.arg(started_at_to)::timestamptz - sqlc.arg(started_at_from)::timestamptz)) / sqlc.arg(max_data_points)::double precision)::bigint, 60) * interval '1 second' AS width
), bucketed AS (
    SELECT time_bucket(settings.width, bucket, sqlc.arg(started_at_from)::timestamptz) AS query_bucket, r.*
    FROM http_result_rollups_1m r, settings
    WHERE probe_id = sqlc.arg(probe_storage_id) AND check_id = sqlc.arg(check_storage_id)
      AND bucket >= sqlc.arg(started_at_from) AND bucket < sqlc.arg(started_at_to)
)
SELECT (extract(epoch FROM query_bucket) * 1000)::bigint AS bucket_ms,
       CASE sqlc.arg(metric)::text
           WHEN 'dns_avg' THEN sum(dns_duration_sum_ms) / NULLIF(sum(dns_duration_count), 0)
           WHEN 'connect_avg' THEN sum(connect_duration_sum_ms) / NULLIF(sum(connect_duration_count), 0)
           WHEN 'tls_avg' THEN sum(tls_duration_sum_ms) / NULLIF(sum(tls_duration_count), 0)
           WHEN 'ttfb_avg' THEN sum(ttfb_duration_sum_ms) / NULLIF(sum(ttfb_duration_count), 0)
           WHEN 'total_avg' THEN sum(duration_sum_ms) / NULLIF(sum(duration_count), 0)
           WHEN 'failure_percent' THEN 100.0 * sum(timeout_count + error_count) / NULLIF(sum(result_count), 0)
       END::double precision AS value
FROM bucketed GROUP BY query_bucket
HAVING CASE sqlc.arg(metric)::text
    WHEN 'dns_avg' THEN sum(dns_duration_count) > 0
    WHEN 'connect_avg' THEN sum(connect_duration_count) > 0
    WHEN 'tls_avg' THEN sum(tls_duration_count) > 0
    WHEN 'ttfb_avg' THEN sum(ttfb_duration_count) > 0
    ELSE sum(result_count) > 0 END
ORDER BY query_bucket ASC;

-- name: GetHTTPInsightSummary :one
SELECT count(*)::bigint AS total_results,
       coalesce(avg(duration_ms), 0)::double precision AS average_total_ms,
       coalesce(max(duration_ms), 0)::double precision AS max_total_ms,
       coalesce(avg(ttfb_duration_ms), 0)::double precision AS average_ttfb_ms,
       coalesce(max(ttfb_duration_ms), 0)::double precision AS max_ttfb_ms,
       coalesce((100.0 * count(*) FILTER (WHERE status IN ('timeout', 'error')) / NULLIF(count(*), 0)), 0)::double precision AS failure_percent,
       coalesce((100.0 * count(*) FILTER (WHERE status = 'successful') / NULLIF(count(*), 0)), 0)::double precision AS success_rate,
       coalesce(min(extract(epoch FROM (certificate_not_after - now())) / 86400.0), 0)::double precision AS certificate_days_remaining,
       count(ttfb_duration_ms)::bigint AS ttfb_count,
       count(certificate_not_after)::bigint AS certificate_count,
       count(*)::bigint AS samples
FROM http_results
WHERE probe_id = sqlc.arg(probe_storage_id) AND check_id = sqlc.arg(check_storage_id)
  AND started_at >= sqlc.arg(started_at_from) AND started_at < sqlc.arg(started_at_to);

-- name: GetHTTPInsightRollupSummary :one
SELECT coalesce(sum(result_count), 0)::bigint AS total_results,
       coalesce(sum(duration_sum_ms) / NULLIF(sum(duration_count), 0), 0)::double precision AS average_total_ms,
       coalesce(max(duration_max_ms), 0)::double precision AS max_total_ms,
       coalesce(sum(ttfb_duration_sum_ms) / NULLIF(sum(ttfb_duration_count), 0), 0)::double precision AS average_ttfb_ms,
       coalesce(max(ttfb_duration_max_ms), 0)::double precision AS max_ttfb_ms,
       coalesce(100.0 * sum(timeout_count + error_count) / NULLIF(sum(result_count), 0), 0)::double precision AS failure_percent,
       coalesce(100.0 * sum(successful_count) / NULLIF(sum(result_count), 0), 0)::double precision AS success_rate,
       coalesce(min(extract(epoch FROM (certificate_not_after_min - now())) / 86400.0), 0)::double precision AS certificate_days_remaining,
       coalesce(sum(ttfb_duration_count), 0)::bigint AS ttfb_count,
       count(certificate_not_after_min)::bigint AS certificate_count,
       coalesce(sum(result_count), 0)::bigint AS samples
FROM http_result_rollups_1m
WHERE probe_id = sqlc.arg(probe_storage_id) AND check_id = sqlc.arg(check_storage_id)
  AND bucket >= sqlc.arg(started_at_from)::timestamptz AND bucket < sqlc.arg(started_at_to)::timestamptz;

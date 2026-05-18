-- name: CreateTracerouteResult :exec
INSERT INTO traceroute_results (
    probe_id,
    check_id,
    started_at,
    finished_at,
    duration_ms,
    status,
    resolved_ip,
    ip_family,
    destination_reached,
    hop_count,
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
    sqlc.narg(resolved_ip),
    sqlc.arg(ip_family),
    sqlc.arg(destination_reached),
    sqlc.arg(hop_count),
    sqlc.narg(error_code),
    sqlc.narg(error_message)
)
ON CONFLICT (probe_id, check_id, started_at) DO NOTHING;

-- name: CreateTracerouteResultHop :exec
INSERT INTO traceroute_result_hops (
    probe_id,
    check_id,
    started_at,
    hop_index,
    address,
    hostname,
    sent_count,
    received_count,
    loss_percent,
    rtt_min_ms,
    rtt_avg_ms,
    rtt_median_ms,
    rtt_max_ms,
    rtt_stddev_ms,
    rtt_samples_ms,
    error_code,
    error_message
)
VALUES (
    sqlc.arg(probe_storage_id),
    sqlc.arg(check_storage_id),
    sqlc.arg(started_at),
    sqlc.arg(hop_index),
    sqlc.narg(address),
    sqlc.narg(hostname),
    sqlc.arg(sent_count),
    sqlc.arg(received_count),
    sqlc.arg(loss_percent),
    sqlc.narg(rtt_min_ms),
    sqlc.narg(rtt_avg_ms),
    sqlc.narg(rtt_median_ms),
    sqlc.narg(rtt_max_ms),
    sqlc.narg(rtt_stddev_ms),
    sqlc.arg(rtt_samples_ms)::double precision[],
    sqlc.narg(error_code),
    sqlc.narg(error_message)
)
ON CONFLICT (probe_id, check_id, started_at, hop_index) DO NOTHING;

-- name: ResolveTracerouteRunStorageIDs :one
SELECT probes.internal_id AS probe_storage_id,
       checks.internal_id AS check_storage_id
FROM probes
JOIN checks
    ON checks.project_id = probes.project_id
    AND checks.id = sqlc.arg(check_id)
    AND checks.check_type = 'traceroute'
    AND checks.deleted_at IS NULL
WHERE probes.project_id = sqlc.arg(project_id)
  AND probes.id = sqlc.arg(probe_id)
  AND probes.deleted_at IS NULL;

-- name: ListTracerouteRunRows :many
WITH selected_runs AS (
    SELECT *
    FROM traceroute_results
    WHERE traceroute_results.probe_id = sqlc.arg(probe_storage_id)
        AND traceroute_results.check_id = sqlc.arg(check_storage_id)
        AND traceroute_results.started_at >= sqlc.arg(started_at_from)
        AND traceroute_results.started_at < sqlc.arg(started_at_to)
        AND (
            sqlc.narg(cursor_started_at)::timestamptz IS NULL
            OR traceroute_results.started_at < sqlc.narg(cursor_started_at)::timestamptz
        )
    ORDER BY traceroute_results.started_at DESC
    LIMIT sqlc.arg(limit_count)
)
SELECT
    selected_runs.started_at,
    selected_runs.finished_at,
    selected_runs.duration_ms,
    selected_runs.status,
    selected_runs.resolved_ip,
    selected_runs.ip_family,
    selected_runs.destination_reached,
    selected_runs.hop_count,
    selected_runs.error_code,
    selected_runs.error_message,
    traceroute_result_hops.hop_index,
    traceroute_result_hops.address,
    traceroute_result_hops.hostname,
    traceroute_result_hops.sent_count,
    traceroute_result_hops.received_count,
    traceroute_result_hops.loss_percent,
    traceroute_result_hops.rtt_min_ms,
    traceroute_result_hops.rtt_avg_ms,
    traceroute_result_hops.rtt_median_ms,
    traceroute_result_hops.rtt_max_ms,
    traceroute_result_hops.rtt_stddev_ms,
    traceroute_result_hops.rtt_samples_ms,
    traceroute_result_hops.error_code AS hop_error_code,
    traceroute_result_hops.error_message AS hop_error_message
FROM selected_runs
LEFT JOIN traceroute_result_hops
    ON traceroute_result_hops.probe_id = selected_runs.probe_id
    AND traceroute_result_hops.check_id = selected_runs.check_id
    AND traceroute_result_hops.started_at = selected_runs.started_at
ORDER BY selected_runs.started_at DESC, traceroute_result_hops.hop_index ASC;

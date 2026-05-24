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

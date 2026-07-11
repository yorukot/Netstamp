-- name: CreateHTTPResult :one
INSERT INTO http_results (
    probe_id, check_id, started_at, finished_at, duration_ms, status,
    dns_duration_ms, connect_duration_ms, tls_duration_ms, ttfb_duration_ms,
    resolved_ip, ip_family, status_code, final_url, redirect_count,
    response_bytes, response_truncated, body_matched, tls_version,
    tls_cipher_suite, certificate_not_before, certificate_not_after,
    error_code, error_message
)
VALUES (
    sqlc.arg(probe_storage_id), sqlc.arg(check_storage_id), sqlc.arg(started_at),
    sqlc.arg(finished_at), sqlc.arg(duration_ms), sqlc.arg(status),
    sqlc.narg(dns_duration_ms), sqlc.narg(connect_duration_ms),
    sqlc.narg(tls_duration_ms), sqlc.narg(ttfb_duration_ms),
    sqlc.narg(resolved_ip), sqlc.narg(ip_family), sqlc.narg(status_code),
    sqlc.narg(final_url), sqlc.arg(redirect_count), sqlc.narg(response_bytes),
    sqlc.arg(response_truncated), sqlc.narg(body_matched), sqlc.narg(tls_version),
    sqlc.narg(tls_cipher_suite), sqlc.narg(certificate_not_before),
    sqlc.narg(certificate_not_after), sqlc.narg(error_code), sqlc.narg(error_message)
)
ON CONFLICT (probe_id, check_id, started_at) DO NOTHING
RETURNING true::boolean AS inserted;

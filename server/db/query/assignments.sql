-- name: ListProjectAssignments :many
SELECT probe_check_assignments.id AS assignment_id,
       probe_check_assignments.project_id,
       probe_check_assignments.probe_id,
       probe_check_assignments.check_id,
       probes.internal_id AS probe_internal_id,
       probes.name AS probe_name,
       probes.enabled AS probe_enabled,
       probes.location AS probe_location,
       probes.location_name AS probe_location_name,
       probes.created_at AS probe_created_at,
       probes.updated_at AS probe_updated_at,
       probes.deleted_at AS probe_deleted_at,
       checks.internal_id AS check_internal_id,
       probe_check_assignments.check_version,
       probe_check_assignments.selector_version,
       checks.name AS check_name,
       checks.check_type,
       checks.target,
       checks.selector,
       checks.description,
       checks.interval_seconds,
       checks.created_at AS check_created_at,
       checks.updated_at AS check_updated_at,
       checks.deleted_at AS check_deleted_at,
       ping_check_configs.packet_count AS ping_packet_count,
       ping_check_configs.packet_size_bytes AS ping_packet_size_bytes,
       ping_check_configs.timeout_ms AS ping_timeout_ms,
       ping_check_configs.ip_family AS ping_ip_family,
       tcp_check_configs.port AS tcp_port,
       tcp_check_configs.timeout_ms AS tcp_timeout_ms,
       tcp_check_configs.ip_family AS tcp_ip_family,
       traceroute_check_configs.protocol AS traceroute_protocol,
       traceroute_check_configs.max_hops AS traceroute_max_hops,
       traceroute_check_configs.timeout_ms AS traceroute_timeout_ms,
       traceroute_check_configs.queries_per_hop AS traceroute_queries_per_hop,
       traceroute_check_configs.packet_size_bytes AS traceroute_packet_size_bytes,
       traceroute_check_configs.port AS traceroute_port,
       traceroute_check_configs.ip_family AS traceroute_ip_family
FROM probe_check_assignments
JOIN probes
    ON probes.project_id = probe_check_assignments.project_id
    AND probes.id = probe_check_assignments.probe_id
JOIN checks
    ON checks.project_id = probe_check_assignments.project_id
    AND checks.id = probe_check_assignments.check_id
LEFT JOIN ping_check_configs ON ping_check_configs.check_id = checks.id
LEFT JOIN tcp_check_configs ON tcp_check_configs.check_id = checks.id
LEFT JOIN traceroute_check_configs ON traceroute_check_configs.check_id = checks.id
WHERE probe_check_assignments.project_id = sqlc.arg(project_id)
  AND (
      sqlc.narg(probe_id)::uuid IS NULL
      OR probe_check_assignments.probe_id = sqlc.narg(probe_id)::uuid
  )
  AND (
      sqlc.narg(check_id)::uuid IS NULL
      OR probe_check_assignments.check_id = sqlc.narg(check_id)::uuid
  )
  AND probe_check_assignments.deleted_at IS NULL
  AND probes.deleted_at IS NULL
  AND checks.deleted_at IS NULL
ORDER BY probes.created_at ASC,
         probes.id ASC,
         checks.created_at ASC,
         checks.id ASC;

-- name: EnqueueAssignmentRefreshJob :one
INSERT INTO assignment_refresh_jobs (
    project_id,
    target_type,
    target_id,
    dedupe_key,
    max_attempts
)
VALUES (
    sqlc.arg(project_id),
    sqlc.arg(target_type),
    sqlc.arg(target_id),
    sqlc.arg(dedupe_key),
    sqlc.arg(max_attempts)
)
ON CONFLICT (dedupe_key) DO UPDATE
SET project_id = EXCLUDED.project_id,
    target_type = EXCLUDED.target_type,
    target_id = EXCLUDED.target_id,
    status = CASE
        WHEN assignment_refresh_jobs.status = 'running' THEN assignment_refresh_jobs.status
        ELSE 'pending'
    END,
    max_attempts = EXCLUDED.max_attempts,
    next_attempt_at = CASE
        WHEN assignment_refresh_jobs.status = 'running' THEN assignment_refresh_jobs.next_attempt_at
        ELSE now()
    END,
    completed_at = CASE
        WHEN assignment_refresh_jobs.status = 'running' THEN assignment_refresh_jobs.completed_at
        ELSE NULL
    END,
    last_error_kind = CASE
        WHEN assignment_refresh_jobs.status = 'running' THEN assignment_refresh_jobs.last_error_kind
        ELSE NULL
    END,
    last_error_code = CASE
        WHEN assignment_refresh_jobs.status = 'running' THEN assignment_refresh_jobs.last_error_code
        ELSE NULL
    END,
    last_error = CASE
        WHEN assignment_refresh_jobs.status = 'running' THEN assignment_refresh_jobs.last_error
        ELSE NULL
    END
RETURNING id;

-- name: RecoverStaleAssignmentRefreshJobs :execrows
UPDATE assignment_refresh_jobs
SET status = 'pending',
    next_attempt_at = now()
WHERE status = 'running'
  AND last_attempt_at < sqlc.arg(stale_before);

-- name: ClaimAssignmentRefreshJobs :many
WITH selected AS (
    SELECT id
    FROM assignment_refresh_jobs
    WHERE status = 'pending'
      AND next_attempt_at <= now()
      AND attempt_count < max_attempts
    ORDER BY next_attempt_at ASC, created_at ASC
    LIMIT sqlc.arg(limit_count)
    FOR UPDATE SKIP LOCKED
)
UPDATE assignment_refresh_jobs
SET status = 'running',
    last_attempt_at = now()
FROM selected
WHERE assignment_refresh_jobs.id = selected.id
RETURNING assignment_refresh_jobs.id,
          assignment_refresh_jobs.project_id,
          assignment_refresh_jobs.target_type,
          assignment_refresh_jobs.target_id,
          assignment_refresh_jobs.status,
          assignment_refresh_jobs.attempt_count,
          assignment_refresh_jobs.max_attempts,
          assignment_refresh_jobs.next_attempt_at,
          assignment_refresh_jobs.last_attempt_at,
          assignment_refresh_jobs.completed_at,
          assignment_refresh_jobs.last_error_kind,
          assignment_refresh_jobs.last_error_code,
          assignment_refresh_jobs.last_error,
          assignment_refresh_jobs.dedupe_key,
          assignment_refresh_jobs.created_at,
          assignment_refresh_jobs.updated_at;

-- name: MarkAssignmentRefreshJobSucceeded :exec
UPDATE assignment_refresh_jobs
SET status = 'succeeded',
    completed_at = sqlc.arg(completed_at),
    last_error_kind = NULL,
    last_error_code = NULL,
    last_error = NULL
WHERE id = sqlc.arg(id);

-- name: MarkAssignmentRefreshJobRetry :exec
UPDATE assignment_refresh_jobs
SET status = 'pending',
    attempt_count = attempt_count + 1,
    next_attempt_at = sqlc.arg(next_attempt_at),
    last_error_kind = sqlc.arg(last_error_kind),
    last_error_code = sqlc.arg(last_error_code),
    last_error = sqlc.arg(last_error)
WHERE id = sqlc.arg(id);

-- name: MarkAssignmentRefreshJobFailed :exec
UPDATE assignment_refresh_jobs
SET status = 'failed',
    attempt_count = attempt_count + 1,
    completed_at = now(),
    last_error_kind = sqlc.arg(last_error_kind),
    last_error_code = sqlc.arg(last_error_code),
    last_error = sqlc.arg(last_error)
WHERE id = sqlc.arg(id);

-- name: MarkAssignmentRefreshJobDiscarded :exec
UPDATE assignment_refresh_jobs
SET status = 'discarded',
    attempt_count = attempt_count + 1,
    completed_at = now(),
    last_error_kind = sqlc.arg(last_error_kind),
    last_error_code = sqlc.arg(last_error_code),
    last_error = sqlc.arg(last_error)
WHERE id = sqlc.arg(id);

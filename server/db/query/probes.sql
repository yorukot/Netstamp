-- name: CreateProbe :one
INSERT INTO probes (project_id, name, enabled, location, city)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, project_id, name, enabled, location, city, created_at, updated_at, deleted_at;

-- name: CreateProbeCredential :one
INSERT INTO probe_credentials (probe_id, secret_hash)
VALUES ($1, $2)
RETURNING probe_id, secret_hash, created_at, last_rotated_at;

-- name: CreateProbeStatus :one
INSERT INTO probe_statuses (probe_id, status)
VALUES ($1, $2)
RETURNING probe_id, status, last_seen_at, agent_version, public_v4, public_v6, addrs, updated_at;

-- name: CreateProbeLabel :exec
INSERT INTO probe_labels (project_id, probe_id, label_id)
VALUES ($1, $2, $3);

-- name: GetActiveProbeCredential :one
SELECT probes.id,
       probes.project_id,
       probes.enabled,
       probe_credentials.secret_hash
FROM probes
JOIN probe_credentials ON probe_credentials.probe_id = probes.id
WHERE probes.id = $1
  AND probes.deleted_at IS NULL;

-- name: UpdateProbeStatus :one
UPDATE probe_statuses
SET status = $2,
    last_seen_at = now(),
    agent_version = $3,
    public_v4 = $4,
    public_v6 = $5,
    addrs = $6,
    updated_at = now()
WHERE probe_id = $1
RETURNING probe_id, status, last_seen_at, agent_version, public_v4, public_v6, addrs, updated_at;

-- name: ListActiveAssignmentsForProbe :many
SELECT effective_probe_checks.id AS assignment_id,
       effective_probe_checks.project_id,
       effective_probe_checks.probe_id,
       effective_probe_checks.check_id,
       effective_probe_checks.check_version,
       effective_probe_checks.selector_version,
       checks.check_type,
       checks.target,
       checks.interval_seconds,
       ping_check_configs.packet_count,
       ping_check_configs.packet_size_bytes,
       ping_check_configs.timeout_ms,
       ping_check_configs.ip_family
FROM effective_probe_checks
JOIN probes
    ON probes.project_id = effective_probe_checks.project_id
    AND probes.id = effective_probe_checks.probe_id
JOIN checks
    ON checks.project_id = effective_probe_checks.project_id
    AND checks.id = effective_probe_checks.check_id
JOIN ping_check_configs ON ping_check_configs.check_id = checks.id
WHERE effective_probe_checks.probe_id = $1
  AND effective_probe_checks.deleted_at IS NULL
  AND probes.enabled = true
  AND probes.deleted_at IS NULL
  AND checks.deleted_at IS NULL
ORDER BY checks.created_at ASC,
         checks.id ASC;

-- name: ListActiveAssignedCheckIDsForProbe :many
SELECT effective_probe_checks.check_id
FROM effective_probe_checks
JOIN probes
    ON probes.project_id = effective_probe_checks.project_id
    AND probes.id = effective_probe_checks.probe_id
JOIN checks
    ON checks.project_id = effective_probe_checks.project_id
    AND checks.id = effective_probe_checks.check_id
WHERE effective_probe_checks.probe_id = $1
  AND effective_probe_checks.deleted_at IS NULL
  AND probes.enabled = true
  AND probes.deleted_at IS NULL
  AND checks.deleted_at IS NULL
ORDER BY checks.id ASC;

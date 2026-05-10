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

-- name: ListActiveProbesForProject :many
SELECT probes.id,
       probes.project_id,
       probes.name,
       probes.enabled,
       probes.location,
       probes.city,
       probes.created_at,
       probes.updated_at,
       probes.deleted_at,
       probe_statuses.status AS status,
       probe_statuses.last_seen_at AS status_last_seen_at,
       probe_statuses.agent_version AS status_agent_version,
       probe_statuses.public_v4 AS status_public_v4,
       probe_statuses.public_v6 AS status_public_v6,
       probe_statuses.addrs AS status_addrs,
       probe_statuses.updated_at AS status_updated_at,
       labels.id AS label_id,
       labels.project_id AS label_project_id,
       labels.key AS label_key,
       labels.value AS label_value,
       labels.created_at AS label_created_at,
       labels.updated_at AS label_updated_at,
       labels.deleted_at AS label_deleted_at
FROM probes
LEFT JOIN probe_statuses ON probe_statuses.probe_id = probes.id
LEFT JOIN probe_labels
    ON probe_labels.project_id = probes.project_id
    AND probe_labels.probe_id = probes.id
LEFT JOIN labels
    ON labels.project_id = probe_labels.project_id
    AND labels.id = probe_labels.label_id
    AND labels.deleted_at IS NULL
WHERE probes.project_id = $1
  AND probes.deleted_at IS NULL
ORDER BY probes.created_at DESC,
         probes.id DESC,
         labels.key ASC NULLS LAST,
         labels.value ASC NULLS LAST,
         labels.id ASC NULLS LAST;

-- name: GetActiveProbeRowsForProject :many
SELECT probes.id,
       probes.project_id,
       probes.name,
       probes.enabled,
       probes.location,
       probes.city,
       probes.created_at,
       probes.updated_at,
       probes.deleted_at,
       probe_statuses.status AS status,
       probe_statuses.last_seen_at AS status_last_seen_at,
       probe_statuses.agent_version AS status_agent_version,
       probe_statuses.public_v4 AS status_public_v4,
       probe_statuses.public_v6 AS status_public_v6,
       probe_statuses.addrs AS status_addrs,
       probe_statuses.updated_at AS status_updated_at,
       labels.id AS label_id,
       labels.project_id AS label_project_id,
       labels.key AS label_key,
       labels.value AS label_value,
       labels.created_at AS label_created_at,
       labels.updated_at AS label_updated_at,
       labels.deleted_at AS label_deleted_at
FROM probes
LEFT JOIN probe_statuses ON probe_statuses.probe_id = probes.id
LEFT JOIN probe_labels
    ON probe_labels.project_id = probes.project_id
    AND probe_labels.probe_id = probes.id
LEFT JOIN labels
    ON labels.project_id = probe_labels.project_id
    AND labels.id = probe_labels.label_id
    AND labels.deleted_at IS NULL
WHERE probes.project_id = $1
  AND probes.id = $2
  AND probes.deleted_at IS NULL
ORDER BY labels.key ASC NULLS LAST,
         labels.value ASC NULLS LAST,
         labels.id ASC NULLS LAST;

-- name: UpdateProbe :one
UPDATE probes
SET name = $3,
    enabled = $4,
    location = $5,
    city = $6
WHERE project_id = $1
  AND id = $2
  AND deleted_at IS NULL
RETURNING id, project_id, name, enabled, location, city, created_at, updated_at, deleted_at;

-- name: SoftDeleteProbe :one
UPDATE probes
SET deleted_at = now()
WHERE project_id = $1
  AND id = $2
  AND deleted_at IS NULL
RETURNING id;

-- name: DeleteProbeLabels :exec
DELETE FROM probe_labels
WHERE project_id = $1
  AND probe_id = $2;

-- name: ListActiveLabelsForProbe :many
SELECT labels.id,
       labels.project_id,
       labels.key,
       labels.value,
       labels.created_at,
       labels.updated_at,
       labels.deleted_at
FROM probe_labels
JOIN labels
    ON labels.project_id = probe_labels.project_id
    AND labels.id = probe_labels.label_id
WHERE probe_labels.project_id = $1
  AND probe_labels.probe_id = $2
  AND labels.deleted_at IS NULL
ORDER BY labels.key ASC, labels.value ASC, labels.id ASC;

-- name: RotateProbeCredential :one
UPDATE probe_credentials
SET secret_hash = $3,
    last_rotated_at = now()
FROM probes
WHERE probe_credentials.probe_id = probes.id
  AND probes.project_id = $1
  AND probes.id = $2
  AND probes.deleted_at IS NULL
RETURNING probe_credentials.probe_id;

-- name: DeleteEffectiveProbeChecksForProbe :exec
DELETE FROM effective_probe_checks
WHERE project_id = $1
  AND probe_id = $2;

-- name: DeleteStaleEffectiveProbeChecksForProbe :exec
DELETE FROM effective_probe_checks
WHERE project_id = sqlc.arg(project_id)
  AND probe_id = sqlc.arg(probe_id)
  AND deleted_at IS NULL
  AND check_id <> ALL(sqlc.arg(check_ids)::uuid[]);

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

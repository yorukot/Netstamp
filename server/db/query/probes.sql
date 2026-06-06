-- name: CreateProbe :one
INSERT INTO probes (project_id, name, enabled, location, location_name)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, internal_id, project_id, name, enabled, location, location_name, created_at, updated_at, deleted_at;

-- name: CreateProbeCredential :one
INSERT INTO probe_credentials (probe_id, secret_hash)
VALUES ($1, $2)
RETURNING probe_id, secret_hash, created_at, last_rotated_at;

-- name: CreateProbeStatus :one
INSERT INTO probe_statuses (probe_id, status)
VALUES ($1, $2)
RETURNING probe_id, status, last_seen_at, agent_version, public_v4, public_v6, "as", addrs, updated_at;

-- name: CreateProbeLabel :exec
INSERT INTO probe_labels (project_id, probe_id, label_id)
VALUES ($1, $2, $3);

-- name: ListActiveProbesForProject :many
SELECT probes.internal_id,
       probes.id,
       probes.project_id,
       probes.name,
       probes.enabled,
       probes.location,
       probes.location_name,
       probes.created_at,
       probes.updated_at,
       probes.deleted_at,
       (CASE
           WHEN probe_statuses.last_seen_at IS NULL THEN 'offline'::probe_state
           WHEN probe_statuses.last_seen_at < now() - interval '35 seconds' THEN 'offline'::probe_state
           ELSE probe_statuses.status
       END)::probe_state AS status,
       probe_statuses.last_seen_at AS status_last_seen_at,
       probe_statuses.agent_version AS status_agent_version,
       probe_statuses.public_v4 AS status_public_v4,
       probe_statuses.public_v6 AS status_public_v6,
       probe_statuses."as" AS status_as,
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
SELECT probes.internal_id,
       probes.id,
       probes.project_id,
       probes.name,
       probes.enabled,
       probes.location,
       probes.location_name,
       probes.created_at,
       probes.updated_at,
       probes.deleted_at,
       (CASE
           WHEN probe_statuses.last_seen_at IS NULL THEN 'offline'::probe_state
           WHEN probe_statuses.last_seen_at < now() - interval '35 seconds' THEN 'offline'::probe_state
           ELSE probe_statuses.status
       END)::probe_state AS status,
       probe_statuses.last_seen_at AS status_last_seen_at,
       probe_statuses.agent_version AS status_agent_version,
       probe_statuses.public_v4 AS status_public_v4,
       probe_statuses.public_v6 AS status_public_v6,
       probe_statuses."as" AS status_as,
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
    location_name = $6
WHERE project_id = $1
  AND id = $2
  AND deleted_at IS NULL
RETURNING id, internal_id, project_id, name, enabled, location, location_name, created_at, updated_at, deleted_at;

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

-- name: ListProbeRefreshTargetsForLabel :many
SELECT probes.id, probes.enabled
FROM probe_labels
JOIN probes
    ON probes.project_id = probe_labels.project_id
    AND probes.id = probe_labels.probe_id
WHERE probe_labels.project_id = $1
  AND probe_labels.label_id = $2
  AND probes.deleted_at IS NULL
ORDER BY probes.id;

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

-- name: DeleteProbeCheckAssignmentsForProbe :exec
DELETE FROM probe_check_assignments
WHERE project_id = $1
  AND probe_id = $2;

-- name: DeleteStaleProbeCheckAssignmentsForProbe :exec
DELETE FROM probe_check_assignments
WHERE project_id = sqlc.arg(project_id)
  AND probe_id = sqlc.arg(probe_id)
  AND deleted_at IS NULL
  AND check_id <> ALL(sqlc.arg(check_ids)::uuid[]);

-- name: GetActiveProbeCredential :one
SELECT probes.id,
       probes.project_id,
       probes.internal_id AS probe_internal_id,
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
    addrs = $4,
    updated_at = now()
WHERE probe_id = $1
RETURNING probe_id, status, last_seen_at, agent_version, public_v4, public_v6, "as", addrs, updated_at;

-- name: UpdateProbeIPFamilyCapabilities :one
UPDATE probe_statuses
SET public_v4 = CASE WHEN sqlc.arg(update_v4)::boolean THEN sqlc.narg(public_v4)::inet ELSE public_v4 END,
    public_v6 = CASE WHEN sqlc.arg(update_v6)::boolean THEN sqlc.narg(public_v6)::inet ELSE public_v6 END,
    updated_at = now()
WHERE probe_id = sqlc.arg(probe_id)
RETURNING probe_id, status, last_seen_at, agent_version, public_v4, public_v6, "as", addrs, updated_at;

-- name: ListActiveAssignmentsForProbe :many
SELECT probe_check_assignments.id AS assignment_id,
       probe_check_assignments.project_id,
       probe_check_assignments.probe_id,
       probe_check_assignments.check_id,
       probes.internal_id AS probe_internal_id,
       checks.internal_id AS check_internal_id,
       probe_check_assignments.check_version,
       probe_check_assignments.selector_version,
       checks.check_type,
       checks.target,
       checks.interval_seconds,
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
WHERE probe_check_assignments.probe_id = $1
  AND probe_check_assignments.deleted_at IS NULL
  AND probes.enabled = true
  AND probes.deleted_at IS NULL
  AND checks.deleted_at IS NULL
ORDER BY checks.created_at ASC,
         checks.id ASC;

-- name: ListActiveAssignmentsForProbeChecks :many
SELECT probe_check_assignments.id AS assignment_id,
       probe_check_assignments.project_id,
       probe_check_assignments.probe_id,
       probe_check_assignments.check_id,
       probes.internal_id AS probe_internal_id,
       checks.internal_id AS check_internal_id,
       probe_check_assignments.check_version,
       probe_check_assignments.selector_version,
       checks.check_type,
       checks.target,
       checks.interval_seconds,
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
WHERE probe_check_assignments.probe_id = sqlc.arg(probe_id)
  AND probe_check_assignments.check_id = ANY(sqlc.arg(check_ids)::uuid[])
  AND probe_check_assignments.deleted_at IS NULL
  AND probes.enabled = true
  AND probes.deleted_at IS NULL
  AND checks.deleted_at IS NULL
ORDER BY checks.created_at ASC,
         checks.id ASC;

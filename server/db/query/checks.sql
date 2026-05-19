-- name: ListActiveChecksForProject :many
SELECT checks.internal_id,
       checks.id,
       checks.project_id,
       checks.name,
       checks.check_type,
       checks.target,
       checks.selector,
       checks.description,
       checks.interval_seconds,
       checks.created_at,
       checks.updated_at,
       checks.deleted_at,
       ping_check_configs.packet_count AS ping_packet_count,
       ping_check_configs.packet_size_bytes AS ping_packet_size_bytes,
       ping_check_configs.timeout_ms AS ping_timeout_ms,
       ping_check_configs.ip_family AS ping_ip_family,
       traceroute_check_configs.protocol AS traceroute_protocol,
       traceroute_check_configs.max_hops AS traceroute_max_hops,
       traceroute_check_configs.timeout_ms AS traceroute_timeout_ms,
       traceroute_check_configs.queries_per_hop AS traceroute_queries_per_hop,
       traceroute_check_configs.packet_size_bytes AS traceroute_packet_size_bytes,
       traceroute_check_configs.port AS traceroute_port,
       traceroute_check_configs.ip_family AS traceroute_ip_family
FROM checks
LEFT JOIN ping_check_configs ON ping_check_configs.check_id = checks.id
LEFT JOIN traceroute_check_configs ON traceroute_check_configs.check_id = checks.id
WHERE checks.project_id = $1
  AND checks.deleted_at IS NULL
ORDER BY checks.created_at DESC, checks.id DESC;

-- name: GetActiveCheckForProject :one
SELECT checks.internal_id,
       checks.id,
       checks.project_id,
       checks.name,
       checks.check_type,
       checks.target,
       checks.selector,
       checks.description,
       checks.interval_seconds,
       checks.created_at,
       checks.updated_at,
       checks.deleted_at,
       ping_check_configs.packet_count AS ping_packet_count,
       ping_check_configs.packet_size_bytes AS ping_packet_size_bytes,
       ping_check_configs.timeout_ms AS ping_timeout_ms,
       ping_check_configs.ip_family AS ping_ip_family,
       traceroute_check_configs.protocol AS traceroute_protocol,
       traceroute_check_configs.max_hops AS traceroute_max_hops,
       traceroute_check_configs.timeout_ms AS traceroute_timeout_ms,
       traceroute_check_configs.queries_per_hop AS traceroute_queries_per_hop,
       traceroute_check_configs.packet_size_bytes AS traceroute_packet_size_bytes,
       traceroute_check_configs.port AS traceroute_port,
       traceroute_check_configs.ip_family AS traceroute_ip_family
FROM checks
LEFT JOIN ping_check_configs ON ping_check_configs.check_id = checks.id
LEFT JOIN traceroute_check_configs ON traceroute_check_configs.check_id = checks.id
WHERE checks.project_id = $1
  AND checks.id = $2
  AND checks.deleted_at IS NULL;

-- name: CreateCheck :one
INSERT INTO checks (project_id, name, check_type, target, selector, description, interval_seconds)
VALUES (
    sqlc.arg(project_id),
    sqlc.arg(name),
    sqlc.arg(check_type),
    sqlc.arg(target),
    sqlc.arg(selector)::jsonb,
    sqlc.narg(description),
    sqlc.arg(interval_seconds)
)
RETURNING id, internal_id, project_id, name, check_type, target, selector, description, interval_seconds, created_at, updated_at, deleted_at;

-- name: UpdateCheck :one
UPDATE checks
SET name = sqlc.arg(name),
    check_type = sqlc.arg(check_type),
    target = sqlc.arg(target),
    selector = sqlc.arg(selector)::jsonb,
    description = sqlc.narg(description),
    interval_seconds = sqlc.arg(interval_seconds)
WHERE project_id = sqlc.arg(project_id)
  AND id = sqlc.arg(id)
  AND deleted_at IS NULL
RETURNING id, internal_id, project_id, name, check_type, target, selector, description, interval_seconds, created_at, updated_at, deleted_at;

-- name: CreatePingCheckConfig :one
INSERT INTO ping_check_configs (check_id, packet_count, packet_size_bytes, timeout_ms, ip_family)
VALUES ($1, $2, $3, $4, $5)
RETURNING check_id, packet_count, packet_size_bytes, timeout_ms, ip_family;

-- name: UpdatePingCheckConfig :one
UPDATE ping_check_configs
SET packet_count = $2,
    packet_size_bytes = $3,
    timeout_ms = $4,
    ip_family = $5
WHERE check_id = $1
RETURNING check_id, packet_count, packet_size_bytes, timeout_ms, ip_family;

-- name: CreateTracerouteCheckConfig :one
INSERT INTO traceroute_check_configs (check_id, protocol, max_hops, timeout_ms, queries_per_hop, packet_size_bytes, port, ip_family)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING check_id, protocol, max_hops, timeout_ms, queries_per_hop, packet_size_bytes, port, ip_family;

-- name: UpdateTracerouteCheckConfig :one
UPDATE traceroute_check_configs
SET protocol = $2,
    max_hops = $3,
    timeout_ms = $4,
    queries_per_hop = $5,
    packet_size_bytes = $6,
    port = $7,
    ip_family = $8
WHERE check_id = $1
RETURNING check_id, protocol, max_hops, timeout_ms, queries_per_hop, packet_size_bytes, port, ip_family;

-- name: CreateCheckLabel :exec
INSERT INTO check_labels (project_id, check_id, label_id)
VALUES ($1, $2, $3);

-- name: ListActiveEnabledProbeLabelsForProject :many
SELECT probes.id AS probe_id,
       probes.internal_id AS probe_internal_id,
       probes.project_id AS probe_project_id,
       probes.name AS probe_name,
       probes.enabled AS probe_enabled,
       labels.id AS label_id,
       labels.project_id AS label_project_id,
       labels.key AS label_key,
       labels.value AS label_value,
       labels.created_at AS label_created_at,
       labels.updated_at AS label_updated_at,
       labels.deleted_at AS label_deleted_at
FROM probes
LEFT JOIN probe_labels
    ON probe_labels.project_id = probes.project_id
    AND probe_labels.probe_id = probes.id
LEFT JOIN labels
    ON labels.project_id = probe_labels.project_id
    AND labels.id = probe_labels.label_id
    AND labels.deleted_at IS NULL
WHERE probes.project_id = $1
  AND probes.enabled = true
  AND probes.deleted_at IS NULL
ORDER BY probes.created_at ASC,
         probes.id ASC,
         labels.key ASC NULLS LAST,
         labels.value ASC NULLS LAST,
         labels.id ASC NULLS LAST;

-- name: CreateProbeCheckAssignment :exec
INSERT INTO probe_check_assignments (project_id, probe_id, check_id, check_version, selector_version)
VALUES ($1, $2, $3, $4, $5);

-- name: UpsertProbeCheckAssignment :exec
INSERT INTO probe_check_assignments (project_id, probe_id, check_id, check_version, selector_version)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (project_id, probe_id, check_id) WHERE deleted_at IS NULL
DO UPDATE SET check_version = EXCLUDED.check_version,
              selector_version = EXCLUDED.selector_version;

-- name: DeleteStaleProbeCheckAssignments :exec
DELETE FROM probe_check_assignments
WHERE project_id = sqlc.arg(project_id)
  AND check_id = sqlc.arg(check_id)
  AND deleted_at IS NULL
  AND (
      check_version <> sqlc.arg(check_version)
      OR selector_version <> sqlc.arg(selector_version)
      OR probe_id <> ALL(sqlc.arg(probe_ids)::uuid[])
  );

-- name: DeleteProbeCheckAssignmentsForCheck :exec
DELETE FROM probe_check_assignments
WHERE project_id = $1
  AND check_id = $2;

-- name: DeleteCheckLabels :exec
DELETE FROM check_labels
WHERE project_id = $1
  AND check_id = $2;

-- name: ListActiveLabelsForCheck :many
SELECT labels.id,
       labels.project_id,
       labels.key,
       labels.value,
       labels.created_at,
       labels.updated_at,
       labels.deleted_at
FROM check_labels
JOIN labels
    ON labels.project_id = check_labels.project_id
    AND labels.id = check_labels.label_id
WHERE check_labels.project_id = $1
  AND check_labels.check_id = $2
  AND labels.deleted_at IS NULL
ORDER BY labels.key ASC, labels.value ASC, labels.id ASC;

-- name: SoftDeleteCheck :one
UPDATE checks
SET deleted_at = now()
WHERE project_id = $1
  AND id = $2
  AND deleted_at IS NULL
RETURNING id;

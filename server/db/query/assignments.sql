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

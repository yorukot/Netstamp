-- name: ListLatestResults :many
WITH active_assignments AS (
    SELECT probes.id AS probe_id,
           checks.id AS check_id,
           probes.internal_id AS probe_storage_id,
           checks.internal_id AS check_storage_id,
           checks.check_type::text AS result_type
    FROM probe_check_assignments
    JOIN probes
        ON probes.project_id = probe_check_assignments.project_id
        AND probes.id = probe_check_assignments.probe_id
    JOIN checks
        ON checks.project_id = probe_check_assignments.project_id
        AND checks.id = probe_check_assignments.check_id
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
      AND (
          sqlc.narg(result_type)::text IS NULL
          OR checks.check_type::text = sqlc.narg(result_type)::text
      )
)
SELECT active_assignments.result_type,
       active_assignments.probe_id,
       active_assignments.check_id,
       latest.started_at AS latest_started_at,
       latest.status AS latest_status
FROM active_assignments
JOIN LATERAL (
    (
        SELECT ping_results.started_at,
               ping_results.status::text AS status
        FROM ping_results
        WHERE active_assignments.result_type = 'ping'
          AND ping_results.probe_id = active_assignments.probe_storage_id
          AND ping_results.check_id = active_assignments.check_storage_id
        ORDER BY ping_results.started_at DESC
        LIMIT 1
    )
    UNION ALL
    (
        SELECT http_results.started_at,
               http_results.status::text AS status
        FROM http_results
        WHERE active_assignments.result_type = 'http'
          AND http_results.probe_id = active_assignments.probe_storage_id
          AND http_results.check_id = active_assignments.check_storage_id
        ORDER BY http_results.started_at DESC
        LIMIT 1
    )
    UNION ALL
    (
        SELECT tcp_results.started_at,
               tcp_results.status::text AS status
        FROM tcp_results
        WHERE active_assignments.result_type = 'tcp'
          AND tcp_results.probe_id = active_assignments.probe_storage_id
          AND tcp_results.check_id = active_assignments.check_storage_id
        ORDER BY tcp_results.started_at DESC
        LIMIT 1
    )
    UNION ALL
    (
        SELECT traceroute_results.started_at,
               traceroute_results.status::text AS status
        FROM traceroute_results
        WHERE active_assignments.result_type = 'traceroute'
          AND traceroute_results.probe_id = active_assignments.probe_storage_id
          AND traceroute_results.check_id = active_assignments.check_storage_id
        ORDER BY traceroute_results.started_at DESC
        LIMIT 1
    )
) latest ON TRUE
ORDER BY latest.started_at DESC,
         active_assignments.result_type ASC,
         active_assignments.probe_id ASC,
         active_assignments.check_id ASC;

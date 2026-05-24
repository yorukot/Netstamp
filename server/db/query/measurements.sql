-- name: ListProjectMeasurements :many
WITH measurements AS (
    SELECT 'ping'::text AS measurement_type,
           ping_results.started_at,
           ping_results.finished_at,
           ping_results.duration_ms,
           ping_results.status::text AS status,
           probes.id AS probe_id,
           checks.id AS check_id,
           ping_results.rtt_avg_ms AS latency_ms,
           ping_results.loss_percent,
           format('%s packets transmitted, %s received', ping_results.sent_count, ping_results.received_count)::text AS metadata,
           ping_results.error_code,
           ping_results.error_message
    FROM ping_results
    JOIN probes ON probes.internal_id = ping_results.probe_id
    JOIN checks ON checks.internal_id = ping_results.check_id
    WHERE probes.project_id = sqlc.arg(project_id)
      AND checks.project_id = sqlc.arg(project_id)
      AND probes.deleted_at IS NULL
      AND checks.deleted_at IS NULL
      AND ping_results.started_at >= sqlc.arg(started_at_from)
      AND ping_results.started_at < sqlc.arg(started_at_to)
      AND (
          sqlc.narg(probe_id)::uuid IS NULL
          OR probes.id = sqlc.narg(probe_id)::uuid
      )
      AND (
          sqlc.narg(check_id)::uuid IS NULL
          OR checks.id = sqlc.narg(check_id)::uuid
      )

    UNION ALL

    SELECT 'tcp'::text AS measurement_type,
           tcp_results.started_at,
           tcp_results.finished_at,
           tcp_results.duration_ms,
           tcp_results.status::text AS status,
           probes.id AS probe_id,
           checks.id AS check_id,
           tcp_results.connect_duration_ms AS latency_ms,
           0::double precision AS loss_percent,
           CASE
               WHEN tcp_results.connect_duration_ms IS NULL THEN NULL::text
               ELSE format('connect duration: %sms', tcp_results.connect_duration_ms)::text
           END AS metadata,
           tcp_results.error_code,
           tcp_results.error_message
    FROM tcp_results
    JOIN probes ON probes.internal_id = tcp_results.probe_id
    JOIN checks ON checks.internal_id = tcp_results.check_id
    WHERE probes.project_id = sqlc.arg(project_id)
      AND checks.project_id = sqlc.arg(project_id)
      AND probes.deleted_at IS NULL
      AND checks.deleted_at IS NULL
      AND tcp_results.started_at >= sqlc.arg(started_at_from)
      AND tcp_results.started_at < sqlc.arg(started_at_to)
      AND (
          sqlc.narg(probe_id)::uuid IS NULL
          OR probes.id = sqlc.narg(probe_id)::uuid
      )
      AND (
          sqlc.narg(check_id)::uuid IS NULL
          OR checks.id = sqlc.narg(check_id)::uuid
      )

    UNION ALL

    SELECT 'traceroute'::text AS measurement_type,
           traceroute_results.started_at,
           traceroute_results.finished_at,
           traceroute_results.duration_ms,
           traceroute_results.status::text AS status,
           probes.id AS probe_id,
           checks.id AS check_id,
           NULL::double precision AS latency_ms,
           0::double precision AS loss_percent,
           format('%s hops, destination reached: %s', traceroute_results.hop_count, traceroute_results.destination_reached)::text AS metadata,
           traceroute_results.error_code,
           traceroute_results.error_message
    FROM traceroute_results
    JOIN probes ON probes.internal_id = traceroute_results.probe_id
    JOIN checks ON checks.internal_id = traceroute_results.check_id
    WHERE probes.project_id = sqlc.arg(project_id)
      AND checks.project_id = sqlc.arg(project_id)
      AND probes.deleted_at IS NULL
      AND checks.deleted_at IS NULL
      AND traceroute_results.started_at >= sqlc.arg(started_at_from)
      AND traceroute_results.started_at < sqlc.arg(started_at_to)
      AND (
          sqlc.narg(probe_id)::uuid IS NULL
          OR probes.id = sqlc.narg(probe_id)::uuid
      )
      AND (
          sqlc.narg(check_id)::uuid IS NULL
          OR checks.id = sqlc.narg(check_id)::uuid
      )
)
SELECT measurement_type,
       started_at,
       finished_at,
       duration_ms,
       status,
       probe_id,
       check_id,
       latency_ms,
       loss_percent,
       metadata,
       error_code,
       error_message
FROM measurements
WHERE (
    sqlc.narg(measurement_type)::text IS NULL
    OR measurement_type = sqlc.narg(measurement_type)::text
)
  AND (
    sqlc.narg(status)::text IS NULL
    OR status = sqlc.narg(status)::text
)
  AND (
    sqlc.narg(cursor_started_at)::timestamptz IS NULL
    OR started_at < sqlc.narg(cursor_started_at)::timestamptz
)
ORDER BY started_at DESC,
         measurement_type ASC,
         probe_id ASC,
         check_id ASC
LIMIT sqlc.arg(limit_count);

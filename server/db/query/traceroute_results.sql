-- name: CreateTracerouteResult :exec
INSERT INTO traceroute_results (
    probe_id,
    check_id,
    started_at,
    finished_at,
    duration_ms,
    status,
    resolved_ip,
    ip_family,
    destination_reached,
    hop_count,
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
    sqlc.narg(resolved_ip),
    sqlc.arg(ip_family),
    sqlc.arg(destination_reached),
    sqlc.arg(hop_count),
    sqlc.narg(error_code),
    sqlc.narg(error_message)
)
ON CONFLICT (probe_id, check_id, started_at) DO NOTHING;

-- name: CreateTracerouteResultHop :exec
INSERT INTO traceroute_result_hops (
    probe_id,
    check_id,
    started_at,
    hop_index,
    address,
    hostname,
    sent_count,
    received_count,
    loss_percent,
    rtt_min_ms,
    rtt_avg_ms,
    rtt_median_ms,
    rtt_max_ms,
    rtt_stddev_ms,
    rtt_samples_ms,
    error_code,
    error_message
)
VALUES (
    sqlc.arg(probe_storage_id),
    sqlc.arg(check_storage_id),
    sqlc.arg(started_at),
    sqlc.arg(hop_index),
    sqlc.narg(address),
    sqlc.narg(hostname),
    sqlc.arg(sent_count),
    sqlc.arg(received_count),
    sqlc.arg(loss_percent),
    sqlc.narg(rtt_min_ms),
    sqlc.narg(rtt_avg_ms),
    sqlc.narg(rtt_median_ms),
    sqlc.narg(rtt_max_ms),
    sqlc.narg(rtt_stddev_ms),
    sqlc.arg(rtt_samples_ms)::double precision[],
    sqlc.narg(error_code),
    sqlc.narg(error_message)
)
ON CONFLICT (probe_id, check_id, started_at, hop_index) DO NOTHING;

-- name: ResolveTracerouteRunStorageIDs :one
SELECT probes.internal_id AS probe_storage_id,
       checks.internal_id AS check_storage_id
FROM probes
JOIN checks
    ON checks.project_id = probes.project_id
    AND checks.id = sqlc.arg(check_id)
    AND checks.check_type = 'traceroute'
    AND checks.deleted_at IS NULL
WHERE probes.project_id = sqlc.arg(project_id)
  AND probes.id = sqlc.arg(probe_id)
  AND probes.deleted_at IS NULL;

-- name: CountTracerouteInsightPoints :one
SELECT count(*)::bigint
FROM traceroute_results
WHERE traceroute_results.probe_id = sqlc.arg(probe_storage_id)
    AND traceroute_results.check_id = sqlc.arg(check_storage_id)
    AND traceroute_results.started_at >= sqlc.arg(started_at_from)
    AND traceroute_results.started_at < sqlc.arg(started_at_to);

-- name: ListTracerouteInsightRawRows :many
WITH runs AS (
    SELECT *
    FROM traceroute_results
    WHERE traceroute_results.probe_id = sqlc.arg(probe_storage_id)
        AND traceroute_results.check_id = sqlc.arg(check_storage_id)
        AND traceroute_results.started_at >= sqlc.arg(started_at_from)
        AND traceroute_results.started_at < sqlc.arg(started_at_to)
),
run_points AS (
    SELECT
        runs.started_at,
        runs.finished_at,
        runs.destination_reached,
        final_hop.rtt_avg_ms AS final_rtt_avg_ms,
        final_hop.loss_percent AS final_loss_percent,
        signature.path_signature
    FROM runs
    LEFT JOIN LATERAL (
        SELECT
            traceroute_result_hops.rtt_avg_ms,
            traceroute_result_hops.loss_percent
        FROM traceroute_result_hops
        WHERE traceroute_result_hops.probe_id = runs.probe_id
            AND traceroute_result_hops.check_id = runs.check_id
            AND traceroute_result_hops.started_at = runs.started_at
            AND (
                traceroute_result_hops.received_count > 0
                OR traceroute_result_hops.rtt_avg_ms IS NOT NULL
            )
        ORDER BY traceroute_result_hops.hop_index DESC
        LIMIT 1
    ) final_hop ON TRUE
    LEFT JOIN LATERAL (
        SELECT string_agg(
            coalesce(
                traceroute_result_hops.address::text,
                traceroute_result_hops.hostname,
                traceroute_result_hops.error_code,
                'unknown:' || traceroute_result_hops.hop_index::text
            ),
            '>' ORDER BY traceroute_result_hops.hop_index
        ) AS path_signature
        FROM traceroute_result_hops
        WHERE traceroute_result_hops.probe_id = runs.probe_id
            AND traceroute_result_hops.check_id = runs.check_id
            AND traceroute_result_hops.started_at = runs.started_at
    ) signature ON TRUE
),
scored AS (
    SELECT
        run_points.*,
        lag(run_points.path_signature) OVER (ORDER BY run_points.started_at ASC) AS previous_path_signature
    FROM run_points
)
SELECT
    (extract(epoch FROM scored.started_at) * 1000)::bigint AS bucket_ms,
    (extract(epoch FROM scored.started_at) * 1000)::bigint AS bucket_from_ms,
    (extract(epoch FROM scored.finished_at) * 1000)::bigint AS bucket_to_ms,
    scored.started_at AS run_started_at,
    1::bigint AS result_count,
    CASE WHEN scored.final_rtt_avg_ms IS NULL THEN 0 ELSE 1 END::bigint AS final_rtt_value_count,
    coalesce(scored.final_rtt_avg_ms, 0)::double precision AS final_rtt_avg_ms,
    CASE WHEN scored.final_loss_percent IS NULL THEN 0 ELSE 1 END::bigint AS final_loss_value_count,
    coalesce(scored.final_loss_percent, 0)::double precision AS final_loss_percent,
    (NOT scored.destination_reached OR coalesce(scored.final_loss_percent, 0) > 0)::boolean AS has_loss,
    (
        scored.previous_path_signature IS NOT NULL
        AND scored.path_signature IS NOT NULL
        AND scored.path_signature <> scored.previous_path_signature
    )::boolean AS has_route_change,
    scored.destination_reached::boolean AS destination_reached
FROM scored
ORDER BY scored.started_at ASC;

-- name: ListTracerouteInsightBucketRows :many
WITH settings AS (
    SELECT (
        ceil(extract(epoch FROM (sqlc.arg(started_at_to)::timestamptz - sqlc.arg(started_at_from)::timestamptz)) / sqlc.arg(max_data_points)::double precision)::bigint * interval '1 second'
    ) AS bucket_width
),
runs AS (
    SELECT *
    FROM traceroute_results
    WHERE traceroute_results.probe_id = sqlc.arg(probe_storage_id)
        AND traceroute_results.check_id = sqlc.arg(check_storage_id)
        AND traceroute_results.started_at >= sqlc.arg(started_at_from)
        AND traceroute_results.started_at < sqlc.arg(started_at_to)
),
run_points AS (
    SELECT
        runs.started_at,
        runs.destination_reached,
        final_hop.rtt_avg_ms AS final_rtt_avg_ms,
        final_hop.loss_percent AS final_loss_percent,
        signature.path_signature
    FROM runs
    LEFT JOIN LATERAL (
        SELECT
            traceroute_result_hops.rtt_avg_ms,
            traceroute_result_hops.loss_percent
        FROM traceroute_result_hops
        WHERE traceroute_result_hops.probe_id = runs.probe_id
            AND traceroute_result_hops.check_id = runs.check_id
            AND traceroute_result_hops.started_at = runs.started_at
            AND (
                traceroute_result_hops.received_count > 0
                OR traceroute_result_hops.rtt_avg_ms IS NOT NULL
            )
        ORDER BY traceroute_result_hops.hop_index DESC
        LIMIT 1
    ) final_hop ON TRUE
    LEFT JOIN LATERAL (
        SELECT string_agg(
            coalesce(
                traceroute_result_hops.address::text,
                traceroute_result_hops.hostname,
                traceroute_result_hops.error_code,
                'unknown:' || traceroute_result_hops.hop_index::text
            ),
            '>' ORDER BY traceroute_result_hops.hop_index
        ) AS path_signature
        FROM traceroute_result_hops
        WHERE traceroute_result_hops.probe_id = runs.probe_id
            AND traceroute_result_hops.check_id = runs.check_id
            AND traceroute_result_hops.started_at = runs.started_at
    ) signature ON TRUE
),
scored AS (
    SELECT
        run_points.*,
        lag(run_points.path_signature) OVER (ORDER BY run_points.started_at ASC) AS previous_path_signature
    FROM run_points
),
bucketed AS (
    SELECT
        time_bucket(settings.bucket_width, scored.started_at, sqlc.arg(started_at_from)::timestamptz) AS bucket,
        settings.bucket_width,
        scored.destination_reached,
        scored.final_rtt_avg_ms,
        scored.final_loss_percent,
        (NOT scored.destination_reached OR coalesce(scored.final_loss_percent, 0) > 0)::boolean AS has_loss,
        (
            scored.previous_path_signature IS NOT NULL
            AND scored.path_signature IS NOT NULL
            AND scored.path_signature <> scored.previous_path_signature
        )::boolean AS has_route_change
    FROM scored
    CROSS JOIN settings
)
SELECT
    (extract(epoch FROM bucketed.bucket) * 1000)::bigint AS bucket_ms,
    (extract(epoch FROM bucketed.bucket) * 1000)::bigint AS bucket_from_ms,
    (extract(epoch FROM (bucketed.bucket + bucketed.bucket_width)) * 1000)::bigint AS bucket_to_ms,
    count(*)::bigint AS result_count,
    count(bucketed.final_rtt_avg_ms)::bigint AS final_rtt_value_count,
    coalesce(avg(bucketed.final_rtt_avg_ms), 0)::double precision AS final_rtt_avg_ms,
    count(bucketed.final_loss_percent)::bigint AS final_loss_value_count,
    coalesce(avg(bucketed.final_loss_percent), 0)::double precision AS final_loss_percent,
    bool_or(bucketed.has_loss)::boolean AS has_loss,
    bool_or(bucketed.has_route_change)::boolean AS has_route_change,
    bool_and(bucketed.destination_reached)::boolean AS destination_reached
FROM bucketed
GROUP BY bucketed.bucket, bucketed.bucket_width
ORDER BY bucketed.bucket ASC;

-- name: ListTracerouteRunRows :many
WITH selected_runs AS (
    SELECT *
    FROM traceroute_results
    WHERE traceroute_results.probe_id = sqlc.arg(probe_storage_id)
        AND traceroute_results.check_id = sqlc.arg(check_storage_id)
        AND traceroute_results.started_at >= sqlc.arg(started_at_from)
        AND traceroute_results.started_at < sqlc.arg(started_at_to)
        AND (
            sqlc.narg(cursor_started_at)::timestamptz IS NULL
            OR traceroute_results.started_at < sqlc.narg(cursor_started_at)::timestamptz
        )
    ORDER BY traceroute_results.started_at DESC
    LIMIT sqlc.arg(limit_count)
)
SELECT
    selected_runs.started_at,
    selected_runs.finished_at,
    selected_runs.duration_ms,
    selected_runs.status,
    selected_runs.resolved_ip,
    selected_runs.ip_family,
    selected_runs.destination_reached,
    selected_runs.hop_count,
    selected_runs.error_code,
    selected_runs.error_message,
    traceroute_result_hops.hop_index,
    traceroute_result_hops.address,
    traceroute_result_hops.hostname,
    traceroute_result_hops.sent_count,
    traceroute_result_hops.received_count,
    traceroute_result_hops.loss_percent,
    traceroute_result_hops.rtt_min_ms,
    traceroute_result_hops.rtt_avg_ms,
    traceroute_result_hops.rtt_median_ms,
    traceroute_result_hops.rtt_max_ms,
    traceroute_result_hops.rtt_stddev_ms,
    traceroute_result_hops.rtt_samples_ms,
    traceroute_result_hops.error_code AS hop_error_code,
    traceroute_result_hops.error_message AS hop_error_message
FROM selected_runs
LEFT JOIN traceroute_result_hops
    ON traceroute_result_hops.probe_id = selected_runs.probe_id
    AND traceroute_result_hops.check_id = selected_runs.check_id
    AND traceroute_result_hops.started_at = selected_runs.started_at
ORDER BY selected_runs.started_at DESC, traceroute_result_hops.hop_index ASC;

-- name: ListTracerouteTopologyRows :many
WITH selected_runs AS (
    SELECT traceroute_results.probe_id,
           traceroute_results.check_id,
           traceroute_results.started_at,
           traceroute_results.resolved_ip,
           probes.id AS probe_public_id,
           probes.name AS probe_name,
           checks.id AS check_public_id,
           checks.name AS check_name,
           checks.target AS check_target
    FROM traceroute_results
    JOIN probes ON probes.internal_id = traceroute_results.probe_id
    JOIN checks ON checks.internal_id = traceroute_results.check_id
    WHERE probes.project_id = sqlc.arg(project_id)
      AND checks.project_id = sqlc.arg(project_id)
      AND checks.check_type = 'traceroute'
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
    ORDER BY traceroute_results.started_at DESC,
             probes.id ASC,
             checks.id ASC
    LIMIT sqlc.arg(limit_count)
)
SELECT selected_runs.started_at,
       selected_runs.probe_public_id,
       selected_runs.probe_name,
       selected_runs.check_public_id,
       selected_runs.check_name,
       selected_runs.check_target,
       selected_runs.resolved_ip,
       traceroute_result_hops.hop_index,
       traceroute_result_hops.address,
       traceroute_result_hops.hostname,
       traceroute_result_hops.loss_percent,
       traceroute_result_hops.rtt_avg_ms
FROM selected_runs
LEFT JOIN traceroute_result_hops
    ON traceroute_result_hops.probe_id = selected_runs.probe_id
    AND traceroute_result_hops.check_id = selected_runs.check_id
    AND traceroute_result_hops.started_at = selected_runs.started_at
ORDER BY selected_runs.started_at DESC,
         selected_runs.probe_public_id ASC,
         selected_runs.check_public_id ASC,
         traceroute_result_hops.hop_index ASC;

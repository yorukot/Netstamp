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
SELECT ((
    SELECT count(*)::bigint
    FROM traceroute_results
    WHERE traceroute_results.probe_id = sqlc.arg(probe_storage_id)
        AND traceroute_results.check_id = sqlc.arg(check_storage_id)
        AND traceroute_results.started_at >= sqlc.arg(started_at_from)
        AND traceroute_results.started_at < sqlc.arg(started_at_to)
        AND traceroute_results.started_at >= sqlc.arg(raw_cutoff)
) + (
    SELECT count(*)::bigint
    FROM traceroute_sampled_runs_1m
    WHERE traceroute_sampled_runs_1m.probe_id = sqlc.arg(probe_storage_id)
        AND traceroute_sampled_runs_1m.check_id = sqlc.arg(check_storage_id)
        AND traceroute_sampled_runs_1m.bucket >= time_bucket(INTERVAL '1 minute', sqlc.arg(started_at_from)::timestamptz)
        AND traceroute_sampled_runs_1m.bucket < time_bucket(INTERVAL '1 minute', sqlc.arg(started_at_to)::timestamptz) + INTERVAL '1 minute'
        AND traceroute_sampled_runs_1m.sampled_started_at >= sqlc.arg(started_at_from)
        AND traceroute_sampled_runs_1m.sampled_started_at < sqlc.arg(started_at_to)
        AND traceroute_sampled_runs_1m.sampled_started_at < sqlc.arg(raw_cutoff)
))::bigint AS total_runs;

-- name: ListTracerouteInsightRawRows :many
WITH unified_runs AS (
    SELECT
        traceroute_results.started_at,
        traceroute_results.finished_at,
        traceroute_results.destination_reached,
        final_hop.rtt_avg_ms AS final_rtt_avg_ms,
        final_hop.loss_percent AS final_loss_percent,
        signature.path_signature
    FROM traceroute_results
    LEFT JOIN LATERAL (
        SELECT
            traceroute_result_hops.rtt_avg_ms,
            traceroute_result_hops.loss_percent
        FROM traceroute_result_hops
        WHERE traceroute_result_hops.probe_id = traceroute_results.probe_id
            AND traceroute_result_hops.check_id = traceroute_results.check_id
            AND traceroute_result_hops.started_at = traceroute_results.started_at
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
        WHERE traceroute_result_hops.probe_id = traceroute_results.probe_id
            AND traceroute_result_hops.check_id = traceroute_results.check_id
            AND traceroute_result_hops.started_at = traceroute_results.started_at
    ) signature ON TRUE
    WHERE traceroute_results.probe_id = sqlc.arg(probe_storage_id)
        AND traceroute_results.check_id = sqlc.arg(check_storage_id)
        AND traceroute_results.started_at >= sqlc.arg(started_at_from)
        AND traceroute_results.started_at < sqlc.arg(started_at_to)
        AND traceroute_results.started_at >= sqlc.arg(raw_cutoff)
    UNION ALL
    SELECT
        traceroute_sampled_runs_1m.sampled_started_at AS started_at,
        traceroute_sampled_runs_1m.finished_at,
        traceroute_sampled_runs_1m.destination_reached,
        final_hop.rtt_avg_ms AS final_rtt_avg_ms,
        final_hop.loss_percent AS final_loss_percent,
        traceroute_sampled_runs_1m.path_signature
    FROM traceroute_sampled_runs_1m
    LEFT JOIN LATERAL (
        SELECT
            parsed_hops.rtt_avg_ms,
            parsed_hops.loss_percent
        FROM (
            SELECT
                (sampled_hops.value ->> 'hopIndex')::integer AS hop_index,
                (sampled_hops.value ->> 'receivedCount')::integer AS received_count,
                (sampled_hops.value ->> 'rttAvgMs')::double precision AS rtt_avg_ms,
                (sampled_hops.value ->> 'lossPercent')::double precision AS loss_percent
            FROM jsonb_array_elements(traceroute_sampled_runs_1m.hops) AS sampled_hops(value)
        ) parsed_hops
        WHERE (
                parsed_hops.received_count > 0
                OR parsed_hops.rtt_avg_ms IS NOT NULL
        )
        ORDER BY parsed_hops.hop_index DESC
        LIMIT 1
    ) final_hop ON TRUE
    WHERE traceroute_sampled_runs_1m.probe_id = sqlc.arg(probe_storage_id)
        AND traceroute_sampled_runs_1m.check_id = sqlc.arg(check_storage_id)
        AND traceroute_sampled_runs_1m.bucket >= time_bucket(INTERVAL '1 minute', sqlc.arg(started_at_from)::timestamptz)
        AND traceroute_sampled_runs_1m.bucket < time_bucket(INTERVAL '1 minute', sqlc.arg(started_at_to)::timestamptz) + INTERVAL '1 minute'
        AND traceroute_sampled_runs_1m.sampled_started_at >= sqlc.arg(started_at_from)
        AND traceroute_sampled_runs_1m.sampled_started_at < sqlc.arg(started_at_to)
        AND traceroute_sampled_runs_1m.sampled_started_at < sqlc.arg(raw_cutoff)
),
scored AS (
    SELECT
        unified_runs.*,
        lag(unified_runs.path_signature) OVER (ORDER BY unified_runs.started_at ASC) AS previous_path_signature
    FROM unified_runs
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
unified_runs AS (
    SELECT
        traceroute_results.started_at,
        traceroute_results.destination_reached,
        final_hop.rtt_avg_ms AS final_rtt_avg_ms,
        final_hop.loss_percent AS final_loss_percent,
        signature.path_signature
    FROM traceroute_results
    LEFT JOIN LATERAL (
        SELECT
            traceroute_result_hops.rtt_avg_ms,
            traceroute_result_hops.loss_percent
        FROM traceroute_result_hops
        WHERE traceroute_result_hops.probe_id = traceroute_results.probe_id
            AND traceroute_result_hops.check_id = traceroute_results.check_id
            AND traceroute_result_hops.started_at = traceroute_results.started_at
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
        WHERE traceroute_result_hops.probe_id = traceroute_results.probe_id
            AND traceroute_result_hops.check_id = traceroute_results.check_id
            AND traceroute_result_hops.started_at = traceroute_results.started_at
    ) signature ON TRUE
    WHERE traceroute_results.probe_id = sqlc.arg(probe_storage_id)
        AND traceroute_results.check_id = sqlc.arg(check_storage_id)
        AND traceroute_results.started_at >= sqlc.arg(started_at_from)
        AND traceroute_results.started_at < sqlc.arg(started_at_to)
        AND traceroute_results.started_at >= sqlc.arg(raw_cutoff)
    UNION ALL
    SELECT
        traceroute_sampled_runs_1m.sampled_started_at AS started_at,
        traceroute_sampled_runs_1m.destination_reached,
        final_hop.rtt_avg_ms AS final_rtt_avg_ms,
        final_hop.loss_percent AS final_loss_percent,
        traceroute_sampled_runs_1m.path_signature
    FROM traceroute_sampled_runs_1m
    LEFT JOIN LATERAL (
        SELECT
            parsed_hops.rtt_avg_ms,
            parsed_hops.loss_percent
        FROM (
            SELECT
                (sampled_hops.value ->> 'hopIndex')::integer AS hop_index,
                (sampled_hops.value ->> 'receivedCount')::integer AS received_count,
                (sampled_hops.value ->> 'rttAvgMs')::double precision AS rtt_avg_ms,
                (sampled_hops.value ->> 'lossPercent')::double precision AS loss_percent
            FROM jsonb_array_elements(traceroute_sampled_runs_1m.hops) AS sampled_hops(value)
        ) parsed_hops
        WHERE (
                parsed_hops.received_count > 0
                OR parsed_hops.rtt_avg_ms IS NOT NULL
        )
        ORDER BY parsed_hops.hop_index DESC
        LIMIT 1
    ) final_hop ON TRUE
    WHERE traceroute_sampled_runs_1m.probe_id = sqlc.arg(probe_storage_id)
        AND traceroute_sampled_runs_1m.check_id = sqlc.arg(check_storage_id)
        AND traceroute_sampled_runs_1m.bucket >= time_bucket(INTERVAL '1 minute', sqlc.arg(started_at_from)::timestamptz)
        AND traceroute_sampled_runs_1m.bucket < time_bucket(INTERVAL '1 minute', sqlc.arg(started_at_to)::timestamptz) + INTERVAL '1 minute'
        AND traceroute_sampled_runs_1m.sampled_started_at >= sqlc.arg(started_at_from)
        AND traceroute_sampled_runs_1m.sampled_started_at < sqlc.arg(started_at_to)
        AND traceroute_sampled_runs_1m.sampled_started_at < sqlc.arg(raw_cutoff)
),
scored AS (
    SELECT
        unified_runs.*,
        lag(unified_runs.path_signature) OVER (ORDER BY unified_runs.started_at ASC) AS previous_path_signature
    FROM unified_runs
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
    FROM (
        SELECT
            'raw'::text AS source,
            traceroute_results.probe_id,
            traceroute_results.check_id,
            traceroute_results.started_at,
            traceroute_results.finished_at,
            traceroute_results.duration_ms,
            traceroute_results.status,
            traceroute_results.resolved_ip,
            traceroute_results.ip_family,
            traceroute_results.destination_reached,
            traceroute_results.hop_count,
            traceroute_results.error_code,
            traceroute_results.error_message,
            NULL::jsonb AS hops
        FROM traceroute_results
        WHERE traceroute_results.probe_id = sqlc.arg(probe_storage_id)
            AND traceroute_results.check_id = sqlc.arg(check_storage_id)
            AND traceroute_results.started_at >= sqlc.arg(started_at_from)
            AND traceroute_results.started_at < sqlc.arg(started_at_to)
            AND traceroute_results.started_at >= sqlc.arg(raw_cutoff)
            AND (
                sqlc.narg(cursor_started_at)::timestamptz IS NULL
                OR traceroute_results.started_at < sqlc.narg(cursor_started_at)::timestamptz
            )
        UNION ALL
        SELECT
            'sampled'::text AS source,
            traceroute_sampled_runs_1m.probe_id,
            traceroute_sampled_runs_1m.check_id,
            traceroute_sampled_runs_1m.sampled_started_at AS started_at,
            traceroute_sampled_runs_1m.finished_at,
            traceroute_sampled_runs_1m.duration_ms,
            traceroute_sampled_runs_1m.status,
            traceroute_sampled_runs_1m.resolved_ip,
            traceroute_sampled_runs_1m.ip_family,
            traceroute_sampled_runs_1m.destination_reached,
            traceroute_sampled_runs_1m.hop_count,
            NULL::text AS error_code,
            NULL::text AS error_message,
            traceroute_sampled_runs_1m.hops
        FROM traceroute_sampled_runs_1m
        WHERE traceroute_sampled_runs_1m.probe_id = sqlc.arg(probe_storage_id)
            AND traceroute_sampled_runs_1m.check_id = sqlc.arg(check_storage_id)
            AND traceroute_sampled_runs_1m.bucket >= time_bucket(INTERVAL '1 minute', sqlc.arg(started_at_from)::timestamptz)
            AND traceroute_sampled_runs_1m.bucket < time_bucket(INTERVAL '1 minute', sqlc.arg(started_at_to)::timestamptz) + INTERVAL '1 minute'
            AND traceroute_sampled_runs_1m.sampled_started_at >= sqlc.arg(started_at_from)
            AND traceroute_sampled_runs_1m.sampled_started_at < sqlc.arg(started_at_to)
            AND traceroute_sampled_runs_1m.sampled_started_at < sqlc.arg(raw_cutoff)
            AND (
                sqlc.narg(cursor_started_at)::timestamptz IS NULL
                OR traceroute_sampled_runs_1m.sampled_started_at < sqlc.narg(cursor_started_at)::timestamptz
            )
    ) unified_runs
    ORDER BY unified_runs.started_at DESC
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
    coalesce(traceroute_hops.hop_index, 0)::integer AS hop_index,
    traceroute_hops.address,
    traceroute_hops.hostname,
    coalesce(traceroute_hops.sent_count, 0)::integer AS sent_count,
    coalesce(traceroute_hops.received_count, 0)::integer AS received_count,
    coalesce(traceroute_hops.loss_percent, 0)::double precision AS loss_percent,
    traceroute_hops.rtt_min_ms,
    traceroute_hops.rtt_avg_ms,
    traceroute_hops.rtt_median_ms,
    traceroute_hops.rtt_max_ms,
    traceroute_hops.rtt_stddev_ms,
    coalesce(traceroute_hops.rtt_samples_ms, ARRAY[]::double precision[]) AS rtt_samples_ms,
    traceroute_hops.error_code AS hop_error_code,
    traceroute_hops.error_message AS hop_error_message
FROM selected_runs
LEFT JOIN LATERAL (
    SELECT
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
        traceroute_result_hops.error_code,
        traceroute_result_hops.error_message
    FROM traceroute_result_hops
    WHERE selected_runs.source = 'raw'
        AND traceroute_result_hops.probe_id = selected_runs.probe_id
        AND traceroute_result_hops.check_id = selected_runs.check_id
        AND traceroute_result_hops.started_at = selected_runs.started_at
    UNION ALL
    SELECT
        (sampled_hops.value ->> 'hopIndex')::integer AS hop_index,
        NULLIF(sampled_hops.value ->> 'address', '')::inet AS address,
        sampled_hops.value ->> 'hostname' AS hostname,
        (sampled_hops.value ->> 'sentCount')::integer AS sent_count,
        (sampled_hops.value ->> 'receivedCount')::integer AS received_count,
        (sampled_hops.value ->> 'lossPercent')::double precision AS loss_percent,
        (sampled_hops.value ->> 'rttMinMs')::double precision AS rtt_min_ms,
        (sampled_hops.value ->> 'rttAvgMs')::double precision AS rtt_avg_ms,
        (sampled_hops.value ->> 'rttMedianMs')::double precision AS rtt_median_ms,
        (sampled_hops.value ->> 'rttMaxMs')::double precision AS rtt_max_ms,
        (sampled_hops.value ->> 'rttStddevMs')::double precision AS rtt_stddev_ms,
        ARRAY(
            SELECT sample.value::double precision
            FROM jsonb_array_elements_text(
                CASE
                    WHEN jsonb_typeof(sampled_hops.value -> 'rttSamplesMs') = 'array' THEN sampled_hops.value -> 'rttSamplesMs'
                    ELSE '[]'::jsonb
                END
            ) AS sample(value)
        )::double precision[] AS rtt_samples_ms,
        sampled_hops.value ->> 'errorCode' AS error_code,
        sampled_hops.value ->> 'errorMessage' AS error_message
    FROM jsonb_array_elements(CASE WHEN selected_runs.source = 'sampled' THEN selected_runs.hops ELSE '[]'::jsonb END) AS sampled_hops(value)
    UNION ALL
    SELECT
        NULL::integer AS hop_index,
        NULL::inet AS address,
        NULL::text AS hostname,
        NULL::integer AS sent_count,
        NULL::integer AS received_count,
        NULL::double precision AS loss_percent,
        NULL::double precision AS rtt_min_ms,
        NULL::double precision AS rtt_avg_ms,
        NULL::double precision AS rtt_median_ms,
        NULL::double precision AS rtt_max_ms,
        NULL::double precision AS rtt_stddev_ms,
        NULL::double precision[] AS rtt_samples_ms,
        NULL::text AS error_code,
        NULL::text AS error_message
    WHERE FALSE
) traceroute_hops ON TRUE
ORDER BY selected_runs.started_at DESC, traceroute_hops.hop_index ASC;

-- name: ListTracerouteTopologyRows :many
WITH selected_runs AS (
    SELECT *
    FROM (
        SELECT
            'raw'::text AS source,
            traceroute_results.probe_id,
            traceroute_results.check_id,
            traceroute_results.started_at,
            traceroute_results.resolved_ip,
            probes.id AS probe_public_id,
            probes.name AS probe_name,
            checks.id AS check_public_id,
            checks.name AS check_name,
            checks.target AS check_target,
            NULL::jsonb AS hops
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
          AND traceroute_results.started_at >= sqlc.arg(raw_cutoff)
          AND (
              sqlc.narg(probe_id)::uuid IS NULL
              OR probes.id = sqlc.narg(probe_id)::uuid
          )
          AND (
              sqlc.narg(check_id)::uuid IS NULL
              OR checks.id = sqlc.narg(check_id)::uuid
          )
        UNION ALL
        SELECT
            'sampled'::text AS source,
            traceroute_sampled_runs_1m.probe_id,
            traceroute_sampled_runs_1m.check_id,
            traceroute_sampled_runs_1m.sampled_started_at AS started_at,
            traceroute_sampled_runs_1m.resolved_ip,
            probes.id AS probe_public_id,
            probes.name AS probe_name,
            checks.id AS check_public_id,
            checks.name AS check_name,
            checks.target AS check_target,
            traceroute_sampled_runs_1m.hops
        FROM traceroute_sampled_runs_1m
        JOIN probes ON probes.internal_id = traceroute_sampled_runs_1m.probe_id
        JOIN checks ON checks.internal_id = traceroute_sampled_runs_1m.check_id
        WHERE probes.project_id = sqlc.arg(project_id)
          AND checks.project_id = sqlc.arg(project_id)
          AND checks.check_type = 'traceroute'
          AND probes.deleted_at IS NULL
          AND checks.deleted_at IS NULL
          AND traceroute_sampled_runs_1m.bucket >= time_bucket(INTERVAL '1 minute', sqlc.arg(started_at_from)::timestamptz)
          AND traceroute_sampled_runs_1m.bucket < time_bucket(INTERVAL '1 minute', sqlc.arg(started_at_to)::timestamptz) + INTERVAL '1 minute'
          AND traceroute_sampled_runs_1m.sampled_started_at >= sqlc.arg(started_at_from)
          AND traceroute_sampled_runs_1m.sampled_started_at < sqlc.arg(started_at_to)
          AND traceroute_sampled_runs_1m.sampled_started_at < sqlc.arg(raw_cutoff)
          AND (
              sqlc.narg(probe_id)::uuid IS NULL
              OR probes.id = sqlc.narg(probe_id)::uuid
          )
          AND (
              sqlc.narg(check_id)::uuid IS NULL
              OR checks.id = sqlc.narg(check_id)::uuid
          )
    ) unified_runs
    ORDER BY unified_runs.started_at DESC,
             unified_runs.probe_public_id ASC,
             unified_runs.check_public_id ASC
    LIMIT sqlc.arg(limit_count)
)
SELECT selected_runs.started_at,
       selected_runs.probe_public_id,
       selected_runs.probe_name,
       selected_runs.check_public_id,
       selected_runs.check_name,
       selected_runs.check_target,
       selected_runs.resolved_ip,
       coalesce(traceroute_hops.hop_index, 0)::integer AS hop_index,
       traceroute_hops.address,
       traceroute_hops.hostname,
       coalesce(traceroute_hops.loss_percent, 0)::double precision AS loss_percent,
       traceroute_hops.rtt_avg_ms
FROM selected_runs
LEFT JOIN LATERAL (
    SELECT
        traceroute_result_hops.hop_index,
        traceroute_result_hops.address,
        traceroute_result_hops.hostname,
        traceroute_result_hops.loss_percent,
        traceroute_result_hops.rtt_avg_ms
    FROM traceroute_result_hops
    WHERE selected_runs.source = 'raw'
        AND traceroute_result_hops.probe_id = selected_runs.probe_id
        AND traceroute_result_hops.check_id = selected_runs.check_id
        AND traceroute_result_hops.started_at = selected_runs.started_at
    UNION ALL
    SELECT
        (sampled_hops.value ->> 'hopIndex')::integer AS hop_index,
        NULLIF(sampled_hops.value ->> 'address', '')::inet AS address,
        sampled_hops.value ->> 'hostname' AS hostname,
        (sampled_hops.value ->> 'lossPercent')::double precision AS loss_percent,
        (sampled_hops.value ->> 'rttAvgMs')::double precision AS rtt_avg_ms
    FROM jsonb_array_elements(CASE WHEN selected_runs.source = 'sampled' THEN selected_runs.hops ELSE '[]'::jsonb END) AS sampled_hops(value)
    UNION ALL
    SELECT
        NULL::integer AS hop_index,
        NULL::inet AS address,
        NULL::text AS hostname,
        NULL::double precision AS loss_percent,
        NULL::double precision AS rtt_avg_ms
    WHERE FALSE
) traceroute_hops ON TRUE
ORDER BY selected_runs.started_at DESC,
         selected_runs.probe_public_id ASC,
         selected_runs.check_public_id ASC,
         traceroute_hops.hop_index ASC;

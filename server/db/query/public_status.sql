-- name: ListPublicStatusPages :many
SELECT id,
       project_id,
       slug,
       title,
       description,
       enabled,
       default_chart_mode,
       default_chart_range,
       created_by_user_id,
       created_at,
       updated_at,
       deleted_at
FROM public_status_pages
WHERE project_id = sqlc.arg(project_id)
  AND deleted_at IS NULL
ORDER BY created_at DESC, id DESC;

-- name: GetPublicStatusPage :one
SELECT id,
       project_id,
       slug,
       title,
       description,
       enabled,
       default_chart_mode,
       default_chart_range,
       created_by_user_id,
       created_at,
       updated_at,
       deleted_at
FROM public_status_pages
WHERE project_id = sqlc.arg(project_id)
  AND id = sqlc.arg(id)
  AND deleted_at IS NULL;

-- name: GetPublicStatusPageBySlug :one
SELECT id,
       project_id,
       slug,
       title,
       description,
       enabled,
       default_chart_mode,
       default_chart_range,
       created_by_user_id,
       created_at,
       updated_at,
       deleted_at
FROM public_status_pages
WHERE slug = sqlc.arg(slug)
  AND enabled = true
  AND deleted_at IS NULL;

-- name: CreatePublicStatusPage :one
INSERT INTO public_status_pages (
    project_id,
    slug,
    title,
    description,
    enabled,
    default_chart_mode,
    default_chart_range,
    created_by_user_id
)
VALUES (
    sqlc.arg(project_id),
    sqlc.arg(slug),
    sqlc.arg(title),
    sqlc.narg(description),
    sqlc.arg(enabled),
    sqlc.arg(default_chart_mode),
    sqlc.arg(default_chart_range),
    sqlc.arg(created_by_user_id)
)
RETURNING id, project_id, slug, title, description, enabled, default_chart_mode, default_chart_range, created_by_user_id, created_at, updated_at, deleted_at;

-- name: UpdatePublicStatusPage :one
UPDATE public_status_pages
SET slug = sqlc.arg(slug),
    title = sqlc.arg(title),
    description = sqlc.narg(description),
    enabled = sqlc.arg(enabled),
    default_chart_mode = sqlc.arg(default_chart_mode),
    default_chart_range = sqlc.arg(default_chart_range)
WHERE project_id = sqlc.arg(project_id)
  AND id = sqlc.arg(id)
  AND deleted_at IS NULL
RETURNING id, project_id, slug, title, description, enabled, default_chart_mode, default_chart_range, created_by_user_id, created_at, updated_at, deleted_at;

-- name: SoftDeletePublicStatusPage :execrows
UPDATE public_status_pages
SET deleted_at = now(),
    enabled = false
WHERE project_id = sqlc.arg(project_id)
  AND id = sqlc.arg(id)
  AND deleted_at IS NULL;

-- name: ListPublicStatusPageElements :many
SELECT public_status_page_elements.id,
       public_status_page_elements.public_page_id,
       public_status_page_elements.project_id,
       public_status_page_elements.parent_element_id,
       public_status_page_elements.kind,
       public_status_page_elements.check_id,
       public_status_page_elements.assignment_selection_mode,
       public_status_page_elements.title,
       public_status_page_elements.description,
       public_status_page_elements.sort_order,
       public_status_page_elements.chart_mode,
       public_status_page_elements.chart_range,
       public_status_page_elements.created_at,
       public_status_page_elements.updated_at,
       checks.name AS check_name,
       checks.check_type AS check_type,
       checks.target AS check_target,
       checks.description AS check_description,
       checks.interval_seconds AS check_interval_seconds
FROM public_status_page_elements
LEFT JOIN checks
  ON checks.project_id = public_status_page_elements.project_id
 AND checks.id = public_status_page_elements.check_id
 AND checks.deleted_at IS NULL
WHERE public_status_page_elements.public_page_id = sqlc.arg(public_page_id)
  AND (
      public_status_page_elements.kind = 'folder'
      OR public_status_page_elements.assignment_selection_mode = 'selected_assignments'
      OR checks.id IS NOT NULL
  )
ORDER BY public_status_page_elements.parent_element_id NULLS FIRST,
         public_status_page_elements.sort_order ASC,
         public_status_page_elements.created_at ASC,
         public_status_page_elements.id ASC;

-- name: GetPublicStatusPageElement :one
SELECT id,
       public_page_id,
       project_id,
       parent_element_id,
       kind,
       check_id,
       assignment_selection_mode,
       title,
       description,
       sort_order,
       chart_mode,
       chart_range,
       created_at,
       updated_at
FROM public_status_page_elements
WHERE public_page_id = sqlc.arg(public_page_id)
  AND project_id = sqlc.arg(project_id)
  AND id = sqlc.arg(id);

-- name: CreatePublicStatusPageElement :one
INSERT INTO public_status_page_elements (
    public_page_id,
    project_id,
    parent_element_id,
    kind,
    check_id,
    assignment_selection_mode,
    title,
    description,
    sort_order,
    chart_mode,
    chart_range
)
VALUES (
    sqlc.arg(public_page_id),
    sqlc.arg(project_id),
    sqlc.narg(parent_element_id),
    sqlc.arg(kind),
    sqlc.narg(check_id),
    sqlc.narg(assignment_selection_mode),
    sqlc.narg(title),
    sqlc.narg(description),
    sqlc.arg(sort_order),
    sqlc.arg(chart_mode),
    sqlc.narg(chart_range)
)
RETURNING id, public_page_id, project_id, parent_element_id, kind, check_id, assignment_selection_mode, title, description, sort_order, chart_mode, chart_range, created_at, updated_at;

-- name: UpdatePublicStatusPageElement :one
UPDATE public_status_page_elements
SET parent_element_id = sqlc.narg(parent_element_id),
    kind = sqlc.arg(kind),
    check_id = sqlc.narg(check_id),
    assignment_selection_mode = sqlc.narg(assignment_selection_mode),
    title = sqlc.narg(title),
    description = sqlc.narg(description),
    sort_order = sqlc.arg(sort_order),
    chart_mode = sqlc.arg(chart_mode),
    chart_range = sqlc.narg(chart_range)
WHERE public_page_id = sqlc.arg(public_page_id)
  AND project_id = sqlc.arg(project_id)
  AND id = sqlc.arg(id)
RETURNING id, public_page_id, project_id, parent_element_id, kind, check_id, assignment_selection_mode, title, description, sort_order, chart_mode, chart_range, created_at, updated_at;

-- name: ListPublicStatusPageElementAssignmentIDs :many
SELECT element_id,
       assignment_id
FROM public_status_page_element_assignments
WHERE public_page_id = sqlc.arg(public_page_id)
ORDER BY element_id ASC,
         sort_order ASC,
         assignment_id ASC;

-- name: DeletePublicStatusPageElementAssignments :exec
DELETE FROM public_status_page_element_assignments
WHERE public_page_id = sqlc.arg(public_page_id)
  AND element_id = sqlc.arg(element_id);

-- name: InsertPublicStatusPageElementAssignments :execrows
INSERT INTO public_status_page_element_assignments (
    element_id,
    public_page_id,
    project_id,
    assignment_id,
    sort_order
)
SELECT sqlc.arg(element_id),
       sqlc.arg(public_page_id),
       sqlc.arg(project_id),
       probe_check_assignments.id,
       (selected.ordinality::integer - 1)
FROM unnest(sqlc.arg(assignment_ids)::uuid[]) WITH ORDINALITY AS selected(assignment_id, ordinality)
JOIN probe_check_assignments
  ON probe_check_assignments.id = selected.assignment_id
 AND probe_check_assignments.project_id = sqlc.arg(project_id)
 AND probe_check_assignments.deleted_at IS NULL
JOIN checks
  ON checks.project_id = probe_check_assignments.project_id
 AND checks.id = probe_check_assignments.check_id
 AND checks.deleted_at IS NULL
 AND checks.check_type IN ('ping', 'tcp', 'http');

-- name: GetPublicStatusAssignableCheck :one
SELECT id,
       check_type
FROM checks
WHERE project_id = sqlc.arg(project_id)
  AND id = sqlc.arg(check_id)
  AND deleted_at IS NULL
  AND check_type IN ('ping', 'tcp', 'http');

-- name: CountPublicStatusAssignableAssignments :one
SELECT count(DISTINCT probe_check_assignments.id)::bigint
FROM probe_check_assignments
JOIN checks
  ON checks.project_id = probe_check_assignments.project_id
 AND checks.id = probe_check_assignments.check_id
 AND checks.deleted_at IS NULL
 AND checks.check_type IN ('ping', 'tcp', 'http')
WHERE probe_check_assignments.project_id = sqlc.arg(project_id)
  AND probe_check_assignments.deleted_at IS NULL
  AND probe_check_assignments.id = ANY(sqlc.arg(assignment_ids)::uuid[]);

-- name: DeletePublicStatusPageElement :execrows
DELETE FROM public_status_page_elements
WHERE public_page_id = sqlc.arg(public_page_id)
  AND project_id = sqlc.arg(project_id)
  AND id = sqlc.arg(id);

-- name: ListPublicStatusAssignments :many
SELECT public_status_page_assignment_scope.element_id,
       public_status_page_assignment_scope.assignment_id,
       checks.id AS check_id,
       checks.name AS check_name,
       checks.check_type,
       checks.target AS check_target,
       checks.interval_seconds,
       probes.id AS probe_id,
       probes.name AS probe_name,
       probes.location_name AS probe_location_name,
       COALESCE(latest.started_at, 'epoch'::timestamptz) AS latest_started_at,
       COALESCE(latest.status, '') AS latest_status,
       latest.latency_avg_ms,
       COALESCE(latest.loss_percent, 0::double precision) AS loss_percent,
       latest.connect_avg_ms,
       latest.failure_percent
FROM public_status_page_assignment_scope
JOIN probes
  ON probes.project_id = public_status_page_assignment_scope.project_id
 AND probes.id = public_status_page_assignment_scope.probe_id
 AND probes.deleted_at IS NULL
JOIN checks
  ON checks.project_id = public_status_page_assignment_scope.project_id
 AND checks.id = public_status_page_assignment_scope.check_id
 AND checks.deleted_at IS NULL
LEFT JOIN LATERAL (
    (
        SELECT ping_results.started_at,
               ping_results.status::text AS status,
               ping_results.rtt_avg_ms AS latency_avg_ms,
               ping_results.loss_percent AS loss_percent,
               NULL::double precision AS connect_avg_ms,
               NULL::double precision AS failure_percent
        FROM ping_results
        WHERE checks.check_type = 'ping'
          AND ping_results.probe_id = probes.internal_id
          AND ping_results.check_id = checks.internal_id
        ORDER BY ping_results.started_at DESC
        LIMIT 1
    )
    UNION ALL
    (
        SELECT tcp_results.started_at,
               tcp_results.status::text AS status,
               NULL::double precision AS latency_avg_ms,
               NULL::double precision AS loss_percent,
               tcp_results.connect_duration_ms AS connect_avg_ms,
               CASE WHEN tcp_results.status = 'successful' THEN 0::double precision ELSE 100::double precision END AS failure_percent
        FROM tcp_results
        WHERE checks.check_type = 'tcp'
          AND tcp_results.probe_id = probes.internal_id
          AND tcp_results.check_id = checks.internal_id
        ORDER BY tcp_results.started_at DESC
        LIMIT 1
    )
    UNION ALL
    (
        SELECT http_results.started_at,
               http_results.status::text AS status,
               http_results.duration_ms::double precision AS latency_avg_ms,
               NULL::double precision AS loss_percent,
               NULL::double precision AS connect_avg_ms,
               CASE WHEN http_results.status = 'successful' THEN 0::double precision ELSE 100::double precision END AS failure_percent
        FROM http_results
        WHERE checks.check_type = 'http'
          AND http_results.probe_id = probes.internal_id
          AND http_results.check_id = checks.internal_id
        ORDER BY http_results.started_at DESC
        LIMIT 1
    )
) latest ON TRUE
WHERE public_status_page_assignment_scope.public_page_id = sqlc.arg(public_page_id)
ORDER BY public_status_page_assignment_scope.element_id ASC,
         probes.name ASC,
         checks.name ASC,
         public_status_page_assignment_scope.assignment_id ASC;

-- name: ListPublicStatusElementAssignments :many
SELECT public_status_page_assignment_scope.element_id,
       public_status_page_assignment_scope.assignment_id,
       checks.id AS check_id,
       checks.name AS check_name,
       checks.check_type,
       checks.target AS check_target,
       checks.interval_seconds,
       probes.id AS probe_id,
       probes.name AS probe_name,
       probes.location_name AS probe_location_name,
       COALESCE(latest.started_at, 'epoch'::timestamptz) AS latest_started_at,
       COALESCE(latest.status, '') AS latest_status,
       latest.latency_avg_ms,
       COALESCE(latest.loss_percent, 0::double precision) AS loss_percent,
       latest.connect_avg_ms,
       latest.failure_percent
FROM public_status_page_assignment_scope
JOIN probes
  ON probes.project_id = public_status_page_assignment_scope.project_id
 AND probes.id = public_status_page_assignment_scope.probe_id
 AND probes.deleted_at IS NULL
JOIN checks
  ON checks.project_id = public_status_page_assignment_scope.project_id
 AND checks.id = public_status_page_assignment_scope.check_id
 AND checks.deleted_at IS NULL
LEFT JOIN LATERAL (
    (
        SELECT ping_results.started_at,
               ping_results.status::text AS status,
               ping_results.rtt_avg_ms AS latency_avg_ms,
               ping_results.loss_percent AS loss_percent,
               NULL::double precision AS connect_avg_ms,
               NULL::double precision AS failure_percent
        FROM ping_results
        WHERE checks.check_type = 'ping'
          AND ping_results.probe_id = probes.internal_id
          AND ping_results.check_id = checks.internal_id
        ORDER BY ping_results.started_at DESC
        LIMIT 1
    )
    UNION ALL
    (
        SELECT tcp_results.started_at,
               tcp_results.status::text AS status,
               NULL::double precision AS latency_avg_ms,
               NULL::double precision AS loss_percent,
               tcp_results.connect_duration_ms AS connect_avg_ms,
               CASE WHEN tcp_results.status = 'successful' THEN 0::double precision ELSE 100::double precision END AS failure_percent
        FROM tcp_results
        WHERE checks.check_type = 'tcp'
          AND tcp_results.probe_id = probes.internal_id
          AND tcp_results.check_id = checks.internal_id
        ORDER BY tcp_results.started_at DESC
        LIMIT 1
    )
    UNION ALL
    (
        SELECT http_results.started_at,
               http_results.status::text AS status,
               http_results.duration_ms::double precision AS latency_avg_ms,
               NULL::double precision AS loss_percent,
               NULL::double precision AS connect_avg_ms,
               CASE WHEN http_results.status = 'successful' THEN 0::double precision ELSE 100::double precision END AS failure_percent
        FROM http_results
        WHERE checks.check_type = 'http'
          AND http_results.probe_id = probes.internal_id
          AND http_results.check_id = checks.internal_id
        ORDER BY http_results.started_at DESC
        LIMIT 1
    )
) latest ON TRUE
WHERE public_status_page_assignment_scope.public_page_id = sqlc.arg(public_page_id)
  AND public_status_page_assignment_scope.element_id = sqlc.arg(element_id)
ORDER BY probes.name ASC,
         checks.name ASC,
         public_status_page_assignment_scope.assignment_id ASC;

-- name: ListPublicStatusIncidents :many
WITH page_scope AS (
    SELECT DISTINCT project_id,
           probe_id,
           check_id
    FROM public_status_page_assignment_scope
    WHERE public_page_id = sqlc.arg(public_page_id)
)
SELECT alert_incidents.id,
       alert_incidents.project_id,
       alert_incidents.rule_id,
       alert_incidents.probe_id,
       alert_incidents.check_id,
       alert_incidents.check_type,
       alert_incidents.status,
       alert_incidents.severity,
       alert_incidents.last_evaluation_state,
       alert_incidents.opened_at,
       alert_incidents.resolved_at,
       alert_incidents.last_triggered_at,
       alert_incidents.last_value,
       alert_incidents.last_summary,
       checks.name AS check_name
FROM alert_incidents
JOIN checks
  ON checks.project_id = alert_incidents.project_id
 AND checks.id = alert_incidents.check_id
WHERE alert_incidents.status IN ('open', 'acknowledged', 'resolved')
  AND EXISTS (
      SELECT 1
      FROM page_scope
      WHERE page_scope.project_id = alert_incidents.project_id
        AND page_scope.check_id = alert_incidents.check_id
        AND (
            alert_incidents.probe_id IS NULL
            OR page_scope.probe_id = alert_incidents.probe_id
        )
  )
ORDER BY CASE WHEN alert_incidents.status IN ('open', 'acknowledged') THEN 0 ELSE 1 END,
         alert_incidents.opened_at DESC,
         alert_incidents.id DESC
LIMIT sqlc.arg(limit_count);

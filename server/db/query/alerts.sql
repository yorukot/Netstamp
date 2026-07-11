-- name: ListAlertRules :many
SELECT alert_rules.id,
       alert_rules.project_id,
       alert_rules.name,
       alert_rules.description,
       alert_rules.status,
       alert_rules.severity,
       alert_rules.check_type,
       alert_rules.probe_id,
       alert_rules.check_id,
       alert_rules.probe_selector,
       alert_rules.condition,
       alert_rules.condition_version,
       alert_rules.cooldown_seconds,
       alert_rules.created_by_user_id,
       alert_rules.created_at,
       alert_rules.updated_at,
       alert_rules.deleted_at
FROM alert_rules
WHERE alert_rules.project_id = sqlc.arg(project_id)
  AND alert_rules.deleted_at IS NULL
  AND (
      sqlc.narg(status)::alert_rule_status IS NULL
      OR alert_rules.status = sqlc.narg(status)::alert_rule_status
  )
  AND (
      sqlc.narg(check_type)::check_type IS NULL
      OR alert_rules.check_type = sqlc.narg(check_type)::check_type
  )
ORDER BY alert_rules.created_at DESC, alert_rules.id DESC;

-- name: GetAlertRule :one
SELECT alert_rules.id,
       alert_rules.project_id,
       alert_rules.name,
       alert_rules.description,
       alert_rules.status,
       alert_rules.severity,
       alert_rules.check_type,
       alert_rules.probe_id,
       alert_rules.check_id,
       alert_rules.probe_selector,
       alert_rules.condition,
       alert_rules.condition_version,
       alert_rules.cooldown_seconds,
       alert_rules.created_by_user_id,
       alert_rules.created_at,
       alert_rules.updated_at,
       alert_rules.deleted_at
FROM alert_rules
WHERE alert_rules.project_id = sqlc.arg(project_id)
  AND alert_rules.id = sqlc.arg(id)
  AND alert_rules.deleted_at IS NULL;

-- name: CreateAlertRule :one
INSERT INTO alert_rules (
    project_id,
    name,
    description,
    status,
    severity,
    check_type,
    probe_id,
    check_id,
    probe_selector,
    condition,
    condition_version,
    cooldown_seconds,
    created_by_user_id
)
VALUES (
    sqlc.arg(project_id),
    sqlc.arg(name),
    sqlc.narg(description),
    sqlc.arg(status),
    sqlc.arg(severity),
    sqlc.arg(check_type),
    sqlc.narg(probe_id),
    sqlc.narg(check_id),
    sqlc.arg(probe_selector)::jsonb,
    sqlc.arg(condition)::jsonb,
    sqlc.arg(condition_version),
    sqlc.arg(cooldown_seconds),
    sqlc.arg(created_by_user_id)
)
RETURNING id, project_id, name, description, status, severity, check_type, probe_id, check_id, probe_selector, condition, condition_version, cooldown_seconds, created_by_user_id, created_at, updated_at, deleted_at;

-- name: UpdateAlertRule :one
UPDATE alert_rules
SET name = sqlc.arg(name),
    description = sqlc.narg(description),
    status = sqlc.arg(status),
    severity = sqlc.arg(severity),
    check_type = sqlc.arg(check_type),
    probe_id = sqlc.narg(probe_id),
    check_id = sqlc.narg(check_id),
    probe_selector = sqlc.arg(probe_selector)::jsonb,
    condition = sqlc.arg(condition)::jsonb,
    condition_version = sqlc.arg(condition_version),
    cooldown_seconds = sqlc.arg(cooldown_seconds)
WHERE project_id = sqlc.arg(project_id)
  AND id = sqlc.arg(id)
  AND deleted_at IS NULL
RETURNING id, project_id, name, description, status, severity, check_type, probe_id, check_id, probe_selector, condition, condition_version, cooldown_seconds, created_by_user_id, created_at, updated_at, deleted_at;

-- name: SoftDeleteAlertRule :execrows
UPDATE alert_rules
SET deleted_at = now(),
    status = 'disabled'
WHERE project_id = sqlc.arg(project_id)
  AND id = sqlc.arg(id)
  AND deleted_at IS NULL;

-- name: ReplaceAlertNotifications :exec
DELETE FROM alert_notifications
WHERE project_id = sqlc.arg(project_id)
  AND rule_id = sqlc.arg(rule_id);

-- name: AddAlertNotification :exec
INSERT INTO alert_notifications (project_id, rule_id, notification_id)
VALUES (sqlc.arg(project_id), sqlc.arg(rule_id), sqlc.arg(notification_id))
ON CONFLICT (rule_id, notification_id) DO NOTHING;

-- name: ListAlertNotificationIDs :many
SELECT notification_id
FROM alert_notifications
WHERE project_id = sqlc.arg(project_id)
  AND rule_id = ANY(sqlc.arg(rule_ids)::uuid[])
ORDER BY created_at ASC, notification_id ASC;

-- name: ListNotifications :many
SELECT id,
       project_id,
       name,
       type,
       enabled,
       config,
       created_by_user_id,
       created_at,
       updated_at,
       deleted_at
FROM notifications
WHERE project_id = sqlc.arg(project_id)
  AND deleted_at IS NULL
  AND (
      sqlc.narg(notification_type)::notification_type IS NULL
      OR type = sqlc.narg(notification_type)::notification_type
  )
ORDER BY created_at DESC, id DESC;

-- name: GetNotification :one
SELECT id,
       project_id,
       name,
       type,
       enabled,
       config,
       created_by_user_id,
       created_at,
       updated_at,
       deleted_at
FROM notifications
WHERE project_id = sqlc.arg(project_id)
  AND id = sqlc.arg(id)
  AND deleted_at IS NULL;

-- name: CreateNotification :one
INSERT INTO notifications (
    project_id,
    name,
    type,
    enabled,
    config,
    created_by_user_id
)
VALUES (
    sqlc.arg(project_id),
    sqlc.arg(name),
    sqlc.arg(notification_type),
    sqlc.arg(enabled),
    sqlc.arg(config)::jsonb,
    sqlc.arg(created_by_user_id)
)
RETURNING id, project_id, name, type, enabled, config, created_by_user_id, created_at, updated_at, deleted_at;

-- name: UpdateNotification :one
UPDATE notifications
SET name = sqlc.arg(name),
    type = sqlc.arg(notification_type),
    enabled = sqlc.arg(enabled),
    config = sqlc.arg(config)::jsonb
WHERE project_id = sqlc.arg(project_id)
  AND id = sqlc.arg(id)
  AND deleted_at IS NULL
RETURNING id, project_id, name, type, enabled, config, created_by_user_id, created_at, updated_at, deleted_at;

-- name: SoftDeleteNotification :execrows
UPDATE notifications
SET deleted_at = now(),
    enabled = false
WHERE project_id = sqlc.arg(project_id)
  AND id = sqlc.arg(id)
  AND deleted_at IS NULL;

-- name: ListAlertIncidents :many
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
       alert_incidents.acknowledged_at,
       alert_incidents.acknowledged_by_user_id,
       alert_incidents.resolved_at,
       alert_incidents.resolved_by_user_id,
       alert_incidents.last_evaluated_at,
       alert_incidents.last_triggered_at,
       alert_incidents.last_value,
       alert_incidents.last_summary,
       alert_incidents.last_notification_sent_at,
       alert_incidents.next_notification_eligible_at,
       alert_incidents.suppressed_notification_count,
       alert_incidents.created_at,
       alert_incidents.updated_at,
       probes.name AS probe_name,
       checks.name AS check_name,
       checks.check_type AS check_summary_type,
       checks.target AS check_target
FROM alert_incidents
JOIN probes
  ON probes.project_id = alert_incidents.project_id
 AND probes.id = alert_incidents.probe_id
JOIN checks
  ON checks.project_id = alert_incidents.project_id
 AND checks.id = alert_incidents.check_id
WHERE alert_incidents.project_id = sqlc.arg(project_id)
  AND (
      sqlc.narg(status)::alert_incident_status IS NULL
      OR alert_incidents.status = sqlc.narg(status)::alert_incident_status
  )
  AND (
      sqlc.narg(rule_id)::uuid IS NULL
      OR alert_incidents.rule_id = sqlc.narg(rule_id)::uuid
  )
  AND (
      sqlc.narg(probe_id)::uuid IS NULL
      OR alert_incidents.probe_id = sqlc.narg(probe_id)::uuid
  )
  AND (
      sqlc.narg(check_id)::uuid IS NULL
      OR alert_incidents.check_id = sqlc.narg(check_id)::uuid
  )
ORDER BY alert_incidents.opened_at DESC, alert_incidents.id DESC
LIMIT sqlc.arg(limit_count);

-- name: GetAlertIncident :one
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
       alert_incidents.acknowledged_at,
       alert_incidents.acknowledged_by_user_id,
       alert_incidents.resolved_at,
       alert_incidents.resolved_by_user_id,
       alert_incidents.last_evaluated_at,
       alert_incidents.last_triggered_at,
       alert_incidents.last_value,
       alert_incidents.last_summary,
       alert_incidents.last_notification_sent_at,
       alert_incidents.next_notification_eligible_at,
       alert_incidents.suppressed_notification_count,
       alert_incidents.created_at,
       alert_incidents.updated_at,
       probes.name AS probe_name,
       checks.name AS check_name,
       checks.check_type AS check_summary_type,
       checks.target AS check_target
FROM alert_incidents
JOIN probes
  ON probes.project_id = alert_incidents.project_id
 AND probes.id = alert_incidents.probe_id
JOIN checks
  ON checks.project_id = alert_incidents.project_id
 AND checks.id = alert_incidents.check_id
WHERE alert_incidents.project_id = sqlc.arg(project_id)
  AND alert_incidents.id = sqlc.arg(id);

-- name: ListEnabledAlertRulesForAssignment :many
SELECT alert_rules.id,
       alert_rules.project_id,
       alert_rules.name,
       alert_rules.description,
       alert_rules.status,
       alert_rules.severity,
       alert_rules.check_type,
       alert_rules.probe_id,
       alert_rules.check_id,
       alert_rules.probe_selector,
       alert_rules.condition,
       alert_rules.condition_version,
       alert_rules.cooldown_seconds,
       alert_rules.created_by_user_id,
       alert_rules.created_at,
       alert_rules.updated_at,
       alert_rules.deleted_at
FROM alert_rules
WHERE alert_rules.project_id = sqlc.arg(project_id)
  AND alert_rules.check_type = sqlc.arg(check_type)
  AND alert_rules.status = 'enabled'
  AND alert_rules.deleted_at IS NULL
  AND (alert_rules.probe_id IS NULL OR alert_rules.probe_id = sqlc.arg(probe_id))
  AND (alert_rules.check_id IS NULL OR alert_rules.check_id = sqlc.arg(check_id))
ORDER BY alert_rules.created_at ASC, alert_rules.id ASC;

-- name: GetActiveAlertIncident :one
SELECT id,
       project_id,
       rule_id,
       probe_id,
       check_id,
       check_type,
       status,
       severity,
       last_evaluation_state,
       opened_at,
       acknowledged_at,
       acknowledged_by_user_id,
       resolved_at,
       resolved_by_user_id,
       last_evaluated_at,
       last_triggered_at,
       last_value,
       last_summary,
       last_notification_sent_at,
       next_notification_eligible_at,
       suppressed_notification_count,
       created_at,
       updated_at
FROM alert_incidents
WHERE rule_id = sqlc.arg(rule_id)
  AND probe_id = sqlc.arg(probe_id)
  AND check_id = sqlc.arg(check_id)
  AND status IN ('open', 'acknowledged')
LIMIT 1;

-- name: GetRecentResolvedAlertIncident :one
SELECT id,
       project_id,
       rule_id,
       probe_id,
       check_id,
       check_type,
       status,
       severity,
       last_evaluation_state,
       opened_at,
       acknowledged_at,
       acknowledged_by_user_id,
       resolved_at,
       resolved_by_user_id,
       last_evaluated_at,
       last_triggered_at,
       last_value,
       last_summary,
       last_notification_sent_at,
       next_notification_eligible_at,
       suppressed_notification_count,
       created_at,
       updated_at
FROM alert_incidents
WHERE rule_id = sqlc.arg(rule_id)
  AND probe_id = sqlc.arg(probe_id)
  AND check_id = sqlc.arg(check_id)
  AND status = 'resolved'
  AND resolved_at >= sqlc.arg(resolved_after)
ORDER BY resolved_at DESC
LIMIT 1;

-- name: CreateAlertIncident :one
INSERT INTO alert_incidents (
    project_id,
    rule_id,
    probe_id,
    check_id,
    check_type,
    status,
    severity,
    last_evaluation_state,
    opened_at,
    last_evaluated_at,
    last_triggered_at,
    last_value,
    last_summary,
    last_notification_sent_at,
    next_notification_eligible_at
)
VALUES (
    sqlc.arg(project_id),
    sqlc.arg(rule_id),
    sqlc.arg(probe_id),
    sqlc.arg(check_id),
    sqlc.arg(check_type),
    'open',
    sqlc.arg(severity),
    sqlc.arg(last_evaluation_state),
    sqlc.arg(opened_at),
    sqlc.arg(last_evaluated_at),
    sqlc.arg(last_triggered_at),
    sqlc.narg(last_value),
    sqlc.arg(last_summary)::jsonb,
    sqlc.narg(last_notification_sent_at),
    sqlc.narg(next_notification_eligible_at)
)
ON CONFLICT (rule_id, probe_id, check_id) WHERE status IN ('open', 'acknowledged') DO NOTHING
RETURNING id, project_id, rule_id, probe_id, check_id, check_type, status, severity, last_evaluation_state, opened_at, acknowledged_at, acknowledged_by_user_id, resolved_at, resolved_by_user_id, last_evaluated_at, last_triggered_at, last_value, last_summary, last_notification_sent_at, next_notification_eligible_at, suppressed_notification_count, created_at, updated_at;

-- name: UpdateActiveAlertIncidentTriggered :one
UPDATE alert_incidents
SET last_evaluation_state = 'firing',
    last_evaluated_at = sqlc.arg(last_evaluated_at),
    last_triggered_at = sqlc.arg(last_triggered_at),
    last_value = sqlc.narg(last_value),
    last_summary = sqlc.arg(last_summary)::jsonb
WHERE id = sqlc.arg(id)
  AND status IN ('open', 'acknowledged')
RETURNING id, project_id, rule_id, probe_id, check_id, check_type, status, severity, last_evaluation_state, opened_at, acknowledged_at, acknowledged_by_user_id, resolved_at, resolved_by_user_id, last_evaluated_at, last_triggered_at, last_value, last_summary, last_notification_sent_at, next_notification_eligible_at, suppressed_notification_count, created_at, updated_at;

-- name: UpdateActiveAlertIncidentInsufficient :one
UPDATE alert_incidents
SET last_evaluation_state = sqlc.arg(last_evaluation_state),
    last_evaluated_at = sqlc.arg(last_evaluated_at),
    last_summary = sqlc.arg(last_summary)::jsonb
WHERE id = sqlc.arg(id)
  AND status IN ('open', 'acknowledged')
RETURNING id, project_id, rule_id, probe_id, check_id, check_type, status, severity, last_evaluation_state, opened_at, acknowledged_at, acknowledged_by_user_id, resolved_at, resolved_by_user_id, last_evaluated_at, last_triggered_at, last_value, last_summary, last_notification_sent_at, next_notification_eligible_at, suppressed_notification_count, created_at, updated_at;

-- name: ResolveActiveAlertIncident :one
UPDATE alert_incidents
SET status = 'resolved',
    last_evaluation_state = 'clear',
    resolved_at = sqlc.arg(resolved_at),
    resolved_by_user_id = NULL,
    last_evaluated_at = sqlc.arg(last_evaluated_at),
    last_summary = sqlc.arg(last_summary)::jsonb
WHERE id = sqlc.arg(id)
  AND status IN ('open', 'acknowledged')
RETURNING id, project_id, rule_id, probe_id, check_id, check_type, status, severity, last_evaluation_state, opened_at, acknowledged_at, acknowledged_by_user_id, resolved_at, resolved_by_user_id, last_evaluated_at, last_triggered_at, last_value, last_summary, last_notification_sent_at, next_notification_eligible_at, suppressed_notification_count, created_at, updated_at;

-- name: UpdateAlertIncidentNotificationSent :exec
UPDATE alert_incidents
SET last_notification_sent_at = sqlc.arg(sent_at),
    next_notification_eligible_at = sqlc.arg(next_eligible_at)
WHERE id = sqlc.arg(id);

-- name: IncrementAlertIncidentSuppressedNotifications :exec
UPDATE alert_incidents
SET suppressed_notification_count = suppressed_notification_count + 1
WHERE id = sqlc.arg(id);

-- name: GetPingAlertMetricSummary :one
SELECT count(*) FILTER (
           WHERE CASE
               WHEN sqlc.arg(metric)::text IN ('ping.average_rtt_ms') THEN rtt_avg_ms IS NOT NULL
               WHEN sqlc.arg(metric)::text IN ('ping.max_rtt_ms') THEN rtt_max_ms IS NOT NULL
               ELSE true
           END
       )::bigint AS samples,
       coalesce(CASE sqlc.arg(metric)::text
           WHEN 'ping.loss_percent' THEN avg(loss_percent)::double precision
           WHEN 'ping.success_rate' THEN (100.0 * count(*) FILTER (WHERE status = 'successful') / NULLIF(count(*), 0))::double precision
           WHEN 'ping.average_rtt_ms' THEN avg(rtt_avg_ms)::double precision
           WHEN 'ping.max_rtt_ms' THEN max(rtt_max_ms)::double precision
           ELSE NULL::double precision
       END, 0)::double precision AS value
FROM ping_results
WHERE probe_id = sqlc.arg(probe_storage_id)
  AND check_id = sqlc.arg(check_storage_id)
  AND started_at >= sqlc.arg(started_at_from)
  AND started_at < sqlc.arg(started_at_to);

-- name: GetTCPAlertMetricSummary :one
SELECT count(*) FILTER (
           WHERE CASE
               WHEN sqlc.arg(metric)::text IN ('tcp.average_connect_ms', 'tcp.max_connect_ms') THEN connect_duration_ms IS NOT NULL
               ELSE true
           END
       )::bigint AS samples,
       coalesce(CASE sqlc.arg(metric)::text
           WHEN 'tcp.failure_percent' THEN (100.0 * count(*) FILTER (WHERE status IN ('timeout', 'error')) / NULLIF(count(*), 0))::double precision
           WHEN 'tcp.success_rate' THEN (100.0 * count(*) FILTER (WHERE status = 'successful') / NULLIF(count(*), 0))::double precision
           WHEN 'tcp.average_connect_ms' THEN avg(connect_duration_ms)::double precision
           WHEN 'tcp.max_connect_ms' THEN max(connect_duration_ms)::double precision
           ELSE NULL::double precision
       END, 0)::double precision AS value
FROM tcp_results
WHERE probe_id = sqlc.arg(probe_storage_id)
  AND check_id = sqlc.arg(check_storage_id)
  AND started_at >= sqlc.arg(started_at_from)
  AND started_at < sqlc.arg(started_at_to);

-- name: GetHTTPAlertMetricSummary :one
SELECT count(*) FILTER (
           WHERE CASE
               WHEN sqlc.arg(metric)::text IN ('http.average_ttfb_ms', 'http.max_ttfb_ms') THEN ttfb_duration_ms IS NOT NULL
               WHEN sqlc.arg(metric)::text = 'http.certificate_days_remaining' THEN certificate_not_after IS NOT NULL
               ELSE true
           END
       )::bigint AS samples,
       coalesce(CASE sqlc.arg(metric)::text
           WHEN 'http.failure_percent' THEN (100.0 * count(*) FILTER (WHERE status IN ('timeout', 'error')) / NULLIF(count(*), 0))::double precision
           WHEN 'http.success_rate' THEN (100.0 * count(*) FILTER (WHERE status = 'successful') / NULLIF(count(*), 0))::double precision
           WHEN 'http.average_total_ms' THEN avg(duration_ms)::double precision
           WHEN 'http.max_total_ms' THEN max(duration_ms)::double precision
           WHEN 'http.average_ttfb_ms' THEN avg(ttfb_duration_ms)::double precision
           WHEN 'http.max_ttfb_ms' THEN max(ttfb_duration_ms)::double precision
           WHEN 'http.certificate_days_remaining' THEN min(extract(epoch FROM (certificate_not_after - now())) / 86400.0)::double precision
           ELSE NULL::double precision
       END, 0)::double precision AS value
FROM http_results
WHERE probe_id = sqlc.arg(probe_storage_id)
  AND check_id = sqlc.arg(check_storage_id)
  AND started_at >= sqlc.arg(started_at_from)
  AND started_at < sqlc.arg(started_at_to);

-- name: EnqueueNotificationOutbox :one
INSERT INTO notification_outbox (
    project_id,
    incident_id,
    rule_id,
    notification_id,
    notification_type,
    event_type,
    payload,
    dedupe_key
)
VALUES (
    sqlc.arg(project_id),
    sqlc.arg(incident_id),
    sqlc.arg(rule_id),
    sqlc.arg(notification_id),
    sqlc.arg(notification_type),
    sqlc.arg(event_type),
    sqlc.arg(payload)::jsonb,
    sqlc.arg(dedupe_key)
)
ON CONFLICT (dedupe_key) DO NOTHING
RETURNING id;

-- name: ListEnabledNotificationsForRule :many
SELECT notifications.id,
       notifications.project_id,
       notifications.name,
       notifications.type,
       notifications.enabled,
       notifications.config,
       notifications.created_by_user_id,
       notifications.created_at,
       notifications.updated_at,
       notifications.deleted_at
FROM alert_notifications
JOIN notifications
    ON notifications.project_id = alert_notifications.project_id
    AND notifications.id = alert_notifications.notification_id
WHERE alert_notifications.project_id = sqlc.arg(project_id)
  AND alert_notifications.rule_id = sqlc.arg(rule_id)
  AND notifications.enabled = true
  AND notifications.deleted_at IS NULL
ORDER BY alert_notifications.created_at ASC, notifications.id ASC;

-- name: RecoverStaleNotificationOutbox :execrows
UPDATE notification_outbox
SET status = 'pending',
    next_attempt_at = now()
WHERE status = 'sending'
  AND last_attempt_at < sqlc.arg(stale_before);

-- name: ClaimNotificationOutbox :many
WITH selected AS (
    SELECT id
    FROM notification_outbox
    WHERE status = 'pending'
      AND next_attempt_at <= now()
    ORDER BY next_attempt_at ASC, created_at ASC
    LIMIT sqlc.arg(limit_count)
    FOR UPDATE SKIP LOCKED
)
UPDATE notification_outbox
SET status = 'sending',
    last_attempt_at = now()
FROM selected
WHERE notification_outbox.id = selected.id
RETURNING notification_outbox.id,
          notification_outbox.project_id,
          notification_outbox.incident_id,
          notification_outbox.rule_id,
          notification_outbox.notification_id,
          notification_outbox.notification_type,
          notification_outbox.event_type,
          notification_outbox.status,
          notification_outbox.payload,
          notification_outbox.attempt_count,
          notification_outbox.max_attempts,
          notification_outbox.next_attempt_at,
          notification_outbox.last_attempt_at,
          notification_outbox.delivered_at,
          notification_outbox.last_error_kind,
          notification_outbox.last_error_code,
          notification_outbox.last_error,
          notification_outbox.dedupe_key,
          notification_outbox.created_at,
          notification_outbox.updated_at;

-- name: MarkNotificationOutboxDelivered :exec
UPDATE notification_outbox
SET status = 'delivered',
    delivered_at = sqlc.arg(delivered_at),
    last_error_kind = NULL,
    last_error_code = NULL,
    last_error = NULL
WHERE id = sqlc.arg(id);

-- name: MarkNotificationOutboxRetry :exec
UPDATE notification_outbox
SET status = 'pending',
    attempt_count = attempt_count + 1,
    next_attempt_at = sqlc.arg(next_attempt_at),
    last_error_kind = sqlc.arg(last_error_kind),
    last_error_code = sqlc.arg(last_error_code),
    last_error = sqlc.arg(last_error)
WHERE id = sqlc.arg(id);

-- name: MarkNotificationOutboxFailed :exec
UPDATE notification_outbox
SET status = 'failed',
    attempt_count = attempt_count + 1,
    last_error_kind = sqlc.arg(last_error_kind),
    last_error_code = sqlc.arg(last_error_code),
    last_error = sqlc.arg(last_error)
WHERE id = sqlc.arg(id);

-- name: MarkNotificationOutboxDiscarded :exec
UPDATE notification_outbox
SET status = 'discarded',
    attempt_count = attempt_count + 1,
    last_error_kind = sqlc.arg(last_error_kind),
    last_error_code = sqlc.arg(last_error_code),
    last_error = sqlc.arg(last_error)
WHERE id = sqlc.arg(id);

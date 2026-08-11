-- +goose Up

CREATE TYPE public.alert_incident_resolution_reason AS ENUM (
    'condition_cleared',
    'target_no_longer_evaluated'
);

ALTER TABLE public.alert_incidents
    ADD COLUMN resolution_reason public.alert_incident_resolution_reason;

UPDATE public.alert_incidents
SET resolution_reason = 'condition_cleared'
WHERE status = 'resolved';

WITH invalid_incidents AS (
    SELECT alert_incidents.id,
           COALESCE(
               GREATEST(
                   probes.deleted_at,
                   checks.deleted_at,
                   alert_rules.deleted_at,
                   CASE WHEN probes.enabled = false THEN probes.updated_at END,
                   CASE WHEN alert_rules.status <> 'enabled' THEN alert_rules.updated_at END
               ),
               now()
           ) AS resolved_at
    FROM public.alert_incidents
    JOIN public.probes
      ON probes.project_id = alert_incidents.project_id
     AND probes.id = alert_incidents.probe_id
    JOIN public.checks
      ON checks.project_id = alert_incidents.project_id
     AND checks.id = alert_incidents.check_id
    JOIN public.alert_rules
      ON alert_rules.project_id = alert_incidents.project_id
     AND alert_rules.id = alert_incidents.rule_id
    WHERE alert_incidents.status IN ('open', 'acknowledged')
      AND (
          probes.deleted_at IS NOT NULL
          OR probes.enabled = false
          OR checks.deleted_at IS NOT NULL
          OR alert_rules.deleted_at IS NOT NULL
          OR alert_rules.status <> 'enabled'
          OR alert_rules.check_type <> checks.check_type
          OR (alert_rules.probe_id IS NOT NULL AND alert_rules.probe_id <> alert_incidents.probe_id)
          OR (alert_rules.check_id IS NOT NULL AND alert_rules.check_id <> alert_incidents.check_id)
          OR NOT EXISTS (
              SELECT 1
              FROM public.probe_check_assignments
              WHERE probe_check_assignments.project_id = alert_incidents.project_id
                AND probe_check_assignments.probe_id = alert_incidents.probe_id
                AND probe_check_assignments.check_id = alert_incidents.check_id
                AND probe_check_assignments.deleted_at IS NULL
          )
      )
)
UPDATE public.alert_incidents
SET status = 'resolved',
    resolution_reason = 'target_no_longer_evaluated',
    last_evaluation_state = 'no_data',
    resolved_at = invalid_incidents.resolved_at,
    resolved_by_user_id = NULL,
    last_summary = jsonb_set(last_summary, '{state}', '"no_data"'::jsonb, true)
FROM invalid_incidents
WHERE alert_incidents.id = invalid_incidents.id;

DELETE FROM public.alert_rule_pending_evaluations
WHERE EXISTS (
    SELECT 1
    FROM public.probes
    JOIN public.checks
      ON checks.project_id = alert_rule_pending_evaluations.project_id
     AND checks.id = alert_rule_pending_evaluations.check_id
    JOIN public.alert_rules
      ON alert_rules.project_id = alert_rule_pending_evaluations.project_id
     AND alert_rules.id = alert_rule_pending_evaluations.rule_id
    WHERE probes.project_id = alert_rule_pending_evaluations.project_id
      AND probes.id = alert_rule_pending_evaluations.probe_id
      AND (
          probes.deleted_at IS NOT NULL
          OR probes.enabled = false
          OR checks.deleted_at IS NOT NULL
          OR alert_rules.deleted_at IS NOT NULL
          OR alert_rules.status <> 'enabled'
          OR alert_rules.check_type <> checks.check_type
          OR (alert_rules.probe_id IS NOT NULL AND alert_rules.probe_id <> alert_rule_pending_evaluations.probe_id)
          OR (alert_rules.check_id IS NOT NULL AND alert_rules.check_id <> alert_rule_pending_evaluations.check_id)
          OR NOT EXISTS (
              SELECT 1
              FROM public.probe_check_assignments
              WHERE probe_check_assignments.project_id = alert_rule_pending_evaluations.project_id
                AND probe_check_assignments.probe_id = alert_rule_pending_evaluations.probe_id
                AND probe_check_assignments.check_id = alert_rule_pending_evaluations.check_id
                AND probe_check_assignments.deleted_at IS NULL
          )
      )
);

ALTER TABLE public.alert_incidents
    ADD CONSTRAINT alert_incidents_resolution_reason_consistency CHECK (
        (status = 'resolved' AND resolution_reason IS NOT NULL)
        OR (status <> 'resolved' AND resolution_reason IS NULL)
    );

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    RAISE EXCEPTION 'alert incident resolution-reason migration is irreversible';
END;
$$;
-- +goose StatementEnd

-- +goose Up
ALTER TABLE alert_rules
    ADD COLUMN trigger_after_seconds integer NOT NULL DEFAULT 60,
    ADD CONSTRAINT alert_rules_trigger_after_seconds_range
        CHECK (trigger_after_seconds >= 60 AND trigger_after_seconds <= 86400),
    ADD CONSTRAINT alert_rules_trigger_after_seconds_whole_minutes
        CHECK (trigger_after_seconds % 60 = 0);

CREATE TABLE alert_rule_pending_evaluations (
    project_id uuid NOT NULL,
    rule_id uuid NOT NULL,
    probe_id uuid NOT NULL,
    check_id uuid NOT NULL,
    firing_since timestamptz NOT NULL,
    PRIMARY KEY (rule_id, probe_id, check_id),
    CONSTRAINT fk_alert_rule_pending_evaluations_project_rule
        FOREIGN KEY (project_id, rule_id) REFERENCES alert_rules(project_id, id) ON DELETE CASCADE,
    CONSTRAINT fk_alert_rule_pending_evaluations_project_probe
        FOREIGN KEY (project_id, probe_id) REFERENCES probes(project_id, id) ON DELETE CASCADE,
    CONSTRAINT fk_alert_rule_pending_evaluations_project_check
        FOREIGN KEY (project_id, check_id) REFERENCES checks(project_id, id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS alert_rule_pending_evaluations;

ALTER TABLE alert_rules
    DROP COLUMN IF EXISTS trigger_after_seconds;

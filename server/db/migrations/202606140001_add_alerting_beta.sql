-- +goose Up
CREATE TYPE alert_severity AS ENUM ('info', 'warning', 'critical');
CREATE TYPE alert_rule_status AS ENUM ('enabled', 'disabled');
CREATE TYPE alert_incident_status AS ENUM ('open', 'acknowledged', 'resolved');
CREATE TYPE alert_evaluation_state AS ENUM ('firing', 'clear', 'insufficient_samples', 'no_data', 'error');
CREATE TYPE notification_channel_type AS ENUM ('webhook', 'email');
CREATE TYPE notification_outbox_status AS ENUM ('pending', 'sending', 'delivered', 'failed', 'discarded');

CREATE TABLE alert_rules (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id uuid NOT NULL REFERENCES projects(id),
    name text NOT NULL,
    description text,
    status alert_rule_status NOT NULL DEFAULT 'enabled',
    severity alert_severity NOT NULL,
    check_type check_type NOT NULL,
    probe_id uuid,
    check_id uuid,
    probe_selector jsonb NOT NULL DEFAULT '{}'::jsonb,
    condition jsonb NOT NULL,
    condition_version text NOT NULL,
    cooldown_seconds integer NOT NULL DEFAULT 900,
    created_by_user_id uuid NOT NULL REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT uq_alert_rules_project_id_id UNIQUE (project_id, id),
    CONSTRAINT alert_rules_name_not_empty CHECK (length(btrim(name)) > 0),
    CONSTRAINT alert_rules_description_not_empty CHECK (description IS NULL OR length(btrim(description)) > 0),
    CONSTRAINT alert_rules_probe_selector_is_object CHECK (jsonb_typeof(probe_selector) = 'object'),
    CONSTRAINT alert_rules_condition_is_object CHECK (jsonb_typeof(condition) = 'object'),
    CONSTRAINT alert_rules_condition_version_not_empty CHECK (length(btrim(condition_version)) > 0),
    CONSTRAINT alert_rules_cooldown_seconds_range CHECK (cooldown_seconds >= 60 AND cooldown_seconds <= 86400),
    CONSTRAINT alert_rules_deleted_at_after_created_at CHECK (deleted_at IS NULL OR deleted_at >= created_at),
    CONSTRAINT fk_alert_rules_project_probe
        FOREIGN KEY (project_id, probe_id) REFERENCES probes(project_id, id),
    CONSTRAINT fk_alert_rules_project_check
        FOREIGN KEY (project_id, check_id) REFERENCES checks(project_id, id)
);

CREATE INDEX idx_alert_rules_project_active
    ON alert_rules (project_id)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_alert_rules_project_check_type_status_active
    ON alert_rules (project_id, check_type, status)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_alert_rules_project_probe_active
    ON alert_rules (project_id, probe_id)
    WHERE deleted_at IS NULL AND probe_id IS NOT NULL;
CREATE INDEX idx_alert_rules_project_check_active
    ON alert_rules (project_id, check_id)
    WHERE deleted_at IS NULL AND check_id IS NOT NULL;

CREATE TRIGGER set_alert_rules_updated_at
    BEFORE UPDATE ON alert_rules
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE notification_channels (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id uuid NOT NULL REFERENCES projects(id),
    name text NOT NULL,
    type notification_channel_type NOT NULL,
    enabled boolean NOT NULL DEFAULT true,
    config jsonb NOT NULL,
    created_by_user_id uuid NOT NULL REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT uq_notification_channels_project_id_id UNIQUE (project_id, id),
    CONSTRAINT notification_channels_name_not_empty CHECK (length(btrim(name)) > 0),
    CONSTRAINT notification_channels_config_is_object CHECK (jsonb_typeof(config) = 'object'),
    CONSTRAINT notification_channels_deleted_at_after_created_at CHECK (deleted_at IS NULL OR deleted_at >= created_at)
);

CREATE INDEX idx_notification_channels_project_active
    ON notification_channels (project_id)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_notification_channels_project_enabled_active
    ON notification_channels (project_id, enabled)
    WHERE deleted_at IS NULL;

CREATE TRIGGER set_notification_channels_updated_at
    BEFORE UPDATE ON notification_channels
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE alert_rule_channels (
    project_id uuid NOT NULL REFERENCES projects(id),
    rule_id uuid NOT NULL,
    channel_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (rule_id, channel_id),
    CONSTRAINT fk_alert_rule_channels_project_rule
        FOREIGN KEY (project_id, rule_id) REFERENCES alert_rules(project_id, id) ON DELETE CASCADE,
    CONSTRAINT fk_alert_rule_channels_project_channel
        FOREIGN KEY (project_id, channel_id) REFERENCES notification_channels(project_id, id)
);

CREATE TABLE alert_incidents (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id uuid NOT NULL REFERENCES projects(id),
    rule_id uuid NOT NULL,
    probe_id uuid NOT NULL,
    check_id uuid NOT NULL,
    check_type check_type NOT NULL,
    status alert_incident_status NOT NULL,
    severity alert_severity NOT NULL,
    last_evaluation_state alert_evaluation_state NOT NULL,
    opened_at timestamptz NOT NULL,
    acknowledged_at timestamptz,
    acknowledged_by_user_id uuid REFERENCES users(id),
    resolved_at timestamptz,
    resolved_by_user_id uuid REFERENCES users(id),
    last_evaluated_at timestamptz NOT NULL,
    last_triggered_at timestamptz NOT NULL,
    last_value double precision,
    last_summary jsonb NOT NULL DEFAULT '{}'::jsonb,
    last_notification_sent_at timestamptz,
    next_notification_eligible_at timestamptz,
    suppressed_notification_count integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uq_alert_incidents_project_id_id UNIQUE (project_id, id),
    CONSTRAINT alert_incidents_last_summary_is_object CHECK (jsonb_typeof(last_summary) = 'object'),
    CONSTRAINT alert_incidents_suppressed_notification_count_non_negative CHECK (suppressed_notification_count >= 0),
    CONSTRAINT alert_incidents_acknowledged_consistency CHECK (
        (status = 'acknowledged' AND acknowledged_at IS NOT NULL)
        OR status <> 'acknowledged'
    ),
    CONSTRAINT alert_incidents_resolved_consistency CHECK (
        (status = 'resolved' AND resolved_at IS NOT NULL)
        OR status <> 'resolved'
    ),
    CONSTRAINT fk_alert_incidents_project_rule
        FOREIGN KEY (project_id, rule_id) REFERENCES alert_rules(project_id, id),
    CONSTRAINT fk_alert_incidents_project_probe
        FOREIGN KEY (project_id, probe_id) REFERENCES probes(project_id, id),
    CONSTRAINT fk_alert_incidents_project_check
        FOREIGN KEY (project_id, check_id) REFERENCES checks(project_id, id)
);

CREATE INDEX idx_alert_incidents_project_status_opened
    ON alert_incidents (project_id, status, opened_at DESC);
CREATE INDEX idx_alert_incidents_project_rule_status
    ON alert_incidents (project_id, rule_id, status);
CREATE INDEX idx_alert_incidents_project_probe_check_status
    ON alert_incidents (project_id, probe_id, check_id, status);
CREATE INDEX idx_alert_incidents_rule_target_resolved
    ON alert_incidents (rule_id, probe_id, check_id, resolved_at DESC)
    WHERE status = 'resolved';
CREATE UNIQUE INDEX uq_alert_incidents_active_rule_probe_check
    ON alert_incidents (rule_id, probe_id, check_id)
    WHERE status IN ('open', 'acknowledged');

CREATE TRIGGER set_alert_incidents_updated_at
    BEFORE UPDATE ON alert_incidents
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE notification_outbox (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id uuid NOT NULL REFERENCES projects(id),
    incident_id uuid NOT NULL,
    rule_id uuid NOT NULL,
    channel_id uuid NOT NULL,
    channel_type notification_channel_type NOT NULL,
    event_type text NOT NULL,
    status notification_outbox_status NOT NULL DEFAULT 'pending',
    payload jsonb NOT NULL,
    attempt_count integer NOT NULL DEFAULT 0,
    max_attempts integer NOT NULL DEFAULT 5,
    next_attempt_at timestamptz NOT NULL DEFAULT now(),
    last_attempt_at timestamptz,
    delivered_at timestamptz,
    last_error_kind text,
    last_error_code text,
    last_error text,
    dedupe_key text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uq_notification_outbox_project_id_id UNIQUE (project_id, id),
    CONSTRAINT notification_outbox_event_type_not_empty CHECK (length(btrim(event_type)) > 0),
    CONSTRAINT notification_outbox_payload_is_object CHECK (jsonb_typeof(payload) = 'object'),
    CONSTRAINT notification_outbox_attempt_count_non_negative CHECK (attempt_count >= 0),
    CONSTRAINT notification_outbox_max_attempts_positive CHECK (max_attempts > 0),
    CONSTRAINT notification_outbox_last_error_kind_not_empty CHECK (last_error_kind IS NULL OR length(btrim(last_error_kind)) > 0),
    CONSTRAINT notification_outbox_last_error_code_not_empty CHECK (last_error_code IS NULL OR length(btrim(last_error_code)) > 0),
    CONSTRAINT notification_outbox_last_error_not_empty CHECK (last_error IS NULL OR length(btrim(last_error)) > 0),
    CONSTRAINT notification_outbox_dedupe_key_not_empty CHECK (length(btrim(dedupe_key)) > 0),
    CONSTRAINT fk_notification_outbox_project_incident
        FOREIGN KEY (project_id, incident_id) REFERENCES alert_incidents(project_id, id),
    CONSTRAINT fk_notification_outbox_project_rule
        FOREIGN KEY (project_id, rule_id) REFERENCES alert_rules(project_id, id),
    CONSTRAINT fk_notification_outbox_project_channel
        FOREIGN KEY (project_id, channel_id) REFERENCES notification_channels(project_id, id)
);

CREATE INDEX idx_notification_outbox_status_next_attempt
    ON notification_outbox (status, next_attempt_at);
CREATE INDEX idx_notification_outbox_status_last_attempt
    ON notification_outbox (status, last_attempt_at);
CREATE INDEX idx_notification_outbox_project_created
    ON notification_outbox (project_id, created_at DESC);
CREATE UNIQUE INDEX uq_notification_outbox_dedupe_key
    ON notification_outbox (dedupe_key);

CREATE TRIGGER set_notification_outbox_updated_at
    BEFORE UPDATE ON notification_outbox
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TABLE IF EXISTS notification_outbox;
DROP TABLE IF EXISTS alert_incidents;
DROP TABLE IF EXISTS alert_rule_channels;
DROP TABLE IF EXISTS notification_channels;
DROP TABLE IF EXISTS alert_rules;

DROP TYPE IF EXISTS notification_outbox_status;
DROP TYPE IF EXISTS notification_channel_type;
DROP TYPE IF EXISTS alert_evaluation_state;
DROP TYPE IF EXISTS alert_incident_status;
DROP TYPE IF EXISTS alert_rule_status;
DROP TYPE IF EXISTS alert_severity;

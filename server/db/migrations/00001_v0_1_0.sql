-- +goose NO TRANSACTION
-- +goose Up

CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Columnstore support for continuous aggregates requires TimescaleDB 2.20.0.
-- Require the latest 2.20 patch release so fresh and upgraded installations use
-- the fixes shipped before Netstamp starts managing these policies.
-- +goose StatementBegin
DO $$
DECLARE
    installed_version text;
    version_parts text[];
BEGIN
    SELECT extversion
    INTO installed_version
    FROM pg_extension
    WHERE extname = 'timescaledb';

    version_parts := regexp_match(installed_version, '^([0-9]+)\.([0-9]+)\.([0-9]+)');

    IF version_parts IS NULL OR
       (version_parts[1]::integer, version_parts[2]::integer, version_parts[3]::integer) < (2, 20, 3) THEN
        RAISE EXCEPTION
            'Netstamp columnstore policies require TimescaleDB 2.20.3 or newer; installed version is %',
            coalesce(installed_version, 'unknown')
            USING HINT = 'Upgrade the TimescaleDB image, run ALTER EXTENSION timescaledb UPDATE, then retry the migration.';
    END IF;
END;
$$;
-- +goose StatementEnd

-- Netstamp v0.1.0 core schema.
-- Prerequisites: the public schema and pgcrypto/citext extensions already exist.
-- Time-series result tables, rollups, jobs, retention, and columnstore follow below.
-- This fresh-install baseline intentionally omits Goose's version table and pre-release rollback artifacts.

-- TYPE: alert_evaluation_state
CREATE TYPE public.alert_evaluation_state AS ENUM (
    'firing',
    'clear',
    'insufficient_samples',
    'no_data',
    'error'
);

-- TYPE: alert_incident_status
CREATE TYPE public.alert_incident_status AS ENUM (
    'open',
    'acknowledged',
    'resolved'
);

-- TYPE: alert_rule_status
CREATE TYPE public.alert_rule_status AS ENUM (
    'enabled',
    'disabled'
);

-- TYPE: alert_severity
CREATE TYPE public.alert_severity AS ENUM (
    'info',
    'warning',
    'critical'
);

-- TYPE: assignment_refresh_job_status
CREATE TYPE public.assignment_refresh_job_status AS ENUM (
    'pending',
    'running',
    'succeeded',
    'failed',
    'discarded'
);

-- TYPE: assignment_refresh_target_type
CREATE TYPE public.assignment_refresh_target_type AS ENUM (
    'project',
    'probe',
    'check',
    'label'
);

-- TYPE: check_type
CREATE TYPE public.check_type AS ENUM (
    'ping',
    'traceroute',
    'tcp',
    'http'
);

-- TYPE: http_method
CREATE TYPE public.http_method AS ENUM (
    'GET',
    'HEAD',
    'POST',
    'PUT',
    'PATCH',
    'DELETE',
    'OPTIONS'
);

-- TYPE: http_status
CREATE TYPE public.http_status AS ENUM (
    'successful',
    'timeout',
    'error'
);

-- TYPE: ip_family
CREATE TYPE public.ip_family AS ENUM (
    'inet',
    'inet6'
);

-- TYPE: notification_outbox_status
CREATE TYPE public.notification_outbox_status AS ENUM (
    'pending',
    'sending',
    'delivered',
    'failed',
    'discarded'
);

-- TYPE: notification_type
CREATE TYPE public.notification_type AS ENUM (
    'webhook',
    'email',
    'discord',
    'telegram',
    'slack'
);

-- TYPE: ping_status
CREATE TYPE public.ping_status AS ENUM (
    'successful',
    'timeout',
    'error'
);

-- TYPE: probe_state
CREATE TYPE public.probe_state AS ENUM (
    'online',
    'offline'
);

-- TYPE: project_invite_status
CREATE TYPE public.project_invite_status AS ENUM (
    'pending',
    'accepted',
    'rejected'
);

-- TYPE: project_member_role
CREATE TYPE public.project_member_role AS ENUM (
    'owner',
    'admin',
    'editor',
    'viewer'
);

-- TYPE: public_status_assignment_selection_mode
-- Keep this type inside a procedural block so sqlc preserves the established
-- opaque mapping for the assignment_selection_mode column.
-- +goose StatementBegin
DO $$
BEGIN
    CREATE TYPE public_status_assignment_selection_mode AS ENUM ('all_check', 'selected_assignments');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END;
$$;
-- +goose StatementEnd

-- TYPE: public_status_chart_mode
CREATE TYPE public.public_status_chart_mode AS ENUM (
    'inherit',
    'off',
    'compact'
);

-- TYPE: public_status_chart_range
CREATE TYPE public.public_status_chart_range AS ENUM (
    '24h',
    '7d',
    '30d'
);

-- TYPE: public_status_element_display_mode
CREATE TYPE public.public_status_element_display_mode AS ENUM (
    'status',
    'history',
    'latency',
    'map'
);

-- TYPE: public_status_element_kind
CREATE TYPE public.public_status_element_kind AS ENUM (
    'folder',
    'check',
    'assignment_group'
);

-- TYPE: public_status_theme
CREATE TYPE public.public_status_theme AS ENUM (
    'light',
    'dark',
    'auto'
);

-- TYPE: system_role
CREATE TYPE public.system_role AS ENUM (
    'admin'
);

-- TYPE: tcp_status
CREATE TYPE public.tcp_status AS ENUM (
    'successful',
    'timeout',
    'error'
);

-- TYPE: traceroute_protocol
CREATE TYPE public.traceroute_protocol AS ENUM (
    'icmp',
    'udp'
);

-- TYPE: traceroute_status
CREATE TYPE public.traceroute_status AS ENUM (
    'successful',
    'timeout',
    'error',
    'partial'
);

-- FUNCTION: set_updated_at()
-- +goose StatementBegin
CREATE FUNCTION public.set_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$;
-- +goose StatementEnd


SET default_tablespace = '';

SET default_table_access_method = heap;

-- TABLE: alert_incidents
CREATE TABLE public.alert_incidents (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    rule_id uuid NOT NULL,
    probe_id uuid NOT NULL,
    check_id uuid NOT NULL,
    check_type public.check_type NOT NULL,
    status public.alert_incident_status NOT NULL,
    severity public.alert_severity NOT NULL,
    last_evaluation_state public.alert_evaluation_state NOT NULL,
    opened_at timestamptz NOT NULL,
    acknowledged_at timestamptz,
    acknowledged_by_user_id uuid,
    resolved_at timestamptz,
    resolved_by_user_id uuid,
    last_evaluated_at timestamptz NOT NULL,
    last_triggered_at timestamptz NOT NULL,
    last_value double precision,
    last_summary jsonb DEFAULT '{}'::jsonb NOT NULL,
    last_notification_sent_at timestamptz,
    next_notification_eligible_at timestamptz,
    suppressed_notification_count integer DEFAULT 0 NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    CONSTRAINT alert_incidents_acknowledged_consistency CHECK ((((status = 'acknowledged'::public.alert_incident_status) AND (acknowledged_at IS NOT NULL)) OR (status <> 'acknowledged'::public.alert_incident_status))),
    CONSTRAINT alert_incidents_last_summary_is_object CHECK ((jsonb_typeof(last_summary) = 'object'::text)),
    CONSTRAINT alert_incidents_resolved_consistency CHECK ((((status = 'resolved'::public.alert_incident_status) AND (resolved_at IS NOT NULL)) OR (status <> 'resolved'::public.alert_incident_status))),
    CONSTRAINT alert_incidents_suppressed_notification_count_non_negative CHECK ((suppressed_notification_count >= 0))
);

-- TABLE: alert_notifications
CREATE TABLE public.alert_notifications (
    project_id uuid NOT NULL,
    rule_id uuid NOT NULL,
    notification_id uuid NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL
);

-- TABLE: alert_rule_pending_evaluations
CREATE TABLE public.alert_rule_pending_evaluations (
    project_id uuid NOT NULL,
    rule_id uuid NOT NULL,
    probe_id uuid NOT NULL,
    check_id uuid NOT NULL,
    firing_since timestamptz NOT NULL
);

-- TABLE: alert_rules
CREATE TABLE public.alert_rules (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    name text NOT NULL,
    description text,
    status public.alert_rule_status DEFAULT 'enabled'::public.alert_rule_status NOT NULL,
    severity public.alert_severity NOT NULL,
    check_type public.check_type NOT NULL,
    probe_id uuid,
    check_id uuid,
    probe_selector jsonb DEFAULT '{}'::jsonb NOT NULL,
    condition jsonb NOT NULL,
    condition_version text NOT NULL,
    cooldown_seconds integer DEFAULT 900 NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    deleted_at timestamptz,
    trigger_after_seconds integer DEFAULT 60 NOT NULL,
    CONSTRAINT alert_rules_condition_is_object CHECK ((jsonb_typeof(condition) = 'object'::text)),
    CONSTRAINT alert_rules_condition_version_not_empty CHECK ((length(btrim(condition_version)) > 0)),
    CONSTRAINT alert_rules_cooldown_seconds_range CHECK (((cooldown_seconds >= 60) AND (cooldown_seconds <= 86400))),
    CONSTRAINT alert_rules_deleted_at_after_created_at CHECK (((deleted_at IS NULL) OR (deleted_at >= created_at))),
    CONSTRAINT alert_rules_description_not_empty CHECK (((description IS NULL) OR (length(btrim(description)) > 0))),
    CONSTRAINT alert_rules_name_not_empty CHECK ((length(btrim(name)) > 0)),
    CONSTRAINT alert_rules_probe_selector_is_object CHECK ((jsonb_typeof(probe_selector) = 'object'::text)),
    CONSTRAINT alert_rules_trigger_after_seconds_range CHECK (((trigger_after_seconds >= 60) AND (trigger_after_seconds <= 86400))),
    CONSTRAINT alert_rules_trigger_after_seconds_whole_minutes CHECK (((trigger_after_seconds % 60) = 0))
);

-- TABLE: api_tokens
CREATE TABLE public.api_tokens (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    name text NOT NULL,
    token_hash bytea NOT NULL,
    token_hint text NOT NULL,
    scopes text[] NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    last_used_at timestamptz,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    revoked_reason text,
    CONSTRAINT api_tokens_expires_after_created CHECK ((expires_at > created_at)),
    CONSTRAINT api_tokens_last_used_after_created CHECK (((last_used_at IS NULL) OR (last_used_at >= created_at))),
    CONSTRAINT api_tokens_name_valid CHECK (((length(btrim(name)) >= 1) AND (length(btrim(name)) <= 100))),
    CONSTRAINT api_tokens_revoked_reason_present CHECK (((revoked_at IS NULL) OR (length(btrim(COALESCE(revoked_reason, ''::text))) > 0))),
    CONSTRAINT api_tokens_scopes_not_empty CHECK ((cardinality(scopes) > 0)),
    CONSTRAINT api_tokens_token_hash_not_empty CHECK ((length(token_hash) > 0)),
    CONSTRAINT api_tokens_token_hint_valid CHECK ((length(token_hint) = 8))
);

-- TABLE: assignment_refresh_jobs
CREATE TABLE public.assignment_refresh_jobs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    target_type public.assignment_refresh_target_type NOT NULL,
    target_id uuid NOT NULL,
    status public.assignment_refresh_job_status DEFAULT 'pending'::public.assignment_refresh_job_status NOT NULL,
    attempt_count integer DEFAULT 0 NOT NULL,
    max_attempts integer DEFAULT 5 NOT NULL,
    next_attempt_at timestamptz DEFAULT now() NOT NULL,
    last_attempt_at timestamptz,
    completed_at timestamptz,
    last_error_kind text,
    last_error_code text,
    last_error text,
    dedupe_key text NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    CONSTRAINT assignment_refresh_jobs_attempt_count_non_negative CHECK ((attempt_count >= 0)),
    CONSTRAINT assignment_refresh_jobs_dedupe_key_not_empty CHECK ((length(btrim(dedupe_key)) > 0)),
    CONSTRAINT assignment_refresh_jobs_last_error_code_not_empty CHECK (((last_error_code IS NULL) OR (length(btrim(last_error_code)) > 0))),
    CONSTRAINT assignment_refresh_jobs_last_error_kind_not_empty CHECK (((last_error_kind IS NULL) OR (length(btrim(last_error_kind)) > 0))),
    CONSTRAINT assignment_refresh_jobs_last_error_not_empty CHECK (((last_error IS NULL) OR (length(btrim(last_error)) > 0))),
    CONSTRAINT assignment_refresh_jobs_max_attempts_positive CHECK ((max_attempts > 0))
);

-- TABLE: auth_sessions
CREATE TABLE public.auth_sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    token_hash bytea NOT NULL,
    csrf_token_hash bytea NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    last_used_at timestamptz DEFAULT now() NOT NULL,
    idle_expires_at timestamptz NOT NULL,
    absolute_expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    revoked_reason text,
    user_agent text DEFAULT ''::text NOT NULL,
    authenticated_at timestamptz NOT NULL,
    authentication_method text NOT NULL,
    identity_id uuid,
    sudo_eligible boolean DEFAULT false NOT NULL,
    CONSTRAINT auth_sessions_absolute_expires_after_created CHECK ((absolute_expires_at > created_at)),
    CONSTRAINT auth_sessions_authenticated_after_created CHECK ((authenticated_at >= created_at)),
    CONSTRAINT auth_sessions_authentication_method_valid CHECK ((authentication_method = ANY (ARRAY['password'::text, 'google'::text, 'github'::text, 'oidc'::text]))),
    CONSTRAINT auth_sessions_csrf_token_hash_not_empty CHECK ((length(csrf_token_hash) > 0)),
    CONSTRAINT auth_sessions_idle_expires_after_created CHECK ((idle_expires_at > created_at)),
    CONSTRAINT auth_sessions_revoked_reason_present CHECK (((revoked_at IS NULL) OR (length(btrim(COALESCE(revoked_reason, ''::text))) > 0))),
    CONSTRAINT auth_sessions_token_hash_not_empty CHECK ((length(token_hash) > 0))
);

-- TABLE: check_labels
CREATE TABLE public.check_labels (
    project_id uuid NOT NULL,
    check_id uuid NOT NULL,
    label_id uuid NOT NULL
);

-- TABLE: checks
CREATE TABLE public.checks (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    internal_id bigint NOT NULL,
    project_id uuid NOT NULL,
    name text NOT NULL,
    check_type public.check_type NOT NULL,
    target text NOT NULL,
    selector jsonb DEFAULT '{}'::jsonb NOT NULL,
    description text,
    interval_seconds integer NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    deleted_at timestamptz,
    CONSTRAINT checks_deleted_at_after_created_at CHECK (((deleted_at IS NULL) OR (deleted_at >= created_at))),
    CONSTRAINT checks_description_not_empty CHECK (((description IS NULL) OR (length(btrim(description)) > 0))),
    CONSTRAINT checks_interval_seconds_positive CHECK ((interval_seconds > 0)),
    CONSTRAINT checks_name_not_empty CHECK ((length(btrim(name)) > 0)),
    CONSTRAINT checks_selector_is_object CHECK ((jsonb_typeof(selector) = 'object'::text)),
    CONSTRAINT checks_target_not_empty CHECK ((length(btrim(target)) > 0))
);

-- SEQUENCE: checks_internal_id_seq
ALTER TABLE public.checks ALTER COLUMN internal_id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.checks_internal_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);

-- TABLE: email_verification_tokens
CREATE TABLE public.email_verification_tokens (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    token_hash text NOT NULL,
    expires_at timestamptz NOT NULL,
    used_at timestamptz,
    created_at timestamptz DEFAULT now() NOT NULL,
    CONSTRAINT email_verification_tokens_expires_after_created CHECK ((expires_at > created_at)),
    CONSTRAINT email_verification_tokens_token_hash_not_empty CHECK ((length(btrim(token_hash)) > 0))
);

-- TABLE: external_auth_flows
CREATE TABLE public.external_auth_flows (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    state_hash bytea NOT NULL,
    browser_token_hash bytea NOT NULL,
    nonce text NOT NULL,
    pkce_verifier text NOT NULL,
    intent text NOT NULL,
    session_id uuid,
    return_to text NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    expires_at timestamptz NOT NULL,
    used_at timestamptz,
    provider text NOT NULL,
    CONSTRAINT external_auth_flows_browser_token_hash_not_empty CHECK ((length(browser_token_hash) > 0)),
    CONSTRAINT external_auth_flows_expires_after_created CHECK ((expires_at > created_at)),
    CONSTRAINT external_auth_flows_intent_valid CHECK ((intent = ANY (ARRAY['login'::text, 'sudo'::text, 'link'::text]))),
    CONSTRAINT external_auth_flows_nonce_not_empty CHECK ((length(nonce) > 0)),
    CONSTRAINT external_auth_flows_pkce_not_empty CHECK ((length(pkce_verifier) > 0)),
    CONSTRAINT external_auth_flows_provider_valid CHECK ((provider = ANY (ARRAY['google'::text, 'github'::text, 'oidc'::text]))),
    CONSTRAINT external_auth_flows_return_to_relative CHECK (((return_to ~~ '/%'::text) AND (return_to !~~ '//%'::text) AND (POSITION((chr(92)) IN (return_to)) = 0) AND (POSITION((chr(10)) IN (return_to)) = 0) AND (POSITION((chr(13)) IN (return_to)) = 0))),
    CONSTRAINT external_auth_flows_session_required CHECK ((((intent = 'login'::text) AND (session_id IS NULL)) OR ((intent = ANY (ARRAY['sudo'::text, 'link'::text])) AND (session_id IS NOT NULL)))),
    CONSTRAINT external_auth_flows_state_hash_not_empty CHECK ((length(state_hash) > 0)),
    CONSTRAINT external_auth_flows_used_after_created CHECK (((used_at IS NULL) OR (used_at >= created_at)))
);

-- TABLE: http_check_configs
CREATE TABLE public.http_check_configs (
    check_id uuid NOT NULL,
    method public.http_method DEFAULT 'GET'::public.http_method NOT NULL,
    headers jsonb DEFAULT '[]'::jsonb NOT NULL,
    body text,
    timeout_ms integer DEFAULT 10000 NOT NULL,
    ip_family public.ip_family,
    follow_redirects boolean DEFAULT true NOT NULL,
    skip_tls_verify boolean DEFAULT false NOT NULL,
    expected_status_codes integer[] DEFAULT '{}'::integer[] NOT NULL,
    expected_status_classes integer[] DEFAULT ARRAY[2, 3] NOT NULL,
    body_contains text,
    CONSTRAINT http_check_configs_body_contains_size CHECK (((body_contains IS NULL) OR ((length(body_contains) >= 1) AND (length(body_contains) <= 1024)))),
    CONSTRAINT http_check_configs_body_size CHECK (((body IS NULL) OR (octet_length(body) <= 65536))),
    CONSTRAINT http_check_configs_expected_status_classes_range CHECK (((1 <= ALL (expected_status_classes)) AND (5 >= ALL (expected_status_classes)))),
    CONSTRAINT http_check_configs_expected_status_codes_range CHECK (((100 <= ALL (expected_status_codes)) AND (599 >= ALL (expected_status_codes)))),
    CONSTRAINT http_check_configs_expected_status_required CHECK (((cardinality(expected_status_codes) > 0) OR (cardinality(expected_status_classes) > 0))),
    CONSTRAINT http_check_configs_headers_array CHECK (((jsonb_typeof(headers) = 'array'::text) AND (jsonb_array_length(headers) <= 50))),
    CONSTRAINT http_check_configs_method_body CHECK (((method <> ALL (ARRAY['GET'::public.http_method, 'HEAD'::public.http_method])) OR (body IS NULL))),
    CONSTRAINT http_check_configs_timeout_range CHECK (((timeout_ms >= 1) AND (timeout_ms <= 60000)))
);

-- TABLE: labels
CREATE TABLE public.labels (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    key text NOT NULL,
    value text NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    deleted_at timestamptz,
    CONSTRAINT labels_deleted_at_after_created_at CHECK (((deleted_at IS NULL) OR (deleted_at >= created_at))),
    CONSTRAINT labels_key_not_empty CHECK ((length(btrim(key)) > 0)),
    CONSTRAINT labels_value_not_empty CHECK ((length(btrim(value)) > 0))
);

-- TABLE: notification_outbox
CREATE TABLE public.notification_outbox (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    incident_id uuid NOT NULL,
    rule_id uuid NOT NULL,
    notification_id uuid NOT NULL,
    notification_type public.notification_type NOT NULL,
    event_type text NOT NULL,
    status public.notification_outbox_status DEFAULT 'pending'::public.notification_outbox_status NOT NULL,
    payload jsonb NOT NULL,
    attempt_count integer DEFAULT 0 NOT NULL,
    max_attempts integer DEFAULT 5 NOT NULL,
    next_attempt_at timestamptz DEFAULT now() NOT NULL,
    last_attempt_at timestamptz,
    delivered_at timestamptz,
    last_error_kind text,
    last_error_code text,
    last_error text,
    dedupe_key text NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    CONSTRAINT notification_outbox_attempt_count_non_negative CHECK ((attempt_count >= 0)),
    CONSTRAINT notification_outbox_dedupe_key_not_empty CHECK ((length(btrim(dedupe_key)) > 0)),
    CONSTRAINT notification_outbox_event_type_not_empty CHECK ((length(btrim(event_type)) > 0)),
    CONSTRAINT notification_outbox_last_error_code_not_empty CHECK (((last_error_code IS NULL) OR (length(btrim(last_error_code)) > 0))),
    CONSTRAINT notification_outbox_last_error_kind_not_empty CHECK (((last_error_kind IS NULL) OR (length(btrim(last_error_kind)) > 0))),
    CONSTRAINT notification_outbox_last_error_not_empty CHECK (((last_error IS NULL) OR (length(btrim(last_error)) > 0))),
    CONSTRAINT notification_outbox_max_attempts_positive CHECK ((max_attempts > 0)),
    CONSTRAINT notification_outbox_payload_is_object CHECK ((jsonb_typeof(payload) = 'object'::text))
);

-- TABLE: notifications
CREATE TABLE public.notifications (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    name text NOT NULL,
    type public.notification_type NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    config jsonb NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    deleted_at timestamptz,
    CONSTRAINT notifications_config_is_object CHECK ((jsonb_typeof(config) = 'object'::text)),
    CONSTRAINT notifications_deleted_at_after_created_at CHECK (((deleted_at IS NULL) OR (deleted_at >= created_at))),
    CONSTRAINT notifications_name_not_empty CHECK ((length(btrim(name)) > 0))
);

-- TABLE: password_credentials
CREATE TABLE public.password_credentials (
    user_id uuid NOT NULL,
    password_hash text NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    CONSTRAINT password_credentials_hash_not_empty CHECK ((length(btrim(password_hash)) > 0))
);

-- TABLE: password_reset_tokens
CREATE TABLE public.password_reset_tokens (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    token_hash text NOT NULL,
    expires_at timestamptz NOT NULL,
    used_at timestamptz,
    created_at timestamptz DEFAULT now() NOT NULL,
    CONSTRAINT password_reset_tokens_expires_after_created CHECK ((expires_at > created_at)),
    CONSTRAINT password_reset_tokens_token_hash_not_empty CHECK ((length(btrim(token_hash)) > 0))
);

-- TABLE: ping_check_configs
CREATE TABLE public.ping_check_configs (
    check_id uuid NOT NULL,
    packet_count integer DEFAULT 4 NOT NULL,
    packet_size_bytes integer DEFAULT 56 NOT NULL,
    timeout_ms integer DEFAULT 3000 NOT NULL,
    ip_family public.ip_family,
    CONSTRAINT ping_check_configs_packet_count_positive CHECK ((packet_count > 0)),
    CONSTRAINT ping_check_configs_packet_size_range CHECK (((packet_size_bytes >= 0) AND (packet_size_bytes <= 65507))),
    CONSTRAINT ping_check_configs_timeout_ms_positive CHECK ((timeout_ms > 0))
);

-- TABLE: probe_check_assignments
CREATE TABLE public.probe_check_assignments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    probe_id uuid NOT NULL,
    check_id uuid NOT NULL,
    check_version text NOT NULL,
    selector_version text NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    deleted_at timestamptz,
    CONSTRAINT probe_check_assignments_check_version_not_empty CHECK ((length(btrim(check_version)) > 0)),
    CONSTRAINT probe_check_assignments_deleted_at_after_created_at CHECK (((deleted_at IS NULL) OR (deleted_at >= created_at))),
    CONSTRAINT probe_check_assignments_selector_version_not_empty CHECK ((length(btrim(selector_version)) > 0))
);

-- TABLE: probe_credentials
CREATE TABLE public.probe_credentials (
    probe_id uuid NOT NULL,
    secret_hash text NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    last_rotated_at timestamptz,
    CONSTRAINT probe_credentials_last_rotated_at_after_created_at CHECK (((last_rotated_at IS NULL) OR (last_rotated_at >= created_at))),
    CONSTRAINT probe_credentials_secret_hash_not_empty CHECK ((length(btrim(secret_hash)) > 0))
);

-- TABLE: probe_labels
CREATE TABLE public.probe_labels (
    project_id uuid NOT NULL,
    probe_id uuid NOT NULL,
    label_id uuid NOT NULL
);

-- TABLE: probe_statuses
CREATE TABLE public.probe_statuses (
    probe_id uuid NOT NULL,
    status public.probe_state NOT NULL,
    last_seen_at timestamptz,
    agent_version text,
    public_v4 inet,
    public_v6 inet,
    "as" text,
    addrs inet[] DEFAULT '{}'::inet[] NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    online_since timestamptz,
    CONSTRAINT probe_statuses_agent_version_not_empty CHECK (((agent_version IS NULL) OR (length(btrim(agent_version)) > 0))),
    CONSTRAINT probe_statuses_as_not_empty CHECK ((("as" IS NULL) OR (length(btrim("as")) > 0)))
);

-- TABLE: probes
CREATE TABLE public.probes (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    internal_id bigint NOT NULL,
    project_id uuid NOT NULL,
    name text NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    location point,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    deleted_at timestamptz,
    location_name text,
    CONSTRAINT probes_deleted_at_after_created_at CHECK (((deleted_at IS NULL) OR (deleted_at >= created_at))),
    CONSTRAINT probes_location_name_valid CHECK (((location_name IS NULL) OR (length(btrim(location_name)) > 0))),
    CONSTRAINT probes_name_not_empty CHECK ((length(btrim(name)) > 0))
);

-- SEQUENCE: probes_internal_id_seq
ALTER TABLE public.probes ALTER COLUMN internal_id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.probes_internal_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);

-- TABLE: project_invites
CREATE TABLE public.project_invites (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    invited_user_id uuid,
    invited_by_user_id uuid NOT NULL,
    role public.project_member_role NOT NULL,
    status public.project_invite_status DEFAULT 'pending'::public.project_invite_status NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    resolved_at timestamptz,
    invited_email citext NOT NULL,
    CONSTRAINT project_invites_invited_email_not_empty CHECK ((length(btrim((invited_email)::text)) > 0)),
    CONSTRAINT project_invites_resolution_consistent CHECK ((((status = 'pending'::public.project_invite_status) AND (resolved_at IS NULL)) OR ((status <> 'pending'::public.project_invite_status) AND (resolved_at IS NOT NULL)))),
    CONSTRAINT project_invites_resolved_at_after_created_at CHECK (((resolved_at IS NULL) OR (resolved_at >= created_at)))
);

-- TABLE: project_members
CREATE TABLE public.project_members (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    user_id uuid NOT NULL,
    role public.project_member_role NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL
);

-- TABLE: projects
CREATE TABLE public.projects (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    slug citext NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    deleted_at timestamptz,
    CONSTRAINT projects_deleted_at_after_created_at CHECK (((deleted_at IS NULL) OR (deleted_at >= created_at))),
    CONSTRAINT projects_name_not_empty CHECK ((length(btrim(name)) > 0)),
    CONSTRAINT projects_slug_not_empty CHECK ((length(btrim((slug)::text)) > 0))
);

-- TABLE: public_status_page_element_assignments
CREATE TABLE public.public_status_page_element_assignments (
    element_id uuid NOT NULL,
    public_page_id uuid NOT NULL,
    project_id uuid NOT NULL,
    assignment_id uuid NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    CONSTRAINT public_status_page_element_assignments_sort_order_non_negative CHECK ((sort_order >= 0))
);

-- TABLE: public_status_page_elements
CREATE TABLE public.public_status_page_elements (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    public_page_id uuid NOT NULL,
    project_id uuid NOT NULL,
    parent_element_id uuid,
    kind public.public_status_element_kind NOT NULL,
    check_id uuid,
    title text,
    description text,
    sort_order integer DEFAULT 0 NOT NULL,
    chart_mode public.public_status_chart_mode DEFAULT 'inherit'::public.public_status_chart_mode NOT NULL,
    chart_range public.public_status_chart_range,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    assignment_selection_mode public.public_status_assignment_selection_mode,
    display_mode public.public_status_element_display_mode DEFAULT 'status'::public.public_status_element_display_mode NOT NULL,
    CONSTRAINT public_status_page_elements_description_not_empty CHECK (((description IS NULL) OR (length(btrim(description)) > 0))),
    CONSTRAINT public_status_page_elements_parent_not_self CHECK (((parent_element_id IS NULL) OR (parent_element_id <> id))),
    CONSTRAINT public_status_page_elements_shape CHECK ((((kind = 'folder'::public.public_status_element_kind) AND (check_id IS NULL) AND (assignment_selection_mode IS NULL)) OR ((kind = 'assignment_group'::public.public_status_element_kind) AND (assignment_selection_mode = 'all_check'::public.public_status_assignment_selection_mode) AND (check_id IS NOT NULL)) OR ((kind = 'assignment_group'::public.public_status_element_kind) AND (assignment_selection_mode = 'selected_assignments'::public.public_status_assignment_selection_mode) AND (check_id IS NULL)))),
    CONSTRAINT public_status_page_elements_sort_order_non_negative CHECK ((sort_order >= 0)),
    CONSTRAINT public_status_page_elements_title_not_empty CHECK (((title IS NULL) OR (length(btrim(title)) > 0)))
);

-- VIEW: public_status_page_assignment_scope
CREATE VIEW public.public_status_page_assignment_scope AS
 SELECT public_status_page_elements.public_page_id,
    public_status_page_elements.id AS element_id,
    probe_check_assignments.id AS assignment_id,
    probe_check_assignments.project_id,
    probe_check_assignments.probe_id,
    probe_check_assignments.check_id
   FROM ((public.public_status_page_elements
     JOIN public.probe_check_assignments ON (((public_status_page_elements.assignment_selection_mode = 'all_check'::public.public_status_assignment_selection_mode) AND (probe_check_assignments.project_id = public_status_page_elements.project_id) AND (probe_check_assignments.check_id = public_status_page_elements.check_id) AND (probe_check_assignments.deleted_at IS NULL))))
     JOIN public.checks ON (((checks.project_id = probe_check_assignments.project_id) AND (checks.id = probe_check_assignments.check_id) AND (checks.deleted_at IS NULL) AND (checks.check_type = ANY (ARRAY['ping'::public.check_type, 'tcp'::public.check_type, 'http'::public.check_type])))))
  WHERE (public_status_page_elements.kind = 'assignment_group'::public.public_status_element_kind)
UNION ALL
 SELECT public_status_page_elements.public_page_id,
    public_status_page_elements.id AS element_id,
    probe_check_assignments.id AS assignment_id,
    probe_check_assignments.project_id,
    probe_check_assignments.probe_id,
    probe_check_assignments.check_id
   FROM (((public.public_status_page_elements
     JOIN public.public_status_page_element_assignments ON (((public_status_page_element_assignments.public_page_id = public_status_page_elements.public_page_id) AND (public_status_page_element_assignments.element_id = public_status_page_elements.id))))
     JOIN public.probe_check_assignments ON (((probe_check_assignments.id = public_status_page_element_assignments.assignment_id) AND (probe_check_assignments.project_id = public_status_page_elements.project_id) AND (probe_check_assignments.deleted_at IS NULL))))
     JOIN public.checks ON (((checks.project_id = probe_check_assignments.project_id) AND (checks.id = probe_check_assignments.check_id) AND (checks.deleted_at IS NULL) AND (checks.check_type = ANY (ARRAY['ping'::public.check_type, 'tcp'::public.check_type, 'http'::public.check_type])))))
  WHERE ((public_status_page_elements.kind = 'assignment_group'::public.public_status_element_kind) AND (public_status_page_elements.assignment_selection_mode = 'selected_assignments'::public.public_status_assignment_selection_mode));

-- TABLE: public_status_pages
CREATE TABLE public.public_status_pages (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    slug citext NOT NULL,
    title text NOT NULL,
    description text,
    enabled boolean DEFAULT true NOT NULL,
    default_chart_mode public.public_status_chart_mode DEFAULT 'off'::public.public_status_chart_mode NOT NULL,
    default_chart_range public.public_status_chart_range DEFAULT '24h'::public.public_status_chart_range NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    deleted_at timestamptz,
    footer_text text,
    banner_image_url text,
    theme public.public_status_theme DEFAULT 'auto'::public.public_status_theme NOT NULL,
    show_targets boolean DEFAULT false NOT NULL,
    show_probe_names boolean DEFAULT false NOT NULL,
    show_probe_locations boolean DEFAULT false NOT NULL,
    show_incident_history boolean DEFAULT true NOT NULL,
    show_generated_at boolean DEFAULT true NOT NULL,
    custom_css text,
    CONSTRAINT public_status_pages_banner_image_url_valid CHECK (((banner_image_url IS NULL) OR ((length(btrim(banner_image_url)) > 0) AND (length(banner_image_url) <= 2048)))),
    CONSTRAINT public_status_pages_custom_css_valid CHECK (((custom_css IS NULL) OR ((length(btrim(custom_css)) > 0) AND (length(custom_css) <= 65536)))),
    CONSTRAINT public_status_pages_deleted_at_after_created_at CHECK (((deleted_at IS NULL) OR (deleted_at >= created_at))),
    CONSTRAINT public_status_pages_description_not_empty CHECK (((description IS NULL) OR (length(btrim(description)) > 0))),
    CONSTRAINT public_status_pages_footer_text_valid CHECK (((footer_text IS NULL) OR ((length(btrim(footer_text)) > 0) AND (length(footer_text) <= 2048)))),
    CONSTRAINT public_status_pages_slug_not_empty CHECK ((length(btrim((slug)::text)) > 0)),
    CONSTRAINT public_status_pages_title_not_empty CHECK ((length(btrim(title)) > 0))
);

-- TABLE: system_setting_audit_events
CREATE TABLE public.system_setting_audit_events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    key text NOT NULL,
    action text NOT NULL,
    updated_by_user_id uuid,
    created_at timestamptz DEFAULT now() NOT NULL,
    CONSTRAINT system_setting_audit_action_not_empty CHECK ((length(btrim(action)) > 0)),
    CONSTRAINT system_setting_audit_key_not_empty CHECK ((length(btrim(key)) > 0))
);

-- TABLE: system_settings
CREATE TABLE public.system_settings (
    key text NOT NULL,
    value jsonb,
    encrypted_value bytea,
    encrypted_value_nonce bytea,
    secret boolean DEFAULT false NOT NULL,
    updated_by_user_id uuid,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    CONSTRAINT system_settings_key_not_empty CHECK ((length(btrim(key)) > 0)),
    CONSTRAINT system_settings_public_or_secret_value CHECK ((((secret = false) AND (value IS NOT NULL) AND (encrypted_value IS NULL) AND (encrypted_value_nonce IS NULL)) OR ((secret = true) AND (value IS NULL) AND (encrypted_value IS NOT NULL) AND (encrypted_value_nonce IS NOT NULL))))
);

-- TABLE: system_user_roles
CREATE TABLE public.system_user_roles (
    user_id uuid NOT NULL,
    role public.system_role NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL
);

-- TABLE: tcp_check_configs
CREATE TABLE public.tcp_check_configs (
    check_id uuid NOT NULL,
    port integer DEFAULT 443 NOT NULL,
    timeout_ms integer DEFAULT 3000 NOT NULL,
    ip_family public.ip_family,
    CONSTRAINT tcp_check_configs_port_range CHECK (((port >= 1) AND (port <= 65535))),
    CONSTRAINT tcp_check_configs_timeout_ms_positive CHECK ((timeout_ms > 0))
);

-- TABLE: traceroute_check_configs
CREATE TABLE public.traceroute_check_configs (
    check_id uuid NOT NULL,
    protocol public.traceroute_protocol DEFAULT 'icmp'::public.traceroute_protocol NOT NULL,
    max_hops integer DEFAULT 30 NOT NULL,
    timeout_ms integer DEFAULT 3000 NOT NULL,
    queries_per_hop integer DEFAULT 3 NOT NULL,
    packet_size_bytes integer DEFAULT 56 NOT NULL,
    port integer DEFAULT 33434 NOT NULL,
    ip_family public.ip_family,
    CONSTRAINT traceroute_check_configs_max_hops_range CHECK (((max_hops >= 1) AND (max_hops <= 64))),
    CONSTRAINT traceroute_check_configs_packet_size_range CHECK (((packet_size_bytes >= 1) AND (packet_size_bytes <= 65507))),
    CONSTRAINT traceroute_check_configs_port_range CHECK (((port >= 1) AND (port <= 65535))),
    CONSTRAINT traceroute_check_configs_queries_per_hop_range CHECK (((queries_per_hop >= 1) AND (queries_per_hop <= 10))),
    CONSTRAINT traceroute_check_configs_timeout_ms_range CHECK (((timeout_ms >= 1) AND (timeout_ms <= 60000)))
);

-- TABLE: user_identities
CREATE TABLE public.user_identities (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    provider text NOT NULL,
    issuer text NOT NULL,
    subject text NOT NULL,
    email citext,
    email_verified boolean DEFAULT false NOT NULL,
    display_name text,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    last_login_at timestamptz,
    username text,
    avatar_url text,
    CONSTRAINT user_identities_avatar_url_not_empty CHECK (((avatar_url IS NULL) OR (length(btrim(avatar_url)) > 0))),
    CONSTRAINT user_identities_display_name_not_empty CHECK (((display_name IS NULL) OR (length(btrim(display_name)) > 0))),
    CONSTRAINT user_identities_email_not_empty CHECK (((email IS NULL) OR (length(btrim((email)::text)) > 0))),
    CONSTRAINT user_identities_issuer_not_empty CHECK ((length(btrim(issuer)) > 0)),
    CONSTRAINT user_identities_last_login_after_created CHECK (((last_login_at IS NULL) OR (last_login_at >= created_at))),
    CONSTRAINT user_identities_provider_not_empty CHECK ((length(btrim(provider)) > 0)),
    CONSTRAINT user_identities_subject_not_empty CHECK ((length(btrim(subject)) > 0)),
    CONSTRAINT user_identities_username_not_empty CHECK (((username IS NULL) OR (length(btrim(username)) > 0)))
);

-- TABLE: users
CREATE TABLE public.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email citext NOT NULL,
    display_name text NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    email_verified_at timestamptz,
    disabled_at timestamptz,
    CONSTRAINT users_disabled_at_after_created_at CHECK (((disabled_at IS NULL) OR (disabled_at >= created_at))),
    CONSTRAINT users_display_name_not_empty CHECK (((length(btrim(display_name)) > 0) AND (length(btrim(display_name)) <= 100))),
    CONSTRAINT users_email_not_empty CHECK ((length(btrim((email)::text)) > 0))
);

-- CONSTRAINT: alert_incidents alert_incidents_pkey
ALTER TABLE ONLY public.alert_incidents
    ADD CONSTRAINT alert_incidents_pkey PRIMARY KEY (id);

-- CONSTRAINT: alert_notifications alert_notifications_pkey
ALTER TABLE ONLY public.alert_notifications
    ADD CONSTRAINT alert_notifications_pkey PRIMARY KEY (rule_id, notification_id);

-- CONSTRAINT: alert_rule_pending_evaluations alert_rule_pending_evaluations_pkey
ALTER TABLE ONLY public.alert_rule_pending_evaluations
    ADD CONSTRAINT alert_rule_pending_evaluations_pkey PRIMARY KEY (rule_id, probe_id, check_id);

-- CONSTRAINT: alert_rules alert_rules_pkey
ALTER TABLE ONLY public.alert_rules
    ADD CONSTRAINT alert_rules_pkey PRIMARY KEY (id);

-- CONSTRAINT: api_tokens api_tokens_pkey
ALTER TABLE ONLY public.api_tokens
    ADD CONSTRAINT api_tokens_pkey PRIMARY KEY (id);

-- CONSTRAINT: assignment_refresh_jobs assignment_refresh_jobs_pkey
ALTER TABLE ONLY public.assignment_refresh_jobs
    ADD CONSTRAINT assignment_refresh_jobs_pkey PRIMARY KEY (id);

-- CONSTRAINT: auth_sessions auth_sessions_pkey
ALTER TABLE ONLY public.auth_sessions
    ADD CONSTRAINT auth_sessions_pkey PRIMARY KEY (id);

-- CONSTRAINT: check_labels check_labels_pkey
ALTER TABLE ONLY public.check_labels
    ADD CONSTRAINT check_labels_pkey PRIMARY KEY (project_id, check_id, label_id);

-- CONSTRAINT: checks checks_internal_id_key
ALTER TABLE ONLY public.checks
    ADD CONSTRAINT checks_internal_id_key UNIQUE (internal_id);

-- CONSTRAINT: checks checks_pkey
ALTER TABLE ONLY public.checks
    ADD CONSTRAINT checks_pkey PRIMARY KEY (id);

-- CONSTRAINT: email_verification_tokens email_verification_tokens_pkey
ALTER TABLE ONLY public.email_verification_tokens
    ADD CONSTRAINT email_verification_tokens_pkey PRIMARY KEY (id);

-- CONSTRAINT: http_check_configs http_check_configs_pkey
ALTER TABLE ONLY public.http_check_configs
    ADD CONSTRAINT http_check_configs_pkey PRIMARY KEY (check_id);

-- CONSTRAINT: labels labels_pkey
ALTER TABLE ONLY public.labels
    ADD CONSTRAINT labels_pkey PRIMARY KEY (id);

-- CONSTRAINT: notification_outbox notification_outbox_pkey
ALTER TABLE ONLY public.notification_outbox
    ADD CONSTRAINT notification_outbox_pkey PRIMARY KEY (id);

-- CONSTRAINT: notifications notifications_pkey
ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);

-- CONSTRAINT: external_auth_flows external_auth_flows_pkey
ALTER TABLE ONLY public.external_auth_flows
    ADD CONSTRAINT external_auth_flows_pkey PRIMARY KEY (id);

-- CONSTRAINT: password_credentials password_credentials_pkey
ALTER TABLE ONLY public.password_credentials
    ADD CONSTRAINT password_credentials_pkey PRIMARY KEY (user_id);

-- CONSTRAINT: password_reset_tokens password_reset_tokens_pkey
ALTER TABLE ONLY public.password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_pkey PRIMARY KEY (id);

-- CONSTRAINT: ping_check_configs ping_check_configs_pkey
ALTER TABLE ONLY public.ping_check_configs
    ADD CONSTRAINT ping_check_configs_pkey PRIMARY KEY (check_id);

-- CONSTRAINT: probe_check_assignments probe_check_assignments_pkey
ALTER TABLE ONLY public.probe_check_assignments
    ADD CONSTRAINT probe_check_assignments_pkey PRIMARY KEY (id);

-- CONSTRAINT: probe_credentials probe_credentials_pkey
ALTER TABLE ONLY public.probe_credentials
    ADD CONSTRAINT probe_credentials_pkey PRIMARY KEY (probe_id);

-- CONSTRAINT: probe_labels probe_labels_pkey
ALTER TABLE ONLY public.probe_labels
    ADD CONSTRAINT probe_labels_pkey PRIMARY KEY (project_id, probe_id, label_id);

-- CONSTRAINT: probe_statuses probe_statuses_pkey
ALTER TABLE ONLY public.probe_statuses
    ADD CONSTRAINT probe_statuses_pkey PRIMARY KEY (probe_id);

-- CONSTRAINT: probes probes_internal_id_key
ALTER TABLE ONLY public.probes
    ADD CONSTRAINT probes_internal_id_key UNIQUE (internal_id);

-- CONSTRAINT: probes probes_pkey
ALTER TABLE ONLY public.probes
    ADD CONSTRAINT probes_pkey PRIMARY KEY (id);

-- CONSTRAINT: project_invites project_invites_pkey
ALTER TABLE ONLY public.project_invites
    ADD CONSTRAINT project_invites_pkey PRIMARY KEY (id);

-- CONSTRAINT: project_members project_members_pkey
ALTER TABLE ONLY public.project_members
    ADD CONSTRAINT project_members_pkey PRIMARY KEY (id);

-- CONSTRAINT: projects projects_pkey
ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (id);

-- CONSTRAINT: public_status_page_element_assignments public_status_page_element_assignments_pkey
ALTER TABLE ONLY public.public_status_page_element_assignments
    ADD CONSTRAINT public_status_page_element_assignments_pkey PRIMARY KEY (element_id, assignment_id);

-- CONSTRAINT: public_status_page_elements public_status_page_elements_pkey
ALTER TABLE ONLY public.public_status_page_elements
    ADD CONSTRAINT public_status_page_elements_pkey PRIMARY KEY (id);

-- CONSTRAINT: public_status_pages public_status_pages_pkey
ALTER TABLE ONLY public.public_status_pages
    ADD CONSTRAINT public_status_pages_pkey PRIMARY KEY (id);

-- CONSTRAINT: system_setting_audit_events system_setting_audit_events_pkey
ALTER TABLE ONLY public.system_setting_audit_events
    ADD CONSTRAINT system_setting_audit_events_pkey PRIMARY KEY (id);

-- CONSTRAINT: system_settings system_settings_pkey
ALTER TABLE ONLY public.system_settings
    ADD CONSTRAINT system_settings_pkey PRIMARY KEY (key);

-- CONSTRAINT: system_user_roles system_user_roles_pkey
ALTER TABLE ONLY public.system_user_roles
    ADD CONSTRAINT system_user_roles_pkey PRIMARY KEY (user_id);

-- CONSTRAINT: tcp_check_configs tcp_check_configs_pkey
ALTER TABLE ONLY public.tcp_check_configs
    ADD CONSTRAINT tcp_check_configs_pkey PRIMARY KEY (check_id);

-- CONSTRAINT: traceroute_check_configs traceroute_check_configs_pkey
ALTER TABLE ONLY public.traceroute_check_configs
    ADD CONSTRAINT traceroute_check_configs_pkey PRIMARY KEY (check_id);

-- CONSTRAINT: alert_incidents uq_alert_incidents_project_id_id
ALTER TABLE ONLY public.alert_incidents
    ADD CONSTRAINT uq_alert_incidents_project_id_id UNIQUE (project_id, id);

-- CONSTRAINT: alert_rules uq_alert_rules_project_id_id
ALTER TABLE ONLY public.alert_rules
    ADD CONSTRAINT uq_alert_rules_project_id_id UNIQUE (project_id, id);

-- CONSTRAINT: assignment_refresh_jobs uq_assignment_refresh_jobs_project_id_id
ALTER TABLE ONLY public.assignment_refresh_jobs
    ADD CONSTRAINT uq_assignment_refresh_jobs_project_id_id UNIQUE (project_id, id);

-- CONSTRAINT: checks uq_checks_project_id_id
ALTER TABLE ONLY public.checks
    ADD CONSTRAINT uq_checks_project_id_id UNIQUE (project_id, id);

-- CONSTRAINT: external_auth_flows uq_external_auth_flows_state_hash
ALTER TABLE ONLY public.external_auth_flows
    ADD CONSTRAINT uq_external_auth_flows_state_hash UNIQUE (state_hash);

-- CONSTRAINT: labels uq_labels_project_id_id
ALTER TABLE ONLY public.labels
    ADD CONSTRAINT uq_labels_project_id_id UNIQUE (project_id, id);

-- CONSTRAINT: notification_outbox uq_notification_outbox_project_id_id
ALTER TABLE ONLY public.notification_outbox
    ADD CONSTRAINT uq_notification_outbox_project_id_id UNIQUE (project_id, id);

-- CONSTRAINT: notifications uq_notifications_project_id_id
ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT uq_notifications_project_id_id UNIQUE (project_id, id);

-- CONSTRAINT: probes uq_probes_project_id_id
ALTER TABLE ONLY public.probes
    ADD CONSTRAINT uq_probes_project_id_id UNIQUE (project_id, id);

-- CONSTRAINT: public_status_page_elements uq_public_status_page_elements_page_id_id
ALTER TABLE ONLY public.public_status_page_elements
    ADD CONSTRAINT uq_public_status_page_elements_page_id_id UNIQUE (public_page_id, id);

-- CONSTRAINT: public_status_pages uq_public_status_pages_id_project
ALTER TABLE ONLY public.public_status_pages
    ADD CONSTRAINT uq_public_status_pages_id_project UNIQUE (id, project_id);

-- CONSTRAINT: user_identities uq_user_identities_provider_issuer_subject
ALTER TABLE ONLY public.user_identities
    ADD CONSTRAINT uq_user_identities_provider_issuer_subject UNIQUE (provider, issuer, subject);

-- CONSTRAINT: user_identities uq_user_identities_user_provider_issuer
ALTER TABLE ONLY public.user_identities
    ADD CONSTRAINT uq_user_identities_user_provider_issuer UNIQUE (user_id, provider, issuer);

-- CONSTRAINT: user_identities user_identities_pkey
ALTER TABLE ONLY public.user_identities
    ADD CONSTRAINT user_identities_pkey PRIMARY KEY (id);

-- CONSTRAINT: users users_pkey
ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

-- INDEX: idx_alert_incidents_project_probe_check_status
CREATE INDEX idx_alert_incidents_project_probe_check_status ON public.alert_incidents USING btree (project_id, probe_id, check_id, status);

-- INDEX: idx_alert_incidents_project_rule_status
CREATE INDEX idx_alert_incidents_project_rule_status ON public.alert_incidents USING btree (project_id, rule_id, status);

-- INDEX: idx_alert_incidents_project_status_opened
CREATE INDEX idx_alert_incidents_project_status_opened ON public.alert_incidents USING btree (project_id, status, opened_at DESC);

-- INDEX: idx_alert_incidents_rule_target_resolved
CREATE INDEX idx_alert_incidents_rule_target_resolved ON public.alert_incidents USING btree (rule_id, probe_id, check_id, resolved_at DESC) WHERE (status = 'resolved'::public.alert_incident_status);

-- INDEX: idx_alert_rules_project_active
CREATE INDEX idx_alert_rules_project_active ON public.alert_rules USING btree (project_id) WHERE (deleted_at IS NULL);

-- INDEX: idx_alert_rules_project_check_active
CREATE INDEX idx_alert_rules_project_check_active ON public.alert_rules USING btree (project_id, check_id) WHERE ((deleted_at IS NULL) AND (check_id IS NOT NULL));

-- INDEX: idx_alert_rules_project_check_type_status_active
CREATE INDEX idx_alert_rules_project_check_type_status_active ON public.alert_rules USING btree (project_id, check_type, status) WHERE (deleted_at IS NULL);

-- INDEX: idx_alert_rules_project_probe_active
CREATE INDEX idx_alert_rules_project_probe_active ON public.alert_rules USING btree (project_id, probe_id) WHERE ((deleted_at IS NULL) AND (probe_id IS NOT NULL));

-- INDEX: idx_assignment_refresh_jobs_project_created
CREATE INDEX idx_assignment_refresh_jobs_project_created ON public.assignment_refresh_jobs USING btree (project_id, created_at DESC);

-- INDEX: idx_assignment_refresh_jobs_status_last_attempt
CREATE INDEX idx_assignment_refresh_jobs_status_last_attempt ON public.assignment_refresh_jobs USING btree (status, last_attempt_at);

-- INDEX: idx_assignment_refresh_jobs_status_next_attempt
CREATE INDEX idx_assignment_refresh_jobs_status_next_attempt ON public.assignment_refresh_jobs USING btree (status, next_attempt_at);

-- INDEX: idx_check_labels_label_id
CREATE INDEX idx_check_labels_label_id ON public.check_labels USING btree (label_id);

-- INDEX: idx_checks_project_id
CREATE INDEX idx_checks_project_id ON public.checks USING btree (project_id);

-- INDEX: idx_checks_project_id_check_type
CREATE INDEX idx_checks_project_id_check_type ON public.checks USING btree (project_id, check_type);

-- INDEX: idx_labels_project_id
CREATE INDEX idx_labels_project_id ON public.labels USING btree (project_id);

-- INDEX: idx_notification_outbox_project_created
CREATE INDEX idx_notification_outbox_project_created ON public.notification_outbox USING btree (project_id, created_at DESC);

-- INDEX: idx_notification_outbox_status_last_attempt
CREATE INDEX idx_notification_outbox_status_last_attempt ON public.notification_outbox USING btree (status, last_attempt_at);

-- INDEX: idx_notification_outbox_status_next_attempt
CREATE INDEX idx_notification_outbox_status_next_attempt ON public.notification_outbox USING btree (status, next_attempt_at);

-- INDEX: idx_notifications_project_active
CREATE INDEX idx_notifications_project_active ON public.notifications USING btree (project_id) WHERE (deleted_at IS NULL);

-- INDEX: idx_notifications_project_enabled_active
CREATE INDEX idx_notifications_project_enabled_active ON public.notifications USING btree (project_id, enabled) WHERE (deleted_at IS NULL);

-- INDEX: idx_probe_check_assignments_check_id
CREATE INDEX idx_probe_check_assignments_check_id ON public.probe_check_assignments USING btree (check_id);

-- INDEX: idx_probe_check_assignments_probe_id
CREATE INDEX idx_probe_check_assignments_probe_id ON public.probe_check_assignments USING btree (probe_id);

-- INDEX: idx_probe_check_assignments_project_id
CREATE INDEX idx_probe_check_assignments_project_id ON public.probe_check_assignments USING btree (project_id);

-- INDEX: idx_probe_labels_label_id
CREATE INDEX idx_probe_labels_label_id ON public.probe_labels USING btree (label_id);

-- INDEX: idx_probes_project_id
CREATE INDEX idx_probes_project_id ON public.probes USING btree (project_id);

-- INDEX: idx_project_invites_invited_by_user_id
CREATE INDEX idx_project_invites_invited_by_user_id ON public.project_invites USING btree (invited_by_user_id);

-- INDEX: idx_project_invites_invited_email_pending
CREATE INDEX idx_project_invites_invited_email_pending ON public.project_invites USING btree (invited_email, created_at DESC, id DESC) WHERE (status = 'pending'::public.project_invite_status);

-- INDEX: idx_project_invites_project_pending
CREATE INDEX idx_project_invites_project_pending ON public.project_invites USING btree (project_id, created_at, id) WHERE (status = 'pending'::public.project_invite_status);

-- INDEX: idx_project_members_project_id
CREATE INDEX idx_project_members_project_id ON public.project_members USING btree (project_id);

-- INDEX: idx_project_members_user_id
CREATE INDEX idx_project_members_user_id ON public.project_members USING btree (user_id);

-- INDEX: idx_public_status_page_element_assignments_assignment
CREATE INDEX idx_public_status_page_element_assignments_assignment ON public.public_status_page_element_assignments USING btree (assignment_id);

-- INDEX: idx_public_status_page_element_assignments_page
CREATE INDEX idx_public_status_page_element_assignments_page ON public.public_status_page_element_assignments USING btree (public_page_id, element_id, sort_order, assignment_id);

-- INDEX: idx_public_status_page_elements_parent_order
CREATE INDEX idx_public_status_page_elements_parent_order ON public.public_status_page_elements USING btree (public_page_id, parent_element_id, sort_order, created_at, id);

-- INDEX: idx_public_status_page_elements_project_check
CREATE INDEX idx_public_status_page_elements_project_check ON public.public_status_page_elements USING btree (project_id, check_id) WHERE (check_id IS NOT NULL);

-- INDEX: idx_public_status_pages_project_active
CREATE INDEX idx_public_status_pages_project_active ON public.public_status_pages USING btree (project_id) WHERE (deleted_at IS NULL);

-- INDEX: idx_system_setting_audit_events_created
CREATE INDEX idx_system_setting_audit_events_created ON public.system_setting_audit_events USING btree (created_at DESC, id DESC);

-- INDEX: idx_system_user_roles_role
CREATE INDEX idx_system_user_roles_role ON public.system_user_roles USING btree (role);

-- INDEX: idx_users_disabled_at
CREATE INDEX idx_users_disabled_at ON public.users USING btree (disabled_at) WHERE (disabled_at IS NOT NULL);

-- INDEX: ix_api_tokens_active_expiry
CREATE INDEX ix_api_tokens_active_expiry ON public.api_tokens USING btree (expires_at) WHERE (revoked_at IS NULL);

-- INDEX: ix_api_tokens_user_id
CREATE INDEX ix_api_tokens_user_id ON public.api_tokens USING btree (user_id);

-- INDEX: ix_auth_sessions_active_expiry
CREATE INDEX ix_auth_sessions_active_expiry ON public.auth_sessions USING btree (idle_expires_at, absolute_expires_at) WHERE (revoked_at IS NULL);

-- INDEX: ix_auth_sessions_identity_id
CREATE INDEX ix_auth_sessions_identity_id ON public.auth_sessions USING btree (identity_id) WHERE (identity_id IS NOT NULL);

-- INDEX: ix_auth_sessions_user_id
CREATE INDEX ix_auth_sessions_user_id ON public.auth_sessions USING btree (user_id);

-- INDEX: ix_email_verification_tokens_user_active
CREATE INDEX ix_email_verification_tokens_user_active ON public.email_verification_tokens USING btree (user_id, expires_at) WHERE (used_at IS NULL);

-- INDEX: ix_external_auth_flows_expires_at
CREATE INDEX ix_external_auth_flows_expires_at ON public.external_auth_flows USING btree (expires_at);

-- INDEX: ix_password_reset_tokens_user_active
CREATE INDEX ix_password_reset_tokens_user_active ON public.password_reset_tokens USING btree (user_id, expires_at) WHERE (used_at IS NULL);

-- INDEX: ix_user_identities_user_id
CREATE INDEX ix_user_identities_user_id ON public.user_identities USING btree (user_id);

-- INDEX: uq_alert_incidents_active_rule_probe_check
CREATE UNIQUE INDEX uq_alert_incidents_active_rule_probe_check ON public.alert_incidents USING btree (rule_id, probe_id, check_id) WHERE (status = ANY (ARRAY['open'::public.alert_incident_status, 'acknowledged'::public.alert_incident_status]));

-- INDEX: uq_api_tokens_token_hash
CREATE UNIQUE INDEX uq_api_tokens_token_hash ON public.api_tokens USING btree (token_hash);

-- INDEX: uq_assignment_refresh_jobs_dedupe_key
CREATE UNIQUE INDEX uq_assignment_refresh_jobs_dedupe_key ON public.assignment_refresh_jobs USING btree (dedupe_key);

-- INDEX: uq_auth_sessions_token_hash
CREATE UNIQUE INDEX uq_auth_sessions_token_hash ON public.auth_sessions USING btree (token_hash);

-- INDEX: uq_email_verification_tokens_token_hash
CREATE UNIQUE INDEX uq_email_verification_tokens_token_hash ON public.email_verification_tokens USING btree (token_hash);

-- INDEX: uq_labels_active_project_key_value
CREATE UNIQUE INDEX uq_labels_active_project_key_value ON public.labels USING btree (project_id, key, value) WHERE (deleted_at IS NULL);

-- INDEX: uq_notification_outbox_dedupe_key
CREATE UNIQUE INDEX uq_notification_outbox_dedupe_key ON public.notification_outbox USING btree (dedupe_key);

-- INDEX: uq_password_reset_tokens_token_hash
CREATE UNIQUE INDEX uq_password_reset_tokens_token_hash ON public.password_reset_tokens USING btree (token_hash);

-- INDEX: uq_probe_check_assignments_active_project_probe_check
CREATE UNIQUE INDEX uq_probe_check_assignments_active_project_probe_check ON public.probe_check_assignments USING btree (project_id, probe_id, check_id) WHERE (deleted_at IS NULL);

-- INDEX: uq_project_invites_pending_project_email
CREATE UNIQUE INDEX uq_project_invites_pending_project_email ON public.project_invites USING btree (project_id, invited_email) WHERE (status = 'pending'::public.project_invite_status);

-- INDEX: uq_project_members_project_user
CREATE UNIQUE INDEX uq_project_members_project_user ON public.project_members USING btree (project_id, user_id);

-- INDEX: uq_projects_slug
CREATE UNIQUE INDEX uq_projects_slug ON public.projects USING btree (slug);

-- INDEX: uq_public_status_pages_slug
CREATE UNIQUE INDEX uq_public_status_pages_slug ON public.public_status_pages USING btree (slug);

-- INDEX: uq_users_email
CREATE UNIQUE INDEX uq_users_email ON public.users USING btree (email);

-- TRIGGER: alert_incidents set_alert_incidents_updated_at
CREATE TRIGGER set_alert_incidents_updated_at BEFORE UPDATE ON public.alert_incidents FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();

-- TRIGGER: alert_rules set_alert_rules_updated_at
CREATE TRIGGER set_alert_rules_updated_at BEFORE UPDATE ON public.alert_rules FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();

-- TRIGGER: assignment_refresh_jobs set_assignment_refresh_jobs_updated_at
CREATE TRIGGER set_assignment_refresh_jobs_updated_at BEFORE UPDATE ON public.assignment_refresh_jobs FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();

-- TRIGGER: checks set_checks_updated_at
CREATE TRIGGER set_checks_updated_at BEFORE UPDATE ON public.checks FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();

-- TRIGGER: labels set_labels_updated_at
CREATE TRIGGER set_labels_updated_at BEFORE UPDATE ON public.labels FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();

-- TRIGGER: notification_outbox set_notification_outbox_updated_at
CREATE TRIGGER set_notification_outbox_updated_at BEFORE UPDATE ON public.notification_outbox FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();

-- TRIGGER: notifications set_notifications_updated_at
CREATE TRIGGER set_notifications_updated_at BEFORE UPDATE ON public.notifications FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();

-- TRIGGER: password_credentials set_password_credentials_updated_at
CREATE TRIGGER set_password_credentials_updated_at BEFORE UPDATE ON public.password_credentials FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();

-- TRIGGER: probe_check_assignments set_probe_check_assignments_updated_at
CREATE TRIGGER set_probe_check_assignments_updated_at BEFORE UPDATE ON public.probe_check_assignments FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();

-- TRIGGER: probes set_probes_updated_at
CREATE TRIGGER set_probes_updated_at BEFORE UPDATE ON public.probes FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();

-- TRIGGER: project_invites set_project_invites_updated_at
CREATE TRIGGER set_project_invites_updated_at BEFORE UPDATE ON public.project_invites FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();

-- TRIGGER: project_members set_project_members_updated_at
CREATE TRIGGER set_project_members_updated_at BEFORE UPDATE ON public.project_members FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();

-- TRIGGER: projects set_projects_updated_at
CREATE TRIGGER set_projects_updated_at BEFORE UPDATE ON public.projects FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();

-- TRIGGER: public_status_page_elements set_public_status_page_elements_updated_at
CREATE TRIGGER set_public_status_page_elements_updated_at BEFORE UPDATE ON public.public_status_page_elements FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();

-- TRIGGER: public_status_pages set_public_status_pages_updated_at
CREATE TRIGGER set_public_status_pages_updated_at BEFORE UPDATE ON public.public_status_pages FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();

-- TRIGGER: system_settings set_system_settings_updated_at
CREATE TRIGGER set_system_settings_updated_at BEFORE UPDATE ON public.system_settings FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();

-- TRIGGER: user_identities set_user_identities_updated_at
CREATE TRIGGER set_user_identities_updated_at BEFORE UPDATE ON public.user_identities FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();

-- TRIGGER: users set_users_updated_at
CREATE TRIGGER set_users_updated_at BEFORE UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();

-- FK CONSTRAINT: alert_incidents alert_incidents_acknowledged_by_user_id_fkey
ALTER TABLE ONLY public.alert_incidents
    ADD CONSTRAINT alert_incidents_acknowledged_by_user_id_fkey FOREIGN KEY (acknowledged_by_user_id) REFERENCES public.users(id);

-- FK CONSTRAINT: alert_incidents alert_incidents_project_id_fkey
ALTER TABLE ONLY public.alert_incidents
    ADD CONSTRAINT alert_incidents_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id);

-- FK CONSTRAINT: alert_incidents alert_incidents_resolved_by_user_id_fkey
ALTER TABLE ONLY public.alert_incidents
    ADD CONSTRAINT alert_incidents_resolved_by_user_id_fkey FOREIGN KEY (resolved_by_user_id) REFERENCES public.users(id);

-- FK CONSTRAINT: alert_notifications alert_notifications_project_id_fkey
ALTER TABLE ONLY public.alert_notifications
    ADD CONSTRAINT alert_notifications_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id);

-- FK CONSTRAINT: alert_rules alert_rules_created_by_user_id_fkey
ALTER TABLE ONLY public.alert_rules
    ADD CONSTRAINT alert_rules_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id);

-- FK CONSTRAINT: alert_rules alert_rules_project_id_fkey
ALTER TABLE ONLY public.alert_rules
    ADD CONSTRAINT alert_rules_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id);

-- FK CONSTRAINT: api_tokens api_tokens_user_id_fkey
ALTER TABLE ONLY public.api_tokens
    ADD CONSTRAINT api_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

-- FK CONSTRAINT: assignment_refresh_jobs assignment_refresh_jobs_project_id_fkey
ALTER TABLE ONLY public.assignment_refresh_jobs
    ADD CONSTRAINT assignment_refresh_jobs_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id);

-- FK CONSTRAINT: auth_sessions auth_sessions_identity_id_fkey
ALTER TABLE ONLY public.auth_sessions
    ADD CONSTRAINT auth_sessions_identity_id_fkey FOREIGN KEY (identity_id) REFERENCES public.user_identities(id) ON DELETE SET NULL;

-- FK CONSTRAINT: auth_sessions auth_sessions_user_id_fkey
ALTER TABLE ONLY public.auth_sessions
    ADD CONSTRAINT auth_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

-- FK CONSTRAINT: checks checks_project_id_fkey
ALTER TABLE ONLY public.checks
    ADD CONSTRAINT checks_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id);

-- FK CONSTRAINT: email_verification_tokens email_verification_tokens_user_id_fkey
ALTER TABLE ONLY public.email_verification_tokens
    ADD CONSTRAINT email_verification_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

-- FK CONSTRAINT: alert_incidents fk_alert_incidents_project_check
ALTER TABLE ONLY public.alert_incidents
    ADD CONSTRAINT fk_alert_incidents_project_check FOREIGN KEY (project_id, check_id) REFERENCES public.checks(project_id, id);

-- FK CONSTRAINT: alert_incidents fk_alert_incidents_project_probe
ALTER TABLE ONLY public.alert_incidents
    ADD CONSTRAINT fk_alert_incidents_project_probe FOREIGN KEY (project_id, probe_id) REFERENCES public.probes(project_id, id);

-- FK CONSTRAINT: alert_incidents fk_alert_incidents_project_rule
ALTER TABLE ONLY public.alert_incidents
    ADD CONSTRAINT fk_alert_incidents_project_rule FOREIGN KEY (project_id, rule_id) REFERENCES public.alert_rules(project_id, id);

-- FK CONSTRAINT: alert_notifications fk_alert_notifications_project_notification
ALTER TABLE ONLY public.alert_notifications
    ADD CONSTRAINT fk_alert_notifications_project_notification FOREIGN KEY (project_id, notification_id) REFERENCES public.notifications(project_id, id);

-- FK CONSTRAINT: alert_notifications fk_alert_notifications_project_rule
ALTER TABLE ONLY public.alert_notifications
    ADD CONSTRAINT fk_alert_notifications_project_rule FOREIGN KEY (project_id, rule_id) REFERENCES public.alert_rules(project_id, id) ON DELETE CASCADE;

-- FK CONSTRAINT: alert_rule_pending_evaluations fk_alert_rule_pending_evaluations_project_check
ALTER TABLE ONLY public.alert_rule_pending_evaluations
    ADD CONSTRAINT fk_alert_rule_pending_evaluations_project_check FOREIGN KEY (project_id, check_id) REFERENCES public.checks(project_id, id) ON DELETE CASCADE;

-- FK CONSTRAINT: alert_rule_pending_evaluations fk_alert_rule_pending_evaluations_project_probe
ALTER TABLE ONLY public.alert_rule_pending_evaluations
    ADD CONSTRAINT fk_alert_rule_pending_evaluations_project_probe FOREIGN KEY (project_id, probe_id) REFERENCES public.probes(project_id, id) ON DELETE CASCADE;

-- FK CONSTRAINT: alert_rule_pending_evaluations fk_alert_rule_pending_evaluations_project_rule
ALTER TABLE ONLY public.alert_rule_pending_evaluations
    ADD CONSTRAINT fk_alert_rule_pending_evaluations_project_rule FOREIGN KEY (project_id, rule_id) REFERENCES public.alert_rules(project_id, id) ON DELETE CASCADE;

-- FK CONSTRAINT: alert_rules fk_alert_rules_project_check
ALTER TABLE ONLY public.alert_rules
    ADD CONSTRAINT fk_alert_rules_project_check FOREIGN KEY (project_id, check_id) REFERENCES public.checks(project_id, id);

-- FK CONSTRAINT: alert_rules fk_alert_rules_project_probe
ALTER TABLE ONLY public.alert_rules
    ADD CONSTRAINT fk_alert_rules_project_probe FOREIGN KEY (project_id, probe_id) REFERENCES public.probes(project_id, id);

-- FK CONSTRAINT: check_labels fk_check_labels_project_check
ALTER TABLE ONLY public.check_labels
    ADD CONSTRAINT fk_check_labels_project_check FOREIGN KEY (project_id, check_id) REFERENCES public.checks(project_id, id) ON DELETE CASCADE;

-- FK CONSTRAINT: check_labels fk_check_labels_project_label
ALTER TABLE ONLY public.check_labels
    ADD CONSTRAINT fk_check_labels_project_label FOREIGN KEY (project_id, label_id) REFERENCES public.labels(project_id, id) ON DELETE CASCADE;

-- FK CONSTRAINT: notification_outbox fk_notification_outbox_project_incident
ALTER TABLE ONLY public.notification_outbox
    ADD CONSTRAINT fk_notification_outbox_project_incident FOREIGN KEY (project_id, incident_id) REFERENCES public.alert_incidents(project_id, id);

-- FK CONSTRAINT: notification_outbox fk_notification_outbox_project_notification
ALTER TABLE ONLY public.notification_outbox
    ADD CONSTRAINT fk_notification_outbox_project_notification FOREIGN KEY (project_id, notification_id) REFERENCES public.notifications(project_id, id);

-- FK CONSTRAINT: notification_outbox fk_notification_outbox_project_rule
ALTER TABLE ONLY public.notification_outbox
    ADD CONSTRAINT fk_notification_outbox_project_rule FOREIGN KEY (project_id, rule_id) REFERENCES public.alert_rules(project_id, id);

-- FK CONSTRAINT: probe_check_assignments fk_probe_check_assignments_project_check
ALTER TABLE ONLY public.probe_check_assignments
    ADD CONSTRAINT fk_probe_check_assignments_project_check FOREIGN KEY (project_id, check_id) REFERENCES public.checks(project_id, id);

-- FK CONSTRAINT: probe_check_assignments fk_probe_check_assignments_project_probe
ALTER TABLE ONLY public.probe_check_assignments
    ADD CONSTRAINT fk_probe_check_assignments_project_probe FOREIGN KEY (project_id, probe_id) REFERENCES public.probes(project_id, id);

-- FK CONSTRAINT: probe_labels fk_probe_labels_project_label
ALTER TABLE ONLY public.probe_labels
    ADD CONSTRAINT fk_probe_labels_project_label FOREIGN KEY (project_id, label_id) REFERENCES public.labels(project_id, id) ON DELETE CASCADE;

-- FK CONSTRAINT: probe_labels fk_probe_labels_project_probe
ALTER TABLE ONLY public.probe_labels
    ADD CONSTRAINT fk_probe_labels_project_probe FOREIGN KEY (project_id, probe_id) REFERENCES public.probes(project_id, id) ON DELETE CASCADE;

-- FK CONSTRAINT: public_status_page_element_assignments fk_public_status_page_element_assignments_element
ALTER TABLE ONLY public.public_status_page_element_assignments
    ADD CONSTRAINT fk_public_status_page_element_assignments_element FOREIGN KEY (public_page_id, element_id) REFERENCES public.public_status_page_elements(public_page_id, id) ON DELETE CASCADE;

-- FK CONSTRAINT: public_status_page_element_assignments fk_public_status_page_element_assignments_page_project
ALTER TABLE ONLY public.public_status_page_element_assignments
    ADD CONSTRAINT fk_public_status_page_element_assignments_page_project FOREIGN KEY (public_page_id, project_id) REFERENCES public.public_status_pages(id, project_id) ON DELETE CASCADE;

-- FK CONSTRAINT: public_status_page_elements fk_public_status_page_elements_check
ALTER TABLE ONLY public.public_status_page_elements
    ADD CONSTRAINT fk_public_status_page_elements_check FOREIGN KEY (project_id, check_id) REFERENCES public.checks(project_id, id) ON DELETE CASCADE;

-- FK CONSTRAINT: public_status_page_elements fk_public_status_page_elements_page_project
ALTER TABLE ONLY public.public_status_page_elements
    ADD CONSTRAINT fk_public_status_page_elements_page_project FOREIGN KEY (public_page_id, project_id) REFERENCES public.public_status_pages(id, project_id) ON DELETE CASCADE;

-- FK CONSTRAINT: public_status_page_elements fk_public_status_page_elements_parent
ALTER TABLE ONLY public.public_status_page_elements
    ADD CONSTRAINT fk_public_status_page_elements_parent FOREIGN KEY (public_page_id, parent_element_id) REFERENCES public.public_status_page_elements(public_page_id, id) ON DELETE CASCADE;

-- FK CONSTRAINT: http_check_configs http_check_configs_check_id_fkey
ALTER TABLE ONLY public.http_check_configs
    ADD CONSTRAINT http_check_configs_check_id_fkey FOREIGN KEY (check_id) REFERENCES public.checks(id) ON DELETE CASCADE;

-- FK CONSTRAINT: labels labels_project_id_fkey
ALTER TABLE ONLY public.labels
    ADD CONSTRAINT labels_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id);

-- FK CONSTRAINT: notification_outbox notification_outbox_project_id_fkey
ALTER TABLE ONLY public.notification_outbox
    ADD CONSTRAINT notification_outbox_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id);

-- FK CONSTRAINT: notifications notifications_created_by_user_id_fkey
ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id);

-- FK CONSTRAINT: notifications notifications_project_id_fkey
ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id);

-- FK CONSTRAINT: external_auth_flows external_auth_flows_session_id_fkey
ALTER TABLE ONLY public.external_auth_flows
    ADD CONSTRAINT external_auth_flows_session_id_fkey FOREIGN KEY (session_id) REFERENCES public.auth_sessions(id) ON DELETE CASCADE;

-- FK CONSTRAINT: password_credentials password_credentials_user_id_fkey
ALTER TABLE ONLY public.password_credentials
    ADD CONSTRAINT password_credentials_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

-- FK CONSTRAINT: password_reset_tokens password_reset_tokens_user_id_fkey
ALTER TABLE ONLY public.password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

-- FK CONSTRAINT: ping_check_configs ping_check_configs_check_id_fkey
ALTER TABLE ONLY public.ping_check_configs
    ADD CONSTRAINT ping_check_configs_check_id_fkey FOREIGN KEY (check_id) REFERENCES public.checks(id) ON DELETE CASCADE;

-- FK CONSTRAINT: probe_check_assignments probe_check_assignments_project_id_fkey
ALTER TABLE ONLY public.probe_check_assignments
    ADD CONSTRAINT probe_check_assignments_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id);

-- FK CONSTRAINT: probe_credentials probe_credentials_probe_id_fkey
ALTER TABLE ONLY public.probe_credentials
    ADD CONSTRAINT probe_credentials_probe_id_fkey FOREIGN KEY (probe_id) REFERENCES public.probes(id) ON DELETE CASCADE;

-- FK CONSTRAINT: probe_statuses probe_statuses_probe_id_fkey
ALTER TABLE ONLY public.probe_statuses
    ADD CONSTRAINT probe_statuses_probe_id_fkey FOREIGN KEY (probe_id) REFERENCES public.probes(id) ON DELETE CASCADE;

-- FK CONSTRAINT: probes probes_project_id_fkey
ALTER TABLE ONLY public.probes
    ADD CONSTRAINT probes_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id);

-- FK CONSTRAINT: project_invites project_invites_invited_by_user_id_fkey
ALTER TABLE ONLY public.project_invites
    ADD CONSTRAINT project_invites_invited_by_user_id_fkey FOREIGN KEY (invited_by_user_id) REFERENCES public.users(id);

-- FK CONSTRAINT: project_invites project_invites_invited_user_id_fkey
ALTER TABLE ONLY public.project_invites
    ADD CONSTRAINT project_invites_invited_user_id_fkey FOREIGN KEY (invited_user_id) REFERENCES public.users(id);

-- FK CONSTRAINT: project_invites project_invites_project_id_fkey
ALTER TABLE ONLY public.project_invites
    ADD CONSTRAINT project_invites_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id);

-- FK CONSTRAINT: project_members project_members_project_id_fkey
ALTER TABLE ONLY public.project_members
    ADD CONSTRAINT project_members_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id);

-- FK CONSTRAINT: project_members project_members_user_id_fkey
ALTER TABLE ONLY public.project_members
    ADD CONSTRAINT project_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);

-- FK CONSTRAINT: projects projects_created_by_user_id_fkey
ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id);

-- FK CONSTRAINT: public_status_page_element_assignments public_status_page_element_assignments_assignment_id_fkey
ALTER TABLE ONLY public.public_status_page_element_assignments
    ADD CONSTRAINT public_status_page_element_assignments_assignment_id_fkey FOREIGN KEY (assignment_id) REFERENCES public.probe_check_assignments(id) ON DELETE CASCADE;

-- FK CONSTRAINT: public_status_pages public_status_pages_created_by_user_id_fkey
ALTER TABLE ONLY public.public_status_pages
    ADD CONSTRAINT public_status_pages_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id);

-- FK CONSTRAINT: public_status_pages public_status_pages_project_id_fkey
ALTER TABLE ONLY public.public_status_pages
    ADD CONSTRAINT public_status_pages_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id);

-- FK CONSTRAINT: system_setting_audit_events system_setting_audit_events_updated_by_user_id_fkey
ALTER TABLE ONLY public.system_setting_audit_events
    ADD CONSTRAINT system_setting_audit_events_updated_by_user_id_fkey FOREIGN KEY (updated_by_user_id) REFERENCES public.users(id);

-- FK CONSTRAINT: system_settings system_settings_updated_by_user_id_fkey
ALTER TABLE ONLY public.system_settings
    ADD CONSTRAINT system_settings_updated_by_user_id_fkey FOREIGN KEY (updated_by_user_id) REFERENCES public.users(id);

-- FK CONSTRAINT: system_user_roles system_user_roles_user_id_fkey
ALTER TABLE ONLY public.system_user_roles
    ADD CONSTRAINT system_user_roles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

-- FK CONSTRAINT: user_identities user_identities_user_id_fkey
ALTER TABLE ONLY public.user_identities
    ADD CONSTRAINT user_identities_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

-- FK CONSTRAINT: tcp_check_configs tcp_check_configs_check_id_fkey
ALTER TABLE ONLY public.tcp_check_configs
    ADD CONSTRAINT tcp_check_configs_check_id_fkey FOREIGN KEY (check_id) REFERENCES public.checks(id) ON DELETE CASCADE;

-- FK CONSTRAINT: traceroute_check_configs traceroute_check_configs_check_id_fkey
ALTER TABLE ONLY public.traceroute_check_configs
    ADD CONSTRAINT traceroute_check_configs_check_id_fkey FOREIGN KEY (check_id) REFERENCES public.checks(id) ON DELETE CASCADE;

-- Final v0.1.0 time-series schema.
-- Prerequisites: TimescaleDB 2.20.3+, the result status/ip-family enum types,
-- and the probes/checks tables with their internal_id unique constraints.

CREATE TABLE ping_results (
    probe_id bigint NOT NULL,
    check_id bigint NOT NULL,
    started_at timestamptz NOT NULL,
    finished_at timestamptz NOT NULL,
    duration_ms integer NOT NULL,
    status ping_status NOT NULL,
    sent_count integer NOT NULL,
    received_count integer NOT NULL,
    loss_percent double precision NOT NULL,
    rtt_min_ms double precision,
    rtt_avg_ms double precision,
    rtt_median_ms double precision,
    rtt_max_ms double precision,
    rtt_stddev_ms double precision,
    rtt_samples_ms double precision[] NOT NULL DEFAULT '{}'::double precision[],
    resolved_ip inet,
    ip_family ip_family,
    error_code text,
    error_message text,
    PRIMARY KEY (probe_id, check_id, started_at),
    CONSTRAINT ping_results_finished_at_after_started_at CHECK (finished_at >= started_at),
    CONSTRAINT ping_results_duration_ms_non_negative CHECK (duration_ms >= 0),
    CONSTRAINT ping_results_sent_count_non_negative CHECK (sent_count >= 0),
    CONSTRAINT ping_results_received_count_range CHECK (received_count >= 0 AND received_count <= sent_count),
    CONSTRAINT ping_results_loss_percent_range CHECK (loss_percent >= 0 AND loss_percent <= 100),
    CONSTRAINT ping_results_rtt_min_ms_non_negative CHECK (rtt_min_ms IS NULL OR rtt_min_ms >= 0),
    CONSTRAINT ping_results_rtt_avg_ms_non_negative CHECK (rtt_avg_ms IS NULL OR rtt_avg_ms >= 0),
    CONSTRAINT ping_results_rtt_median_ms_non_negative CHECK (rtt_median_ms IS NULL OR rtt_median_ms >= 0),
    CONSTRAINT ping_results_rtt_max_ms_non_negative CHECK (rtt_max_ms IS NULL OR rtt_max_ms >= 0),
    CONSTRAINT ping_results_rtt_stddev_ms_non_negative CHECK (rtt_stddev_ms IS NULL OR rtt_stddev_ms >= 0),
    CONSTRAINT ping_results_rtt_order CHECK (
        (rtt_min_ms IS NULL OR rtt_max_ms IS NULL OR rtt_min_ms <= rtt_max_ms) AND
        (rtt_min_ms IS NULL OR rtt_avg_ms IS NULL OR rtt_min_ms <= rtt_avg_ms) AND
        (rtt_avg_ms IS NULL OR rtt_max_ms IS NULL OR rtt_avg_ms <= rtt_max_ms)
    ),
    CONSTRAINT ping_results_error_code_not_empty CHECK (error_code IS NULL OR length(btrim(error_code)) > 0),
    CONSTRAINT ping_results_error_message_not_empty CHECK (error_message IS NULL OR length(btrim(error_message)) > 0),
    CONSTRAINT fk_ping_results_probe FOREIGN KEY (probe_id) REFERENCES probes(internal_id),
    CONSTRAINT fk_ping_results_check FOREIGN KEY (check_id) REFERENCES checks(internal_id)
);

SELECT create_hypertable('ping_results', 'started_at', if_not_exists => TRUE);
SELECT set_chunk_time_interval('ping_results', INTERVAL '1 day');

CREATE INDEX idx_ping_results_probe_check_started_at
    ON ping_results (probe_id, check_id, started_at DESC);
CREATE INDEX idx_ping_results_check_id_started_at ON ping_results (check_id, started_at DESC);
CREATE INDEX idx_ping_results_probe_id_started_at ON ping_results (probe_id, started_at DESC);
CREATE INDEX idx_ping_results_status_started_at ON ping_results (status, started_at DESC);

CREATE TABLE tcp_results (
    probe_id bigint NOT NULL,
    check_id bigint NOT NULL,
    started_at timestamptz NOT NULL,
    finished_at timestamptz NOT NULL,
    duration_ms integer NOT NULL,
    status tcp_status NOT NULL,
    connect_duration_ms double precision,
    resolved_ip inet,
    ip_family ip_family,
    error_code text,
    error_message text,
    PRIMARY KEY (probe_id, check_id, started_at),
    CONSTRAINT tcp_results_finished_at_after_started_at CHECK (finished_at >= started_at),
    CONSTRAINT tcp_results_duration_ms_non_negative CHECK (duration_ms >= 0),
    CONSTRAINT tcp_results_connect_duration_ms_non_negative CHECK (
        connect_duration_ms IS NULL OR connect_duration_ms >= 0
    ),
    CONSTRAINT tcp_results_error_code_not_empty CHECK (error_code IS NULL OR length(btrim(error_code)) > 0),
    CONSTRAINT tcp_results_error_message_not_empty CHECK (error_message IS NULL OR length(btrim(error_message)) > 0),
    CONSTRAINT fk_tcp_results_probe FOREIGN KEY (probe_id) REFERENCES probes(internal_id),
    CONSTRAINT fk_tcp_results_check FOREIGN KEY (check_id) REFERENCES checks(internal_id)
);

SELECT create_hypertable('tcp_results', 'started_at', if_not_exists => TRUE);
SELECT set_chunk_time_interval('tcp_results', INTERVAL '1 day');

CREATE INDEX idx_tcp_results_probe_check_started_at
    ON tcp_results (probe_id, check_id, started_at DESC);
CREATE INDEX idx_tcp_results_check_id_started_at ON tcp_results (check_id, started_at DESC);
CREATE INDEX idx_tcp_results_probe_id_started_at ON tcp_results (probe_id, started_at DESC);
CREATE INDEX idx_tcp_results_status_started_at ON tcp_results (status, started_at DESC);

CREATE TABLE http_results (
    probe_id bigint NOT NULL,
    check_id bigint NOT NULL,
    started_at timestamptz NOT NULL,
    finished_at timestamptz NOT NULL,
    duration_ms integer NOT NULL,
    status http_status NOT NULL,
    dns_duration_ms double precision,
    connect_duration_ms double precision,
    tls_duration_ms double precision,
    ttfb_duration_ms double precision,
    resolved_ip inet,
    ip_family ip_family,
    status_code integer,
    final_url text,
    redirect_count integer NOT NULL DEFAULT 0,
    response_bytes bigint,
    response_truncated boolean NOT NULL DEFAULT false,
    body_matched boolean,
    tls_version text,
    tls_cipher_suite text,
    certificate_not_before timestamptz,
    certificate_not_after timestamptz,
    error_code text,
    error_message text,
    PRIMARY KEY (probe_id, check_id, started_at),
    CONSTRAINT fk_http_results_probe FOREIGN KEY (probe_id) REFERENCES probes(internal_id),
    CONSTRAINT fk_http_results_check FOREIGN KEY (check_id) REFERENCES checks(internal_id),
    CONSTRAINT http_results_finished_at_after_started_at CHECK (finished_at >= started_at),
    CONSTRAINT http_results_duration_ms_non_negative CHECK (duration_ms >= 0),
    CONSTRAINT http_results_dns_duration_ms_non_negative CHECK (dns_duration_ms IS NULL OR dns_duration_ms >= 0),
    CONSTRAINT http_results_connect_duration_ms_non_negative CHECK (connect_duration_ms IS NULL OR connect_duration_ms >= 0),
    CONSTRAINT http_results_tls_duration_ms_non_negative CHECK (tls_duration_ms IS NULL OR tls_duration_ms >= 0),
    CONSTRAINT http_results_ttfb_duration_ms_non_negative CHECK (ttfb_duration_ms IS NULL OR ttfb_duration_ms >= 0),
    CONSTRAINT http_results_status_code_range CHECK (status_code IS NULL OR status_code BETWEEN 100 AND 599),
    CONSTRAINT http_results_redirect_count_range CHECK (redirect_count BETWEEN 0 AND 10),
    CONSTRAINT http_results_response_bytes_non_negative CHECK (response_bytes IS NULL OR response_bytes >= 0),
    CONSTRAINT http_results_error_code_not_empty CHECK (error_code IS NULL OR length(btrim(error_code)) > 0),
    CONSTRAINT http_results_error_message_not_empty CHECK (error_message IS NULL OR length(btrim(error_message)) > 0)
);

SELECT create_hypertable('http_results', 'started_at', if_not_exists => TRUE);
SELECT set_chunk_time_interval('http_results', INTERVAL '1 day');

CREATE INDEX idx_http_results_probe_check_started_at ON http_results (probe_id, check_id, started_at DESC);
CREATE INDEX idx_http_results_check_id_started_at ON http_results (check_id, started_at DESC);
CREATE INDEX idx_http_results_probe_id_started_at ON http_results (probe_id, started_at DESC);
CREATE INDEX idx_http_results_status_started_at ON http_results (status, started_at DESC);

CREATE TABLE traceroute_results (
    probe_id bigint NOT NULL,
    check_id bigint NOT NULL,
    started_at timestamptz NOT NULL,
    finished_at timestamptz NOT NULL,
    duration_ms integer NOT NULL,
    status traceroute_status NOT NULL,
    resolved_ip inet,
    ip_family ip_family,
    destination_reached boolean NOT NULL DEFAULT false,
    hop_count integer NOT NULL,
    error_code text,
    error_message text,
    PRIMARY KEY (probe_id, check_id, started_at),
    CONSTRAINT traceroute_results_finished_at_after_started_at CHECK (finished_at >= started_at),
    CONSTRAINT traceroute_results_duration_ms_non_negative CHECK (duration_ms >= 0),
    CONSTRAINT traceroute_results_hop_count_non_negative CHECK (hop_count >= 0),
    CONSTRAINT traceroute_results_error_code_not_empty CHECK (error_code IS NULL OR length(btrim(error_code)) > 0),
    CONSTRAINT traceroute_results_error_message_not_empty CHECK (error_message IS NULL OR length(btrim(error_message)) > 0),
    CONSTRAINT fk_traceroute_results_probe FOREIGN KEY (probe_id) REFERENCES probes(internal_id),
    CONSTRAINT fk_traceroute_results_check FOREIGN KEY (check_id) REFERENCES checks(internal_id)
);

SELECT create_hypertable('traceroute_results', 'started_at', if_not_exists => TRUE);
SELECT set_chunk_time_interval('traceroute_results', INTERVAL '1 day');

CREATE INDEX idx_traceroute_results_probe_check_started_at
    ON traceroute_results (probe_id, check_id, started_at DESC);
CREATE INDEX idx_traceroute_results_check_id_started_at
    ON traceroute_results (check_id, started_at DESC);
CREATE INDEX idx_traceroute_results_probe_id_started_at
    ON traceroute_results (probe_id, started_at DESC);
CREATE INDEX idx_traceroute_results_status_started_at
    ON traceroute_results (status, started_at DESC);

-- This is a hypertable, so it intentionally has no foreign key to the
-- traceroute_results hypertable. Runtime writes both in one transaction.
CREATE TABLE traceroute_result_hops (
    probe_id bigint NOT NULL,
    check_id bigint NOT NULL,
    started_at timestamptz NOT NULL,
    hop_index integer NOT NULL,
    address inet,
    hostname text,
    sent_count integer NOT NULL,
    received_count integer NOT NULL,
    loss_percent double precision NOT NULL,
    rtt_min_ms double precision,
    rtt_avg_ms double precision,
    rtt_median_ms double precision,
    rtt_max_ms double precision,
    rtt_stddev_ms double precision,
    rtt_samples_ms double precision[] NOT NULL DEFAULT '{}'::double precision[],
    error_code text,
    error_message text,
    CONSTRAINT traceroute_result_hops_pkey PRIMARY KEY (probe_id, check_id, started_at, hop_index),
    CONSTRAINT traceroute_result_hops_hop_index_positive CHECK (hop_index > 0),
    CONSTRAINT traceroute_result_hops_sent_count_non_negative CHECK (sent_count >= 0),
    CONSTRAINT traceroute_result_hops_received_count_range CHECK (received_count >= 0 AND received_count <= sent_count),
    CONSTRAINT traceroute_result_hops_loss_percent_range CHECK (loss_percent >= 0 AND loss_percent <= 100),
    CONSTRAINT traceroute_result_hops_rtt_min_ms_non_negative CHECK (rtt_min_ms IS NULL OR rtt_min_ms >= 0),
    CONSTRAINT traceroute_result_hops_rtt_avg_ms_non_negative CHECK (rtt_avg_ms IS NULL OR rtt_avg_ms >= 0),
    CONSTRAINT traceroute_result_hops_rtt_median_ms_non_negative CHECK (rtt_median_ms IS NULL OR rtt_median_ms >= 0),
    CONSTRAINT traceroute_result_hops_rtt_max_ms_non_negative CHECK (rtt_max_ms IS NULL OR rtt_max_ms >= 0),
    CONSTRAINT traceroute_result_hops_rtt_stddev_ms_non_negative CHECK (rtt_stddev_ms IS NULL OR rtt_stddev_ms >= 0),
    CONSTRAINT traceroute_result_hops_rtt_order CHECK (
        (rtt_min_ms IS NULL OR rtt_max_ms IS NULL OR rtt_min_ms <= rtt_max_ms) AND
        (rtt_min_ms IS NULL OR rtt_avg_ms IS NULL OR rtt_min_ms <= rtt_avg_ms) AND
        (rtt_avg_ms IS NULL OR rtt_max_ms IS NULL OR rtt_avg_ms <= rtt_max_ms)
    ),
    CONSTRAINT traceroute_result_hops_hostname_not_empty CHECK (hostname IS NULL OR length(btrim(hostname)) > 0),
    CONSTRAINT traceroute_result_hops_error_code_not_empty CHECK (error_code IS NULL OR length(btrim(error_code)) > 0),
    CONSTRAINT traceroute_result_hops_error_message_not_empty CHECK (error_message IS NULL OR length(btrim(error_message)) > 0)
);

SELECT create_hypertable(
    'traceroute_result_hops',
    'started_at',
    chunk_time_interval => INTERVAL '1 day',
    create_default_indexes => FALSE,
    if_not_exists => TRUE
);

SELECT add_retention_policy('ping_results', INTERVAL '3 days', if_not_exists => TRUE);
SELECT add_retention_policy('tcp_results', INTERVAL '3 days', if_not_exists => TRUE);
SELECT add_retention_policy('http_results', INTERVAL '3 days', if_not_exists => TRUE);
SELECT add_retention_policy('traceroute_results', INTERVAL '3 days', if_not_exists => TRUE);
SELECT add_retention_policy('traceroute_result_hops', INTERVAL '3 days', if_not_exists => TRUE);

CREATE MATERIALIZED VIEW ping_result_rollups_1m
WITH (
    timescaledb.continuous,
    timescaledb.materialized_only = true
) AS
SELECT
    time_bucket('1 minute', started_at) AS bucket,
    probe_id,
    check_id,
    count(*)::bigint AS result_count,
    sum(CASE WHEN status = 'successful' THEN 1 ELSE 0 END)::bigint AS successful_count,
    sum(sent_count)::bigint AS sent_count,
    sum(received_count)::bigint AS received_count,
    coalesce(sum(rtt_avg_ms), 0)::double precision AS rtt_avg_sum_ms,
    count(rtt_avg_ms)::bigint AS rtt_avg_count,
    min(rtt_min_ms)::double precision AS rtt_min_ms,
    max(rtt_max_ms)::double precision AS rtt_max_ms
FROM ping_results
GROUP BY bucket, probe_id, check_id
WITH DATA;

SELECT add_continuous_aggregate_policy(
    'ping_result_rollups_1m',
    start_offset => INTERVAL '3 days',
    end_offset => INTERVAL '2 minutes',
    schedule_interval => INTERVAL '1 minute',
    if_not_exists => TRUE
);

CREATE MATERIALIZED VIEW tcp_result_rollups_1m
WITH (
    timescaledb.continuous,
    timescaledb.materialized_only = true
) AS
SELECT
    time_bucket('1 minute', started_at) AS bucket,
    probe_id,
    check_id,
    count(*)::bigint AS result_count,
    sum(CASE WHEN status = 'successful' THEN 1 ELSE 0 END)::bigint AS successful_count,
    sum(CASE WHEN status = 'timeout' THEN 1 ELSE 0 END)::bigint AS timeout_count,
    sum(CASE WHEN status = 'error' THEN 1 ELSE 0 END)::bigint AS error_count,
    coalesce(sum(connect_duration_ms), 0)::double precision AS connect_duration_sum_ms,
    count(connect_duration_ms)::bigint AS connect_duration_count,
    min(connect_duration_ms)::double precision AS connect_duration_min_ms,
    max(connect_duration_ms)::double precision AS connect_duration_max_ms
FROM tcp_results
GROUP BY bucket, probe_id, check_id
WITH DATA;

SELECT add_continuous_aggregate_policy(
    'tcp_result_rollups_1m',
    start_offset => INTERVAL '3 days',
    end_offset => INTERVAL '2 minutes',
    schedule_interval => INTERVAL '1 minute',
    if_not_exists => TRUE
);

CREATE MATERIALIZED VIEW http_result_rollups_1m
WITH (
    timescaledb.continuous,
    timescaledb.materialized_only = false
) AS
SELECT
    time_bucket('1 minute', started_at) AS bucket,
    probe_id,
    check_id,
    count(*)::bigint AS result_count,
    count(*) FILTER (WHERE status = 'successful')::bigint AS successful_count,
    count(*) FILTER (WHERE status = 'timeout')::bigint AS timeout_count,
    count(*) FILTER (WHERE status = 'error')::bigint AS error_count,
    sum(duration_ms)::double precision AS duration_sum_ms,
    count(duration_ms)::bigint AS duration_count,
    min(duration_ms)::double precision AS duration_min_ms,
    max(duration_ms)::double precision AS duration_max_ms,
    sum(dns_duration_ms)::double precision AS dns_duration_sum_ms,
    count(dns_duration_ms)::bigint AS dns_duration_count,
    min(dns_duration_ms)::double precision AS dns_duration_min_ms,
    max(dns_duration_ms)::double precision AS dns_duration_max_ms,
    sum(connect_duration_ms)::double precision AS connect_duration_sum_ms,
    count(connect_duration_ms)::bigint AS connect_duration_count,
    min(connect_duration_ms)::double precision AS connect_duration_min_ms,
    max(connect_duration_ms)::double precision AS connect_duration_max_ms,
    sum(tls_duration_ms)::double precision AS tls_duration_sum_ms,
    count(tls_duration_ms)::bigint AS tls_duration_count,
    min(tls_duration_ms)::double precision AS tls_duration_min_ms,
    max(tls_duration_ms)::double precision AS tls_duration_max_ms,
    sum(ttfb_duration_ms)::double precision AS ttfb_duration_sum_ms,
    count(ttfb_duration_ms)::bigint AS ttfb_duration_count,
    min(ttfb_duration_ms)::double precision AS ttfb_duration_min_ms,
    max(ttfb_duration_ms)::double precision AS ttfb_duration_max_ms,
    sum(response_bytes)::double precision AS response_bytes_sum,
    count(response_bytes)::bigint AS response_bytes_count,
    min(certificate_not_after) AS certificate_not_after_min
FROM http_results
GROUP BY bucket, probe_id, check_id
WITH DATA;

SELECT add_continuous_aggregate_policy(
    'http_result_rollups_1m',
    start_offset => INTERVAL '3 days',
    end_offset => INTERVAL '2 minutes',
    schedule_interval => INTERVAL '1 minute',
    if_not_exists => TRUE
);

CREATE TABLE traceroute_sampled_runs_1m (
    bucket timestamptz NOT NULL,
    probe_id bigint NOT NULL,
    check_id bigint NOT NULL,
    sampled_started_at timestamptz NOT NULL,
    finished_at timestamptz NOT NULL,
    duration_ms integer NOT NULL,
    status traceroute_status NOT NULL,
    resolved_ip inet,
    ip_family ip_family,
    destination_reached boolean NOT NULL,
    hop_count integer NOT NULL,
    path_signature text,
    hops jsonb NOT NULL DEFAULT '[]'::jsonb,
    CONSTRAINT traceroute_sampled_runs_1m_pkey PRIMARY KEY (probe_id, check_id, bucket),
    CONSTRAINT traceroute_sampled_runs_1m_finished_at_after_sampled_started_at CHECK (finished_at >= sampled_started_at),
    CONSTRAINT traceroute_sampled_runs_1m_duration_ms_non_negative CHECK (duration_ms >= 0),
    CONSTRAINT traceroute_sampled_runs_1m_hop_count_non_negative CHECK (hop_count >= 0),
    CONSTRAINT traceroute_sampled_runs_1m_path_signature_not_empty CHECK (path_signature IS NULL OR length(btrim(path_signature)) > 0),
    CONSTRAINT traceroute_sampled_runs_1m_hops_is_array CHECK (jsonb_typeof(hops) = 'array'),
    CONSTRAINT fk_traceroute_sampled_runs_1m_probe FOREIGN KEY (probe_id) REFERENCES probes(internal_id),
    CONSTRAINT fk_traceroute_sampled_runs_1m_check FOREIGN KEY (check_id) REFERENCES checks(internal_id)
);

SELECT create_hypertable(
    'traceroute_sampled_runs_1m',
    'bucket',
    chunk_time_interval => INTERVAL '7 days',
    create_default_indexes => FALSE,
    if_not_exists => TRUE
);

CREATE INDEX idx_traceroute_sampled_runs_1m_probe_check_bucket
    ON traceroute_sampled_runs_1m (probe_id, check_id, bucket DESC);
CREATE INDEX idx_traceroute_sampled_runs_1m_check_bucket
    ON traceroute_sampled_runs_1m (check_id, bucket DESC);

-- +goose StatementBegin
CREATE OR REPLACE PROCEDURE refresh_traceroute_sampled_runs_1m(
    job_id integer DEFAULT NULL,
    config jsonb DEFAULT '{}'::jsonb
)
LANGUAGE plpgsql
AS $$
DECLARE
    refresh_lookback interval;
    refresh_lag interval;
    window_from timestamptz;
    window_to timestamptz;
BEGIN
    refresh_lookback := COALESCE((config ->> 'lookback')::interval, INTERVAL '10 minutes');
    refresh_lag := COALESCE((config ->> 'refresh_lag')::interval, INTERVAL '1 minute');
    window_to := COALESCE((config ->> 'to')::timestamptz, now() - refresh_lag);
    window_from := COALESCE((config ->> 'from')::timestamptz, window_to - refresh_lookback);

    IF window_from >= window_to THEN
        RAISE EXCEPTION 'refresh window start must be before end: from %, to %', window_from, window_to;
    END IF;

    INSERT INTO traceroute_sampled_runs_1m (
        bucket,
        probe_id,
        check_id,
        sampled_started_at,
        finished_at,
        duration_ms,
        status,
        resolved_ip,
        ip_family,
        destination_reached,
        hop_count,
        path_signature,
        hops
    )
    WITH candidate_runs AS (
        SELECT
            time_bucket(INTERVAL '1 minute', traceroute_results.started_at) AS bucket,
            traceroute_results.probe_id,
            traceroute_results.check_id,
            traceroute_results.started_at,
            traceroute_results.finished_at,
            traceroute_results.duration_ms,
            traceroute_results.status,
            traceroute_results.resolved_ip,
            traceroute_results.ip_family,
            traceroute_results.destination_reached,
            traceroute_results.hop_count
        FROM traceroute_results
        WHERE traceroute_results.started_at >= window_from
            AND traceroute_results.started_at < window_to
    ),
    sampled_runs AS (
        SELECT DISTINCT ON (candidate_runs.probe_id, candidate_runs.check_id, candidate_runs.bucket)
            candidate_runs.bucket,
            candidate_runs.probe_id,
            candidate_runs.check_id,
            candidate_runs.started_at,
            candidate_runs.finished_at,
            candidate_runs.duration_ms,
            candidate_runs.status,
            candidate_runs.resolved_ip,
            candidate_runs.ip_family,
            candidate_runs.destination_reached,
            candidate_runs.hop_count
        FROM candidate_runs
        ORDER BY
            candidate_runs.probe_id,
            candidate_runs.check_id,
            candidate_runs.bucket,
            candidate_runs.started_at DESC
    ),
    sampled_hops AS (
        SELECT
            sampled_runs.bucket,
            sampled_runs.probe_id,
            sampled_runs.check_id,
            sampled_runs.started_at,
            sampled_runs.finished_at,
            sampled_runs.duration_ms,
            sampled_runs.status,
            sampled_runs.resolved_ip,
            sampled_runs.ip_family,
            sampled_runs.destination_reached,
            sampled_runs.hop_count,
            string_agg(
                COALESCE(
                    traceroute_result_hops.address::text,
                    traceroute_result_hops.hostname,
                    traceroute_result_hops.error_code,
                    'unknown:' || traceroute_result_hops.hop_index::text
                ),
                '>' ORDER BY traceroute_result_hops.hop_index
            ) FILTER (WHERE traceroute_result_hops.hop_index IS NOT NULL) AS path_signature,
            COALESCE(
                jsonb_agg(
                    jsonb_build_object(
                        'hopIndex', traceroute_result_hops.hop_index,
                        'address', traceroute_result_hops.address::text,
                        'hostname', traceroute_result_hops.hostname,
                        'sentCount', traceroute_result_hops.sent_count,
                        'receivedCount', traceroute_result_hops.received_count,
                        'lossPercent', traceroute_result_hops.loss_percent,
                        'rttMinMs', traceroute_result_hops.rtt_min_ms,
                        'rttAvgMs', traceroute_result_hops.rtt_avg_ms,
                        'rttMedianMs', traceroute_result_hops.rtt_median_ms,
                        'rttMaxMs', traceroute_result_hops.rtt_max_ms,
                        'rttStddevMs', traceroute_result_hops.rtt_stddev_ms,
                        'rttSamplesMs', traceroute_result_hops.rtt_samples_ms,
                        'errorCode', traceroute_result_hops.error_code,
                        'errorMessage', traceroute_result_hops.error_message
                    )
                    ORDER BY traceroute_result_hops.hop_index
                ) FILTER (WHERE traceroute_result_hops.hop_index IS NOT NULL),
                '[]'::jsonb
            ) AS hops
        FROM sampled_runs
        LEFT JOIN traceroute_result_hops
            ON traceroute_result_hops.probe_id = sampled_runs.probe_id
            AND traceroute_result_hops.check_id = sampled_runs.check_id
            AND traceroute_result_hops.started_at = sampled_runs.started_at
        GROUP BY
            sampled_runs.bucket,
            sampled_runs.probe_id,
            sampled_runs.check_id,
            sampled_runs.started_at,
            sampled_runs.finished_at,
            sampled_runs.duration_ms,
            sampled_runs.status,
            sampled_runs.resolved_ip,
            sampled_runs.ip_family,
            sampled_runs.destination_reached,
            sampled_runs.hop_count
    )
    SELECT
        sampled_hops.bucket,
        sampled_hops.probe_id,
        sampled_hops.check_id,
        sampled_hops.started_at,
        sampled_hops.finished_at,
        sampled_hops.duration_ms,
        sampled_hops.status,
        sampled_hops.resolved_ip,
        sampled_hops.ip_family,
        sampled_hops.destination_reached,
        sampled_hops.hop_count,
        sampled_hops.path_signature,
        sampled_hops.hops
    FROM sampled_hops
    ON CONFLICT (probe_id, check_id, bucket) DO UPDATE
    SET
        sampled_started_at = EXCLUDED.sampled_started_at,
        finished_at = EXCLUDED.finished_at,
        duration_ms = EXCLUDED.duration_ms,
        status = EXCLUDED.status,
        resolved_ip = EXCLUDED.resolved_ip,
        ip_family = EXCLUDED.ip_family,
        destination_reached = EXCLUDED.destination_reached,
        hop_count = EXCLUDED.hop_count,
        path_signature = EXCLUDED.path_signature,
        hops = EXCLUDED.hops
    WHERE traceroute_sampled_runs_1m.sampled_started_at <= EXCLUDED.sampled_started_at;
END;
$$;
-- +goose StatementEnd

SELECT add_job(
    'refresh_traceroute_sampled_runs_1m',
    INTERVAL '1 minute',
    config => '{"lookback":"10 minutes","refresh_lag":"1 minute"}'::jsonb,
    fixed_schedule => TRUE
)
WHERE NOT EXISTS (
    SELECT 1
    FROM timescaledb_information.jobs
    WHERE proc_schema = 'public'
        AND proc_name = 'refresh_traceroute_sampled_runs_1m'
);

ALTER TABLE ping_results SET (
    timescaledb.enable_columnstore = true,
    timescaledb.segmentby = 'probe_id, check_id',
    timescaledb.orderby = 'started_at DESC'
);

ALTER TABLE tcp_results SET (
    timescaledb.enable_columnstore = true,
    timescaledb.segmentby = 'probe_id, check_id',
    timescaledb.orderby = 'started_at DESC'
);

ALTER TABLE http_results SET (
    timescaledb.enable_columnstore = true,
    timescaledb.segmentby = 'probe_id, check_id',
    timescaledb.orderby = 'started_at DESC'
);

ALTER TABLE traceroute_results SET (
    timescaledb.enable_columnstore = true,
    timescaledb.segmentby = 'probe_id, check_id',
    timescaledb.orderby = 'started_at DESC'
);

ALTER TABLE traceroute_result_hops SET (
    timescaledb.enable_columnstore = true,
    timescaledb.segmentby = 'probe_id, check_id',
    timescaledb.orderby = 'started_at DESC, hop_index ASC'
);

ALTER TABLE traceroute_sampled_runs_1m SET (
    timescaledb.enable_columnstore = true,
    timescaledb.segmentby = 'probe_id, check_id',
    timescaledb.orderby = 'bucket DESC'
);

ALTER MATERIALIZED VIEW ping_result_rollups_1m SET (
    timescaledb.enable_columnstore = true,
    timescaledb.segmentby = 'probe_id, check_id',
    timescaledb.orderby = 'bucket DESC'
);

ALTER MATERIALIZED VIEW tcp_result_rollups_1m SET (
    timescaledb.enable_columnstore = true,
    timescaledb.segmentby = 'probe_id, check_id',
    timescaledb.orderby = 'bucket DESC'
);

ALTER MATERIALIZED VIEW http_result_rollups_1m SET (
    timescaledb.enable_columnstore = true,
    timescaledb.segmentby = 'probe_id, check_id',
    timescaledb.orderby = 'bucket DESC'
);

CALL add_columnstore_policy(
    'ping_results',
    after => INTERVAL '1 day',
    schedule_interval => INTERVAL '6 hours',
    initial_start => now() + INTERVAL '10 minutes',
    if_not_exists => true
);

CALL add_columnstore_policy(
    'tcp_results',
    after => INTERVAL '1 day',
    schedule_interval => INTERVAL '6 hours',
    initial_start => now() + INTERVAL '20 minutes',
    if_not_exists => true
);

CALL add_columnstore_policy(
    'http_results',
    after => INTERVAL '1 day',
    schedule_interval => INTERVAL '6 hours',
    initial_start => now() + INTERVAL '30 minutes',
    if_not_exists => true
);

CALL add_columnstore_policy(
    'traceroute_results',
    after => INTERVAL '1 day',
    schedule_interval => INTERVAL '6 hours',
    initial_start => now() + INTERVAL '40 minutes',
    if_not_exists => true
);

CALL add_columnstore_policy(
    'traceroute_result_hops',
    after => INTERVAL '1 day',
    schedule_interval => INTERVAL '6 hours',
    initial_start => now() + INTERVAL '50 minutes',
    if_not_exists => true
);

CALL add_columnstore_policy(
    'traceroute_sampled_runs_1m',
    after => INTERVAL '7 days',
    schedule_interval => INTERVAL '12 hours',
    initial_start => now() + INTERVAL '60 minutes',
    if_not_exists => true
);

CALL add_columnstore_policy(
    'ping_result_rollups_1m',
    after => INTERVAL '7 days',
    schedule_interval => INTERVAL '12 hours',
    initial_start => now() + INTERVAL '70 minutes',
    if_not_exists => true
);

CALL add_columnstore_policy(
    'tcp_result_rollups_1m',
    after => INTERVAL '7 days',
    schedule_interval => INTERVAL '12 hours',
    initial_start => now() + INTERVAL '80 minutes',
    if_not_exists => true
);

CALL add_columnstore_policy(
    'http_result_rollups_1m',
    after => INTERVAL '7 days',
    schedule_interval => INTERVAL '12 hours',
    initial_start => now() + INTERVAL '90 minutes',
    if_not_exists => true
);

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    RAISE EXCEPTION 'Netstamp v0.1.0 baseline is irreversible; restore a backup or recreate the database';
END;
$$;
-- +goose StatementEnd

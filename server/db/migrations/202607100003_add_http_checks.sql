-- +goose NO TRANSACTION
-- +goose Up
ALTER TYPE check_type ADD VALUE IF NOT EXISTS 'http';

CREATE OR REPLACE VIEW public_status_page_assignment_scope AS
SELECT public_status_page_elements.public_page_id,
       public_status_page_elements.id AS element_id,
       probe_check_assignments.id AS assignment_id,
       probe_check_assignments.project_id,
       probe_check_assignments.probe_id,
       probe_check_assignments.check_id
FROM public_status_page_elements
JOIN probe_check_assignments
  ON public_status_page_elements.assignment_selection_mode = 'all_check'
 AND probe_check_assignments.project_id = public_status_page_elements.project_id
 AND probe_check_assignments.check_id = public_status_page_elements.check_id
 AND probe_check_assignments.deleted_at IS NULL
JOIN checks
  ON checks.project_id = probe_check_assignments.project_id
 AND checks.id = probe_check_assignments.check_id
 AND checks.deleted_at IS NULL
 AND checks.check_type IN ('ping', 'tcp', 'http')
WHERE public_status_page_elements.kind = 'assignment_group'
UNION ALL
SELECT public_status_page_elements.public_page_id,
       public_status_page_elements.id AS element_id,
       probe_check_assignments.id AS assignment_id,
       probe_check_assignments.project_id,
       probe_check_assignments.probe_id,
       probe_check_assignments.check_id
FROM public_status_page_elements
JOIN public_status_page_element_assignments
  ON public_status_page_element_assignments.public_page_id = public_status_page_elements.public_page_id
 AND public_status_page_element_assignments.element_id = public_status_page_elements.id
JOIN probe_check_assignments
  ON probe_check_assignments.id = public_status_page_element_assignments.assignment_id
 AND probe_check_assignments.project_id = public_status_page_elements.project_id
 AND probe_check_assignments.deleted_at IS NULL
JOIN checks
  ON checks.project_id = probe_check_assignments.project_id
 AND checks.id = probe_check_assignments.check_id
 AND checks.deleted_at IS NULL
 AND checks.check_type IN ('ping', 'tcp', 'http')
WHERE public_status_page_elements.kind = 'assignment_group'
  AND public_status_page_elements.assignment_selection_mode = 'selected_assignments';

CREATE TYPE http_method AS ENUM ('GET', 'HEAD', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS');
CREATE TYPE http_status AS ENUM ('successful', 'timeout', 'error');

CREATE TABLE http_check_configs (
    check_id uuid PRIMARY KEY REFERENCES checks(id) ON DELETE CASCADE,
    method http_method NOT NULL DEFAULT 'GET',
    headers jsonb NOT NULL DEFAULT '[]'::jsonb,
    body text,
    timeout_ms integer NOT NULL DEFAULT 10000,
    ip_family ip_family,
    follow_redirects boolean NOT NULL DEFAULT true,
    skip_tls_verify boolean NOT NULL DEFAULT false,
    expected_status_codes integer[] NOT NULL DEFAULT '{}'::integer[],
    expected_status_classes integer[] NOT NULL DEFAULT ARRAY[2, 3],
    body_contains text,
    CONSTRAINT http_check_configs_headers_array CHECK (
        jsonb_typeof(headers) = 'array' AND jsonb_array_length(headers) <= 50
    ),
    CONSTRAINT http_check_configs_timeout_range CHECK (timeout_ms BETWEEN 1 AND 60000),
    CONSTRAINT http_check_configs_body_size CHECK (body IS NULL OR octet_length(body) <= 65536),
    CONSTRAINT http_check_configs_method_body CHECK (method NOT IN ('GET', 'HEAD') OR body IS NULL),
    CONSTRAINT http_check_configs_body_contains_size CHECK (body_contains IS NULL OR length(body_contains) BETWEEN 1 AND 1024),
    CONSTRAINT http_check_configs_expected_status_codes_range CHECK (
        100 <= ALL(expected_status_codes) AND 599 >= ALL(expected_status_codes)
    ),
    CONSTRAINT http_check_configs_expected_status_classes_range CHECK (
        1 <= ALL(expected_status_classes) AND 5 >= ALL(expected_status_classes)
    ),
    CONSTRAINT http_check_configs_expected_status_required CHECK (
        cardinality(expected_status_codes) > 0 OR cardinality(expected_status_classes) > 0
    )
);

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
SELECT add_retention_policy('http_results', INTERVAL '3 days', if_not_exists => TRUE);

CREATE INDEX idx_http_results_probe_check_started_at ON http_results (probe_id, check_id, started_at DESC);
CREATE INDEX idx_http_results_check_id_started_at ON http_results (check_id, started_at DESC);
CREATE INDEX idx_http_results_probe_id_started_at ON http_results (probe_id, started_at DESC);
CREATE INDEX idx_http_results_status_started_at ON http_results (status, started_at DESC);

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

-- +goose Down
SELECT remove_continuous_aggregate_policy('http_result_rollups_1m', if_exists => TRUE);
DROP MATERIALIZED VIEW IF EXISTS http_result_rollups_1m;
SELECT remove_retention_policy('http_results', if_exists => TRUE);
DROP TABLE IF EXISTS http_results;
DROP TABLE IF EXISTS http_check_configs;
DROP TYPE IF EXISTS http_status;
DROP TYPE IF EXISTS http_method;

DELETE FROM notification_outbox
WHERE incident_id IN (SELECT id FROM alert_incidents WHERE check_type = 'http')
   OR rule_id IN (SELECT id FROM alert_rules WHERE check_type = 'http');
DELETE FROM alert_incidents WHERE check_type = 'http';
DELETE FROM alert_rules WHERE check_type = 'http';
DELETE FROM probe_check_assignments
WHERE check_id IN (SELECT id FROM checks WHERE check_type = 'http');
DELETE FROM checks WHERE check_type = 'http';

DROP VIEW IF EXISTS public_status_page_assignment_scope;

ALTER TYPE check_type RENAME TO check_type_with_http;
CREATE TYPE check_type AS ENUM ('ping', 'traceroute', 'tcp');
ALTER TABLE checks ALTER COLUMN check_type TYPE check_type USING check_type::text::check_type;
ALTER TABLE alert_rules ALTER COLUMN check_type TYPE check_type USING check_type::text::check_type;
ALTER TABLE alert_incidents ALTER COLUMN check_type TYPE check_type USING check_type::text::check_type;
DROP TYPE check_type_with_http;

CREATE VIEW public_status_page_assignment_scope AS
SELECT public_status_page_elements.public_page_id,
       public_status_page_elements.id AS element_id,
       probe_check_assignments.id AS assignment_id,
       probe_check_assignments.project_id,
       probe_check_assignments.probe_id,
       probe_check_assignments.check_id
FROM public_status_page_elements
JOIN probe_check_assignments
  ON public_status_page_elements.assignment_selection_mode = 'all_check'
 AND probe_check_assignments.project_id = public_status_page_elements.project_id
 AND probe_check_assignments.check_id = public_status_page_elements.check_id
 AND probe_check_assignments.deleted_at IS NULL
JOIN checks
  ON checks.project_id = probe_check_assignments.project_id
 AND checks.id = probe_check_assignments.check_id
 AND checks.deleted_at IS NULL
 AND checks.check_type IN ('ping', 'tcp')
WHERE public_status_page_elements.kind = 'assignment_group'
UNION ALL
SELECT public_status_page_elements.public_page_id,
       public_status_page_elements.id AS element_id,
       probe_check_assignments.id AS assignment_id,
       probe_check_assignments.project_id,
       probe_check_assignments.probe_id,
       probe_check_assignments.check_id
FROM public_status_page_elements
JOIN public_status_page_element_assignments
  ON public_status_page_element_assignments.public_page_id = public_status_page_elements.public_page_id
 AND public_status_page_element_assignments.element_id = public_status_page_elements.id
JOIN probe_check_assignments
  ON probe_check_assignments.id = public_status_page_element_assignments.assignment_id
 AND probe_check_assignments.project_id = public_status_page_elements.project_id
 AND probe_check_assignments.deleted_at IS NULL
JOIN checks
  ON checks.project_id = probe_check_assignments.project_id
 AND checks.id = probe_check_assignments.check_id
 AND checks.deleted_at IS NULL
 AND checks.check_type IN ('ping', 'tcp')
WHERE public_status_page_elements.kind = 'assignment_group'
  AND public_status_page_elements.assignment_selection_mode = 'selected_assignments';

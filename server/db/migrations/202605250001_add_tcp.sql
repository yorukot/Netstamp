-- +goose NO TRANSACTION
-- +goose Up
ALTER TYPE check_type ADD VALUE IF NOT EXISTS 'tcp';

CREATE TYPE tcp_status AS ENUM ('successful', 'timeout', 'error');

CREATE TABLE tcp_check_configs (
    check_id uuid PRIMARY KEY REFERENCES checks(id) ON DELETE CASCADE,
    port integer NOT NULL DEFAULT 443,
    timeout_ms integer NOT NULL DEFAULT 3000,
    ip_family ip_family,
    CONSTRAINT tcp_check_configs_port_range CHECK (port >= 1 AND port <= 65535),
    CONSTRAINT tcp_check_configs_timeout_ms_positive CHECK (timeout_ms > 0)
);

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
    CONSTRAINT fk_tcp_results_probe
        FOREIGN KEY (probe_id) REFERENCES probes(internal_id),
    CONSTRAINT fk_tcp_results_check
        FOREIGN KEY (check_id) REFERENCES checks(internal_id)
);

SELECT create_hypertable('tcp_results', 'started_at', if_not_exists => TRUE);

CREATE INDEX idx_tcp_results_probe_check_started_at
    ON tcp_results (probe_id, check_id, started_at DESC);
CREATE INDEX idx_tcp_results_check_id_started_at ON tcp_results (check_id, started_at DESC);
CREATE INDEX idx_tcp_results_probe_id_started_at ON tcp_results (probe_id, started_at DESC);
CREATE INDEX idx_tcp_results_status_started_at ON tcp_results (status, started_at DESC);

-- +goose Down
DROP TABLE IF EXISTS tcp_results;
DROP TABLE IF EXISTS tcp_check_configs;

DROP TYPE IF EXISTS tcp_status;

ALTER TYPE check_type RENAME TO check_type_with_tcp;
CREATE TYPE check_type AS ENUM ('ping', 'traceroute');
ALTER TABLE checks
    ALTER COLUMN check_type TYPE check_type
    USING check_type::text::check_type;
DROP TYPE check_type_with_tcp;

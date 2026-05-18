-- +goose NO TRANSACTION
-- +goose Up
ALTER TYPE check_type ADD VALUE IF NOT EXISTS 'traceroute';

CREATE TYPE traceroute_protocol AS ENUM ('icmp', 'udp', 'tcp');
CREATE TYPE traceroute_status AS ENUM ('successful', 'timeout', 'error', 'partial');

CREATE TABLE traceroute_check_configs (
    check_id uuid PRIMARY KEY REFERENCES checks(id) ON DELETE CASCADE,
    protocol traceroute_protocol NOT NULL DEFAULT 'icmp',
    max_hops integer NOT NULL DEFAULT 30,
    timeout_ms integer NOT NULL DEFAULT 3000,
    queries_per_hop integer NOT NULL DEFAULT 3,
    packet_size_bytes integer NOT NULL DEFAULT 56,
    port integer NOT NULL DEFAULT 33434,
    ip_family ip_family,
    CONSTRAINT traceroute_check_configs_max_hops_range CHECK (max_hops >= 1 AND max_hops <= 64),
    CONSTRAINT traceroute_check_configs_timeout_ms_range CHECK (timeout_ms >= 1 AND timeout_ms <= 60000),
    CONSTRAINT traceroute_check_configs_queries_per_hop_range CHECK (queries_per_hop >= 1 AND queries_per_hop <= 10),
    CONSTRAINT traceroute_check_configs_packet_size_range CHECK (packet_size_bytes >= 0 AND packet_size_bytes <= 65507),
    CONSTRAINT traceroute_check_configs_port_range CHECK (port >= 1 AND port <= 65535)
);

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
    CONSTRAINT fk_traceroute_results_probe
        FOREIGN KEY (probe_id) REFERENCES probes(internal_id),
    CONSTRAINT fk_traceroute_results_check
        FOREIGN KEY (check_id) REFERENCES checks(internal_id)
);

SELECT create_hypertable('traceroute_results', 'started_at', if_not_exists => TRUE);

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
    PRIMARY KEY (probe_id, check_id, started_at, hop_index),
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
    CONSTRAINT traceroute_result_hops_error_message_not_empty CHECK (error_message IS NULL OR length(btrim(error_message)) > 0),
    CONSTRAINT fk_traceroute_result_hops_result
        FOREIGN KEY (probe_id, check_id, started_at)
        REFERENCES traceroute_results(probe_id, check_id, started_at)
        ON DELETE CASCADE
);

CREATE INDEX idx_traceroute_results_probe_check_started_at
    ON traceroute_results (probe_id, check_id, started_at DESC);
CREATE INDEX idx_traceroute_results_check_id_started_at
    ON traceroute_results (check_id, started_at DESC);
CREATE INDEX idx_traceroute_results_probe_id_started_at
    ON traceroute_results (probe_id, started_at DESC);
CREATE INDEX idx_traceroute_results_status_started_at
    ON traceroute_results (status, started_at DESC);
CREATE INDEX idx_traceroute_result_hops_result
    ON traceroute_result_hops (probe_id, check_id, started_at, hop_index);

-- +goose Down
DROP TABLE IF EXISTS traceroute_result_hops;
DROP TABLE IF EXISTS traceroute_results;
DROP TABLE IF EXISTS traceroute_check_configs;

DROP TYPE IF EXISTS traceroute_status;
DROP TYPE IF EXISTS traceroute_protocol;

ALTER TYPE check_type RENAME TO check_type_with_traceroute;
CREATE TYPE check_type AS ENUM ('ping');
ALTER TABLE checks
    ALTER COLUMN check_type TYPE check_type
    USING check_type::text::check_type;
DROP TYPE check_type_with_traceroute;

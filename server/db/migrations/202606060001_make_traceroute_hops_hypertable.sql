-- +goose NO TRANSACTION
-- +goose Up
DROP TABLE IF EXISTS traceroute_result_hops_retained;

CREATE TABLE traceroute_result_hops_retained (
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
    CONSTRAINT traceroute_result_hops_retained_pkey PRIMARY KEY (probe_id, check_id, started_at, hop_index),
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
    'traceroute_result_hops_retained',
    'started_at',
    chunk_time_interval => INTERVAL '1 day',
    create_default_indexes => FALSE,
    if_not_exists => TRUE
);

INSERT INTO traceroute_result_hops_retained (
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
SELECT
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
FROM traceroute_result_hops
WHERE started_at >= now() - INTERVAL '3 days'
ON CONFLICT (probe_id, check_id, started_at, hop_index) DO NOTHING;

DROP TABLE traceroute_result_hops;

ALTER TABLE traceroute_result_hops_retained RENAME TO traceroute_result_hops;
ALTER TABLE traceroute_result_hops RENAME CONSTRAINT traceroute_result_hops_retained_pkey TO traceroute_result_hops_pkey;

SELECT add_retention_policy('traceroute_result_hops', INTERVAL '3 days', if_not_exists => TRUE);

-- +goose Down
SELECT remove_retention_policy('traceroute_result_hops', if_exists => TRUE);

DROP TABLE IF EXISTS traceroute_result_hops_plain;

CREATE TABLE traceroute_result_hops_plain (
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
    CONSTRAINT traceroute_result_hops_plain_pkey PRIMARY KEY (probe_id, check_id, started_at, hop_index),
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

INSERT INTO traceroute_result_hops_plain (
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
SELECT
    hops.probe_id,
    hops.check_id,
    hops.started_at,
    hops.hop_index,
    hops.address,
    hops.hostname,
    hops.sent_count,
    hops.received_count,
    hops.loss_percent,
    hops.rtt_min_ms,
    hops.rtt_avg_ms,
    hops.rtt_median_ms,
    hops.rtt_max_ms,
    hops.rtt_stddev_ms,
    hops.rtt_samples_ms,
    hops.error_code,
    hops.error_message
FROM traceroute_result_hops AS hops
JOIN traceroute_results AS results
    ON results.probe_id = hops.probe_id
    AND results.check_id = hops.check_id
    AND results.started_at = hops.started_at
ON CONFLICT (probe_id, check_id, started_at, hop_index) DO NOTHING;

DROP TABLE traceroute_result_hops;

ALTER TABLE traceroute_result_hops_plain RENAME TO traceroute_result_hops;
ALTER TABLE traceroute_result_hops RENAME CONSTRAINT traceroute_result_hops_plain_pkey TO traceroute_result_hops_pkey;

CREATE INDEX idx_traceroute_result_hops_result
    ON traceroute_result_hops (probe_id, check_id, started_at, hop_index);

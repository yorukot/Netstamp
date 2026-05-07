-- +goose NO TRANSACTION
-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS timescaledb;

CREATE TYPE project_member_role AS ENUM ('owner', 'admin', 'editor', 'viewer');
CREATE TYPE check_type AS ENUM ('ping', 'traceroute');
CREATE TYPE target_type AS ENUM ('host', 'ip');
CREATE TYPE ip_family AS ENUM ('ipv4', 'ipv6');
CREATE TYPE probe_state AS ENUM ('online', 'offline', 'degraded');
CREATE TYPE check_result_status AS ENUM ('success', 'partial', 'timeout', 'error');
CREATE TYPE traceroute_protocol AS ENUM ('icmp', 'udp', 'tcp');

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email citext NOT NULL,
    password_hash text NOT NULL,
    is_active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT users_email_not_empty CHECK (length(btrim(email::text)) > 0),
    CONSTRAINT users_password_hash_not_empty CHECK (length(btrim(password_hash)) > 0)
);

CREATE UNIQUE INDEX uq_users_email ON users (email);

CREATE TRIGGER set_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE projects (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    slug citext NOT NULL,
    created_by_user_id uuid NOT NULL REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT projects_name_not_empty CHECK (length(btrim(name)) > 0),
    CONSTRAINT projects_slug_not_empty CHECK (length(btrim(slug::text)) > 0),
    CONSTRAINT projects_deleted_at_after_created_at CHECK (deleted_at IS NULL OR deleted_at >= created_at)
);

CREATE UNIQUE INDEX uq_projects_slug ON projects (slug);

CREATE TRIGGER set_projects_updated_at
    BEFORE UPDATE ON projects
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE project_members (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id uuid NOT NULL REFERENCES projects(id),
    user_id uuid NOT NULL REFERENCES users(id),
    role project_member_role NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT project_members_deleted_at_after_created_at CHECK (deleted_at IS NULL OR deleted_at >= created_at)
);

CREATE UNIQUE INDEX uq_project_members_active_project_user
    ON project_members (project_id, user_id)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_project_members_project_id ON project_members (project_id);
CREATE INDEX idx_project_members_user_id ON project_members (user_id);

CREATE TRIGGER set_project_members_updated_at
    BEFORE UPDATE ON project_members
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE probes (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id uuid NOT NULL REFERENCES projects(id),
    name text NOT NULL,
    hostname text,
    latitude double precision,
    longitude double precision,
    enabled boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT uq_probes_project_id_id UNIQUE (project_id, id),
    CONSTRAINT probes_name_not_empty CHECK (length(btrim(name)) > 0),
    CONSTRAINT probes_hostname_not_empty CHECK (hostname IS NULL OR length(btrim(hostname)) > 0),
    CONSTRAINT probes_latitude_range CHECK (latitude IS NULL OR (latitude >= -90 AND latitude <= 90)),
    CONSTRAINT probes_longitude_range CHECK (longitude IS NULL OR (longitude >= -180 AND longitude <= 180)),
    CONSTRAINT probes_location_all_or_none CHECK ((latitude IS NULL) = (longitude IS NULL)),
    CONSTRAINT probes_deleted_at_after_created_at CHECK (deleted_at IS NULL OR deleted_at >= created_at)
);

CREATE INDEX idx_probes_project_id ON probes (project_id);

CREATE TRIGGER set_probes_updated_at
    BEFORE UPDATE ON probes
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE probe_credentials (
    probe_id uuid PRIMARY KEY REFERENCES probes(id) ON DELETE CASCADE,
    secret_hash text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    last_rotated_at timestamptz,
    CONSTRAINT probe_credentials_secret_hash_not_empty CHECK (length(btrim(secret_hash)) > 0),
    CONSTRAINT probe_credentials_last_rotated_at_after_created_at CHECK (
        last_rotated_at IS NULL OR last_rotated_at >= created_at
    )
);

CREATE TABLE probe_statuses (
    probe_id uuid PRIMARY KEY REFERENCES probes(id) ON DELETE CASCADE,
    status probe_state NOT NULL,
    last_seen_at timestamptz,
    agent_version text,
    ip_families ip_family[] NOT NULL DEFAULT '{}'::ip_family[],
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT probe_statuses_agent_version_not_empty CHECK (
        agent_version IS NULL OR length(btrim(agent_version)) > 0
    )
);

CREATE TABLE checks (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id uuid NOT NULL REFERENCES projects(id),
    name text NOT NULL,
    check_type check_type NOT NULL,
    target text NOT NULL,
    target_type target_type NOT NULL,
    description text,
    enabled boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT uq_checks_project_id_id UNIQUE (project_id, id),
    CONSTRAINT uq_checks_id_check_type UNIQUE (id, check_type),
    CONSTRAINT checks_name_not_empty CHECK (length(btrim(name)) > 0),
    CONSTRAINT checks_target_not_empty CHECK (length(btrim(target)) > 0),
    CONSTRAINT checks_description_not_empty CHECK (description IS NULL OR length(btrim(description)) > 0),
    CONSTRAINT checks_deleted_at_after_created_at CHECK (deleted_at IS NULL OR deleted_at >= created_at)
);

CREATE INDEX idx_checks_project_id ON checks (project_id);
CREATE INDEX idx_checks_project_id_check_type ON checks (project_id, check_type);

CREATE TRIGGER set_checks_updated_at
    BEFORE UPDATE ON checks
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE ping_check_configs (
    check_id uuid PRIMARY KEY,
    check_type check_type NOT NULL DEFAULT 'ping',
    packet_count integer NOT NULL DEFAULT 4,
    packet_size_bytes integer NOT NULL DEFAULT 56,
    timeout_ms integer NOT NULL DEFAULT 3000,
    ip_family ip_family,
    CONSTRAINT ping_check_configs_check_type_is_ping CHECK (check_type = 'ping'),
    CONSTRAINT ping_check_configs_packet_count_positive CHECK (packet_count > 0),
    CONSTRAINT ping_check_configs_packet_size_range CHECK (packet_size_bytes >= 0 AND packet_size_bytes <= 65507),
    CONSTRAINT ping_check_configs_timeout_ms_positive CHECK (timeout_ms > 0),
    CONSTRAINT fk_ping_check_configs_check_type
        FOREIGN KEY (check_id, check_type) REFERENCES checks(id, check_type) ON DELETE CASCADE
);

CREATE TABLE traceroute_check_configs (
    check_id uuid PRIMARY KEY,
    check_type check_type NOT NULL DEFAULT 'traceroute',
    protocol traceroute_protocol NOT NULL DEFAULT 'icmp',
    max_hops integer NOT NULL DEFAULT 30,
    timeout_ms integer NOT NULL DEFAULT 3000,
    queries_per_hop integer NOT NULL DEFAULT 3,
    port integer,
    CONSTRAINT traceroute_check_configs_check_type_is_traceroute CHECK (check_type = 'traceroute'),
    CONSTRAINT traceroute_check_configs_max_hops_range CHECK (max_hops > 0 AND max_hops <= 255),
    CONSTRAINT traceroute_check_configs_timeout_ms_positive CHECK (timeout_ms > 0),
    CONSTRAINT traceroute_check_configs_queries_per_hop_positive CHECK (queries_per_hop > 0),
    CONSTRAINT traceroute_check_configs_port_range CHECK (port IS NULL OR (port > 0 AND port <= 65535)),
    CONSTRAINT traceroute_check_configs_port_matches_protocol CHECK (
        (protocol = 'icmp' AND port IS NULL) OR
        (protocol IN ('udp', 'tcp') AND port IS NOT NULL)
    ),
    CONSTRAINT fk_traceroute_check_configs_check_type
        FOREIGN KEY (check_id, check_type) REFERENCES checks(id, check_type) ON DELETE CASCADE
);

CREATE TABLE probe_checks (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id uuid NOT NULL REFERENCES projects(id),
    probe_id uuid NOT NULL,
    check_id uuid NOT NULL,
    name text,
    enabled boolean NOT NULL DEFAULT true,
    interval_seconds integer NOT NULL,
    jitter_seconds integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT uq_probe_checks_project_id_id UNIQUE (project_id, id),
    CONSTRAINT probe_checks_name_not_empty CHECK (name IS NULL OR length(btrim(name)) > 0),
    CONSTRAINT probe_checks_interval_seconds_positive CHECK (interval_seconds > 0),
    CONSTRAINT probe_checks_jitter_seconds_range CHECK (jitter_seconds >= 0 AND jitter_seconds < interval_seconds),
    CONSTRAINT probe_checks_deleted_at_after_created_at CHECK (deleted_at IS NULL OR deleted_at >= created_at),
    CONSTRAINT fk_probe_checks_project_probe
        FOREIGN KEY (project_id, probe_id) REFERENCES probes(project_id, id),
    CONSTRAINT fk_probe_checks_project_check
        FOREIGN KEY (project_id, check_id) REFERENCES checks(project_id, id)
);

CREATE UNIQUE INDEX uq_probe_checks_active_project_probe_check
    ON probe_checks (project_id, probe_id, check_id)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_probe_checks_project_id ON probe_checks (project_id);
CREATE INDEX idx_probe_checks_probe_id ON probe_checks (probe_id);
CREATE INDEX idx_probe_checks_check_id ON probe_checks (check_id);

CREATE TRIGGER set_probe_checks_updated_at
    BEFORE UPDATE ON probe_checks
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE labels (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id uuid NOT NULL REFERENCES projects(id),
    name citext NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT uq_labels_project_id_id UNIQUE (project_id, id),
    CONSTRAINT labels_name_not_empty CHECK (length(btrim(name::text)) > 0),
    CONSTRAINT labels_deleted_at_after_created_at CHECK (deleted_at IS NULL OR deleted_at >= created_at)
);

CREATE UNIQUE INDEX uq_labels_active_project_name
    ON labels (project_id, name)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_labels_project_id ON labels (project_id);

CREATE TRIGGER set_labels_updated_at
    BEFORE UPDATE ON labels
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE check_labels (
    project_id uuid NOT NULL,
    check_id uuid NOT NULL,
    label_id uuid NOT NULL,
    PRIMARY KEY (project_id, check_id, label_id),
    CONSTRAINT fk_check_labels_project_check
        FOREIGN KEY (project_id, check_id) REFERENCES checks(project_id, id) ON DELETE CASCADE,
    CONSTRAINT fk_check_labels_project_label
        FOREIGN KEY (project_id, label_id) REFERENCES labels(project_id, id) ON DELETE CASCADE
);

CREATE INDEX idx_check_labels_label_id ON check_labels (label_id);

CREATE TABLE probe_labels (
    project_id uuid NOT NULL,
    probe_id uuid NOT NULL,
    label_id uuid NOT NULL,
    PRIMARY KEY (project_id, probe_id, label_id),
    CONSTRAINT fk_probe_labels_project_probe
        FOREIGN KEY (project_id, probe_id) REFERENCES probes(project_id, id) ON DELETE CASCADE,
    CONSTRAINT fk_probe_labels_project_label
        FOREIGN KEY (project_id, label_id) REFERENCES labels(project_id, id) ON DELETE CASCADE
);

CREATE INDEX idx_probe_labels_label_id ON probe_labels (label_id);

CREATE TABLE ping_results (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    project_id uuid NOT NULL,
    probe_check_id uuid NOT NULL,
    started_at timestamptz NOT NULL,
    finished_at timestamptz NOT NULL,
    duration_ms integer NOT NULL,
    status check_result_status NOT NULL,
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
    raw jsonb NOT NULL DEFAULT '{}'::jsonb,
    error_code text,
    error_message text,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id, started_at),
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
    CONSTRAINT fk_ping_results_project_probe_check
        FOREIGN KEY (project_id, probe_check_id) REFERENCES probe_checks(project_id, id)
);

SELECT create_hypertable('ping_results', 'started_at', if_not_exists => TRUE);

CREATE INDEX idx_ping_results_project_id_started_at ON ping_results (project_id, started_at DESC);
CREATE INDEX idx_ping_results_probe_check_id_started_at ON ping_results (probe_check_id, started_at DESC);
CREATE INDEX idx_ping_results_status_started_at ON ping_results (status, started_at DESC);

CREATE TABLE traceroute_results (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    project_id uuid NOT NULL,
    probe_check_id uuid NOT NULL,
    started_at timestamptz NOT NULL,
    finished_at timestamptz NOT NULL,
    duration_ms integer NOT NULL,
    status check_result_status NOT NULL,
    resolved_ip inet,
    reached boolean NOT NULL DEFAULT false,
    hop_count integer NOT NULL DEFAULT 0,
    path_hash text,
    protocol traceroute_protocol NOT NULL,
    raw jsonb NOT NULL DEFAULT '{}'::jsonb,
    error_code text,
    error_message text,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id, started_at),
    CONSTRAINT traceroute_results_finished_at_after_started_at CHECK (finished_at >= started_at),
    CONSTRAINT traceroute_results_duration_ms_non_negative CHECK (duration_ms >= 0),
    CONSTRAINT traceroute_results_hop_count_non_negative CHECK (hop_count >= 0),
    CONSTRAINT traceroute_results_path_hash_not_empty CHECK (path_hash IS NULL OR length(btrim(path_hash)) > 0),
    CONSTRAINT traceroute_results_error_code_not_empty CHECK (error_code IS NULL OR length(btrim(error_code)) > 0),
    CONSTRAINT traceroute_results_error_message_not_empty CHECK (
        error_message IS NULL OR length(btrim(error_message)) > 0
    ),
    CONSTRAINT fk_traceroute_results_project_probe_check
        FOREIGN KEY (project_id, probe_check_id) REFERENCES probe_checks(project_id, id)
);

SELECT create_hypertable('traceroute_results', 'started_at', if_not_exists => TRUE);

CREATE INDEX idx_traceroute_results_project_id_started_at ON traceroute_results (project_id, started_at DESC);
CREATE INDEX idx_traceroute_results_probe_check_id_started_at ON traceroute_results (probe_check_id, started_at DESC);
CREATE INDEX idx_traceroute_results_status_started_at ON traceroute_results (status, started_at DESC);
CREATE INDEX idx_traceroute_results_project_path_hash_started_at
    ON traceroute_results (project_id, path_hash, started_at DESC);

CREATE TABLE traceroute_hops (
    traceroute_result_id uuid NOT NULL,
    project_id uuid NOT NULL,
    probe_check_id uuid NOT NULL,
    started_at timestamptz NOT NULL,
    hop_number integer NOT NULL,
    hop_ip inet,
    hostname text,
    rtts_ms double precision[] NOT NULL DEFAULT '{}'::double precision[],
    loss_percent double precision NOT NULL,
    error_code text,
    error_message text,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (traceroute_result_id, started_at, hop_number),
    CONSTRAINT traceroute_hops_hop_number_positive CHECK (hop_number > 0),
    CONSTRAINT traceroute_hops_loss_percent_range CHECK (loss_percent >= 0 AND loss_percent <= 100),
    CONSTRAINT traceroute_hops_hostname_not_empty CHECK (hostname IS NULL OR length(btrim(hostname)) > 0),
    CONSTRAINT traceroute_hops_error_code_not_empty CHECK (error_code IS NULL OR length(btrim(error_code)) > 0),
    CONSTRAINT traceroute_hops_error_message_not_empty CHECK (
        error_message IS NULL OR length(btrim(error_message)) > 0
    ),
    CONSTRAINT fk_traceroute_hops_result
        FOREIGN KEY (traceroute_result_id, started_at) REFERENCES traceroute_results(id, started_at) ON DELETE CASCADE,
    CONSTRAINT fk_traceroute_hops_project_probe_check
        FOREIGN KEY (project_id, probe_check_id) REFERENCES probe_checks(project_id, id)
);

CREATE INDEX idx_traceroute_hops_project_id_started_at ON traceroute_hops (project_id, started_at DESC);
CREATE INDEX idx_traceroute_hops_probe_check_id_started_at ON traceroute_hops (probe_check_id, started_at DESC);
CREATE INDEX idx_traceroute_hops_result_id_started_at
    ON traceroute_hops (traceroute_result_id, started_at DESC);
CREATE INDEX idx_traceroute_hops_project_hop_ip_started_at ON traceroute_hops (project_id, hop_ip, started_at DESC);
CREATE INDEX idx_traceroute_hops_project_probe_check_hop_started_at
    ON traceroute_hops (project_id, probe_check_id, hop_number, started_at DESC);

-- +goose Down
DROP TABLE IF EXISTS traceroute_hops;
DROP TABLE IF EXISTS traceroute_results;
DROP TABLE IF EXISTS ping_results;
DROP TABLE IF EXISTS probe_labels;
DROP TABLE IF EXISTS check_labels;
DROP TABLE IF EXISTS labels;
DROP TABLE IF EXISTS probe_checks;
DROP TABLE IF EXISTS traceroute_check_configs;
DROP TABLE IF EXISTS ping_check_configs;
DROP TABLE IF EXISTS checks;
DROP TABLE IF EXISTS probe_statuses;
DROP TABLE IF EXISTS probe_credentials;
DROP TABLE IF EXISTS probes;
DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS users;

DROP FUNCTION IF EXISTS set_updated_at();

DROP TYPE IF EXISTS traceroute_protocol;
DROP TYPE IF EXISTS check_result_status;
DROP TYPE IF EXISTS probe_state;
DROP TYPE IF EXISTS ip_family;
DROP TYPE IF EXISTS target_type;
DROP TYPE IF EXISTS check_type;
DROP TYPE IF EXISTS project_member_role;

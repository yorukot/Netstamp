-- +goose Up
CREATE TYPE assignment_refresh_target_type AS ENUM ('project', 'probe', 'check', 'label');
CREATE TYPE assignment_refresh_job_status AS ENUM ('pending', 'running', 'succeeded', 'failed', 'discarded');

CREATE TABLE assignment_refresh_jobs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id uuid NOT NULL REFERENCES projects(id),
    target_type assignment_refresh_target_type NOT NULL,
    target_id uuid NOT NULL,
    status assignment_refresh_job_status NOT NULL DEFAULT 'pending',
    attempt_count integer NOT NULL DEFAULT 0,
    max_attempts integer NOT NULL DEFAULT 5,
    next_attempt_at timestamptz NOT NULL DEFAULT now(),
    last_attempt_at timestamptz,
    completed_at timestamptz,
    last_error_kind text,
    last_error_code text,
    last_error text,
    dedupe_key text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uq_assignment_refresh_jobs_project_id_id UNIQUE (project_id, id),
    CONSTRAINT assignment_refresh_jobs_attempt_count_non_negative CHECK (attempt_count >= 0),
    CONSTRAINT assignment_refresh_jobs_max_attempts_positive CHECK (max_attempts > 0),
    CONSTRAINT assignment_refresh_jobs_last_error_kind_not_empty CHECK (last_error_kind IS NULL OR length(btrim(last_error_kind)) > 0),
    CONSTRAINT assignment_refresh_jobs_last_error_code_not_empty CHECK (last_error_code IS NULL OR length(btrim(last_error_code)) > 0),
    CONSTRAINT assignment_refresh_jobs_last_error_not_empty CHECK (last_error IS NULL OR length(btrim(last_error)) > 0),
    CONSTRAINT assignment_refresh_jobs_dedupe_key_not_empty CHECK (length(btrim(dedupe_key)) > 0)
);

CREATE INDEX idx_assignment_refresh_jobs_status_next_attempt
    ON assignment_refresh_jobs (status, next_attempt_at);
CREATE INDEX idx_assignment_refresh_jobs_status_last_attempt
    ON assignment_refresh_jobs (status, last_attempt_at);
CREATE INDEX idx_assignment_refresh_jobs_project_created
    ON assignment_refresh_jobs (project_id, created_at DESC);
CREATE UNIQUE INDEX uq_assignment_refresh_jobs_dedupe_key
    ON assignment_refresh_jobs (dedupe_key);

CREATE TRIGGER set_assignment_refresh_jobs_updated_at
    BEFORE UPDATE ON assignment_refresh_jobs
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TABLE IF EXISTS assignment_refresh_jobs;

DROP TYPE IF EXISTS assignment_refresh_job_status;
DROP TYPE IF EXISTS assignment_refresh_target_type;

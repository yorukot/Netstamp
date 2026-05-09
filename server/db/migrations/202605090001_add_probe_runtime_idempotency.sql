-- +goose Up
ALTER TABLE ping_results
    ADD COLUMN external_id text;

ALTER TABLE ping_results
    ADD CONSTRAINT ping_results_external_id_not_empty CHECK (
        external_id IS NULL OR length(btrim(external_id)) > 0
    );

CREATE TABLE ping_result_external_ids (
    project_id uuid NOT NULL,
    probe_id uuid NOT NULL,
    external_id text NOT NULL,
    result_id uuid NOT NULL DEFAULT gen_random_uuid(),
    started_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (project_id, probe_id, external_id),
    CONSTRAINT ping_result_external_ids_external_id_not_empty CHECK (length(btrim(external_id)) > 0),
    CONSTRAINT fk_ping_result_external_ids_project_probe
        FOREIGN KEY (project_id, probe_id) REFERENCES probes(project_id, id)
);

-- +goose Down
DROP TABLE IF EXISTS ping_result_external_ids;

ALTER TABLE ping_results
    DROP CONSTRAINT IF EXISTS ping_results_external_id_not_empty;

ALTER TABLE ping_results
    DROP COLUMN IF EXISTS external_id;

-- +goose Up
ALTER TABLE probes
    ADD COLUMN location_name text;

ALTER TABLE probes
    DROP CONSTRAINT probes_subdivision_code_valid,
    DROP COLUMN subdivision_code;

ALTER TABLE probes
    ADD CONSTRAINT probes_location_name_valid CHECK (location_name IS NULL OR length(btrim(location_name)) > 0);

-- +goose Down
ALTER TABLE probes
    ADD COLUMN subdivision_code text;

ALTER TABLE probes
    DROP CONSTRAINT probes_location_name_valid,
    DROP COLUMN location_name;

ALTER TABLE probes
    ADD CONSTRAINT probes_subdivision_code_valid CHECK (subdivision_code IS NULL OR length(btrim(subdivision_code)) > 0);

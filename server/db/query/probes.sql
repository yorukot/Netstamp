-- name: CreateProbe :one
INSERT INTO probes (project_id, name, enabled, location, city)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, project_id, name, enabled, location, city, created_at, updated_at, deleted_at;

-- name: CreateProbeCredential :one
INSERT INTO probe_credentials (probe_id, secret_hash)
VALUES ($1, $2)
RETURNING probe_id, secret_hash, created_at, last_rotated_at;

-- name: CreateProbeStatus :one
INSERT INTO probe_statuses (probe_id, status)
VALUES ($1, $2)
RETURNING probe_id, status, last_seen_at, agent_version, public_v4, public_v6, addrs, updated_at;

-- name: CreateProbeLabel :exec
INSERT INTO probe_labels (project_id, probe_id, label_id)
VALUES ($1, $2, $3);

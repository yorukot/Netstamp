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

-- name: GetActiveLabelsByIDsForProject :many
WITH requested AS (
    SELECT label_id, position
    FROM unnest(sqlc.arg(label_ids)::uuid[]) WITH ORDINALITY AS requested(label_id, position)
)
SELECT labels.id,
       labels.project_id,
       labels.key,
       labels.value,
       labels.created_at,
       labels.updated_at,
       labels.deleted_at
FROM requested
JOIN labels
    ON labels.project_id = sqlc.arg(project_id)
    AND labels.id = requested.label_id
WHERE labels.deleted_at IS NULL
ORDER BY requested.position;

-- name: CreateProbeLabel :exec
INSERT INTO probe_labels (project_id, probe_id, label_id)
VALUES ($1, $2, $3);

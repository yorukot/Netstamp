-- name: ListActiveLabelsForProject :many
SELECT id, project_id, key, value, created_at, updated_at, deleted_at
FROM labels
WHERE project_id = $1
  AND deleted_at IS NULL
ORDER BY key ASC, value ASC, id ASC;

-- name: GetActiveLabelForProject :one
SELECT id, project_id, key, value, created_at, updated_at, deleted_at
FROM labels
WHERE project_id = $1
  AND id = $2
  AND deleted_at IS NULL;

-- name: CreateLabel :one
INSERT INTO labels (project_id, key, value)
VALUES ($1, $2, $3)
RETURNING id, project_id, key, value, created_at, updated_at, deleted_at;

-- name: UpdateLabel :one
UPDATE labels
SET key = $3,
    value = $4
WHERE project_id = $1
  AND id = $2
  AND deleted_at IS NULL
RETURNING id, project_id, key, value, created_at, updated_at, deleted_at;

-- name: SoftDeleteLabel :one
UPDATE labels
SET deleted_at = now()
WHERE project_id = $1
  AND id = $2
  AND deleted_at IS NULL
RETURNING id;

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

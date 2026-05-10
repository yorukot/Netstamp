-- name: CreateProject :one
INSERT INTO projects (name, slug, created_by_user_id)
VALUES ($1, $2, $3)
RETURNING id, name, slug, created_by_user_id, created_at, updated_at, deleted_at;

-- name: CreateProjectMember :one
WITH inserted AS (
    INSERT INTO project_members (project_id, user_id, role)
    VALUES ($1, $2, $3)
    RETURNING id, project_id, user_id, role, created_at, updated_at
)
SELECT inserted.id,
       inserted.project_id,
       inserted.user_id,
       users.email,
       inserted.role,
       inserted.created_at,
       inserted.updated_at
FROM inserted
JOIN users ON users.id = inserted.user_id;

-- name: ListProjectsForUser :many
SELECT projects.id, projects.name, projects.slug, projects.created_by_user_id, projects.created_at, projects.updated_at, projects.deleted_at
FROM projects
JOIN project_members
    ON project_members.project_id = projects.id
    AND project_members.user_id = $1
WHERE projects.deleted_at IS NULL
ORDER BY projects.created_at DESC, projects.id DESC;

-- name: GetProjectForUser :one
SELECT projects.id, projects.name, projects.slug, projects.created_by_user_id, projects.created_at, projects.updated_at, projects.deleted_at
FROM projects
JOIN project_members
    ON project_members.project_id = projects.id
    AND project_members.user_id = $2
WHERE projects.id = $1
  AND projects.deleted_at IS NULL;

-- name: GetProjectBySlugForUser :one
SELECT projects.id, projects.name, projects.slug, projects.created_by_user_id, projects.created_at, projects.updated_at, projects.deleted_at
FROM projects
JOIN project_members
    ON project_members.project_id = projects.id
    AND project_members.user_id = $2
WHERE projects.slug = $1
  AND projects.deleted_at IS NULL;

-- name: GetActiveProjectMemberRole :one
SELECT project_members.role
FROM project_members
JOIN projects ON projects.id = project_members.project_id
WHERE project_members.project_id = $1
  AND project_members.user_id = $2
  AND projects.deleted_at IS NULL;

-- name: UpdateProject :one
UPDATE projects
SET name = $2,
    slug = $3
WHERE id = $1
  AND deleted_at IS NULL
RETURNING id, name, slug, created_by_user_id, created_at, updated_at, deleted_at;

-- name: SoftDeleteProject :one
UPDATE projects
SET deleted_at = now()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING id;

-- name: ListActiveProjectMembers :many
SELECT project_members.id,
       project_members.project_id,
       project_members.user_id,
       users.email,
       project_members.role,
       project_members.created_at,
       project_members.updated_at
FROM project_members
JOIN users ON users.id = project_members.user_id
JOIN projects ON projects.id = project_members.project_id
WHERE project_members.project_id = $1
  AND projects.deleted_at IS NULL
ORDER BY project_members.created_at ASC, project_members.id ASC;

-- name: GetActiveProjectMember :one
SELECT project_members.id,
       project_members.project_id,
       project_members.user_id,
       users.email,
       project_members.role,
       project_members.created_at,
       project_members.updated_at
FROM project_members
JOIN users ON users.id = project_members.user_id
JOIN projects ON projects.id = project_members.project_id
WHERE project_members.project_id = $1
  AND project_members.user_id = $2
  AND projects.deleted_at IS NULL;

-- name: UpdateProjectMemberRole :one
WITH updated AS (
    UPDATE project_members
    SET role = $3
    WHERE project_id = $1
      AND user_id = $2
    RETURNING id, project_id, user_id, role, created_at, updated_at
)
SELECT updated.id,
       updated.project_id,
       updated.user_id,
       users.email,
       updated.role,
       updated.created_at,
       updated.updated_at
FROM updated
JOIN users ON users.id = updated.user_id;

-- name: DeleteProjectMember :one
DELETE FROM project_members
WHERE project_id = $1
  AND user_id = $2
RETURNING id;

-- name: CountActiveProjectOwners :one
SELECT count(*)::int4
FROM project_members
WHERE project_id = $1
  AND role = 'owner';

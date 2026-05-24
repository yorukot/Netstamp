-- name: CreateProjectInvite :one
WITH inserted AS (
    INSERT INTO project_invites (project_id, invited_user_id, invited_by_user_id, role)
    SELECT $1, $2, $3, $4
    WHERE NOT EXISTS (
        SELECT 1
        FROM project_members
        WHERE project_members.project_id = $1
          AND project_members.user_id = $2
    )
    RETURNING id, project_id, invited_user_id, invited_by_user_id, role, status, created_at, updated_at, resolved_at
)
SELECT inserted.id,
       inserted.project_id,
       inserted.invited_user_id,
       inserted.invited_by_user_id,
       inserted.role,
       inserted.status,
       inserted.created_at,
       inserted.updated_at,
       inserted.resolved_at,
       projects.name AS project_name,
       projects.slug AS project_slug,
       invited_user.email AS invited_user_email,
       invited_user.display_name AS invited_user_display_name,
       inviter.email AS invited_by_user_email,
       inviter.display_name AS invited_by_user_display_name
FROM inserted
JOIN projects ON projects.id = inserted.project_id
JOIN users AS invited_user ON invited_user.id = inserted.invited_user_id
JOIN users AS inviter ON inviter.id = inserted.invited_by_user_id;

-- name: ListPendingProjectInvites :many
SELECT project_invites.id,
       project_invites.project_id,
       project_invites.invited_user_id,
       project_invites.invited_by_user_id,
       project_invites.role,
       project_invites.status,
       project_invites.created_at,
       project_invites.updated_at,
       project_invites.resolved_at,
       projects.name AS project_name,
       projects.slug AS project_slug,
       invited_user.email AS invited_user_email,
       invited_user.display_name AS invited_user_display_name,
       inviter.email AS invited_by_user_email,
       inviter.display_name AS invited_by_user_display_name
FROM project_invites
JOIN projects ON projects.id = project_invites.project_id
JOIN users AS invited_user ON invited_user.id = project_invites.invited_user_id
JOIN users AS inviter ON inviter.id = project_invites.invited_by_user_id
WHERE project_invites.project_id = $1
  AND project_invites.status = 'pending'
  AND projects.deleted_at IS NULL
ORDER BY project_invites.created_at ASC, project_invites.id ASC;

-- name: ListPendingProjectInvitesForUser :many
SELECT project_invites.id,
       project_invites.project_id,
       project_invites.invited_user_id,
       project_invites.invited_by_user_id,
       project_invites.role,
       project_invites.status,
       project_invites.created_at,
       project_invites.updated_at,
       project_invites.resolved_at,
       projects.name AS project_name,
       projects.slug AS project_slug,
       invited_user.email AS invited_user_email,
       invited_user.display_name AS invited_user_display_name,
       inviter.email AS invited_by_user_email,
       inviter.display_name AS invited_by_user_display_name
FROM project_invites
JOIN projects ON projects.id = project_invites.project_id
JOIN users AS invited_user ON invited_user.id = project_invites.invited_user_id
JOIN users AS inviter ON inviter.id = project_invites.invited_by_user_id
WHERE project_invites.invited_user_id = $1
  AND project_invites.status = 'pending'
  AND projects.deleted_at IS NULL
ORDER BY project_invites.created_at DESC, project_invites.id DESC;

-- name: AcceptPendingProjectInvite :one
WITH updated AS (
    UPDATE project_invites
    SET status = 'accepted',
        resolved_at = now()
    WHERE project_invites.id = $1
      AND project_invites.invited_user_id = $2
      AND project_invites.status = 'pending'
      AND EXISTS (
          SELECT 1
          FROM projects
          WHERE projects.id = project_invites.project_id
            AND projects.deleted_at IS NULL
      )
    RETURNING id, project_id, invited_user_id, invited_by_user_id, role, status, created_at, updated_at, resolved_at
)
SELECT updated.id,
       updated.project_id,
       updated.invited_user_id,
       updated.invited_by_user_id,
       updated.role,
       updated.status,
       updated.created_at,
       updated.updated_at,
       updated.resolved_at,
       projects.name AS project_name,
       projects.slug AS project_slug,
       invited_user.email AS invited_user_email,
       invited_user.display_name AS invited_user_display_name,
       inviter.email AS invited_by_user_email,
       inviter.display_name AS invited_by_user_display_name
FROM updated
JOIN projects ON projects.id = updated.project_id
JOIN users AS invited_user ON invited_user.id = updated.invited_user_id
JOIN users AS inviter ON inviter.id = updated.invited_by_user_id;

-- name: RejectPendingProjectInvite :one
WITH updated AS (
    UPDATE project_invites
    SET status = 'rejected',
        resolved_at = now()
    WHERE project_invites.id = $1
      AND project_invites.invited_user_id = $2
      AND project_invites.status = 'pending'
      AND EXISTS (
          SELECT 1
          FROM projects
          WHERE projects.id = project_invites.project_id
            AND projects.deleted_at IS NULL
      )
    RETURNING id, project_id, invited_user_id, invited_by_user_id, role, status, created_at, updated_at, resolved_at
)
SELECT updated.id,
       updated.project_id,
       updated.invited_user_id,
       updated.invited_by_user_id,
       updated.role,
       updated.status,
       updated.created_at,
       updated.updated_at,
       updated.resolved_at,
       projects.name AS project_name,
       projects.slug AS project_slug,
       invited_user.email AS invited_user_email,
       invited_user.display_name AS invited_user_display_name,
       inviter.email AS invited_by_user_email,
       inviter.display_name AS invited_by_user_display_name
FROM updated
JOIN projects ON projects.id = updated.project_id
JOIN users AS invited_user ON invited_user.id = updated.invited_user_id
JOIN users AS inviter ON inviter.id = updated.invited_by_user_id;

-- name: ListPublicPagesForProject :many
SELECT id, project_id, slug, title, description, enabled, created_at, updated_at, deleted_at
FROM public_pages
WHERE project_id = $1
  AND deleted_at IS NULL
ORDER BY created_at DESC, id DESC;

-- name: GetActivePublicPageForProject :one
SELECT id, project_id, slug, title, description, enabled, created_at, updated_at, deleted_at
FROM public_pages
WHERE project_id = $1
  AND id = $2
  AND deleted_at IS NULL;

-- name: GetEnabledPublicPageBySlug :one
SELECT id, project_id, slug, title, description, enabled, created_at, updated_at, deleted_at
FROM public_pages
WHERE slug = $1
  AND enabled = true
  AND deleted_at IS NULL;

-- name: CreatePublicPage :one
INSERT INTO public_pages (project_id, slug, title, description, enabled)
VALUES (sqlc.arg(project_id), sqlc.arg(slug), sqlc.arg(title), sqlc.narg(description), sqlc.arg(enabled))
RETURNING id, project_id, slug, title, description, enabled, created_at, updated_at, deleted_at;

-- name: UpdatePublicPage :one
UPDATE public_pages
SET slug = coalesce(sqlc.narg(slug), slug),
    title = coalesce(sqlc.narg(title), title),
    description = CASE
        WHEN sqlc.arg(description_set)::boolean THEN sqlc.narg(description)
        ELSE description
    END,
    enabled = coalesce(sqlc.narg(enabled), enabled)
WHERE project_id = sqlc.arg(project_id)
  AND id = sqlc.arg(id)
  AND deleted_at IS NULL
RETURNING id, project_id, slug, title, description, enabled, created_at, updated_at, deleted_at;

-- name: SoftDeletePublicPage :one
UPDATE public_pages
SET deleted_at = now()
WHERE project_id = $1
  AND id = $2
  AND deleted_at IS NULL
RETURNING id;

-- name: ListPublicPageFoldersForProjectPage :many
SELECT public_page_folders.id,
       public_page_folders.public_page_id,
       public_page_folders.parent_id,
       public_page_folders.name,
       public_page_folders.description,
       public_page_folders.sort_order,
       public_page_folders.created_at,
       public_page_folders.updated_at
FROM public_page_folders
JOIN public_pages ON public_pages.id = public_page_folders.public_page_id
WHERE public_pages.project_id = sqlc.arg(project_id)
  AND public_page_folders.public_page_id = sqlc.arg(public_page_id)
  AND public_pages.deleted_at IS NULL
ORDER BY public_page_folders.sort_order ASC,
         public_page_folders.created_at ASC,
         public_page_folders.id ASC;

-- name: CreatePublicPageFolder :one
INSERT INTO public_page_folders (public_page_id, parent_id, name, description, sort_order)
SELECT sqlc.arg(public_page_id), sqlc.narg(parent_id), sqlc.arg(name), sqlc.narg(description), sqlc.arg(sort_order)
FROM public_pages
WHERE public_pages.project_id = sqlc.arg(project_id)
  AND public_pages.id = sqlc.arg(public_page_id)
  AND public_pages.deleted_at IS NULL
RETURNING id, public_page_id, parent_id, name, description, sort_order, created_at, updated_at;

-- name: UpdatePublicPageFolder :one
UPDATE public_page_folders
SET parent_id = CASE
        WHEN sqlc.arg(parent_id_set)::boolean THEN sqlc.narg(parent_id)
        ELSE public_page_folders.parent_id
    END,
    name = coalesce(sqlc.narg(name), public_page_folders.name),
    description = CASE
        WHEN sqlc.arg(description_set)::boolean THEN sqlc.narg(description)
        ELSE public_page_folders.description
    END,
    sort_order = coalesce(sqlc.narg(sort_order), public_page_folders.sort_order)
FROM public_pages
WHERE public_pages.id = public_page_folders.public_page_id
  AND public_pages.project_id = sqlc.arg(project_id)
  AND public_page_folders.public_page_id = sqlc.arg(public_page_id)
  AND public_page_folders.id = sqlc.arg(id)
  AND public_pages.deleted_at IS NULL
RETURNING public_page_folders.id,
          public_page_folders.public_page_id,
          public_page_folders.parent_id,
          public_page_folders.name,
          public_page_folders.description,
          public_page_folders.sort_order,
          public_page_folders.created_at,
          public_page_folders.updated_at;

-- name: DeletePublicPageFolder :one
DELETE FROM public_page_folders
USING public_pages
WHERE public_pages.id = public_page_folders.public_page_id
  AND public_pages.project_id = sqlc.arg(project_id)
  AND public_page_folders.public_page_id = sqlc.arg(public_page_id)
  AND public_page_folders.id = sqlc.arg(id)
  AND public_pages.deleted_at IS NULL
RETURNING public_page_folders.id;

-- name: DeletePublicPageFolderChecks :exec
DELETE FROM public_page_folder_checks
WHERE public_page_id = $1
  AND project_id = $2
  AND folder_id = $3;

-- name: CreatePublicPageFolderCheck :one
INSERT INTO public_page_folder_checks (public_page_id, project_id, folder_id, check_id, sort_order)
SELECT sqlc.arg(public_page_id), sqlc.arg(project_id), sqlc.arg(folder_id), sqlc.arg(check_id), sqlc.arg(sort_order)
FROM public_pages
JOIN public_page_folders
    ON public_page_folders.public_page_id = public_pages.id
JOIN checks
    ON checks.project_id = public_pages.project_id
WHERE public_pages.id = sqlc.arg(public_page_id)
  AND public_pages.project_id = sqlc.arg(project_id)
  AND public_page_folders.id = sqlc.arg(folder_id)
  AND checks.id = sqlc.arg(check_id)
  AND checks.check_type = 'ping'
  AND public_pages.deleted_at IS NULL
  AND checks.deleted_at IS NULL
RETURNING public_page_id, project_id, folder_id, check_id, sort_order, created_at;

-- name: ListPublicPageFolderChecksForProjectPage :many
SELECT public_page_folder_checks.public_page_id,
       public_page_folder_checks.folder_id,
       public_page_folder_checks.check_id,
       public_page_folder_checks.sort_order,
       checks.name AS check_name,
       checks.description AS check_description,
       checks.interval_seconds AS check_interval_seconds,
       checks.created_at AS check_created_at,
       checks.updated_at AS check_updated_at
FROM public_page_folder_checks
JOIN public_pages
    ON public_pages.id = public_page_folder_checks.public_page_id
    AND public_pages.project_id = public_page_folder_checks.project_id
JOIN checks
    ON checks.project_id = public_page_folder_checks.project_id
    AND checks.id = public_page_folder_checks.check_id
WHERE public_pages.project_id = sqlc.arg(project_id)
  AND public_pages.id = sqlc.arg(public_page_id)
  AND public_pages.deleted_at IS NULL
  AND checks.deleted_at IS NULL
ORDER BY public_page_folder_checks.sort_order ASC,
         checks.name ASC,
         checks.id ASC;

-- name: ListPublicPagePingPairs :many
SELECT public_page_folder_checks.folder_id,
       probes.id AS probe_id,
       probes.name AS probe_name,
       probes.location_name AS probe_location_name,
       (CASE
           WHEN probe_statuses.last_seen_at IS NULL THEN 'offline'::probe_state
           WHEN probe_statuses.last_seen_at < now() - interval '35 seconds' THEN 'offline'::probe_state
           ELSE probe_statuses.status
       END)::probe_state AS probe_status,
       checks.id AS check_id,
       checks.name AS check_name,
       checks.description AS check_description,
       checks.interval_seconds AS check_interval_seconds
FROM public_page_folder_checks
JOIN checks
    ON checks.project_id = public_page_folder_checks.project_id
    AND checks.id = public_page_folder_checks.check_id
JOIN probe_check_assignments
    ON probe_check_assignments.project_id = public_page_folder_checks.project_id
    AND probe_check_assignments.check_id = public_page_folder_checks.check_id
JOIN probes
    ON probes.project_id = probe_check_assignments.project_id
    AND probes.id = probe_check_assignments.probe_id
LEFT JOIN probe_statuses ON probe_statuses.probe_id = probes.id
WHERE public_page_folder_checks.public_page_id = sqlc.arg(public_page_id)
  AND public_page_folder_checks.project_id = sqlc.arg(project_id)
  AND checks.check_type = 'ping'
  AND checks.deleted_at IS NULL
  AND probe_check_assignments.deleted_at IS NULL
  AND probes.enabled = true
  AND probes.deleted_at IS NULL
ORDER BY public_page_folder_checks.sort_order ASC,
         checks.name ASC,
         probes.name ASC,
         probes.id ASC;

-- name: ResolvePublicPingPairProjectID :one
SELECT public_pages.project_id
FROM public_pages
JOIN public_page_folder_checks
    ON public_page_folder_checks.public_page_id = public_pages.id
JOIN checks
    ON checks.project_id = public_page_folder_checks.project_id
    AND checks.id = public_page_folder_checks.check_id
JOIN probe_check_assignments
    ON probe_check_assignments.project_id = public_page_folder_checks.project_id
    AND probe_check_assignments.check_id = public_page_folder_checks.check_id
JOIN probes
    ON probes.project_id = probe_check_assignments.project_id
    AND probes.id = probe_check_assignments.probe_id
WHERE public_pages.slug = sqlc.arg(slug)
  AND public_pages.enabled = true
  AND public_pages.deleted_at IS NULL
  AND checks.id = sqlc.arg(check_id)
  AND checks.check_type = 'ping'
  AND checks.deleted_at IS NULL
  AND probes.id = sqlc.arg(probe_id)
  AND probes.enabled = true
  AND probes.deleted_at IS NULL
  AND probe_check_assignments.deleted_at IS NULL
LIMIT 1;

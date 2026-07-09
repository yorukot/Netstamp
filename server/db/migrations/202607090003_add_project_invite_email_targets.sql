-- +goose Up
ALTER TABLE project_invites
    ADD COLUMN invited_email citext;

UPDATE project_invites
SET invited_email = users.email
FROM users
WHERE users.id = project_invites.invited_user_id;

ALTER TABLE project_invites
    ALTER COLUMN invited_email SET NOT NULL,
    ALTER COLUMN invited_user_id DROP NOT NULL,
    ADD CONSTRAINT project_invites_invited_email_not_empty CHECK (length(btrim(invited_email::text)) > 0);

DROP INDEX uq_project_invites_pending_project_user;
DROP INDEX idx_project_invites_invited_user_pending;

CREATE UNIQUE INDEX uq_project_invites_pending_project_email
    ON project_invites (project_id, invited_email)
    WHERE status = 'pending';
CREATE INDEX idx_project_invites_invited_email_pending
    ON project_invites (invited_email, created_at DESC, id DESC)
    WHERE status = 'pending';

-- +goose Down
UPDATE project_invites
SET invited_user_id = users.id
FROM users
WHERE project_invites.invited_user_id IS NULL
  AND users.email = project_invites.invited_email;

DELETE FROM project_invites
WHERE invited_user_id IS NULL;

DROP INDEX idx_project_invites_invited_email_pending;
DROP INDEX uq_project_invites_pending_project_email;

ALTER TABLE project_invites
    DROP CONSTRAINT project_invites_invited_email_not_empty,
    ALTER COLUMN invited_user_id SET NOT NULL,
    DROP COLUMN invited_email;

CREATE UNIQUE INDEX uq_project_invites_pending_project_user
    ON project_invites (project_id, invited_user_id)
    WHERE status = 'pending';
CREATE INDEX idx_project_invites_invited_user_pending
    ON project_invites (invited_user_id, created_at DESC, id DESC)
    WHERE status = 'pending';

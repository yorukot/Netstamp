-- +goose Up
CREATE TYPE project_invite_status AS ENUM ('pending', 'accepted', 'rejected');

CREATE TABLE project_invites (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id uuid NOT NULL REFERENCES projects(id),
    invited_user_id uuid NOT NULL REFERENCES users(id),
    invited_by_user_id uuid NOT NULL REFERENCES users(id),
    role project_member_role NOT NULL,
    status project_invite_status NOT NULL DEFAULT 'pending',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    resolved_at timestamptz,
    CONSTRAINT project_invites_resolved_at_after_created_at CHECK (resolved_at IS NULL OR resolved_at >= created_at),
    CONSTRAINT project_invites_resolution_consistent CHECK (
        (status = 'pending' AND resolved_at IS NULL) OR
        (status <> 'pending' AND resolved_at IS NOT NULL)
    )
);

CREATE UNIQUE INDEX uq_project_invites_pending_project_user
    ON project_invites (project_id, invited_user_id)
    WHERE status = 'pending';
CREATE INDEX idx_project_invites_project_pending
    ON project_invites (project_id, created_at ASC, id ASC)
    WHERE status = 'pending';
CREATE INDEX idx_project_invites_invited_user_pending
    ON project_invites (invited_user_id, created_at DESC, id DESC)
    WHERE status = 'pending';
CREATE INDEX idx_project_invites_invited_by_user_id ON project_invites (invited_by_user_id);

CREATE TRIGGER set_project_invites_updated_at
    BEFORE UPDATE ON project_invites
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TABLE IF EXISTS project_invites;
DROP TYPE IF EXISTS project_invite_status;

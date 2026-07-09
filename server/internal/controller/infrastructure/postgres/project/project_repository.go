package pgproject

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/trace"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	"github.com/yorukot/netstamp/internal/domain/identity"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type ProjectRepository struct {
	queries *sqlc.Queries
	tx      *postgres.Transactor
}

func NewProjectRepository(pool *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{
		queries: sqlc.New(pool),
		tx:      postgres.NewTransactor(pool),
	}
}

func (r *ProjectRepository) CreateProjectWithOwner(ctx context.Context, input domainproject.Project) (domainproject.Project, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprojectTracer, "projects", "postgres.projects.create_with_owner", "INSERT", "INSERT projects and owner membership")
	defer span.End()

	userID, err := postgres.ParseUUID(input.CreatedByUserID, identity.ErrUserNotFound)
	if err != nil {
		return domainproject.Project{}, err
	}

	var project domainproject.Project
	err = r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		row, createErr := q.CreateProject(ctx, sqlc.CreateProjectParams{
			Name:            input.Name,
			Slug:            input.Slug,
			CreatedByUserID: userID,
		})
		if createErr != nil {
			_, mapped := mapCreateProjectError(createErr)
			return mapped
		}
		project = mapProject(row)

		if _, memberErr := q.CreateProjectMember(ctx, sqlc.CreateProjectMemberParams{
			ProjectID: row.ID,
			UserID:    userID,
			Role:      sqlc.ProjectMemberRoleOwner,
		}); memberErr != nil {
			_, mapped := mapCreateProjectMemberError(memberErr)
			return mapped
		}

		return nil
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domainproject.Project{}, err
	}

	return project, nil
}

func (r *ProjectRepository) ListProjectsForUser(ctx context.Context, userIDValue string) ([]domainproject.Project, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprojectTracer, "projects", "postgres.projects.list_for_user", "SELECT", "SELECT projects for member")
	defer span.End()

	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListProjectsForUser(ctx, userID)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}

	projects := make([]domainproject.Project, 0, len(rows))
	for _, row := range rows {
		projects = append(projects, mapProject(row))
	}

	return projects, nil
}

func (r *ProjectRepository) GetProjectForUser(ctx context.Context, projectRef, userIDValue string) (domainproject.Project, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprojectTracer, "projects", "postgres.projects.select_for_user", "SELECT", "SELECT project for member")
	defer span.End()

	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return domainproject.Project{}, err
	}

	var row sqlc.Project
	projectID, parseErr := uuid.Parse(projectRef)
	if parseErr == nil {
		row, err = r.queries.GetProjectForUser(ctx, sqlc.GetProjectForUserParams{
			ID:     projectID,
			UserID: userID,
		})
	} else {
		row, err = r.queries.GetProjectBySlugForUser(ctx, sqlc.GetProjectBySlugForUserParams{
			Slug:   projectRef,
			UserID: userID,
		})
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainproject.Project{}, domainproject.ErrProjectNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return domainproject.Project{}, err
	}

	return mapProject(row), nil
}

func (r *ProjectRepository) GetMemberRole(ctx context.Context, projectIDValue, userIDValue string) (domainproject.Role, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprojectTracer, "project_members", "postgres.project_members.select_role", "SELECT", "SELECT active project member role")
	defer span.End()

	projectID, userID, err := parseProjectAndUserIDs(projectIDValue, userIDValue)
	if err != nil {
		return "", err
	}

	role, err := r.queries.GetActiveProjectMemberRole(ctx, sqlc.GetActiveProjectMemberRoleParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domainproject.ErrMemberNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return "", err
	}

	return domainproject.Role(role), nil
}

func (r *ProjectRepository) UpdateProject(ctx context.Context, input domainproject.Project) (domainproject.Project, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprojectTracer, "projects", "postgres.projects.update", "UPDATE", "UPDATE project")
	defer span.End()

	projectID, err := postgres.ParseUUID(input.ID, domainproject.ErrProjectNotFound)
	if err != nil {
		return domainproject.Project{}, err
	}

	row, err := r.queries.UpdateProject(ctx, sqlc.UpdateProjectParams{
		ID:   projectID,
		Name: input.Name,
		Slug: input.Slug,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainproject.Project{}, domainproject.ErrProjectNotFound
		}
		if ok, mapped := mapCreateProjectError(err); ok {
			return domainproject.Project{}, mapped
		}
		postgres.RecordDBSpanError(span, err)
		return domainproject.Project{}, err
	}

	return mapProject(row), nil
}

func (r *ProjectRepository) SoftDeleteProject(ctx context.Context, projectIDValue string) error {
	ctx, span := postgres.StartDBSpan(ctx, pgprojectTracer, "projects", "postgres.projects.soft_delete", "UPDATE", "SOFT DELETE project")
	defer span.End()

	projectID, err := postgres.ParseUUID(projectIDValue, domainproject.ErrProjectNotFound)
	if err != nil {
		return err
	}

	if _, err := r.queries.SoftDeleteProject(ctx, projectID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainproject.ErrProjectNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
}

func (r *ProjectRepository) ListMembers(ctx context.Context, projectIDValue string) ([]domainproject.Member, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprojectTracer, "project_members", "postgres.project_members.list", "SELECT", "SELECT active project members")
	defer span.End()

	projectID, err := postgres.ParseUUID(projectIDValue, domainproject.ErrProjectNotFound)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListActiveProjectMembers(ctx, projectID)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}

	members := make([]domainproject.Member, 0, len(rows))
	for _, row := range rows {
		members = append(members, mapListMember(row))
	}

	return members, nil
}

func (r *ProjectRepository) GetMember(ctx context.Context, projectIDValue, userIDValue string) (domainproject.Member, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprojectTracer, "project_members", "postgres.project_members.select", "SELECT", "SELECT active project member")
	defer span.End()

	projectID, userID, err := parseProjectAndUserIDs(projectIDValue, userIDValue)
	if err != nil {
		return domainproject.Member{}, err
	}

	row, err := queryProjectRow(span, domainproject.ErrMemberNotFound, func() (sqlc.GetActiveProjectMemberRow, error) {
		return r.queries.GetActiveProjectMember(ctx, sqlc.GetActiveProjectMemberParams{
			ProjectID: projectID,
			UserID:    userID,
		})
	})
	if err != nil {
		return domainproject.Member{}, err
	}

	return mapGetMember(row), nil
}

func (r *ProjectRepository) UpdateMemberRole(ctx context.Context, input domainproject.Member) (domainproject.Member, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprojectTracer, "project_members", "postgres.project_members.update_role", "UPDATE", "UPDATE project member role")
	defer span.End()

	projectID, userID, err := parseProjectAndUserIDs(input.ProjectID, input.UserID)
	if err != nil {
		return domainproject.Member{}, err
	}

	row, err := r.queries.UpdateProjectMemberRole(ctx, sqlc.UpdateProjectMemberRoleParams{
		ProjectID: projectID,
		UserID:    userID,
		Role:      sqlc.ProjectMemberRole(input.Role),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainproject.Member{}, domainproject.ErrMemberNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return domainproject.Member{}, err
	}

	return mapUpdateMember(row), nil
}

func (r *ProjectRepository) DeleteMember(ctx context.Context, projectIDValue, userIDValue string) error {
	ctx, span := postgres.StartDBSpan(ctx, pgprojectTracer, "project_members", "postgres.project_members.delete", "DELETE", "DELETE project member")
	defer span.End()

	projectID, userID, err := parseProjectAndUserIDs(projectIDValue, userIDValue)
	if err != nil {
		return err
	}

	if _, err := r.queries.DeleteProjectMember(ctx, sqlc.DeleteProjectMemberParams{
		ProjectID: projectID,
		UserID:    userID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainproject.ErrMemberNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
}

func (r *ProjectRepository) CountOwners(ctx context.Context, projectIDValue string) (int, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprojectTracer, "project_members", "postgres.project_members.count_owners", "SELECT", "COUNT active project owners")
	defer span.End()

	projectID, err := postgres.ParseUUID(projectIDValue, domainproject.ErrProjectNotFound)
	if err != nil {
		return 0, err
	}

	count, err := r.queries.CountActiveProjectOwners(ctx, projectID)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return 0, err
	}

	return int(count), nil
}

func (r *ProjectRepository) CreateInvite(ctx context.Context, input domainproject.Invite) (domainproject.Invite, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprojectTracer, "project_invites", "postgres.project_invites.insert", "INSERT", "INSERT project invite")
	defer span.End()

	projectID, invitedByUserID, err := parseProjectAndUserIDs(input.ProjectID, input.InvitedByUserID)
	if err != nil {
		return domainproject.Invite{}, err
	}

	row, err := r.queries.CreateProjectInvite(ctx, sqlc.CreateProjectInviteParams{
		ProjectID:       projectID,
		InvitedEmail:    input.InvitedEmail,
		InvitedByUserID: invitedByUserID,
		Role:            sqlc.ProjectMemberRole(input.Role),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainproject.Invite{}, domainproject.ErrMemberAlreadyExists
		}
		ok, mapped := mapCreateProjectInviteError(err)
		if !ok {
			postgres.RecordDBSpanError(span, err)
		}
		return domainproject.Invite{}, mapped
	}

	return mapCreateInvite(row), nil
}

func (r *ProjectRepository) ListProjectInvites(ctx context.Context, projectIDValue string) ([]domainproject.Invite, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprojectTracer, "project_invites", "postgres.project_invites.list_for_project", "SELECT", "SELECT pending project invites")
	defer span.End()

	projectID, err := postgres.ParseUUID(projectIDValue, domainproject.ErrProjectNotFound)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListPendingProjectInvites(ctx, projectID)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}

	invites := make([]domainproject.Invite, 0, len(rows))
	for _, row := range rows {
		invites = append(invites, mapListProjectInvite(row))
	}

	return invites, nil
}

func (r *ProjectRepository) ListUserInvites(ctx context.Context, userIDValue string) ([]domainproject.Invite, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprojectTracer, "project_invites", "postgres.project_invites.list_for_user", "SELECT", "SELECT pending user invites")
	defer span.End()

	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListPendingProjectInvitesForUser(ctx, userID)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}

	invites := make([]domainproject.Invite, 0, len(rows))
	for _, row := range rows {
		invites = append(invites, mapListUserInvite(row))
	}

	return invites, nil
}

func (r *ProjectRepository) CancelInvite(ctx context.Context, projectIDValue, inviteIDValue string) (domainproject.Invite, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprojectTracer, "project_invites", "postgres.project_invites.cancel", "UPDATE", "CANCEL pending project invite")
	defer span.End()

	projectID, inviteID, err := parseProjectAndInviteIDs(projectIDValue, inviteIDValue)
	if err != nil {
		return domainproject.Invite{}, err
	}

	row, err := queryProjectRow(span, domainproject.ErrInviteNotFound, func() (sqlc.CancelPendingProjectInviteRow, error) {
		return r.queries.CancelPendingProjectInvite(ctx, sqlc.CancelPendingProjectInviteParams{
			ID:        inviteID,
			ProjectID: projectID,
		})
	})
	if err != nil {
		return domainproject.Invite{}, err
	}

	return mapCancelInvite(row), nil
}

func (r *ProjectRepository) AcceptInvite(ctx context.Context, inviteIDValue, userIDValue string) (domainproject.Invite, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprojectTracer, "project_invites", "postgres.project_invites.accept", "UPDATE", "ACCEPT pending project invite")
	defer span.End()

	inviteID, userID, err := parseInviteAndUserIDs(inviteIDValue, userIDValue)
	if err != nil {
		return domainproject.Invite{}, err
	}

	var invite domainproject.Invite
	err = r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		row, acceptErr := q.AcceptPendingProjectInvite(ctx, sqlc.AcceptPendingProjectInviteParams{
			ID:            inviteID,
			CurrentUserID: userID,
		})
		if acceptErr != nil {
			if errors.Is(acceptErr, pgx.ErrNoRows) {
				return domainproject.ErrInviteNotFound
			}
			return acceptErr
		}
		invite = mapAcceptInvite(row)

		if _, memberErr := q.CreateProjectMember(ctx, sqlc.CreateProjectMemberParams{
			ProjectID: row.ProjectID,
			UserID:    userID,
			Role:      row.Role,
		}); memberErr != nil {
			_, mapped := mapCreateProjectMemberError(memberErr)
			return mapped
		}

		return nil
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domainproject.Invite{}, err
	}

	return invite, nil
}

func (r *ProjectRepository) RejectInvite(ctx context.Context, inviteIDValue, userIDValue string) (domainproject.Invite, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprojectTracer, "project_invites", "postgres.project_invites.reject", "UPDATE", "REJECT pending project invite")
	defer span.End()

	inviteID, userID, err := parseInviteAndUserIDs(inviteIDValue, userIDValue)
	if err != nil {
		return domainproject.Invite{}, err
	}

	row, err := queryProjectRow(span, domainproject.ErrInviteNotFound, func() (sqlc.RejectPendingProjectInviteRow, error) {
		return r.queries.RejectPendingProjectInvite(ctx, sqlc.RejectPendingProjectInviteParams{
			ID:            inviteID,
			CurrentUserID: userID,
		})
	})
	if err != nil {
		return domainproject.Invite{}, err
	}

	return mapRejectInvite(row), nil
}

func queryProjectRow[T any](span trace.Span, notFound error, query func() (T, error)) (T, error) {
	row, err := query()
	if err != nil {
		var zero T
		if errors.Is(err, pgx.ErrNoRows) {
			return zero, notFound
		}
		postgres.RecordDBSpanError(span, err)
		return zero, err
	}

	return row, nil
}

func parseProjectAndUserIDs(projectIDValue, userIDValue string) (uuid.UUID, uuid.UUID, error) {
	projectID, err := postgres.ParseUUID(projectIDValue, domainproject.ErrProjectNotFound)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	return projectID, userID, nil
}

func parseInviteAndUserIDs(inviteIDValue, userIDValue string) (uuid.UUID, uuid.UUID, error) {
	inviteID, err := postgres.ParseUUID(inviteIDValue, domainproject.ErrInviteNotFound)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	return inviteID, userID, nil
}

func parseProjectAndInviteIDs(projectIDValue, inviteIDValue string) (uuid.UUID, uuid.UUID, error) {
	projectID, err := postgres.ParseUUID(projectIDValue, domainproject.ErrProjectNotFound)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	inviteID, err := postgres.ParseUUID(inviteIDValue, domainproject.ErrInviteNotFound)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	return projectID, inviteID, nil
}

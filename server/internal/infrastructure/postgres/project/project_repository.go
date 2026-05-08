package pgproject

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yorukot/netstamp/internal/domain/identity"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	"github.com/yorukot/netstamp/internal/infrastructure/postgres"
	"github.com/yorukot/netstamp/internal/infrastructure/postgres/sqlc"
	"github.com/yorukot/netstamp/internal/normalize"
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

func (r *ProjectRepository) CreateProjectWithOwner(ctx context.Context, input domainproject.CreateProjectStorageInput) (domainproject.Project, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprojectTracer, "projects", "postgres.projects.create_with_owner", "INSERT", "INSERT projects and owner membership")
	defer span.End()

	userID, err := postgres.ParseUUID(input.CreatedByUserID, identity.ErrUserNotFound)
	if err != nil {
		return domainproject.Project{}, err
	}

	var project domainproject.Project
	err = r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		row, err := q.CreateProject(ctx, sqlc.CreateProjectParams{
			Name:            input.Name,
			Slug:            input.Slug,
			CreatedByUserID: userID,
		})
		if err != nil {
			return mapCreateProjectError(err)
		}
		project = mapProject(row)

		if _, err := q.CreateProjectMember(ctx, sqlc.CreateProjectMemberParams{
			ProjectID: row.ID,
			UserID:    userID,
			Role:      sqlc.ProjectMemberRoleOwner,
		}); err != nil {
			return mapCreateProjectMemberError(err)
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

	userID, err := postgres.ParseUUID(userIDValue, domainproject.ErrProjectNotFound)
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

func (r *ProjectRepository) GetProjectForUser(ctx context.Context, projectRef string, userIDValue string) (domainproject.Project, error) {
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
		if !normalize.IsProjectSlug(projectRef) {
			return domainproject.Project{}, domainproject.ErrProjectNotFound
		}
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

func (r *ProjectRepository) GetMemberRole(ctx context.Context, projectIDValue string, userIDValue string) (domainproject.Role, error) {
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
			return "", domainproject.ErrProjectNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return "", err
	}

	return domainproject.Role(role), nil
}

func (r *ProjectRepository) UpdateProject(ctx context.Context, input domainproject.UpdateProjectStorageInput) (domainproject.Project, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprojectTracer, "projects", "postgres.projects.update", "UPDATE", "UPDATE project")
	defer span.End()

	projectID, err := postgres.ParseUUID(input.ProjectID, domainproject.ErrProjectNotFound)
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
		if mapped := mapCreateProjectError(err); mapped != err {
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

func (r *ProjectRepository) GetMember(ctx context.Context, projectIDValue string, userIDValue string) (domainproject.Member, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprojectTracer, "project_members", "postgres.project_members.select", "SELECT", "SELECT active project member")
	defer span.End()

	projectID, userID, err := parseProjectAndUserIDs(projectIDValue, userIDValue)
	if err != nil {
		return domainproject.Member{}, domainproject.ErrMemberNotFound
	}

	row, err := r.queries.GetActiveProjectMember(ctx, sqlc.GetActiveProjectMemberParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainproject.Member{}, domainproject.ErrMemberNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return domainproject.Member{}, err
	}

	return mapGetMember(row), nil
}

func (r *ProjectRepository) AddMember(ctx context.Context, input domainproject.AddMemberStorageInput) (domainproject.Member, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprojectTracer, "project_members", "postgres.project_members.insert", "INSERT", "INSERT project member")
	defer span.End()

	projectID, userID, err := parseProjectAndUserIDs(input.ProjectID, input.UserID)
	if err != nil {
		return domainproject.Member{}, err
	}

	row, err := r.queries.CreateProjectMember(ctx, sqlc.CreateProjectMemberParams{
		ProjectID: projectID,
		UserID:    userID,
		Role:      sqlc.ProjectMemberRole(input.Role),
	})
	if err != nil {
		mapped := mapCreateProjectMemberError(err)
		if mapped == err {
			postgres.RecordDBSpanError(span, err)
		}
		return domainproject.Member{}, mapped
	}

	return mapCreateMember(row), nil
}

func (r *ProjectRepository) UpdateMemberRole(ctx context.Context, input domainproject.UpdateMemberRoleStorageInput) (domainproject.Member, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprojectTracer, "project_members", "postgres.project_members.update_role", "UPDATE", "UPDATE project member role")
	defer span.End()

	projectID, userID, err := parseProjectAndUserIDs(input.ProjectID, input.UserID)
	if err != nil {
		return domainproject.Member{}, domainproject.ErrMemberNotFound
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

func parseProjectAndUserIDs(projectIDValue string, userIDValue string) (uuid.UUID, uuid.UUID, error) {
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

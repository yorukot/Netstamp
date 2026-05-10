package pgproject

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func mapProject(row sqlc.Project) domainproject.Project {
	return domainproject.Project{
		ID:              row.ID.String(),
		Name:            row.Name,
		Slug:            row.Slug,
		CreatedByUserID: row.CreatedByUserID.String(),
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
		DeletedAt:       timePtr(row.DeletedAt),
	}
}

func mapCreateMember(row sqlc.CreateProjectMemberRow) domainproject.Member {
	return domainproject.Member{
		ID:        row.ID.String(),
		ProjectID: row.ProjectID.String(),
		UserID:    row.UserID.String(),
		Email:     row.Email,
		Role:      domainproject.Role(row.Role),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}

func timePtr(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}

	return &value.Time
}

func mapListMember(row sqlc.ListActiveProjectMembersRow) domainproject.Member {
	return domainproject.Member{
		ID:        row.ID.String(),
		ProjectID: row.ProjectID.String(),
		UserID:    row.UserID.String(),
		Email:     row.Email,
		Role:      domainproject.Role(row.Role),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}

func mapGetMember(row sqlc.GetActiveProjectMemberRow) domainproject.Member {
	return domainproject.Member{
		ID:        row.ID.String(),
		ProjectID: row.ProjectID.String(),
		UserID:    row.UserID.String(),
		Email:     row.Email,
		Role:      domainproject.Role(row.Role),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}

func mapUpdateMember(row sqlc.UpdateProjectMemberRoleRow) domainproject.Member {
	return domainproject.Member{
		ID:        row.ID.String(),
		ProjectID: row.ProjectID.String(),
		UserID:    row.UserID.String(),
		Email:     row.Email,
		Role:      domainproject.Role(row.Role),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}

package pgproject

import (
	"time"

	"github.com/google/uuid"
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
	return mapMemberFields(row.ID, row.ProjectID, row.UserID, row.Role, row.CreatedAt, row.UpdatedAt, row.UserEmail, row.UserDisplayName)
}

func timePtr(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}

	return &value.Time
}

func mapListMember(row sqlc.ListActiveProjectMembersRow) domainproject.Member {
	return mapMemberFields(row.ID, row.ProjectID, row.UserID, row.Role, row.CreatedAt, row.UpdatedAt, row.UserEmail, row.UserDisplayName)
}

func mapGetMember(row sqlc.GetActiveProjectMemberRow) domainproject.Member {
	return mapMemberFields(row.ID, row.ProjectID, row.UserID, row.Role, row.CreatedAt, row.UpdatedAt, row.UserEmail, row.UserDisplayName)
}

func mapUpdateMember(row sqlc.UpdateProjectMemberRoleRow) domainproject.Member {
	return mapMemberFields(row.ID, row.ProjectID, row.UserID, row.Role, row.CreatedAt, row.UpdatedAt, row.UserEmail, row.UserDisplayName)
}

func mapMemberFields(
	id uuid.UUID,
	projectID uuid.UUID,
	userID uuid.UUID,
	role sqlc.ProjectMemberRole,
	createdAt pgtype.Timestamptz,
	updatedAt pgtype.Timestamptz,
	userEmail string,
	userDisplayName string,
) domainproject.Member {
	return domainproject.Member{
		ID:        id.String(),
		ProjectID: projectID.String(),
		UserID:    userID.String(),
		Role:      domainproject.Role(role),
		User: domainproject.MemberUser{
			ID:          userID.String(),
			Email:       userEmail,
			DisplayName: userDisplayName,
		},
		CreatedAt: createdAt.Time,
		UpdatedAt: updatedAt.Time,
	}
}

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

func mapCreateInvite(row sqlc.CreateProjectInviteRow) domainproject.Invite {
	return mapInviteFields(
		row.ID,
		row.ProjectID,
		row.InvitedUserID,
		row.InvitedByUserID,
		row.Role,
		row.Status,
		row.CreatedAt,
		row.UpdatedAt,
		row.ResolvedAt,
		row.ProjectName,
		row.ProjectSlug,
		row.InvitedUserEmail,
		row.InvitedUserDisplayName,
		row.InvitedByUserEmail,
		row.InvitedByUserDisplayName,
	)
}

func mapListProjectInvite(row sqlc.ListPendingProjectInvitesRow) domainproject.Invite {
	return mapInviteFields(
		row.ID,
		row.ProjectID,
		row.InvitedUserID,
		row.InvitedByUserID,
		row.Role,
		row.Status,
		row.CreatedAt,
		row.UpdatedAt,
		row.ResolvedAt,
		row.ProjectName,
		row.ProjectSlug,
		row.InvitedUserEmail,
		row.InvitedUserDisplayName,
		row.InvitedByUserEmail,
		row.InvitedByUserDisplayName,
	)
}

func mapListUserInvite(row sqlc.ListPendingProjectInvitesForUserRow) domainproject.Invite {
	return mapInviteFields(
		row.ID,
		row.ProjectID,
		row.InvitedUserID,
		row.InvitedByUserID,
		row.Role,
		row.Status,
		row.CreatedAt,
		row.UpdatedAt,
		row.ResolvedAt,
		row.ProjectName,
		row.ProjectSlug,
		row.InvitedUserEmail,
		row.InvitedUserDisplayName,
		row.InvitedByUserEmail,
		row.InvitedByUserDisplayName,
	)
}

func mapAcceptInvite(row sqlc.AcceptPendingProjectInviteRow) domainproject.Invite {
	return mapInviteFields(
		row.ID,
		row.ProjectID,
		row.InvitedUserID,
		row.InvitedByUserID,
		row.Role,
		row.Status,
		row.CreatedAt,
		row.UpdatedAt,
		row.ResolvedAt,
		row.ProjectName,
		row.ProjectSlug,
		row.InvitedUserEmail,
		row.InvitedUserDisplayName,
		row.InvitedByUserEmail,
		row.InvitedByUserDisplayName,
	)
}

func mapRejectInvite(row sqlc.RejectPendingProjectInviteRow) domainproject.Invite {
	return mapInviteFields(
		row.ID,
		row.ProjectID,
		row.InvitedUserID,
		row.InvitedByUserID,
		row.Role,
		row.Status,
		row.CreatedAt,
		row.UpdatedAt,
		row.ResolvedAt,
		row.ProjectName,
		row.ProjectSlug,
		row.InvitedUserEmail,
		row.InvitedUserDisplayName,
		row.InvitedByUserEmail,
		row.InvitedByUserDisplayName,
	)
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

func mapInviteFields(
	id uuid.UUID,
	projectID uuid.UUID,
	invitedUserID uuid.UUID,
	invitedByUserID uuid.UUID,
	role sqlc.ProjectMemberRole,
	status sqlc.ProjectInviteStatus,
	createdAt pgtype.Timestamptz,
	updatedAt pgtype.Timestamptz,
	resolvedAt pgtype.Timestamptz,
	projectName string,
	projectSlug string,
	invitedUserEmail string,
	invitedUserDisplayName string,
	invitedByUserEmail string,
	invitedByUserDisplayName string,
) domainproject.Invite {
	return domainproject.Invite{
		ID:              id.String(),
		ProjectID:       projectID.String(),
		InvitedUserID:   invitedUserID.String(),
		InvitedByUserID: invitedByUserID.String(),
		Role:            domainproject.Role(role),
		Status:          domainproject.InviteStatus(status),
		Project: domainproject.InviteProject{
			ID:   projectID.String(),
			Name: projectName,
			Slug: projectSlug,
		},
		InvitedUser: domainproject.MemberUser{
			ID:          invitedUserID.String(),
			Email:       invitedUserEmail,
			DisplayName: invitedUserDisplayName,
		},
		InvitedByUser: domainproject.MemberUser{
			ID:          invitedByUserID.String(),
			Email:       invitedByUserEmail,
			DisplayName: invitedByUserDisplayName,
		},
		CreatedAt:  createdAt.Time,
		UpdatedAt:  updatedAt.Time,
		ResolvedAt: timePtr(resolvedAt),
	}
}

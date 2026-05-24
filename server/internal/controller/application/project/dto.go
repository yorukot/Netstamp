package project

import domainproject "github.com/yorukot/netstamp/internal/domain/project"

type CreateProjectInput struct {
	CurrentUserID string
	Name          string
	Slug          string
}

type ListProjectsInput struct {
	CurrentUserID string
}

type GetProjectInput struct {
	CurrentUserID string
	ProjectRef    string
}

type UpdateProjectInput struct {
	CurrentUserID string
	ProjectRef    string
	Name          *string
	Slug          *string
}

type DeleteProjectInput struct {
	CurrentUserID string
	ProjectRef    string
}

type ListMembersInput struct {
	CurrentUserID string
	ProjectRef    string
}

type CreateInviteInput struct {
	CurrentUserID string
	ProjectRef    string
	Email         string
	Role          domainproject.Role
}

type ListProjectInvitesInput struct {
	CurrentUserID string
	ProjectRef    string
}

type ListUserInvitesInput struct {
	CurrentUserID string
}

type ResolveInviteInput struct {
	CurrentUserID string
	InviteID      string
}

type UpdateMemberRoleInput struct {
	CurrentUserID string
	ProjectRef    string
	UserID        string
	Role          domainproject.Role
}

type RemoveMemberInput struct {
	CurrentUserID string
	ProjectRef    string
	UserID        string
}

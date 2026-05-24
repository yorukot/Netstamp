package project

import (
	"context"

	"github.com/yorukot/netstamp/internal/domain/identity"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type Repository interface {
	CreateProjectWithOwner(ctx context.Context, input domainproject.Project) (domainproject.Project, error)
	ListProjectsForUser(ctx context.Context, userID string) ([]domainproject.Project, error)
	GetProjectForUser(ctx context.Context, projectRef, userID string) (domainproject.Project, error)
	GetMemberRole(ctx context.Context, projectID, userID string) (domainproject.Role, error)
	UpdateProject(ctx context.Context, input domainproject.Project) (domainproject.Project, error)
	SoftDeleteProject(ctx context.Context, projectID string) error
	ListMembers(ctx context.Context, projectID string) ([]domainproject.Member, error)
	GetMember(ctx context.Context, projectID, userID string) (domainproject.Member, error)
	UpdateMemberRole(ctx context.Context, input domainproject.Member) (domainproject.Member, error)
	DeleteMember(ctx context.Context, projectID, userID string) error
	CountOwners(ctx context.Context, projectID string) (int, error)
	CreateInvite(ctx context.Context, input domainproject.Invite) (domainproject.Invite, error)
	ListProjectInvites(ctx context.Context, projectID string) ([]domainproject.Invite, error)
	ListUserInvites(ctx context.Context, userID string) ([]domainproject.Invite, error)
	AcceptInvite(ctx context.Context, inviteID, userID string) (domainproject.Invite, error)
	RejectInvite(ctx context.Context, inviteID, userID string) (domainproject.Invite, error)
}

type UserLookup interface {
	GetUserByEmail(ctx context.Context, email string) (identity.User, error)
}

type EventRecorder interface {
	RecordProjectEvent(ctx context.Context, event ProjectEvent)
}

type ProjectEventName string

const (
	ProjectEventCreateSuccess           ProjectEventName = "project.create.success"
	ProjectEventCreateFailure           ProjectEventName = "project.create.failure"
	ProjectEventListFailure             ProjectEventName = "project.list.failure"
	ProjectEventGetFailure              ProjectEventName = "project.get.failure"
	ProjectEventUpdateFailure           ProjectEventName = "project.update.failure"
	ProjectEventDeleteSuccess           ProjectEventName = "project.delete.success"
	ProjectEventDeleteFailure           ProjectEventName = "project.delete.failure"
	ProjectEventListMembersFailure      ProjectEventName = "project.members.list.failure"
	ProjectEventUpdateMemberRoleSuccess ProjectEventName = "project.member.role_update.success"
	ProjectEventUpdateMemberRoleFailure ProjectEventName = "project.member.role_update.failure"
	ProjectEventRemoveMemberSuccess     ProjectEventName = "project.member.remove.success"
	ProjectEventRemoveMemberFailure     ProjectEventName = "project.member.remove.failure"
	ProjectEventCreateInviteSuccess     ProjectEventName = "project.invite.create.success"
	ProjectEventCreateInviteFailure     ProjectEventName = "project.invite.create.failure"
	ProjectEventListInvitesFailure      ProjectEventName = "project.invites.list.failure"
	ProjectEventListUserInvitesFailure  ProjectEventName = "project.invites.list_user.failure"
	ProjectEventAcceptInviteSuccess     ProjectEventName = "project.invite.accept.success"
	ProjectEventAcceptInviteFailure     ProjectEventName = "project.invite.accept.failure"
	ProjectEventRejectInviteSuccess     ProjectEventName = "project.invite.reject.success"
	ProjectEventRejectInviteFailure     ProjectEventName = "project.invite.reject.failure"
)

type ProjectEventAction string

const (
	ProjectActionCreate           ProjectEventAction = "create"
	ProjectActionList             ProjectEventAction = "list"
	ProjectActionGet              ProjectEventAction = "get"
	ProjectActionUpdate           ProjectEventAction = "update"
	ProjectActionDelete           ProjectEventAction = "delete"
	ProjectActionListMembers      ProjectEventAction = "list_members"
	ProjectActionUpdateMemberRole ProjectEventAction = "update_member_role"
	ProjectActionRemoveMember     ProjectEventAction = "remove_member"
	ProjectActionCreateInvite     ProjectEventAction = "create_invite"
	ProjectActionListInvites      ProjectEventAction = "list_invites"
	ProjectActionListUserInvites  ProjectEventAction = "list_user_invites"
	ProjectActionAcceptInvite     ProjectEventAction = "accept_invite"
	ProjectActionRejectInvite     ProjectEventAction = "reject_invite"
)

type ProjectEventOutcome string

const (
	ProjectOutcomeSuccess ProjectEventOutcome = "success"
	ProjectOutcomeFailure ProjectEventOutcome = "failure"
)

type ProjectEventReason string

const (
	ProjectReasonInvalidInput           ProjectEventReason = "invalid_input"
	ProjectReasonInvalidRole            ProjectEventReason = "invalid_role"
	ProjectReasonForbidden              ProjectEventReason = "forbidden"
	ProjectReasonProjectNotFound        ProjectEventReason = "project_not_found"
	ProjectReasonSlugAlreadyExists      ProjectEventReason = "slug_already_exists"
	ProjectReasonMemberAlreadyExists    ProjectEventReason = "member_already_exists"
	ProjectReasonMemberNotFound         ProjectEventReason = "member_not_found"
	ProjectReasonInviteAlreadyExists    ProjectEventReason = "invite_already_exists"
	ProjectReasonInviteNotFound         ProjectEventReason = "invite_not_found"
	ProjectReasonUserNotFound           ProjectEventReason = "user_not_found"
	ProjectReasonLastOwner              ProjectEventReason = "last_owner"
	ProjectReasonProjectCreateFailed    ProjectEventReason = "project_create_failed"
	ProjectReasonProjectListFailed      ProjectEventReason = "project_list_failed"
	ProjectReasonProjectLookupFailed    ProjectEventReason = "project_lookup_failed"
	ProjectReasonProjectUpdateFailed    ProjectEventReason = "project_update_failed"
	ProjectReasonProjectDeleteFailed    ProjectEventReason = "project_delete_failed"
	ProjectReasonRoleLookupFailed       ProjectEventReason = "role_lookup_failed"
	ProjectReasonMembersListFailed      ProjectEventReason = "members_list_failed"
	ProjectReasonMemberLookupFailed     ProjectEventReason = "member_lookup_failed"
	ProjectReasonMemberAddFailed        ProjectEventReason = "member_add_failed"
	ProjectReasonMemberRoleUpdateFailed ProjectEventReason = "member_role_update_failed"
	ProjectReasonMemberRemoveFailed     ProjectEventReason = "member_remove_failed"
	ProjectReasonOwnerCountFailed       ProjectEventReason = "owner_count_failed"
	ProjectReasonInvitesListFailed      ProjectEventReason = "invites_list_failed"
	ProjectReasonInviteCreateFailed     ProjectEventReason = "invite_create_failed"
	ProjectReasonInviteResolveFailed    ProjectEventReason = "invite_resolve_failed"
)

type ProjectEvent struct {
	Name         ProjectEventName
	Action       ProjectEventAction
	Outcome      ProjectEventOutcome
	Reason       ProjectEventReason
	ActorUserID  string
	ProjectID    string
	ProjectRef   string
	ProjectSlug  string
	InviteID     string
	TargetUserID string
	Role         domainproject.Role
	InviteStatus domainproject.InviteStatus
	Err          error
}

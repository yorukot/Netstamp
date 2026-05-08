package project

import (
	"context"

	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type Repository interface {
	CreateProjectWithOwner(ctx context.Context, input domainproject.CreateProjectStorageInput) (domainproject.Project, error)
	ListProjectsForUser(ctx context.Context, userID string) ([]domainproject.Project, error)
	GetProjectForUser(ctx context.Context, projectRef string, userID string) (domainproject.Project, error)
	GetMemberRole(ctx context.Context, projectID string, userID string) (domainproject.Role, error)
	UpdateProject(ctx context.Context, input domainproject.UpdateProjectStorageInput) (domainproject.Project, error)
	SoftDeleteProject(ctx context.Context, projectID string) error
	ListMembers(ctx context.Context, projectID string) ([]domainproject.Member, error)
	GetMember(ctx context.Context, projectID string, userID string) (domainproject.Member, error)
	AddMember(ctx context.Context, input domainproject.AddMemberStorageInput) (domainproject.Member, error)
	UpdateMemberRole(ctx context.Context, input domainproject.UpdateMemberRoleStorageInput) (domainproject.Member, error)
	CountOwners(ctx context.Context, projectID string) (int, error)
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
	ProjectEventAddMemberSuccess        ProjectEventName = "project.member.add.success"
	ProjectEventAddMemberFailure        ProjectEventName = "project.member.add.failure"
	ProjectEventUpdateMemberRoleSuccess ProjectEventName = "project.member.role_update.success"
	ProjectEventUpdateMemberRoleFailure ProjectEventName = "project.member.role_update.failure"
)

type ProjectEventAction string

const (
	ProjectActionCreate           ProjectEventAction = "create"
	ProjectActionList             ProjectEventAction = "list"
	ProjectActionGet              ProjectEventAction = "get"
	ProjectActionUpdate           ProjectEventAction = "update"
	ProjectActionDelete           ProjectEventAction = "delete"
	ProjectActionListMembers      ProjectEventAction = "list_members"
	ProjectActionAddMember        ProjectEventAction = "add_member"
	ProjectActionUpdateMemberRole ProjectEventAction = "update_member_role"
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
	ProjectReasonOwnerCountFailed       ProjectEventReason = "owner_count_failed"
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
	TargetUserID string
	Role         domainproject.Role
	Err          error
}

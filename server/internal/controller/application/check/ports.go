package check

import (
	"context"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type Repository interface {
	ListChecks(ctx context.Context, projectID string) ([]domaincheck.Check, error)
	GetCheck(ctx context.Context, projectID, checkID string) (domaincheck.Check, error)
	CreateCheck(ctx context.Context, input domaincheck.Check, labelIDs []string) (domaincheck.Check, error)
	UpdateCheck(ctx context.Context, input domaincheck.Check, replaceLabels bool, labelIDs []string) (domaincheck.Check, error)
	SoftDeleteCheck(ctx context.Context, projectID, checkID string) error
}

type ProjectAccess interface {
	GetProjectForUser(ctx context.Context, projectRef, userID string) (domainproject.Project, error)
	GetMemberRole(ctx context.Context, projectID, userID string) (domainproject.Role, error)
}

type LabelAccess interface {
	GetActiveLabelsByIDsForProject(ctx context.Context, projectID string, labelIDs []string) ([]domainlabel.Label, error)
}

type AssignmentRefresher interface {
	RefreshProbeCheckAssignmentsForCheck(ctx context.Context, projectID, checkID string) error
	DeleteProbeCheckAssignmentsForCheck(ctx context.Context, projectID, checkID string) error
}

type EventRecorder interface {
	RecordCheckEvent(ctx context.Context, event CheckEvent)
}

type CheckEventName string

const (
	CheckEventListFailure   CheckEventName = "check.list.failure"
	CheckEventGetFailure    CheckEventName = "check.get.failure"
	CheckEventCreateSuccess CheckEventName = "check.create.success"
	CheckEventCreateFailure CheckEventName = "check.create.failure"
	CheckEventUpdateSuccess CheckEventName = "check.update.success"
	CheckEventUpdateFailure CheckEventName = "check.update.failure"
	CheckEventDeleteSuccess CheckEventName = "check.delete.success"
	CheckEventDeleteFailure CheckEventName = "check.delete.failure"
)

type CheckEventAction string

const (
	CheckActionList   CheckEventAction = "list"
	CheckActionGet    CheckEventAction = "get"
	CheckActionCreate CheckEventAction = "create"
	CheckActionUpdate CheckEventAction = "update"
	CheckActionDelete CheckEventAction = "delete"
)

type CheckEventOutcome string

const (
	CheckOutcomeSuccess CheckEventOutcome = "success"
	CheckOutcomeFailure CheckEventOutcome = "failure"
)

type CheckEventReason string

const (
	CheckReasonInvalidInput            CheckEventReason = "invalid_input"
	CheckReasonForbidden               CheckEventReason = "forbidden"
	CheckReasonProjectNotFound         CheckEventReason = "project_not_found"
	CheckReasonUserNotFound            CheckEventReason = "user_not_found"
	CheckReasonCheckNotFound           CheckEventReason = "check_not_found"
	CheckReasonLabelNotFound           CheckEventReason = "label_not_found"
	CheckReasonProjectLookupFailed     CheckEventReason = "project_lookup_failed"
	CheckReasonRoleLookupFailed        CheckEventReason = "role_lookup_failed"
	CheckReasonCheckListFailed         CheckEventReason = "check_list_failed"
	CheckReasonCheckLookupFailed       CheckEventReason = "check_lookup_failed"
	CheckReasonLabelLookupFailed       CheckEventReason = "label_lookup_failed"
	CheckReasonCheckCreateFailed       CheckEventReason = "check_create_failed"
	CheckReasonCheckUpdateFailed       CheckEventReason = "check_update_failed"
	CheckReasonCheckDeleteFailed       CheckEventReason = "check_delete_failed"
	CheckReasonAssignmentRefreshFailed CheckEventReason = "assignment_refresh_failed"
	CheckReasonAssignmentDeleteFailed  CheckEventReason = "assignment_delete_failed"
)

type CheckEvent struct {
	Name        CheckEventName
	Action      CheckEventAction
	Outcome     CheckEventOutcome
	Reason      CheckEventReason
	ActorUserID string
	ProjectID   string
	ProjectRef  string
	ProjectSlug string
	CheckID     string
	Err         error
}

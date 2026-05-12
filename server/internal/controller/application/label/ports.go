package label

import (
	"context"

	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type Repository interface {
	ListLabels(ctx context.Context, projectID string) ([]domainlabel.Label, error)
	GetLabel(ctx context.Context, projectID, labelID string) (domainlabel.Label, error)
	CreateLabel(ctx context.Context, input domainlabel.Label) (domainlabel.Label, error)
	UpdateLabel(ctx context.Context, input domainlabel.Label) (domainlabel.Label, error)
	SoftDeleteLabel(ctx context.Context, projectID, labelID string) error
	GetActiveLabelsByIDsForProject(ctx context.Context, projectID string, labelIDs []string) ([]domainlabel.Label, error)
}

type ProjectAccess interface {
	GetProjectForUser(ctx context.Context, projectRef, userID string) (domainproject.Project, error)
	GetMemberRole(ctx context.Context, projectID, userID string) (domainproject.Role, error)
}

type EventRecorder interface {
	RecordLabelEvent(ctx context.Context, event LabelEvent)
}

type LabelEventName string

const (
	LabelEventListFailure    LabelEventName = "label.list.failure"
	LabelEventCreateSuccess  LabelEventName = "label.create.success"
	LabelEventCreateFailure  LabelEventName = "label.create.failure"
	LabelEventUpdateSuccess  LabelEventName = "label.update.success"
	LabelEventUpdateFailure  LabelEventName = "label.update.failure"
	LabelEventDeleteSuccess  LabelEventName = "label.delete.success"
	LabelEventDeleteFailure  LabelEventName = "label.delete.failure"
	LabelEventResolveFailure LabelEventName = "label.resolve.failure"
)

type LabelEventAction string

const (
	LabelActionList    LabelEventAction = "list"
	LabelActionCreate  LabelEventAction = "create"
	LabelActionUpdate  LabelEventAction = "update"
	LabelActionDelete  LabelEventAction = "delete"
	LabelActionResolve LabelEventAction = "resolve"
)

type LabelEventOutcome string

const (
	LabelOutcomeSuccess LabelEventOutcome = "success"
	LabelOutcomeFailure LabelEventOutcome = "failure"
)

type LabelEventReason string

const (
	LabelReasonInvalidInput        LabelEventReason = "invalid_input"
	LabelReasonForbidden           LabelEventReason = "forbidden"
	LabelReasonProjectNotFound     LabelEventReason = "project_not_found"
	LabelReasonUserNotFound        LabelEventReason = "user_not_found"
	LabelReasonLabelNotFound       LabelEventReason = "label_not_found"
	LabelReasonLabelAlreadyExists  LabelEventReason = "label_already_exists"
	LabelReasonProjectLookupFailed LabelEventReason = "project_lookup_failed"
	LabelReasonRoleLookupFailed    LabelEventReason = "role_lookup_failed"
	LabelReasonLabelListFailed     LabelEventReason = "label_list_failed"
	LabelReasonLabelLookupFailed   LabelEventReason = "label_lookup_failed"
	LabelReasonLabelCreateFailed   LabelEventReason = "label_create_failed"
	LabelReasonLabelUpdateFailed   LabelEventReason = "label_update_failed"
	LabelReasonLabelDeleteFailed   LabelEventReason = "label_delete_failed"
	LabelReasonLabelResolveFailed  LabelEventReason = "label_resolve_failed"
)

type LabelEvent struct {
	Name        LabelEventName
	Action      LabelEventAction
	Outcome     LabelEventOutcome
	Reason      LabelEventReason
	ActorUserID string
	ProjectID   string
	ProjectRef  string
	ProjectSlug string
	LabelID     string
	Err         error
}

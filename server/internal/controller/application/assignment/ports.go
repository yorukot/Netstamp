package assignment

import "context"

type Repository interface {
	RefreshProbeCheckAssignmentsForProbe(ctx context.Context, projectID, probeID string) error
	RefreshProbeCheckAssignmentsForCheck(ctx context.Context, projectID, checkID string) error
	RefreshProbeCheckAssignmentsForLabel(ctx context.Context, projectID, labelID string) error
	DeleteProbeCheckAssignmentsForProbe(ctx context.Context, projectID, probeID string) error
	DeleteProbeCheckAssignmentsForCheck(ctx context.Context, projectID, checkID string) error
}

type EventRecorder interface {
	RecordAssignmentEvent(ctx context.Context, event AssignmentEvent)
}

type AssignmentEventName string

const (
	AssignmentEventRefreshProbeSuccess AssignmentEventName = "assignment.probe.refresh.success"
	AssignmentEventRefreshProbeFailure AssignmentEventName = "assignment.probe.refresh.failure"
	AssignmentEventRefreshCheckSuccess AssignmentEventName = "assignment.check.refresh.success"
	AssignmentEventRefreshCheckFailure AssignmentEventName = "assignment.check.refresh.failure"
	AssignmentEventRefreshLabelSuccess AssignmentEventName = "assignment.label.refresh.success"
	AssignmentEventRefreshLabelFailure AssignmentEventName = "assignment.label.refresh.failure"
	AssignmentEventDeleteProbeSuccess  AssignmentEventName = "assignment.probe.delete.success"
	AssignmentEventDeleteProbeFailure  AssignmentEventName = "assignment.probe.delete.failure"
	AssignmentEventDeleteCheckSuccess  AssignmentEventName = "assignment.check.delete.success"
	AssignmentEventDeleteCheckFailure  AssignmentEventName = "assignment.check.delete.failure"
)

type AssignmentEventAction string

const (
	AssignmentActionRefreshProbe AssignmentEventAction = "probe.refresh"
	AssignmentActionRefreshCheck AssignmentEventAction = "check.refresh"
	AssignmentActionRefreshLabel AssignmentEventAction = "label.refresh"
	AssignmentActionDeleteProbe  AssignmentEventAction = "probe.delete"
	AssignmentActionDeleteCheck  AssignmentEventAction = "check.delete"
)

type AssignmentEventOutcome string

const (
	AssignmentOutcomeSuccess AssignmentEventOutcome = "success"
	AssignmentOutcomeFailure AssignmentEventOutcome = "failure"
)

type AssignmentEventReason string

const (
	AssignmentReasonInvalidInput    AssignmentEventReason = "invalid_input"
	AssignmentReasonProjectNotFound AssignmentEventReason = "project_not_found"
	AssignmentReasonProbeNotFound   AssignmentEventReason = "probe_not_found"
	AssignmentReasonCheckNotFound   AssignmentEventReason = "check_not_found"
	AssignmentReasonLabelNotFound   AssignmentEventReason = "label_not_found"
	AssignmentReasonRefreshFailed   AssignmentEventReason = "refresh_failed"
	AssignmentReasonDeleteFailed    AssignmentEventReason = "delete_failed"
)

type AssignmentEvent struct {
	Name      AssignmentEventName
	Action    AssignmentEventAction
	Outcome   AssignmentEventOutcome
	Reason    AssignmentEventReason
	ProjectID string
	ProbeID   string
	CheckID   string
	LabelID   string
	Err       error
}

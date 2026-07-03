package assignment

import (
	"context"

	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainselector "github.com/yorukot/netstamp/internal/domain/selector"
)

type Repository interface {
	RefreshProbeCheckAssignmentsForProject(ctx context.Context, projectID string) error
	RefreshProbeCheckAssignmentsForProbe(ctx context.Context, projectID, probeID string) error
	RefreshProbeCheckAssignmentsForCheck(ctx context.Context, projectID, checkID string) error
	RefreshProbeCheckAssignmentsForLabel(ctx context.Context, projectID, labelID string) error
	DeleteProbeCheckAssignmentsForProbe(ctx context.Context, projectID, probeID string) error
	DeleteProbeCheckAssignmentsForCheck(ctx context.Context, projectID, checkID string) error
	ListSelectorPreviewProbes(ctx context.Context, projectID string, selector domainselector.Selector) ([]domainprobe.Probe, error)
	ListProjectAssignments(ctx context.Context, input domainassignment.Query) ([]domainassignment.Assignment, error)
}

type ProjectAccess interface {
	GetProjectForUser(ctx context.Context, projectRef, userID string) (domainproject.Project, error)
}

type EventRecorder interface {
	RecordAssignmentEvent(ctx context.Context, event AssignmentEvent)
}

type AssignmentEventName string

const (
	AssignmentEventRefreshProjectSuccess AssignmentEventName = "assignment.project.refresh.success"
	AssignmentEventRefreshProjectFailure AssignmentEventName = "assignment.project.refresh.failure"
	AssignmentEventRefreshProbeSuccess   AssignmentEventName = "assignment.probe.refresh.success"
	AssignmentEventRefreshProbeFailure   AssignmentEventName = "assignment.probe.refresh.failure"
	AssignmentEventRefreshCheckSuccess   AssignmentEventName = "assignment.check.refresh.success"
	AssignmentEventRefreshCheckFailure   AssignmentEventName = "assignment.check.refresh.failure"
	AssignmentEventRefreshLabelSuccess   AssignmentEventName = "assignment.label.refresh.success"
	AssignmentEventRefreshLabelFailure   AssignmentEventName = "assignment.label.refresh.failure"
	AssignmentEventDeleteProbeSuccess    AssignmentEventName = "assignment.probe.delete.success"
	AssignmentEventDeleteProbeFailure    AssignmentEventName = "assignment.probe.delete.failure"
	AssignmentEventDeleteCheckSuccess    AssignmentEventName = "assignment.check.delete.success"
	AssignmentEventDeleteCheckFailure    AssignmentEventName = "assignment.check.delete.failure"
	AssignmentEventPreviewFailure        AssignmentEventName = "assignment.selector_preview.failure"
	AssignmentEventListFailure           AssignmentEventName = "assignment.list.failure"
)

type AssignmentEventAction string

const (
	AssignmentActionRefreshProject AssignmentEventAction = "project.refresh"
	AssignmentActionRefreshProbe   AssignmentEventAction = "probe.refresh"
	AssignmentActionRefreshCheck   AssignmentEventAction = "check.refresh"
	AssignmentActionRefreshLabel   AssignmentEventAction = "label.refresh"
	AssignmentActionDeleteProbe    AssignmentEventAction = "probe.delete"
	AssignmentActionDeleteCheck    AssignmentEventAction = "check.delete"
	AssignmentActionPreview        AssignmentEventAction = "selector_preview"
	AssignmentActionList           AssignmentEventAction = "list"
)

type AssignmentEventOutcome string

const (
	AssignmentOutcomeSuccess AssignmentEventOutcome = "success"
	AssignmentOutcomeFailure AssignmentEventOutcome = "failure"
)

type AssignmentEventReason string

const (
	AssignmentReasonInvalidInput        AssignmentEventReason = "invalid_input"
	AssignmentReasonProjectNotFound     AssignmentEventReason = "project_not_found"
	AssignmentReasonProbeNotFound       AssignmentEventReason = "probe_not_found"
	AssignmentReasonCheckNotFound       AssignmentEventReason = "check_not_found"
	AssignmentReasonLabelNotFound       AssignmentEventReason = "label_not_found"
	AssignmentReasonForbidden           AssignmentEventReason = "forbidden"
	AssignmentReasonProjectLookupFailed AssignmentEventReason = "project_lookup_failed"
	AssignmentReasonRefreshFailed       AssignmentEventReason = "refresh_failed"
	AssignmentReasonDeleteFailed        AssignmentEventReason = "delete_failed"
	AssignmentReasonListFailed          AssignmentEventReason = "list_failed"
	AssignmentReasonPreviewFailed       AssignmentEventReason = "preview_failed"
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

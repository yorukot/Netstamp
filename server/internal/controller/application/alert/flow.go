package alert

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/trace"

	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
	"github.com/yorukot/netstamp/internal/domain/alertcondition"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	"github.com/yorukot/netstamp/internal/domain/identity"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type alertFlow struct {
	service        *Service
	ctx            context.Context
	span           trace.Span
	action         AlertAction
	actorUserID    string
	projectID      string
	projectRef     string
	projectSlug    string
	ruleID         string
	notificationID string
}

func (s *Service) startAlertFlow(ctx context.Context, spanName string, action AlertAction, actorUserID string) (context.Context, *alertFlow) {
	ctx, span := alertTracer.Start(ctx, spanName, trace.WithAttributes(
		attrAlertAction.String(string(action)),
	))
	if actorUserID != "" {
		span.SetAttributes(attrUserID.String(actorUserID))
	}

	return ctx, &alertFlow{
		service:     s,
		ctx:         ctx,
		span:        span,
		action:      action,
		actorUserID: actorUserID,
	}
}

func (f *alertFlow) end() {
	f.span.End()
}

func (f *alertFlow) setProjectRef(projectRef string) {
	f.projectRef = projectRef
	if projectRef != "" {
		f.span.SetAttributes(attrProjectRef.String(projectRef))
	}
}

func (f *alertFlow) setProject(project domainproject.Project) {
	f.projectID = project.ID
	if project.ID != "" {
		f.span.SetAttributes(attrProjectID.String(project.ID))
	}
	f.projectSlug = project.Slug
	if project.Slug != "" {
		f.span.SetAttributes(attrProjectSlug.String(project.Slug))
	}
}

func (f *alertFlow) setRuleID(ruleID string) {
	f.ruleID = ruleID
	if ruleID != "" {
		f.span.SetAttributes(attrAlertRuleID.String(ruleID))
	}
}

func (f *alertFlow) setNotificationID(notificationID string) {
	f.notificationID = notificationID
	if notificationID != "" {
		f.span.SetAttributes(attrAlertNotificationID.String(notificationID))
	}
}

func (f *alertFlow) success() {
	f.span.SetAttributes(attrAlertOutcome.String(string(AlertOutcomeSuccess)))
	f.record(AlertOutcomeSuccess, "", nil)
}

func (f *alertFlow) businessFailure(reason AlertReason, returnErr error) error {
	f.span.SetAttributes(
		attrAlertOutcome.String(string(AlertOutcomeFailure)),
		attrAlertFailureReason.String(string(reason)),
	)
	f.record(AlertOutcomeFailure, reason, nil)
	return returnErr
}

func (f *alertFlow) businessResult(reason AlertReason) {
	f.span.SetAttributes(
		attrAlertOutcome.String(string(AlertOutcomeFailure)),
		attrAlertFailureReason.String(string(reason)),
	)
	f.record(AlertOutcomeFailure, reason, nil)
}

func (f *alertFlow) technicalFailure(reason AlertReason, err error) error {
	f.span.SetAttributes(attrAlertOutcome.String(string(AlertOutcomeFailure)))
	recordSpanError(f.span, err, reason)
	f.record(AlertOutcomeFailure, reason, err)
	return err
}

func (f *alertFlow) projectLookupFailure(err error) error {
	switch {
	case errors.Is(err, domainproject.ErrProjectNotFound), errors.Is(err, domainproject.ErrMemberNotFound):
		return f.businessFailure(AlertReasonProjectNotFound, err)
	case errors.Is(err, identity.ErrUserNotFound):
		return f.businessFailure(AlertReasonUserNotFound, err)
	default:
		return f.technicalFailure(AlertReasonProjectLookupFailed, err)
	}
}

func (f *alertFlow) roleLookupFailure(err error) error {
	switch {
	case errors.Is(err, domainproject.ErrProjectNotFound), errors.Is(err, domainproject.ErrMemberNotFound):
		return f.businessFailure(AlertReasonProjectNotFound, err)
	case errors.Is(err, identity.ErrUserNotFound):
		return f.businessFailure(AlertReasonUserNotFound, err)
	default:
		return f.technicalFailure(AlertReasonRoleLookupFailed, err)
	}
}

func (f *alertFlow) writeFailure(technicalReason AlertReason, err error) error {
	switch {
	case errors.Is(err, ErrInvalidInput),
		errors.Is(err, domainalert.ErrInvalidInput),
		errors.Is(err, domaincheck.ErrInvalidInput),
		errors.Is(err, alertcondition.ErrInvalidCondition):
		return f.businessFailure(AlertReasonInvalidInput, err)
	case errors.Is(err, ErrForbidden):
		return f.businessFailure(AlertReasonForbidden, err)
	case errors.Is(err, domainproject.ErrProjectNotFound), errors.Is(err, domainproject.ErrMemberNotFound):
		return f.businessFailure(AlertReasonProjectNotFound, err)
	case errors.Is(err, identity.ErrUserNotFound):
		return f.businessFailure(AlertReasonUserNotFound, err)
	case errors.Is(err, domainalert.ErrRuleNotFound):
		return f.businessFailure(AlertReasonRuleNotFound, err)
	case errors.Is(err, domainalert.ErrNotificationNotFound):
		return f.businessFailure(AlertReasonNotificationNotFound, err)
	case errors.Is(err, domainalert.ErrIncidentNotFound):
		return f.businessFailure(AlertReasonIncidentNotFound, err)
	default:
		return f.technicalFailure(technicalReason, err)
	}
}

func recordAlertQueryFailure(span trace.Span, reason AlertReason, err error) error {
	span.SetAttributes(
		attrAlertOutcome.String(string(AlertOutcomeFailure)),
		attrAlertFailureReason.String(string(reason)),
	)
	recordSpanError(span, err, reason)
	return err
}

func (f *alertFlow) record(outcome AlertOutcome, reason AlertReason, err error) {
	if f.service.events == nil {
		return
	}
	f.service.events.RecordAlertEvent(f.ctx, AlertEvent{
		Name:           alertEventName(f.action, outcome),
		Action:         f.action,
		Outcome:        outcome,
		Reason:         reason,
		ActorUserID:    f.actorUserID,
		ProjectID:      f.projectID,
		ProjectRef:     f.projectRef,
		ProjectSlug:    f.projectSlug,
		RuleID:         f.ruleID,
		NotificationID: f.notificationID,
		Err:            err,
	})
}

var alertEventNames = map[AlertAction]struct {
	success AlertEventName
	failure AlertEventName
}{
	AlertActionCreateRule: {
		success: AlertEventCreateRuleSuccess,
		failure: AlertEventCreateRuleFailure,
	},
	AlertActionUpdateRule: {
		success: AlertEventUpdateRuleSuccess,
		failure: AlertEventUpdateRuleFailure,
	},
	AlertActionDeleteRule: {
		success: AlertEventDeleteRuleSuccess,
		failure: AlertEventDeleteRuleFailure,
	},
	AlertActionCreateNotification: {
		success: AlertEventCreateNotificationSuccess,
		failure: AlertEventCreateNotificationFailure,
	},
	AlertActionUpdateNotification: {
		success: AlertEventUpdateNotificationSuccess,
		failure: AlertEventUpdateNotificationFailure,
	},
	AlertActionDeleteNotification: {
		success: AlertEventDeleteNotificationSuccess,
		failure: AlertEventDeleteNotificationFailure,
	},
	AlertActionTestNotification: {
		success: AlertEventTestNotificationSuccess,
		failure: AlertEventTestNotificationFailure,
	},
}

func alertEventName(action AlertAction, outcome AlertOutcome) AlertEventName {
	names, ok := alertEventNames[action]
	if !ok {
		return AlertEventName("alert.unknown." + string(outcome))
	}
	if outcome == AlertOutcomeSuccess {
		return names.success
	}
	return names.failure
}

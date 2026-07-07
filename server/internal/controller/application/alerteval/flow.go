package alerteval

import (
	"context"

	"go.opentelemetry.io/otel/trace"

	appproberuntime "github.com/yorukot/netstamp/internal/controller/application/proberuntime"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
)

type alertEvalFlow struct {
	service    *Service
	ctx        context.Context
	span       trace.Span
	action     AlertEvalAction
	projectID  string
	probeID    string
	checkID    string
	checkType  string
	ruleID     string
	incidentID string
}

func (s *Service) startAssignmentFlow(ctx context.Context, input appproberuntime.ChangedAssignmentInput) (context.Context, *alertEvalFlow) {
	ctx, span := alertEvalTracer.Start(ctx, "alert_eval.assignment.evaluate", trace.WithAttributes(
		attrAlertEvalAction.String(string(AlertEvalActionEvaluateAssignment)),
		attrProjectID.String(input.ProjectID),
		attrProbeID.String(input.ProbeID),
		attrCheckID.String(input.CheckID),
		attrCheckType.String(input.CheckType),
	))

	return ctx, &alertEvalFlow{
		service:   s,
		ctx:       ctx,
		span:      span,
		action:    AlertEvalActionEvaluateAssignment,
		projectID: input.ProjectID,
		probeID:   input.ProbeID,
		checkID:   input.CheckID,
		checkType: input.CheckType,
	}
}

func (s *Service) startRuleFlow(ctx context.Context, input appproberuntime.ChangedAssignmentInput, rule domainalert.Rule) (context.Context, *alertEvalFlow) {
	ctx, span := alertEvalTracer.Start(ctx, "alert_eval.rule.evaluate", trace.WithAttributes(
		attrAlertEvalAction.String(string(AlertEvalActionEvaluateRule)),
		attrProjectID.String(input.ProjectID),
		attrProbeID.String(input.ProbeID),
		attrCheckID.String(input.CheckID),
		attrCheckType.String(input.CheckType),
		attrAlertRuleID.String(rule.ID),
	))

	return ctx, &alertEvalFlow{
		service:   s,
		ctx:       ctx,
		span:      span,
		action:    AlertEvalActionEvaluateRule,
		projectID: input.ProjectID,
		probeID:   input.ProbeID,
		checkID:   input.CheckID,
		checkType: input.CheckType,
		ruleID:    rule.ID,
	}
}

func (f *alertEvalFlow) end() {
	f.span.End()
}

func (f *alertEvalFlow) setIncidentID(incidentID string) {
	f.incidentID = incidentID
	if incidentID != "" {
		f.span.SetAttributes(attrAlertIncidentID.String(incidentID))
	}
}

func (f *alertEvalFlow) success() {
	f.span.SetAttributes(attrAlertEvalOutcome.String(string(AlertEvalOutcomeSuccess)))
}

func (f *alertEvalFlow) failure(name AlertEvalEventName, reason AlertEvalReason, err error) error {
	f.recordFailure(name, reason, err)
	return err
}

func (f *alertEvalFlow) recordFailure(name AlertEvalEventName, reason AlertEvalReason, err error) {
	f.span.SetAttributes(attrAlertEvalOutcome.String(string(AlertEvalOutcomeFailure)))
	recordSpanError(f.span, err, reason)
	if f.service.events == nil {
		return
	}
	f.service.events.RecordAlertEvalEvent(f.ctx, AlertEvalEvent{
		Name:       name,
		Action:     f.action,
		Outcome:    AlertEvalOutcomeFailure,
		Reason:     reason,
		ProjectID:  f.projectID,
		ProbeID:    f.probeID,
		CheckID:    f.checkID,
		CheckType:  f.checkType,
		RuleID:     f.ruleID,
		IncidentID: f.incidentID,
		Err:        err,
	})
}

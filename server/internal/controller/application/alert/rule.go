package alert

import (
	"context"

	"go.opentelemetry.io/otel/trace"

	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func (s *Service) ListRules(ctx context.Context, input ListRulesInput) ([]domainalert.Rule, error) {
	ctx, span := alertTracer.Start(ctx, "alert.rule.list", trace.WithAttributes(
		attrAlertAction.String(string(AlertActionListRules)),
		attrProjectRef.String(input.ProjectRef),
	))
	defer span.End()

	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return nil, recordAlertQueryFailure(span, AlertReasonProjectLookupFailed, err)
	}
	span.SetAttributes(attrProjectID.String(project.ID))
	rules, err := s.repo.ListRules(ctx, project.ID, input.Status, input.CheckType)
	if err != nil {
		return nil, recordAlertQueryFailure(span, AlertReasonRuleListFailed, err)
	}
	span.SetAttributes(attrAlertOutcome.String(string(AlertOutcomeSuccess)))
	return rules, nil
}

func (s *Service) GetRule(ctx context.Context, input GetRuleInput) (domainalert.Rule, error) {
	ctx, span := alertTracer.Start(ctx, "alert.rule.get", trace.WithAttributes(
		attrAlertAction.String(string(AlertActionGetRule)),
		attrProjectRef.String(input.ProjectRef),
		attrAlertRuleID.String(input.RuleID),
	))
	defer span.End()

	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainalert.Rule{}, recordAlertQueryFailure(span, AlertReasonProjectLookupFailed, err)
	}
	span.SetAttributes(attrProjectID.String(project.ID))
	rule, err := s.repo.GetRule(ctx, project.ID, input.RuleID)
	if err != nil {
		return domainalert.Rule{}, recordAlertQueryFailure(span, AlertReasonRuleLookupFailed, err)
	}
	span.SetAttributes(attrAlertOutcome.String(string(AlertOutcomeSuccess)))
	return rule, nil
}

func (s *Service) CreateRule(ctx context.Context, input CreateRuleInput) (domainalert.Rule, error) {
	ctx, flow := s.startAlertFlow(ctx, "alert.rule.create", AlertActionCreateRule, input.CurrentUserID)
	defer flow.end()
	flow.setProjectRef(input.ProjectRef)

	project, err := s.loadProjectForFlow(ctx, flow, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainalert.Rule{}, err
	}
	if actionErr := s.requireProjectActionForFlow(ctx, flow, project.ID, input.CurrentUserID, domainproject.ActionManageAlerts); actionErr != nil {
		return domainalert.Rule{}, actionErr
	}
	rule, err := normalizeCreateRule(project.ID, input)
	if err != nil {
		return domainalert.Rule{}, flow.writeFailure(AlertReasonRuleCreateFailed, err)
	}
	created, err := s.repo.CreateRule(ctx, rule)
	if err != nil {
		return domainalert.Rule{}, flow.writeFailure(AlertReasonRuleCreateFailed, err)
	}
	flow.setRuleID(created.ID)
	flow.success()
	return created, nil
}

func (s *Service) UpdateRule(ctx context.Context, input UpdateRuleInput) (domainalert.Rule, error) {
	ctx, flow := s.startAlertFlow(ctx, "alert.rule.update", AlertActionUpdateRule, input.CurrentUserID)
	defer flow.end()
	flow.setProjectRef(input.ProjectRef)
	flow.setRuleID(input.RuleID)

	project, err := s.loadProjectForFlow(ctx, flow, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainalert.Rule{}, err
	}
	if actionErr := s.requireProjectActionForFlow(ctx, flow, project.ID, input.CurrentUserID, domainproject.ActionManageAlerts); actionErr != nil {
		return domainalert.Rule{}, actionErr
	}
	rule, err := normalizeUpdateRule(project.ID, input)
	if err != nil {
		return domainalert.Rule{}, flow.writeFailure(AlertReasonRuleUpdateFailed, err)
	}
	updated, err := s.repo.UpdateRule(ctx, rule)
	if err != nil {
		return domainalert.Rule{}, flow.writeFailure(AlertReasonRuleUpdateFailed, err)
	}
	flow.setRuleID(updated.ID)
	flow.success()
	return updated, nil
}

func (s *Service) DeleteRule(ctx context.Context, input DeleteRuleInput) error {
	ctx, flow := s.startAlertFlow(ctx, "alert.rule.delete", AlertActionDeleteRule, input.CurrentUserID)
	defer flow.end()
	flow.setProjectRef(input.ProjectRef)
	flow.setRuleID(input.RuleID)

	project, err := s.loadProjectForFlow(ctx, flow, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return err
	}
	if actionErr := s.requireProjectActionForFlow(ctx, flow, project.ID, input.CurrentUserID, domainproject.ActionManageAlerts); actionErr != nil {
		return actionErr
	}
	if err := s.repo.DeleteRule(ctx, project.ID, input.RuleID); err != nil {
		return flow.writeFailure(AlertReasonRuleDeleteFailed, err)
	}
	flow.success()
	return nil
}

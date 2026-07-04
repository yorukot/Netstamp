package alert

import (
	"context"

	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func (s *Service) ListRules(ctx context.Context, input ListRulesInput) ([]domainalert.Rule, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListRules(ctx, project.ID, input.Status, input.CheckType)
}

func (s *Service) GetRule(ctx context.Context, input GetRuleInput) (domainalert.Rule, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainalert.Rule{}, err
	}
	return s.repo.GetRule(ctx, project.ID, input.RuleID)
}

func (s *Service) CreateRule(ctx context.Context, input CreateRuleInput) (domainalert.Rule, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainalert.Rule{}, err
	}
	if actionErr := s.requireProjectAction(ctx, project.ID, input.CurrentUserID, domainproject.ActionManageAlerts); actionErr != nil {
		return domainalert.Rule{}, actionErr
	}
	rule, err := normalizeCreateRule(project.ID, input)
	if err != nil {
		return domainalert.Rule{}, err
	}
	return s.repo.CreateRule(ctx, rule)
}

func (s *Service) UpdateRule(ctx context.Context, input UpdateRuleInput) (domainalert.Rule, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainalert.Rule{}, err
	}
	if actionErr := s.requireProjectAction(ctx, project.ID, input.CurrentUserID, domainproject.ActionManageAlerts); actionErr != nil {
		return domainalert.Rule{}, actionErr
	}
	rule, err := normalizeUpdateRule(project.ID, input)
	if err != nil {
		return domainalert.Rule{}, err
	}
	return s.repo.UpdateRule(ctx, rule)
}

func (s *Service) DeleteRule(ctx context.Context, input DeleteRuleInput) error {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return err
	}
	if actionErr := s.requireProjectAction(ctx, project.ID, input.CurrentUserID, domainproject.ActionManageAlerts); actionErr != nil {
		return actionErr
	}
	return s.repo.DeleteRule(ctx, project.ID, input.RuleID)
}

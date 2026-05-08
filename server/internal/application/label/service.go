package label

import (
	"context"
	"errors"

	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	"github.com/yorukot/netstamp/internal/normalize"
	"go.opentelemetry.io/otel/trace"
)

type Service struct {
	repo          Repository
	projectAccess ProjectAccess
}

func NewService(repo Repository, projectAccess ProjectAccess) *Service {
	return &Service{
		repo:          repo,
		projectAccess: projectAccess,
	}
}

func (s *Service) ListLabels(ctx context.Context, input ListLabelsInput) ([]domainlabel.Label, error) {
	ctx, span := s.startLabelSpan(ctx, "label.list", "list", input.CurrentUserID, input.ProjectRef, "")
	defer span.End()

	project, err := s.loadProject(ctx, span, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return nil, err
	}

	labels, err := s.repo.ListLabels(ctx, project.ID)
	if err != nil {
		recordSpanError(span, err, "label_list_failed")
		return nil, err
	}

	return labels, nil
}

func (s *Service) CreateLabel(ctx context.Context, input CreateLabelInput) (domainlabel.Label, error) {
	ctx, span := s.startLabelSpan(ctx, "label.create", "create", input.CurrentUserID, input.ProjectRef, "")
	defer span.End()

	project, err := s.loadProject(ctx, span, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainlabel.Label{}, err
	}
	if err := s.requireManager(ctx, span, project.ID, input.CurrentUserID); err != nil {
		return domainlabel.Label{}, err
	}

	key, value, err := normalizeLabelKeyValue(input.Key, input.Value)
	if err != nil {
		return domainlabel.Label{}, err
	}

	label, err := s.repo.CreateLabel(ctx, domainlabel.CreateLabelStorageInput{
		ProjectID: project.ID,
		Key:       key,
		Value:     value,
	})
	if err != nil {
		recordLabelMutationError(span, err, "label_create_failed")
		return domainlabel.Label{}, err
	}

	return label, nil
}

func (s *Service) UpdateLabel(ctx context.Context, input UpdateLabelInput) (domainlabel.Label, error) {
	ctx, span := s.startLabelSpan(ctx, "label.update", "update", input.CurrentUserID, input.ProjectRef, input.LabelID)
	defer span.End()

	project, err := s.loadProject(ctx, span, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainlabel.Label{}, err
	}
	if err := s.requireManager(ctx, span, project.ID, input.CurrentUserID); err != nil {
		return domainlabel.Label{}, err
	}
	if input.Key == nil && input.Value == nil {
		return domainlabel.Label{}, ErrInvalidInput
	}

	current, err := s.repo.GetLabel(ctx, project.ID, input.LabelID)
	if err != nil {
		recordLabelMutationError(span, err, "label_lookup_failed")
		return domainlabel.Label{}, err
	}

	key := current.Key
	if input.Key != nil {
		key, err = normalize.RequiredString(*input.Key, ErrInvalidInput)
		if err != nil {
			return domainlabel.Label{}, err
		}
	}
	value := current.Value
	if input.Value != nil {
		value, err = normalize.RequiredString(*input.Value, ErrInvalidInput)
		if err != nil {
			return domainlabel.Label{}, err
		}
	}

	label, err := s.repo.UpdateLabel(ctx, domainlabel.UpdateLabelStorageInput{
		ProjectID: project.ID,
		LabelID:   input.LabelID,
		Key:       key,
		Value:     value,
	})
	if err != nil {
		recordLabelMutationError(span, err, "label_update_failed")
		return domainlabel.Label{}, err
	}

	return label, nil
}

func (s *Service) DeleteLabel(ctx context.Context, input DeleteLabelInput) error {
	ctx, span := s.startLabelSpan(ctx, "label.delete", "delete", input.CurrentUserID, input.ProjectRef, input.LabelID)
	defer span.End()

	project, err := s.loadProject(ctx, span, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return err
	}
	if err := s.requireManager(ctx, span, project.ID, input.CurrentUserID); err != nil {
		return err
	}

	if err := s.repo.SoftDeleteLabel(ctx, project.ID, input.LabelID); err != nil {
		recordLabelMutationError(span, err, "label_delete_failed")
		return err
	}

	return nil
}

func (s *Service) GetActiveLabelsByIDsForProject(ctx context.Context, projectID string, labelIDs []string) ([]domainlabel.Label, error) {
	ctx, span := labelTracer.Start(ctx, "label.resolve", trace.WithAttributes(
		attrLabelAction.String("resolve"),
		attrProjectID.String(projectID),
	))
	defer span.End()

	labels, err := s.repo.GetActiveLabelsByIDsForProject(ctx, projectID, labelIDs)
	if err != nil {
		recordLabelMutationError(span, err, "label_resolve_failed")
		return nil, err
	}

	return labels, nil
}

func (s *Service) startLabelSpan(ctx context.Context, spanName string, action string, userID string, projectRef string, labelID string) (context.Context, trace.Span) {
	attrs := []trace.SpanStartOption{trace.WithAttributes(attrLabelAction.String(action))}
	ctx, span := labelTracer.Start(ctx, spanName, attrs...)
	if userID != "" {
		span.SetAttributes(attrUserID.String(userID))
	}
	if projectRef != "" {
		span.SetAttributes(attrProjectRef.String(projectRef))
	}
	if labelID != "" {
		span.SetAttributes(attrLabelID.String(labelID))
	}

	return ctx, span
}

func (s *Service) loadProject(ctx context.Context, span trace.Span, projectRef string, userID string) (domainproject.Project, error) {
	project, err := s.projectAccess.GetProjectForUser(ctx, projectRef, userID)
	if err != nil {
		if !errors.Is(err, ErrProjectNotFound) && !errors.Is(err, ErrUserNotFound) {
			recordSpanError(span, err, "project_lookup_failed")
		}
		return domainproject.Project{}, err
	}
	if project.ID != "" {
		span.SetAttributes(attrProjectID.String(project.ID))
	}

	return project, nil
}

func (s *Service) requireManager(ctx context.Context, span trace.Span, projectID string, userID string) error {
	role, err := s.projectAccess.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		if !errors.Is(err, ErrProjectNotFound) && !errors.Is(err, ErrUserNotFound) {
			recordSpanError(span, err, "role_lookup_failed")
		}
		return err
	}
	if !canManageLabels(role) {
		return ErrForbidden
	}

	return nil
}

func normalizeLabelKeyValue(keyValue string, labelValue string) (string, string, error) {
	key, err := normalize.RequiredString(keyValue, ErrInvalidInput)
	if err != nil {
		return "", "", err
	}
	value, err := normalize.RequiredString(labelValue, ErrInvalidInput)
	if err != nil {
		return "", "", err
	}

	return key, value, nil
}

func canManageLabels(role domainproject.Role) bool {
	return role == domainproject.RoleOwner || role == domainproject.RoleAdmin || role == domainproject.RoleEditor
}

func recordLabelMutationError(span trace.Span, err error, fallbackReason string) {
	switch {
	case errors.Is(err, ErrLabelNotFound),
		errors.Is(err, ErrLabelAlreadyExists),
		errors.Is(err, ErrInvalidInput),
		errors.Is(err, ErrProjectNotFound):
		return
	default:
		recordSpanError(span, err, fallbackReason)
	}
}

package project

import (
	"context"

	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	"go.opentelemetry.io/otel/trace"
)

type projectFlow struct {
	service      *Service
	ctx          context.Context
	span         trace.Span
	action       ProjectEventAction
	actorUserID  string
	projectID    string
	projectRef   string
	projectSlug  string
	targetUserID string
	role         domainproject.Role
}

func (s *Service) startProjectFlow(ctx context.Context, spanName string, action ProjectEventAction, actorUserID string) (context.Context, *projectFlow) {
	ctx, span := projectTracer.Start(ctx, spanName, trace.WithAttributes(
		attrProjectAction.String(string(action)),
	))
	if actorUserID != "" {
		span.SetAttributes(attrUserID.String(actorUserID))
	}

	return ctx, &projectFlow{
		service:     s,
		ctx:         ctx,
		span:        span,
		action:      action,
		actorUserID: actorUserID,
	}
}

func (f *projectFlow) end() {
	f.span.End()
}

func (f *projectFlow) setProjectRef(projectRef string) {
	f.projectRef = projectRef
	if projectRef != "" {
		f.span.SetAttributes(attrProjectRef.String(projectRef))
	}
}

func (f *projectFlow) setProject(project domainproject.Project) {
	f.projectID = project.ID
	if project.ID != "" {
		f.span.SetAttributes(attrProjectID.String(project.ID))
	}
	f.setProjectSlug(project.Slug)
}

func (f *projectFlow) setProjectSlug(slug string) {
	f.projectSlug = slug
	if slug != "" {
		f.span.SetAttributes(attrProjectSlug.String(slug))
	}
}

func (f *projectFlow) setTargetUser(userID string) {
	f.targetUserID = userID
	if userID != "" {
		f.span.SetAttributes(attrProjectMemberUserID.String(userID))
	}
}

func (f *projectFlow) setRole(role domainproject.Role) {
	f.role = role
	if role != "" {
		f.span.SetAttributes(attrProjectMemberRole.String(string(role)))
	}
}

func (f *projectFlow) success(name ProjectEventName) {
	f.span.SetAttributes(attrProjectOutcome.String(string(ProjectOutcomeSuccess)))
	f.service.events.RecordProjectEvent(f.ctx, f.projectEvent(name, ProjectOutcomeSuccess, "", nil))
}

func (f *projectFlow) businessFailure(name ProjectEventName, reason ProjectEventReason, returnErr error) error {
	f.span.SetAttributes(
		attrProjectOutcome.String(string(ProjectOutcomeFailure)),
		attrProjectFailureReason.String(string(reason)),
	)
	f.service.events.RecordProjectEvent(f.ctx, f.projectEvent(name, ProjectOutcomeFailure, reason, nil))
	return returnErr
}

func (f *projectFlow) technicalFailure(name ProjectEventName, reason ProjectEventReason, err error) error {
	f.span.SetAttributes(attrProjectOutcome.String(string(ProjectOutcomeFailure)))
	recordSpanError(f.span, err, reason)
	f.service.events.RecordProjectEvent(f.ctx, f.projectEvent(name, ProjectOutcomeFailure, reason, err))
	return err
}

func (f *projectFlow) projectEvent(name ProjectEventName, outcome ProjectEventOutcome, reason ProjectEventReason, err error) ProjectEvent {
	return ProjectEvent{
		Name:         name,
		Action:       f.action,
		Outcome:      outcome,
		Reason:       reason,
		ActorUserID:  f.actorUserID,
		ProjectID:    f.projectID,
		ProjectRef:   f.projectRef,
		ProjectSlug:  f.projectSlug,
		TargetUserID: f.targetUserID,
		Role:         f.role,
		Err:          err,
	}
}

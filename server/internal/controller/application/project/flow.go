package project

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/trace"

	"github.com/yorukot/netstamp/internal/domain/identity"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
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
	inviteID     string
	targetUserID string
	role         domainproject.Role
	inviteStatus domainproject.InviteStatus
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

func (f *projectFlow) setInvite(invite domainproject.Invite) {
	f.setInviteID(invite.ID)
	f.setProjectID(invite.ProjectID)
	f.setTargetUser(invite.InvitedUserID)
	f.setRole(invite.Role)
	f.setInviteStatus(invite.Status)
}

func (f *projectFlow) setProjectID(projectID string) {
	f.projectID = projectID
	if projectID != "" {
		f.span.SetAttributes(attrProjectID.String(projectID))
	}
}

func (f *projectFlow) setInviteID(inviteID string) {
	f.inviteID = inviteID
	if inviteID != "" {
		f.span.SetAttributes(attrProjectInviteID.String(inviteID))
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

func (f *projectFlow) setInviteStatus(status domainproject.InviteStatus) {
	f.inviteStatus = status
	if status != "" {
		f.span.SetAttributes(attrProjectInviteStatus.String(string(status)))
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

func (f *projectFlow) projectCreateFailure(err error) error {
	switch {
	case errors.Is(err, domainproject.ErrProjectSlugAlreadyExists):
		return f.businessFailure(ProjectEventCreateFailure, ProjectReasonSlugAlreadyExists, err)
	case errors.Is(err, identity.ErrUserNotFound):
		return f.businessFailure(ProjectEventCreateFailure, ProjectReasonUserNotFound, err)
	default:
		return f.technicalFailure(ProjectEventCreateFailure, ProjectReasonProjectCreateFailed, err)
	}
}

func (f *projectFlow) projectListFailure(err error) error {
	if errors.Is(err, domainproject.ErrProjectNotFound) || errors.Is(err, identity.ErrUserNotFound) {
		return err
	}

	return f.technicalFailure(ProjectEventListFailure, ProjectReasonProjectListFailed, err)
}

func (f *projectFlow) projectLookupFailure(event ProjectEventName, err error) error {
	switch {
	case errors.Is(err, domainproject.ErrProjectNotFound):
		return f.businessFailure(event, ProjectReasonProjectNotFound, err)
	case errors.Is(err, identity.ErrUserNotFound):
		return f.businessFailure(event, ProjectReasonUserNotFound, err)
	default:
		return f.technicalFailure(event, ProjectReasonProjectLookupFailed, err)
	}
}

func (f *projectFlow) projectReadLookupFailure(event ProjectEventName, err error) error {
	if errors.Is(err, domainproject.ErrProjectNotFound) || errors.Is(err, identity.ErrUserNotFound) {
		return err
	}

	return f.technicalFailure(event, ProjectReasonProjectLookupFailed, err)
}

func (f *projectFlow) roleLookupFailure(event ProjectEventName, err error) error {
	switch {
	case errors.Is(err, domainproject.ErrProjectNotFound):
		return f.businessFailure(event, ProjectReasonProjectNotFound, err)
	case errors.Is(err, identity.ErrUserNotFound):
		return f.businessFailure(event, ProjectReasonUserNotFound, err)
	case errors.Is(err, domainproject.ErrMemberNotFound):
		// User is not a member of the project; treat as forbidden to avoid leaking project existence.
		return f.businessFailure(event, ProjectReasonForbidden, ErrForbidden)
	default:
		return f.technicalFailure(event, ProjectReasonRoleLookupFailed, err)
	}
}

func (f *projectFlow) projectUpdateFailure(err error) error {
	switch {
	case errors.Is(err, domainproject.ErrProjectNotFound):
		return f.businessFailure(ProjectEventUpdateFailure, ProjectReasonProjectNotFound, err)
	case errors.Is(err, domainproject.ErrProjectSlugAlreadyExists):
		return f.businessFailure(ProjectEventUpdateFailure, ProjectReasonSlugAlreadyExists, err)
	default:
		return f.technicalFailure(ProjectEventUpdateFailure, ProjectReasonProjectUpdateFailed, err)
	}
}

func (f *projectFlow) projectDeleteFailure(err error) error {
	if errors.Is(err, domainproject.ErrProjectNotFound) {
		return f.businessFailure(ProjectEventDeleteFailure, ProjectReasonProjectNotFound, err)
	}

	return f.technicalFailure(ProjectEventDeleteFailure, ProjectReasonProjectDeleteFailed, err)
}

func (f *projectFlow) membersListFailure(err error) error {
	if errors.Is(err, domainproject.ErrProjectNotFound) {
		return err
	}

	return f.technicalFailure(ProjectEventListMembersFailure, ProjectReasonMembersListFailed, err)
}

func (f *projectFlow) invitesListFailure(event ProjectEventName, err error) error {
	if errors.Is(err, domainproject.ErrProjectNotFound) || errors.Is(err, identity.ErrUserNotFound) {
		return err
	}

	return f.technicalFailure(event, ProjectReasonInvitesListFailed, err)
}

func (f *projectFlow) inviteCreateFailure(err error) error {
	switch {
	case errors.Is(err, domainproject.ErrInviteAlreadyExists):
		return f.businessFailure(ProjectEventCreateInviteFailure, ProjectReasonInviteAlreadyExists, err)
	case errors.Is(err, domainproject.ErrMemberAlreadyExists):
		return f.businessFailure(ProjectEventCreateInviteFailure, ProjectReasonMemberAlreadyExists, err)
	case errors.Is(err, domainproject.ErrProjectNotFound):
		return f.businessFailure(ProjectEventCreateInviteFailure, ProjectReasonProjectNotFound, err)
	case errors.Is(err, identity.ErrUserNotFound):
		return f.businessFailure(ProjectEventCreateInviteFailure, ProjectReasonUserNotFound, err)
	default:
		return f.technicalFailure(ProjectEventCreateInviteFailure, ProjectReasonInviteCreateFailed, err)
	}
}

func (f *projectFlow) inviteResolveFailure(event ProjectEventName, err error) error {
	switch {
	case errors.Is(err, domainproject.ErrInviteNotFound):
		return f.businessFailure(event, ProjectReasonInviteNotFound, err)
	case errors.Is(err, domainproject.ErrMemberAlreadyExists):
		return f.businessFailure(event, ProjectReasonMemberAlreadyExists, err)
	case errors.Is(err, domainproject.ErrProjectNotFound):
		return f.businessFailure(event, ProjectReasonProjectNotFound, err)
	case errors.Is(err, identity.ErrUserNotFound):
		return f.businessFailure(event, ProjectReasonUserNotFound, err)
	default:
		return f.technicalFailure(event, ProjectReasonInviteResolveFailed, err)
	}
}

func (f *projectFlow) inviteCancelFailure(err error) error {
	switch {
	case errors.Is(err, domainproject.ErrInviteNotFound):
		return f.businessFailure(ProjectEventCancelInviteFailure, ProjectReasonInviteNotFound, err)
	case errors.Is(err, domainproject.ErrProjectNotFound):
		return f.businessFailure(ProjectEventCancelInviteFailure, ProjectReasonProjectNotFound, err)
	case errors.Is(err, identity.ErrUserNotFound):
		return f.businessFailure(ProjectEventCancelInviteFailure, ProjectReasonUserNotFound, err)
	default:
		return f.technicalFailure(ProjectEventCancelInviteFailure, ProjectReasonInviteCancelFailed, err)
	}
}

func (f *projectFlow) memberLookupFailure(event ProjectEventName, err error) error {
	if errors.Is(err, domainproject.ErrMemberNotFound) {
		return f.businessFailure(event, ProjectReasonMemberNotFound, err)
	}

	return f.technicalFailure(event, ProjectReasonMemberLookupFailed, err)
}

func (f *projectFlow) ownerCountFailure(event ProjectEventName, err error) error {
	if errors.Is(err, domainproject.ErrProjectNotFound) {
		return f.businessFailure(event, ProjectReasonProjectNotFound, err)
	}

	return f.technicalFailure(event, ProjectReasonOwnerCountFailed, err)
}

func (f *projectFlow) memberRoleUpdateFailure(err error) error {
	if errors.Is(err, domainproject.ErrMemberNotFound) {
		return f.businessFailure(ProjectEventUpdateMemberRoleFailure, ProjectReasonMemberNotFound, err)
	}

	return f.technicalFailure(ProjectEventUpdateMemberRoleFailure, ProjectReasonMemberRoleUpdateFailed, err)
}

func (f *projectFlow) memberRemoveFailure(err error) error {
	if errors.Is(err, domainproject.ErrMemberNotFound) {
		return f.businessFailure(ProjectEventRemoveMemberFailure, ProjectReasonMemberNotFound, err)
	}

	return f.technicalFailure(ProjectEventRemoveMemberFailure, ProjectReasonMemberRemoveFailed, err)
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
		InviteID:     f.inviteID,
		TargetUserID: f.targetUserID,
		Role:         f.role,
		InviteStatus: f.inviteStatus,
		Err:          err,
	}
}

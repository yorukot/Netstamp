package account

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/trace"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

type userFlow struct {
	service *Service
	ctx     context.Context
	span    trace.Span
	action  UserEventAction
	userID  string
	email   string
}

func (s *Service) startUserFlow(ctx context.Context, spanName string, action UserEventAction, userID string) (context.Context, *userFlow) {
	ctx, span := userTracer.Start(ctx, spanName, trace.WithAttributes(
		attrUserAction.String(string(action)),
	))
	if userID != "" {
		span.SetAttributes(attrUserID.String(userID))
	}

	return ctx, &userFlow{
		service: s,
		ctx:     ctx,
		span:    span,
		action:  action,
		userID:  userID,
	}
}

func (f *userFlow) end() {
	f.span.End()
}

func (f *userFlow) setUser(user identity.User) {
	f.userID = user.ID
	f.email = user.Email
	if user.ID != "" {
		f.span.SetAttributes(attrUserID.String(user.ID))
	}
}

func (f *userFlow) success(name UserEventName) {
	f.span.SetAttributes(attrUserOutcome.String(string(UserOutcomeSuccess)))
	f.record(name, UserOutcomeSuccess, "", nil)
}

func (f *userFlow) businessFailure(name UserEventName, reason UserEventReason, returnErr error) error {
	f.span.SetAttributes(
		attrUserOutcome.String(string(UserOutcomeFailure)),
		attrUserFailureReason.String(string(reason)),
	)
	f.record(name, UserOutcomeFailure, reason, nil)
	return returnErr
}

func (f *userFlow) technicalFailure(name UserEventName, reason UserEventReason, err error) error {
	f.span.SetAttributes(attrUserOutcome.String(string(UserOutcomeFailure)))
	recordSpanError(f.span, err, reason)
	f.record(name, UserOutcomeFailure, reason, err)
	return err
}

func (f *userFlow) lookupFailure(event UserEventName, err error) error {
	if errors.Is(err, identity.ErrUserNotFound) {
		return f.businessFailure(event, UserReasonUserNotFound, err)
	}

	return f.technicalFailure(event, UserReasonUserLookupFailed, err)
}

func (f *userFlow) updateFailure(event UserEventName, err error) error {
	switch {
	case errors.Is(err, identity.ErrUserNotFound):
		return f.businessFailure(event, UserReasonUserNotFound, err)
	case errors.Is(err, identity.ErrEmailAlreadyExists):
		return f.businessFailure(event, UserReasonEmailAlreadyExists, err)
	default:
		return f.technicalFailure(event, UserReasonUserUpdateFailed, err)
	}
}

func (f *userFlow) record(name UserEventName, outcome UserEventOutcome, reason UserEventReason, err error) {
	if f.service.events == nil {
		return
	}
	f.service.events.RecordUserEvent(f.ctx, UserEvent{
		Name:    name,
		Action:  f.action,
		Outcome: outcome,
		Reason:  reason,
		UserID:  f.userID,
		Email:   f.email,
		Err:     err,
	})
}

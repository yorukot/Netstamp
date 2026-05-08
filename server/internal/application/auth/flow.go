package auth

import (
	"context"

	"go.opentelemetry.io/otel/trace"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

type authFlow struct {
	service *Service
	ctx     context.Context
	span    trace.Span
	action  AuthEventAction
	email   string
	userID  string
}

func (s *Service) startAuthFlow(ctx context.Context, spanName string, action AuthEventAction, email string) (context.Context, *authFlow) {
	ctx, span := authTracer.Start(ctx, spanName, trace.WithAttributes(
		attrAuthAction.String(string(action)),
	))

	return ctx, &authFlow{
		service: s,
		ctx:     ctx,
		span:    span,
		action:  action,
		email:   email,
	}
}

func (f *authFlow) end() {
	f.span.End()
}

func (f *authFlow) setUser(user identity.User) {
	f.userID = user.ID
	f.email = user.Email
	f.span.SetAttributes(attrUserID.String(user.ID))
}

func (f *authFlow) success(name AuthEventName) {
	f.span.SetAttributes(attrAuthOutcome.String(string(AuthOutcomeSuccess)))
	f.service.events.RecordAuthEvent(f.ctx, f.authEvent(name, AuthOutcomeSuccess, "", nil))
}

func (f *authFlow) businessFailure(name AuthEventName, reason AuthEventReason, returnErr error) error {
	f.span.SetAttributes(
		attrAuthOutcome.String(string(AuthOutcomeFailure)),
		attrAuthFailureReason.String(string(reason)),
	)
	f.service.events.RecordAuthEvent(f.ctx, f.authEvent(name, AuthOutcomeFailure, reason, nil))
	return returnErr
}

func (f *authFlow) technicalFailure(name AuthEventName, reason AuthEventReason, err error) error {
	f.span.SetAttributes(attrAuthOutcome.String(string(AuthOutcomeFailure)))
	markSpanTechnicalFailure(f.span, reason)
	f.service.events.RecordAuthEvent(f.ctx, f.authEvent(name, AuthOutcomeFailure, reason, err))
	return err
}

func (f *authFlow) authEvent(name AuthEventName, outcome AuthEventOutcome, reason AuthEventReason, err error) AuthEvent {
	return AuthEvent{
		Name:    name,
		Action:  f.action,
		Outcome: outcome,
		Reason:  reason,
		UserID:  f.userID,
		Email:   f.email,
		Err:     err,
	}
}

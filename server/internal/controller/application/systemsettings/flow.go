package systemsettings

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/trace"
)

// settingsFlow centralizes the authorization shared by every settings command.
type settingsFlow struct {
	service  *Service
	resource Resource
	actor    string
	span     trace.Span
}

func (s *Service) startSettingsFlow(
	ctx context.Context,
	action string,
	resource Resource,
	actor string,
) (context.Context, settingsFlow) {
	ctx, span := systemSettingsTracer.Start(ctx, "systemsettings."+action, trace.WithAttributes(
		attrSettingsAction.String(action),
		attrSettingsResource.String(string(resource)),
	))
	if actor != "" {
		span.SetAttributes(attrSettingsActorID.String(actor))
	}
	return ctx, settingsFlow{
		service: s, resource: resource, actor: actor, span: span,
	}
}

func (f settingsFlow) end() {
	f.span.End()
}

func (f settingsFlow) authorize(ctx context.Context) error {
	if f.service == nil || f.service.repo == nil {
		return errors.New("system settings repository is unavailable")
	}
	return f.service.requireSystemAdmin(ctx, f.actor)
}

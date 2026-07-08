package alert

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func getAlertResource[T any](
	ctx context.Context,
	service *Service,
	spanName string,
	action AlertAction,
	projectRef string,
	currentUserID string,
	resourceAttribute attribute.KeyValue,
	lookupFailure AlertReason,
	lookup func(context.Context, string) (T, error),
) (T, error) {
	ctx, span := alertTracer.Start(ctx, spanName, trace.WithAttributes(
		attrAlertAction.String(string(action)),
		attrProjectRef.String(projectRef),
		resourceAttribute,
	))
	defer span.End()

	var zero T
	project, err := service.loadProject(ctx, projectRef, currentUserID)
	if err != nil {
		return zero, recordAlertQueryFailure(span, AlertReasonProjectLookupFailed, err)
	}
	span.SetAttributes(attrProjectID.String(project.ID))
	resource, err := lookup(ctx, project.ID)
	if err != nil {
		return zero, recordAlertQueryFailure(span, lookupFailure, err)
	}
	span.SetAttributes(attrAlertOutcome.String(string(AlertOutcomeSuccess)))
	return resource, nil
}

func getAlertList[T any](
	ctx context.Context,
	service *Service,
	spanName string,
	action AlertAction,
	projectRef string,
	currentUserID string,
	listFailure AlertReason,
	list func(context.Context, string) ([]T, error),
) ([]T, error) {
	ctx, span := alertTracer.Start(ctx, spanName, trace.WithAttributes(
		attrAlertAction.String(string(action)),
		attrProjectRef.String(projectRef),
	))
	defer span.End()

	project, err := service.loadProject(ctx, projectRef, currentUserID)
	if err != nil {
		return nil, recordAlertQueryFailure(span, AlertReasonProjectLookupFailed, err)
	}
	span.SetAttributes(attrProjectID.String(project.ID))
	resources, err := list(ctx, project.ID)
	if err != nil {
		return nil, recordAlertQueryFailure(span, listFailure, err)
	}
	span.SetAttributes(attrAlertOutcome.String(string(AlertOutcomeSuccess)))
	return resources, nil
}

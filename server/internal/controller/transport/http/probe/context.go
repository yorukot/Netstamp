package probe

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"

	appprobe "github.com/yorukot/netstamp/internal/controller/application/probe"
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	"github.com/yorukot/netstamp/internal/domain/label"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func currentUserID(ctx context.Context) (string, error) {
	claims, ok := httpmiddleware.AccessTokenClaimsFromContext(ctx)
	if !ok || claims.Subject == "" {
		return "", huma.Error401Unauthorized("missing bearer token")
	}

	return claims.Subject, nil
}

func mapProbeError(err error, fallback string) error {
	switch {
	case errors.Is(err, domainproject.ErrProjectNotFound), errors.Is(err, label.ErrLabelNotFound), errors.Is(err, domainprobe.ErrProbeNotFound):
		return huma.Error404NotFound("not found")
	case errors.Is(err, appprobe.ErrForbidden):
		return huma.Error403Forbidden("forbidden")
	case errors.Is(err, appprobe.ErrInvalidInput), errors.Is(err, label.ErrInvalidInput), errors.Is(err, domainprobe.ErrInvalidInput):
		return invalidProbeInputError(err)
	default:
		return huma.Error500InternalServerError(fallback)
	}
}

func invalidProbeInputError(err error) error {
	fieldErrors, ok := appvalidation.FieldErrors(err)
	if !ok {
		return huma.Error422UnprocessableEntity("invalid probe input")
	}

	details := make([]error, 0, len(fieldErrors))
	for _, fieldErr := range fieldErrors {
		details = append(details, &huma.ErrorDetail{
			Message:  fieldErr.Message,
			Location: probeErrorLocation(fieldErr.Field),
			Value:    fieldErr.Value,
		})
	}

	return huma.Error422UnprocessableEntity("invalid probe input", details...)
}

func probeErrorLocation(field string) string {
	switch field {
	case "":
		return "body"
	case "projectRef":
		return "path.ref"
	case "probeId":
		return "path.probe_id"
	default:
		return "body." + field
	}
}

package probe

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"

	appprobe "github.com/yorukot/netstamp/internal/application/probe"
	appvalidation "github.com/yorukot/netstamp/internal/application/validation"
	httpmiddleware "github.com/yorukot/netstamp/internal/transport/http/middleware"
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
	case errors.Is(err, appprobe.ErrProjectNotFound), errors.Is(err, appprobe.ErrLabelNotFound):
		return huma.Error404NotFound("not found")
	case errors.Is(err, appprobe.ErrForbidden):
		return huma.Error403Forbidden("forbidden")
	case errors.Is(err, appprobe.ErrInvalidInput):
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
	default:
		return "body." + field
	}
}

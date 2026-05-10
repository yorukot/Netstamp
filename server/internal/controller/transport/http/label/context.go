package label

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"

	applabel "github.com/yorukot/netstamp/internal/controller/application/label"
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
)

func currentUserID(ctx context.Context) (string, error) {
	claims, ok := httpmiddleware.AccessTokenClaimsFromContext(ctx)
	if !ok || claims.Subject == "" {
		return "", huma.Error401Unauthorized("missing bearer token")
	}

	return claims.Subject, nil
}

func mapLabelError(err error, fallback string) error {
	switch {
	case errors.Is(err, applabel.ErrProjectNotFound), errors.Is(err, applabel.ErrUserNotFound), errors.Is(err, applabel.ErrLabelNotFound):
		return huma.Error404NotFound("not found")
	case errors.Is(err, applabel.ErrForbidden):
		return huma.Error403Forbidden("forbidden")
	case errors.Is(err, applabel.ErrLabelAlreadyExists):
		return huma.Error409Conflict("label already exists")
	case errors.Is(err, applabel.ErrInvalidInput):
		return invalidLabelInputError(err)
	default:
		return huma.Error500InternalServerError(fallback)
	}
}

func invalidLabelInputError(err error) error {
	fieldErrors, ok := appvalidation.FieldErrors(err)
	if !ok {
		return huma.Error422UnprocessableEntity("invalid label input")
	}

	details := make([]error, 0, len(fieldErrors))
	for _, fieldErr := range fieldErrors {
		details = append(details, &huma.ErrorDetail{
			Message:  fieldErr.Message,
			Location: labelErrorLocation(fieldErr.Field),
			Value:    fieldErr.Value,
		})
	}

	return huma.Error422UnprocessableEntity("invalid label input", details...)
}

func labelErrorLocation(field string) string {
	switch field {
	case "":
		return "body"
	case "projectRef":
		return "path.ref"
	case "labelId":
		return "path.label_id"
	default:
		return "body." + field
	}
}

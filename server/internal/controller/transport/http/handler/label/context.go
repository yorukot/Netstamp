package label

import (
	"context"
	"errors"

	applabel "github.com/yorukot/netstamp/internal/controller/application/label"
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	handlerproblem "github.com/yorukot/netstamp/internal/controller/transport/http/handler/problem"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	"github.com/yorukot/netstamp/internal/domain/label"
)

func currentUserID(ctx context.Context) (string, error) {
	claims, ok := httpmiddleware.AccessTokenClaimsFromContext(ctx)
	if !ok || claims.Subject == "" {
		return "", httpx.Unauthorized("missing auth cookie")
	}

	return claims.Subject, nil
}

func mapLabelError(err error, fallback string) error {
	if ok, mapped := handlerproblem.NotFound(err); ok {
		return mapped
	}

	switch {
	case errors.Is(err, applabel.ErrForbidden):
		return httpx.Forbidden("current user does not have the required project role for labels")
	case errors.Is(err, label.ErrLabelAlreadyExists):
		return httpx.Conflict("label already exists")
	case errors.Is(err, applabel.ErrInvalidInput), errors.Is(err, label.ErrInvalidInput):
		return invalidLabelInputError(err)
	default:
		return httpx.InternalServerError(fallback)
	}
}

func invalidLabelInputError(err error) error {
	fieldErrors, ok := appvalidation.FieldErrors(err)
	if !ok {
		return httpx.UnprocessableEntity("invalid label input")
	}

	details := make([]httpx.ErrorDetail, 0, len(fieldErrors))
	for _, fieldErr := range fieldErrors {
		details = append(details, httpx.ErrorDetail{
			Message:  fieldErr.Message,
			Location: labelErrorLocation(fieldErr.Field),
			Value:    fieldErr.Value,
		})
	}

	return httpx.UnprocessableEntity("invalid label input", details...)
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

package project

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"

	appproject "github.com/yorukot/netstamp/internal/controller/application/project"
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	"github.com/yorukot/netstamp/internal/domain/identity"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func currentUserID(ctx context.Context) (string, error) {
	claims, ok := httpmiddleware.AccessTokenClaimsFromContext(ctx)
	if !ok || claims.Subject == "" {
		return "", huma.Error401Unauthorized("missing auth cookie")
	}

	return claims.Subject, nil
}

func mapProjectError(err error, fallback string) error {
	switch {
	case errors.Is(err, domainproject.ErrProjectNotFound), errors.Is(err, domainproject.ErrMemberNotFound), errors.Is(err, identity.ErrUserNotFound):
		return huma.Error404NotFound("not found")
	case errors.Is(err, appproject.ErrForbidden):
		return huma.Error403Forbidden("forbidden")
	case errors.Is(err, domainproject.ErrProjectSlugAlreadyExists):
		return huma.Error409Conflict("project slug already exists")
	case errors.Is(err, domainproject.ErrMemberAlreadyExists):
		return huma.Error409Conflict("project member already exists")
	case errors.Is(err, appproject.ErrLastOwner):
		return huma.Error409Conflict("project must keep an owner")
	case errors.Is(err, appproject.ErrInvalidInput):
		return invalidProjectInputError(err)
	default:
		return huma.Error500InternalServerError(fallback)
	}
}

func invalidProjectInputError(err error) error {
	fieldErrors, ok := appvalidation.FieldErrors(err)
	if !ok {
		return huma.Error422UnprocessableEntity("invalid project input")
	}

	details := make([]error, 0, len(fieldErrors))
	for _, fieldErr := range fieldErrors {
		details = append(details, &huma.ErrorDetail{
			Message:  fieldErr.Message,
			Location: projectErrorLocation(fieldErr.Field),
			Value:    fieldErr.Value,
		})
	}

	return huma.Error422UnprocessableEntity("invalid project input", details...)
}

func projectErrorLocation(field string) string {
	switch field {
	case "":
		return "body"
	case "projectRef":
		return "path.ref"
	case "memberUserId":
		return "path.user_id"
	case "userId":
		return "body.userId"
	default:
		return "body." + field
	}
}

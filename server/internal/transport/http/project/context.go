package project

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"

	appproject "github.com/yorukot/netstamp/internal/application/project"
	httpmiddleware "github.com/yorukot/netstamp/internal/transport/http/middleware"
)

func currentUserID(ctx context.Context) (string, error) {
	claims, ok := httpmiddleware.AccessTokenClaimsFromContext(ctx)
	if !ok || claims.Subject == "" {
		return "", huma.Error401Unauthorized("missing bearer token")
	}

	return claims.Subject, nil
}

func mapProjectError(err error, fallback string) error {
	switch {
	case errors.Is(err, appproject.ErrProjectNotFound), errors.Is(err, appproject.ErrMemberNotFound), errors.Is(err, appproject.ErrUserNotFound):
		return huma.Error404NotFound("not found")
	case errors.Is(err, appproject.ErrForbidden):
		return huma.Error403Forbidden("forbidden")
	case errors.Is(err, appproject.ErrProjectSlugAlreadyExists):
		return huma.Error409Conflict("project slug already exists")
	case errors.Is(err, appproject.ErrMemberAlreadyExists):
		return huma.Error409Conflict("project member already exists")
	case errors.Is(err, appproject.ErrLastOwner):
		return huma.Error409Conflict("project must keep an owner")
	case errors.Is(err, appproject.ErrInvalidInput):
		return huma.Error422UnprocessableEntity("invalid project input")
	case errors.Is(err, appproject.ErrInvalidRole):
		return huma.Error422UnprocessableEntity("invalid project member role")
	default:
		return huma.Error500InternalServerError(fallback)
	}
}

package check

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"

	appcheck "github.com/yorukot/netstamp/internal/application/check"
	httpmiddleware "github.com/yorukot/netstamp/internal/transport/http/middleware"
)

func currentUserID(ctx context.Context) (string, error) {
	claims, ok := httpmiddleware.AccessTokenClaimsFromContext(ctx)
	if !ok || claims.Subject == "" {
		return "", huma.Error401Unauthorized("missing bearer token")
	}

	return claims.Subject, nil
}

func mapCheckError(err error, fallback string) error {
	switch {
	case errors.Is(err, appcheck.ErrProjectNotFound), errors.Is(err, appcheck.ErrUserNotFound), errors.Is(err, appcheck.ErrCheckNotFound), errors.Is(err, appcheck.ErrLabelNotFound):
		return huma.Error404NotFound("not found")
	case errors.Is(err, appcheck.ErrForbidden):
		return huma.Error403Forbidden("forbidden")
	case errors.Is(err, appcheck.ErrInvalidInput):
		return huma.Error422UnprocessableEntity("invalid check input")
	default:
		return huma.Error500InternalServerError(fallback)
	}
}

package probe

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"

	appprobe "github.com/yorukot/netstamp/internal/application/probe"
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
	case errors.Is(err, appprobe.ErrInvalidInput):
		return huma.Error422UnprocessableEntity("invalid probe input")
	default:
		return huma.Error500InternalServerError(fallback)
	}
}

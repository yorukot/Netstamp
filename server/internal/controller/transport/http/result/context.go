package result

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"

	appresult "github.com/yorukot/netstamp/internal/controller/application/result"
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	"github.com/yorukot/netstamp/internal/domain/identity"
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

func mapResultError(err error, fallback string) error {
	switch {
	case errors.Is(err, domainproject.ErrProjectNotFound), errors.Is(err, domainproject.ErrMemberNotFound), errors.Is(err, identity.ErrUserNotFound), errors.Is(err, domainprobe.ErrProbeNotFound), errors.Is(err, domaincheck.ErrCheckNotFound):
		return huma.Error404NotFound("not found")
	case errors.Is(err, appresult.ErrInvalidInput), errors.Is(err, domainprobe.ErrInvalidInput), errors.Is(err, domaincheck.ErrInvalidInput):
		return invalidResultInputError(err)
	default:
		return huma.Error500InternalServerError(fallback)
	}
}

func invalidResultInputError(err error) error {
	fieldErrors, ok := appvalidation.FieldErrors(err)
	if !ok {
		return huma.Error422UnprocessableEntity("invalid result query input")
	}

	details := make([]error, 0, len(fieldErrors))
	for _, fieldErr := range fieldErrors {
		details = append(details, &huma.ErrorDetail{
			Message:  fieldErr.Message,
			Location: resultErrorLocation(fieldErr.Field),
			Value:    fieldErr.Value,
		})
	}

	return huma.Error422UnprocessableEntity("invalid result query input", details...)
}

func resultErrorLocation(field string) string {
	switch field {
	case "projectRef":
		return "path.ref"
	case "probeId":
		return "query.probeId"
	case "checkId":
		return "query.checkId"
	case "from", "to", "metric", "maxDataPoints":
		return "query." + field
	default:
		return "query"
	}
}

package result

import (
	"context"
	"errors"

	appresult "github.com/yorukot/netstamp/internal/controller/application/result"
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	handlerproblem "github.com/yorukot/netstamp/internal/controller/transport/http/handler/problem"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

func currentUserID(ctx context.Context) (string, error) {
	claims, ok := httpmiddleware.AccessTokenClaimsFromContext(ctx)
	if !ok || claims.Subject == "" {
		return "", httpx.UnauthorizedCode(httpx.CodeAuthMissingSession, "missing auth cookie")
	}

	return claims.Subject, nil
}

func mapResultError(err error, fallback string) error {
	if ok, mapped := handlerproblem.NotFound(err); ok {
		return mapped
	}

	switch {
	case errors.Is(err, appresult.ErrInvalidInput), errors.Is(err, domainprobe.ErrInvalidInput), errors.Is(err, domaincheck.ErrInvalidInput):
		return invalidResultInputError(err)
	default:
		return httpx.InternalServerError(fallback)
	}
}

func invalidResultInputError(err error) error {
	fieldErrors, ok := appvalidation.FieldErrors(err)
	if !ok {
		return httpx.UnprocessableEntity("invalid result query input")
	}

	details := make([]httpx.ErrorDetail, 0, len(fieldErrors))
	for _, fieldErr := range fieldErrors {
		details = append(details, httpx.ErrorDetail{
			Code:     fieldErr.Code,
			Message:  fieldErr.Message,
			Location: resultErrorLocation(fieldErr.Field),
			Value:    fieldErr.Value,
		})
	}

	return httpx.UnprocessableEntity("invalid result query input", details...)
}

func resultErrorLocation(field string) string {
	switch field {
	case "projectRef":
		return "path.ref"
	case "probeId":
		return "query.probeId"
	case "checkId":
		return "query.checkId"
	case "from", "to", "series", "maxDataPoints", "limit", "cursor", "type", "status":
		return "query." + field
	default:
		return "query"
	}
}

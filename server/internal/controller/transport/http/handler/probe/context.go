package probe

import (
	"context"
	"errors"

	appprobe "github.com/yorukot/netstamp/internal/controller/application/probe"
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	handlerproblem "github.com/yorukot/netstamp/internal/controller/transport/http/handler/problem"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	"github.com/yorukot/netstamp/internal/domain/label"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

func currentUserID(ctx context.Context) (string, error) {
	claims, ok := httpmiddleware.AccessTokenClaimsFromContext(ctx)
	if !ok || claims.Subject == "" {
		return "", httpx.UnauthorizedCode(httpx.CodeAuthMissingSession, "missing auth cookie")
	}

	return claims.Subject, nil
}

func mapProbeError(err error, fallback string) error {
	if ok, mapped := handlerproblem.NotFound(err); ok {
		return mapped
	}

	switch {
	case errors.Is(err, appprobe.ErrForbidden):
		return httpx.ForbiddenCode(httpx.CodeProjectRoleRequired, "current user does not have the required project role for probes")
	case errors.Is(err, appprobe.ErrInvalidInput), errors.Is(err, label.ErrInvalidInput), errors.Is(err, domainprobe.ErrInvalidInput):
		return invalidProbeInputError(err)
	default:
		return httpx.InternalServerError(fallback)
	}
}

func invalidProbeInputError(err error) error {
	fieldErrors, ok := appvalidation.FieldErrors(err)
	if !ok {
		return httpx.UnprocessableEntity("invalid probe input")
	}

	details := make([]httpx.ErrorDetail, 0, len(fieldErrors))
	for _, fieldErr := range fieldErrors {
		details = append(details, httpx.ErrorDetail{
			Code:     fieldErr.Code,
			Message:  fieldErr.Message,
			Location: probeErrorLocation(fieldErr.Field),
			Value:    fieldErr.Value,
		})
	}

	return httpx.UnprocessableEntity("invalid probe input", details...)
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

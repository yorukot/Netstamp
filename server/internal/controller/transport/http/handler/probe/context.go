package probe

import (
	"context"
	"errors"

	appprobe "github.com/yorukot/netstamp/internal/controller/application/probe"
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	"github.com/yorukot/netstamp/internal/domain/label"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func currentUserID(ctx context.Context) (string, error) {
	claims, ok := httpmiddleware.AccessTokenClaimsFromContext(ctx)
	if !ok || claims.Subject == "" {
		return "", httpx.Unauthorized("missing auth cookie")
	}

	return claims.Subject, nil
}

func mapProbeError(err error, fallback string) error {
	switch {
	case errors.Is(err, domainproject.ErrProjectNotFound), errors.Is(err, domainproject.ErrMemberNotFound), errors.Is(err, label.ErrLabelNotFound), errors.Is(err, domainprobe.ErrProbeNotFound):
		return httpx.NotFound("not found")
	case errors.Is(err, appprobe.ErrForbidden):
		return httpx.Forbidden("forbidden")
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

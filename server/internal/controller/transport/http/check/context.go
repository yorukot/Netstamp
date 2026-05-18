package check

import (
	"context"
	"errors"
	"strings"

	appcheck "github.com/yorukot/netstamp/internal/controller/application/check"
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	"github.com/yorukot/netstamp/internal/domain/identity"
	"github.com/yorukot/netstamp/internal/domain/label"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func currentUserID(ctx context.Context) (string, error) {
	claims, ok := httpmiddleware.AccessTokenClaimsFromContext(ctx)
	if !ok || claims.Subject == "" {
		return "", httpx.Unauthorized("missing auth cookie")
	}

	return claims.Subject, nil
}

func mapCheckError(err error, fallback string) error {
	switch {
	case errors.Is(err, domainproject.ErrProjectNotFound), errors.Is(err, domainproject.ErrMemberNotFound), errors.Is(err, identity.ErrUserNotFound), errors.Is(err, domaincheck.ErrCheckNotFound), errors.Is(err, label.ErrLabelNotFound):
		return httpx.NotFound("not found")
	case errors.Is(err, appcheck.ErrForbidden):
		return httpx.Forbidden("forbidden")
	case errors.Is(err, appcheck.ErrInvalidInput), errors.Is(err, domaincheck.ErrInvalidInput), errors.Is(err, label.ErrInvalidInput):
		return invalidCheckInputError(err)
	default:
		return httpx.InternalServerError(fallback)
	}
}

func invalidCheckInputError(err error) error {
	fieldErrors, ok := appvalidation.FieldErrors(err)
	if !ok {
		return httpx.UnprocessableEntity("invalid check input")
	}

	details := make([]httpx.ErrorDetail, 0, len(fieldErrors))
	for _, fieldErr := range fieldErrors {
		details = append(details, httpx.ErrorDetail{
			Message:  fieldErr.Message,
			Location: checkErrorLocation(fieldErr.Field),
			Value:    fieldErr.Value,
		})
	}

	return httpx.UnprocessableEntity("invalid check input", details...)
}

func checkErrorLocation(field string) string {
	switch field {
	case "":
		return "body"
	case "projectRef":
		return "path.ref"
	case "checkId":
		return "path.check_id"
	}
	if isPingConfigField(field) {
		return "body.pingConfig." + field
	}
	if strings.HasPrefix(field, "tracerouteConfig.") {
		return "body." + field
	}

	return "body." + field
}

func isPingConfigField(field string) bool {
	switch field {
	case "packetCount", "packetSizeBytes", "timeoutMs", "ipFamily":
		return true
	default:
		return false
	}
}

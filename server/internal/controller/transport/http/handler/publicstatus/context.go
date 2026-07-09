package publicstatus

import (
	"context"
	"errors"

	apppublic "github.com/yorukot/netstamp/internal/controller/application/publicstatus"
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	handlerproblem "github.com/yorukot/netstamp/internal/controller/transport/http/handler/problem"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	domainpublic "github.com/yorukot/netstamp/internal/domain/publicstatus"
)

func currentUserID(ctx context.Context) (string, error) {
	claims, ok := httpmiddleware.AccessTokenClaimsFromContext(ctx)
	if !ok || claims.Subject == "" {
		return "", httpx.Unauthorized("missing auth cookie")
	}
	return claims.Subject, nil
}

func mapPublicStatusError(err error, fallback string) error {
	if mapped, ok := handlerproblem.NotFound(err); ok {
		return mapped
	}

	switch {
	case errors.Is(err, apppublic.ErrForbidden):
		return httpx.Forbidden("current user does not have the required project role for public status")
	case errors.Is(err, domainpublic.ErrSlugAlreadyExist):
		return httpx.Conflict("public status page slug already exists")
	case errors.Is(err, apppublic.ErrInvalidInput), errors.Is(err, domainpublic.ErrInvalidInput):
		return invalidPublicStatusInputError(err)
	default:
		return httpx.InternalServerError(fallback)
	}
}

func invalidPublicStatusInputError(err error) error {
	fieldErrors, ok := appvalidation.FieldErrors(err)
	if !ok {
		return httpx.UnprocessableEntity("invalid public status page input")
	}
	details := make([]httpx.ErrorDetail, 0, len(fieldErrors))
	for _, fieldErr := range fieldErrors {
		details = append(details, httpx.ErrorDetail{
			Message:  fieldErr.Message,
			Location: publicStatusErrorLocation(fieldErr.Field),
			Value:    fieldErr.Value,
		})
	}
	return httpx.UnprocessableEntity("invalid public status page input", details...)
}

func publicStatusErrorLocation(field string) string {
	switch field {
	case "":
		return "body"
	case "projectRef":
		return "path.ref"
	case "pageId":
		return "path.page_id"
	case "elementId":
		return "path.element_id"
	case "slug":
		return "path.slug"
	default:
		return "body." + field
	}
}

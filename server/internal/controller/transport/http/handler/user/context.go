package userhttp

import (
	"context"
	"errors"

	appuser "github.com/yorukot/netstamp/internal/controller/application/user"
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

func currentUserID(ctx context.Context) (string, error) {
	claims, ok := httpmiddleware.AccessTokenClaimsFromContext(ctx)
	if !ok || claims.Subject == "" {
		return "", httpx.Unauthorized("missing auth cookie")
	}

	return claims.Subject, nil
}

func mapUserError(err error, fallback string) error {
	switch {
	case errors.Is(err, identity.ErrUserNotFound):
		return httpx.Unauthorized("invalid session")
	case errors.Is(err, appuser.ErrCredentialsInvalid):
		return httpx.Unauthorized("credentials invalid")
	case errors.Is(err, identity.ErrEmailAlreadyExists):
		return httpx.Conflict("email already exists")
	case errors.Is(err, appuser.ErrInvalidInput):
		return invalidUserInputError(err)
	default:
		return httpx.InternalServerError(fallback)
	}
}

func invalidUserInputError(err error) error {
	fieldErrors, ok := appvalidation.FieldErrors(err)
	if !ok {
		return httpx.UnprocessableEntity("invalid user input")
	}

	details := make([]httpx.ErrorDetail, 0, len(fieldErrors))
	for _, fieldErr := range fieldErrors {
		details = append(details, httpx.ErrorDetail{
			Message:  fieldErr.Message,
			Location: userErrorLocation(fieldErr.Field),
			Value:    fieldErr.Value,
		})
	}

	return httpx.UnprocessableEntity("invalid user input", details...)
}

func userErrorLocation(field string) string {
	switch field {
	case "":
		return "body"
	case "currentUserId":
		return "auth.subject"
	default:
		return "body." + field
	}
}

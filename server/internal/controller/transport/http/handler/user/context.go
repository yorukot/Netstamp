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
	userID, ok := httpmiddleware.CurrentUserIDFromContext(ctx)
	if !ok {
		return "", httpx.UnauthorizedCode(httpx.CodeAuthMissingSession, "missing auth cookie")
	}
	return userID, nil
}

func currentSessionID(ctx context.Context) string {
	claims, _ := httpmiddleware.SessionClaimsFromContext(ctx)
	return claims.SessionID
}

func mapUserError(err error, fallback string) error {
	switch {
	case errors.Is(err, identity.ErrUserNotFound):
		return httpx.UnauthorizedCode(httpx.CodeAuthInvalidSession, "invalid session")
	case errors.Is(err, appuser.ErrCredentialsInvalid):
		return httpx.UnauthorizedCode(httpx.CodeAuthInvalidCredentials, "credentials invalid")
	case errors.Is(err, identity.ErrEmailAlreadyExists):
		return httpx.ConflictCode(httpx.CodeEmailAlreadyExists, "email already exists")
	case errors.Is(err, appuser.ErrLastSystemAdmin):
		return httpx.ConflictCode(httpx.CodeLastSystemAdmin, "system must keep at least one administrator")
	case errors.Is(err, appuser.ErrLastCredential):
		return httpx.ConflictCode(httpx.CodeAuthLastCredential, "account must keep at least one authentication method")
	case errors.Is(err, appuser.ErrIdentityNotFound):
		return httpx.NotFoundCode(httpx.CodeAuthIdentityNotFound, "identity not found")
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
			Code:     fieldErr.Code,
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

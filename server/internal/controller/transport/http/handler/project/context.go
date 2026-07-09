package project

import (
	"context"
	"errors"

	appproject "github.com/yorukot/netstamp/internal/controller/application/project"
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	handlerproblem "github.com/yorukot/netstamp/internal/controller/transport/http/handler/problem"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func currentUserID(ctx context.Context) (string, error) {
	claims, ok := httpmiddleware.AccessTokenClaimsFromContext(ctx)
	if !ok || claims.Subject == "" {
		return "", httpx.UnauthorizedCode(httpx.CodeAuthMissingSession, "missing auth cookie")
	}

	return claims.Subject, nil
}

func mapProjectError(err error, fallback string) error {
	if ok, mapped := handlerproblem.NotFound(err); ok {
		return mapped
	}

	switch {
	case errors.Is(err, appproject.ErrForbidden):
		return httpx.ForbiddenCode(httpx.CodeProjectRoleRequired, "current user does not have the required project role")
	case errors.Is(err, domainproject.ErrProjectSlugAlreadyExists):
		return httpx.ConflictCode(httpx.CodeProjectSlugAlreadyExists, "project slug already exists")
	case errors.Is(err, domainproject.ErrMemberAlreadyExists):
		return httpx.ConflictCode(httpx.CodeProjectMemberAlreadyExists, "project member already exists")
	case errors.Is(err, domainproject.ErrInviteAlreadyExists):
		return httpx.ConflictCode(httpx.CodeProjectInviteAlreadyExists, "project invite already exists")
	case errors.Is(err, appproject.ErrLastOwner):
		return httpx.ConflictCode(httpx.CodeProjectLastOwner, "project must keep an owner")
	case errors.Is(err, appproject.ErrInvalidInput):
		return invalidProjectInputError(err)
	default:
		return httpx.InternalServerError(fallback)
	}
}

func invalidProjectInputError(err error) error {
	fieldErrors, ok := appvalidation.FieldErrors(err)
	if !ok {
		return httpx.UnprocessableEntity("invalid project input")
	}

	details := make([]httpx.ErrorDetail, 0, len(fieldErrors))
	for _, fieldErr := range fieldErrors {
		details = append(details, httpx.ErrorDetail{
			Code:     fieldErr.Code,
			Message:  fieldErr.Message,
			Location: projectErrorLocation(fieldErr.Field),
			Value:    fieldErr.Value,
		})
	}

	return httpx.UnprocessableEntity("invalid project input", details...)
}

func projectErrorLocation(field string) string {
	switch field {
	case "":
		return "body"
	case "projectRef":
		return "path.ref"
	case "memberUserId":
		return "path.user_id"
	case "inviteId":
		return "path.invite_id"
	case "email":
		return "body.email"
	default:
		return "body." + field
	}
}

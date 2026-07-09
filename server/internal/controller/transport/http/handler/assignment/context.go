package assignment

import (
	"context"
	"errors"

	appassignment "github.com/yorukot/netstamp/internal/controller/application/assignment"
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

func mapAssignmentError(err error, fallback string) error {
	if ok, mapped := handlerproblem.NotFound(err); ok {
		return mapped
	}

	switch {
	case errors.Is(err, appassignment.ErrInvalidInput), errors.Is(err, domainprobe.ErrInvalidInput), errors.Is(err, domaincheck.ErrInvalidInput):
		return invalidAssignmentInputError(err)
	default:
		return httpx.InternalServerError(fallback)
	}
}

func invalidAssignmentInputError(err error) error {
	fieldErrors, ok := appvalidation.FieldErrors(err)
	if !ok {
		return httpx.UnprocessableEntity("invalid assignment input")
	}

	details := make([]httpx.ErrorDetail, 0, len(fieldErrors))
	for _, fieldErr := range fieldErrors {
		details = append(details, httpx.ErrorDetail{
			Code:     fieldErr.Code,
			Message:  fieldErr.Message,
			Location: assignmentErrorLocation(fieldErr.Field),
			Value:    fieldErr.Value,
		})
	}

	return httpx.UnprocessableEntity("invalid assignment input", details...)
}

func assignmentErrorLocation(field string) string {
	switch field {
	case "projectRef":
		return "path.ref"
	case "selector":
		return "body.selector"
	case "probeId", "checkId":
		return "query." + field
	default:
		return "body"
	}
}

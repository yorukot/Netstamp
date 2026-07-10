package alert

import (
	"errors"
	"net/http"

	appalert "github.com/yorukot/netstamp/internal/controller/application/alert"
	handlerproblem "github.com/yorukot/netstamp/internal/controller/transport/http/handler/problem"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
	"github.com/yorukot/netstamp/internal/domain/alertcondition"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
)

func currentUserID(r *http.Request) (string, error) {
	userID, ok := httpmiddleware.CurrentUserIDFromContext(r.Context())
	if !ok {
		return "", httpx.UnauthorizedCode(httpx.CodeAuthMissingSession, "missing auth cookie")
	}
	return userID, nil
}

func projectInput(r *http.Request, userID string) appalert.ProjectInput {
	return appalert.ProjectInput{ProjectRef: httpx.Path(r, "ref"), CurrentUserID: userID}
}

func writeJSONOrProblem(w http.ResponseWriter, r *http.Request, status int, body any, err error) {
	if err != nil {
		httpx.WriteProblem(w, r, mapAlertError(err, "alert request failed"))
		return
	}
	httpx.WriteJSON(w, status, body)
}

func mapAlertError(err error, fallback string) error {
	if ok, mapped := handlerproblem.NotFound(err); ok {
		return mapped
	}

	switch {
	case errors.Is(err, appalert.ErrForbidden):
		return httpx.ForbiddenCode(httpx.CodeProjectRoleRequired, "current user does not have the required project role for alerts")
	case errors.Is(err, appalert.ErrInvalidInput),
		errors.Is(err, domainalert.ErrInvalidInput),
		errors.Is(err, alertcondition.ErrInvalidCondition),
		errors.Is(err, domaincheck.ErrInvalidInput):
		return httpx.UnprocessableEntity("invalid alert input")
	default:
		return httpx.InternalServerError(fallback)
	}
}

package alert

import (
	"errors"
	"net/http"

	appalert "github.com/yorukot/netstamp/internal/controller/application/alert"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
	"github.com/yorukot/netstamp/internal/domain/alertcondition"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	"github.com/yorukot/netstamp/internal/domain/identity"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func currentUserID(r *http.Request) (string, error) {
	claims, ok := httpmiddleware.AccessTokenClaimsFromContext(r.Context())
	if !ok || claims.Subject == "" {
		return "", httpx.Unauthorized("missing auth cookie")
	}
	return claims.Subject, nil
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
	switch {
	case errors.Is(err, domainproject.ErrProjectNotFound),
		errors.Is(err, domainproject.ErrMemberNotFound),
		errors.Is(err, identity.ErrUserNotFound),
		errors.Is(err, domainalert.ErrRuleNotFound),
		errors.Is(err, domainalert.ErrIncidentNotFound),
		errors.Is(err, domainalert.ErrChannelNotFound):
		return httpx.NotFound("not found")
	case errors.Is(err, appalert.ErrForbidden):
		return httpx.Forbidden("forbidden")
	case errors.Is(err, appalert.ErrInvalidInput),
		errors.Is(err, domainalert.ErrInvalidInput),
		errors.Is(err, alertcondition.ErrInvalidCondition),
		errors.Is(err, domaincheck.ErrInvalidInput):
		return httpx.UnprocessableEntity("invalid alert input")
	default:
		return httpx.InternalServerError(fallback)
	}
}

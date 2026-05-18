package result

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	appresult "github.com/yorukot/netstamp/internal/controller/application/result"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
)

type Handler struct {
	service  *appresult.Service
	verifier appauth.TokenVerifier
}

func NewHandler(service *appresult.Service, verifier appauth.TokenVerifier) *Handler {
	return &Handler{
		service:  service,
		verifier: verifier,
	}
}

func (h *Handler) RegisterRoutes(api huma.API) {
	authMiddleware := httpmiddleware.RequireAuth(h.verifier)
	security := []map[string][]string{{httpmiddleware.SessionCookieSecurityScheme: {}}}
	middlewares := huma.Middlewares{authMiddleware}

	huma.Register(api, huma.Operation{
		OperationID: "queryProjectPingResultSeries",
		Method:      http.MethodGet,
		Path:        "/projects/{ref}/results/ping/series",
		Summary:     "Query project ping result series",
		Tags:        []string{"Results"},
		Security:    security,
		Middlewares: middlewares,
		Errors:      []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.queryPingSeries)

	huma.Register(api, huma.Operation{
		OperationID: "queryProjectTracerouteResultRuns",
		Method:      http.MethodGet,
		Path:        "/projects/{ref}/results/traceroute/runs",
		Summary:     "Query project traceroute result runs",
		Tags:        []string{"Results"},
		Security:    security,
		Middlewares: middlewares,
		Errors:      []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.queryTracerouteRuns)
}

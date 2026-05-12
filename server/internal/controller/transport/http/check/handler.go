package check

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	appcheck "github.com/yorukot/netstamp/internal/controller/application/check"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
)

type Handler struct {
	service  *appcheck.Service
	verifier appauth.TokenVerifier
}

func NewHandler(service *appcheck.Service, verifier appauth.TokenVerifier) *Handler {
	return &Handler{
		service:  service,
		verifier: verifier,
	}
}

func (h *Handler) RegisterRoutes(api huma.API) {
	authMiddleware := httpmiddleware.RequireAuth(h.verifier)
	security := []map[string][]string{{"bearerAuth": {}}}
	middlewares := huma.Middlewares{authMiddleware}

	huma.Register(api, huma.Operation{
		OperationID: "listProjectChecks",
		Method:      http.MethodGet,
		Path:        "/projects/{ref}/checks",
		Summary:     "List project checks",
		Tags:        []string{"Checks"},
		Security:    security,
		Middlewares: middlewares,
		Errors:      []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.listChecks)

	huma.Register(api, huma.Operation{
		OperationID:   "createProjectCheck",
		Method:        http.MethodPost,
		Path:          "/projects/{ref}/checks",
		DefaultStatus: http.StatusCreated,
		Summary:       "Create project check",
		Tags:          []string{"Checks"},
		Security:      security,
		Middlewares:   middlewares,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.createCheck)

	huma.Register(api, huma.Operation{
		OperationID: "getProjectCheck",
		Method:      http.MethodGet,
		Path:        "/projects/{ref}/checks/{check_id}",
		Summary:     "Get project check",
		Tags:        []string{"Checks"},
		Security:    security,
		Middlewares: middlewares,
		Errors:      []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.getCheck)

	huma.Register(api, huma.Operation{
		OperationID: "updateProjectCheck",
		Method:      http.MethodPatch,
		Path:        "/projects/{ref}/checks/{check_id}",
		Summary:     "Update project check",
		Tags:        []string{"Checks"},
		Security:    security,
		Middlewares: middlewares,
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.updateCheck)

	huma.Register(api, huma.Operation{
		OperationID:   "deleteProjectCheck",
		Method:        http.MethodDelete,
		Path:          "/projects/{ref}/checks/{check_id}",
		DefaultStatus: http.StatusNoContent,
		Summary:       "Delete project check",
		Tags:          []string{"Checks"},
		Security:      security,
		Middlewares:   middlewares,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.deleteCheck)
}

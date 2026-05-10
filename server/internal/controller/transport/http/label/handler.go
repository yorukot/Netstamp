package label

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	applabel "github.com/yorukot/netstamp/internal/controller/application/label"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
)

type Handler struct {
	service  *applabel.Service
	verifier appauth.TokenVerifier
}

func NewHandler(service *applabel.Service, verifier appauth.TokenVerifier) *Handler {
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
		OperationID: "listProjectLabels",
		Method:      http.MethodGet,
		Path:        "/projects/{ref}/labels",
		Summary:     "List project labels",
		Tags:        []string{"Labels"},
		Security:    security,
		Middlewares: middlewares,
		Errors:      []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusInternalServerError},
	}, h.listLabels)

	huma.Register(api, huma.Operation{
		OperationID:   "createProjectLabel",
		Method:        http.MethodPost,
		Path:          "/projects/{ref}/labels",
		DefaultStatus: http.StatusCreated,
		Summary:       "Create project label",
		Tags:          []string{"Labels"},
		Security:      security,
		Middlewares:   middlewares,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.createLabel)

	huma.Register(api, huma.Operation{
		OperationID: "updateProjectLabel",
		Method:      http.MethodPatch,
		Path:        "/projects/{ref}/labels/{label_id}",
		Summary:     "Update project label",
		Tags:        []string{"Labels"},
		Security:    security,
		Middlewares: middlewares,
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.updateLabel)

	huma.Register(api, huma.Operation{
		OperationID:   "deleteProjectLabel",
		Method:        http.MethodDelete,
		Path:          "/projects/{ref}/labels/{label_id}",
		DefaultStatus: http.StatusNoContent,
		Summary:       "Delete project label",
		Tags:          []string{"Labels"},
		Security:      security,
		Middlewares:   middlewares,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusInternalServerError},
	}, h.deleteLabel)
}

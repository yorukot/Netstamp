package probe

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	appprobe "github.com/yorukot/netstamp/internal/controller/application/proberegistry"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
)

type Handler struct {
	service  *appprobe.Service
	verifier appauth.TokenVerifier
}

func NewHandler(service *appprobe.Service, verifier appauth.TokenVerifier) *Handler {
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
		OperationID: "listProjectProbes",
		Method:      http.MethodGet,
		Path:        "/projects/{ref}/probes",
		Summary:     "List project probes",
		Tags:        []string{"Probes"},
		Security:    security,
		Middlewares: middlewares,
		Errors:      []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.listProbes)

	huma.Register(api, huma.Operation{
		OperationID:   "createProjectProbe",
		Method:        http.MethodPost,
		Path:          "/projects/{ref}/probes",
		DefaultStatus: http.StatusCreated,
		Summary:       "Create project probe",
		Tags:          []string{"Probes"},
		Security:      security,
		Middlewares:   middlewares,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.createProbe)

	huma.Register(api, huma.Operation{
		OperationID: "getProjectProbe",
		Method:      http.MethodGet,
		Path:        "/projects/{ref}/probes/{probe_id}",
		Summary:     "Get project probe",
		Tags:        []string{"Probes"},
		Security:    security,
		Middlewares: middlewares,
		Errors:      []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.getProbe)

	huma.Register(api, huma.Operation{
		OperationID: "updateProjectProbe",
		Method:      http.MethodPatch,
		Path:        "/projects/{ref}/probes/{probe_id}",
		Summary:     "Update project probe",
		Tags:        []string{"Probes"},
		Security:    security,
		Middlewares: middlewares,
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.updateProbe)

	huma.Register(api, huma.Operation{
		OperationID:   "deleteProjectProbe",
		Method:        http.MethodDelete,
		Path:          "/projects/{ref}/probes/{probe_id}",
		DefaultStatus: http.StatusNoContent,
		Summary:       "Delete project probe",
		Tags:          []string{"Probes"},
		Security:      security,
		Middlewares:   middlewares,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.deleteProbe)

	huma.Register(api, huma.Operation{
		OperationID: "rotateProjectProbeSecret",
		Method:      http.MethodPost,
		Path:        "/projects/{ref}/probes/{probe_id}/secret-rotations",
		Summary:     "Rotate project probe secret",
		Tags:        []string{"Probes"},
		Security:    security,
		Middlewares: middlewares,
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.rotateSecret)
}

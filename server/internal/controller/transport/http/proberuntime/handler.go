package proberuntime

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	appproberuntime "github.com/yorukot/netstamp/internal/controller/application/proberuntime"
)

type Handler struct {
	service *appproberuntime.Service
}

func NewHandler(service *appproberuntime.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(api huma.API) {
	security := []map[string][]string{{"probeAuth": {}}}
	middlewares := huma.Middlewares{requireRuntimeAuth}

	huma.Register(api, huma.Operation{
		OperationID: "probeRuntimeHello",
		Method:      http.MethodPost,
		Path:        "/runtime/probes/{probe_id}/hello",
		Summary:     "Start probe runtime session",
		Tags:        []string{"Probe Runtime"},
		Security:    security,
		Middlewares: middlewares,
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.hello)

	huma.Register(api, huma.Operation{
		OperationID: "probeRuntimeHeartbeat",
		Method:      http.MethodPost,
		Path:        "/runtime/probes/{probe_id}/heartbeat",
		Summary:     "Update probe runtime status",
		Tags:        []string{"Probe Runtime"},
		Security:    security,
		Middlewares: middlewares,
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.heartbeat)

	huma.Register(api, huma.Operation{
		OperationID: "listProbeRuntimeAssignments",
		Method:      http.MethodGet,
		Path:        "/runtime/probes/{probe_id}/assignments",
		Summary:     "List probe runtime assignments",
		Tags:        []string{"Probe Runtime"},
		Security:    security,
		Middlewares: middlewares,
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.listAssignments)

	huma.Register(api, huma.Operation{
		OperationID: "submitProbeRuntimeResults",
		Method:      http.MethodPost,
		Path:        "/runtime/probes/{probe_id}/results",
		Summary:     "Submit probe runtime results",
		Tags:        []string{"Probe Runtime"},
		Security:    security,
		Middlewares: middlewares,
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.submitResults)
}

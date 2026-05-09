package proberuntime

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	appproberuntime "github.com/yorukot/netstamp/internal/application/proberuntime"
)

type Handler struct {
	service *appproberuntime.Service
}

func NewHandler(service *appproberuntime.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(api huma.API) {
	security := []map[string][]string{{"probeAuth": {}}}

	huma.Register(api, huma.Operation{
		OperationID: "probeRuntimeHello",
		Method:      http.MethodPost,
		Path:        "/probes/{probe_id}/runtime/hello",
		Summary:     "Start probe runtime session",
		Tags:        []string{"Probe Runtime"},
		Security:    security,
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.hello)

	huma.Register(api, huma.Operation{
		OperationID: "probeRuntimeHeartbeat",
		Method:      http.MethodPost,
		Path:        "/probes/{probe_id}/runtime/heartbeat",
		Summary:     "Update probe runtime status",
		Tags:        []string{"Probe Runtime"},
		Security:    security,
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.heartbeat)

	huma.Register(api, huma.Operation{
		OperationID: "listProbeRuntimeAssignments",
		Method:      http.MethodGet,
		Path:        "/probes/{probe_id}/runtime/assignments",
		Summary:     "List probe runtime assignments",
		Tags:        []string{"Probe Runtime"},
		Security:    security,
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.listAssignments)

	huma.Register(api, huma.Operation{
		OperationID: "submitProbeRuntimePingResults",
		Method:      http.MethodPost,
		Path:        "/probes/{probe_id}/runtime/results/ping",
		Summary:     "Submit probe ping result batch",
		Tags:        []string{"Probe Runtime"},
		Security:    security,
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.submitPingResults)
}

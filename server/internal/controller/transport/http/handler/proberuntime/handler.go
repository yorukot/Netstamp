package proberuntime

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	appproberuntime "github.com/yorukot/netstamp/internal/controller/application/proberuntime"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

type Handler struct {
	service *appproberuntime.Service
}

func NewHandler(service *appproberuntime.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(api chi.Router) {
	api.Group(func(r chi.Router) {
		r.Use(requireRuntimeAuth)

		r.Post("/runtime/probes/{probe_id}/hello", h.handleHello)
		r.Post("/runtime/probes/{probe_id}/heartbeat", h.handleHeartbeat)
		r.Put("/runtime/probes/{probe_id}/ip-family-capabilities", h.handleUpdateIPFamilyCapabilities)
		r.Get("/runtime/probes/{probe_id}/assignments", h.handleListAssignments)
		r.Post("/runtime/probes/{probe_id}/results", h.handleSubmitResults)
	})
}

func (h *Handler) handleHello(w http.ResponseWriter, r *http.Request) {
	output, err := h.hello(r.Context(), &helloInput{ProbeID: httpx.Path(r, "probe_id")})
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func (h *Handler) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	var body runtimeStatusBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.heartbeat(r.Context(), &heartbeatInput{ProbeID: httpx.Path(r, "probe_id"), Body: body})
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func (h *Handler) handleListAssignments(w http.ResponseWriter, r *http.Request) {
	output, err := h.listAssignments(r.Context(), &listAssignmentsInput{ProbeID: httpx.Path(r, "probe_id")})
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func (h *Handler) handleSubmitResults(w http.ResponseWriter, r *http.Request) {
	var body submitResultsBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.submitResults(r.Context(), &submitResultsInput{ProbeID: httpx.Path(r, "probe_id"), Body: body})
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

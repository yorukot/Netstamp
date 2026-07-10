package assignment

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	appassignment "github.com/yorukot/netstamp/internal/controller/application/assignment"
	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
)

type Handler struct {
	service    *appassignment.Service
	verifier   appauth.SessionManager
	cookieName string
}

func NewHandler(service *appassignment.Service, verifier appauth.SessionManager, cookieName string) *Handler {
	return &Handler{
		service:    service,
		verifier:   verifier,
		cookieName: cookieName,
	}
}

func (h *Handler) RegisterRoutes(api chi.Router) {
	api.Group(func(r chi.Router) {
		r.Use(httpmiddleware.RequireAuth(h.verifier, h.cookieName))

		r.Post("/projects/{ref}/selector-previews", h.handlePreviewSelector)
		r.Get("/projects/{ref}/assignments", h.handleListProjectAssignments)
	})
}

func (h *Handler) handlePreviewSelector(w http.ResponseWriter, r *http.Request) {
	var body previewSelectorInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.previewSelector(r.Context(), &previewSelectorInput{Ref: httpx.Path(r, "ref"), Body: body})
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func (h *Handler) handleListProjectAssignments(w http.ResponseWriter, r *http.Request) {
	output, err := h.listProjectAssignments(r.Context(), &listProjectAssignmentsInput{
		Ref:     httpx.Path(r, "ref"),
		ProbeID: httpx.QueryString(r, "probeId"),
		CheckID: httpx.QueryString(r, "checkId"),
	})
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

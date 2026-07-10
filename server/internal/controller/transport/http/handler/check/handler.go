package check

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	appcheck "github.com/yorukot/netstamp/internal/controller/application/check"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
)

type Handler struct {
	service    *appcheck.Service
	verifier   appauth.SessionManager
	cookieName string
}

func NewHandler(service *appcheck.Service, verifier appauth.SessionManager, cookieName string) *Handler {
	return &Handler{
		service:    service,
		verifier:   verifier,
		cookieName: cookieName,
	}
}

func (h *Handler) RegisterRoutes(api chi.Router) {
	api.Group(func(r chi.Router) {
		r.Use(httpmiddleware.RequireAuth(h.verifier, h.cookieName))

		r.Get("/projects/{ref}/checks", h.handleListChecks)
		r.Post("/projects/{ref}/checks", h.handleCreateCheck)
		r.Get("/projects/{ref}/checks/{check_id}", h.handleGetCheck)
		r.Patch("/projects/{ref}/checks/{check_id}", h.handleUpdateCheck)
		r.Delete("/projects/{ref}/checks/{check_id}", h.handleDeleteCheck)
	})
}

func (h *Handler) handleListChecks(w http.ResponseWriter, r *http.Request) {
	output, err := h.listChecks(r.Context(), &listChecksInput{Ref: httpx.Path(r, "ref")})
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func (h *Handler) handleCreateCheck(w http.ResponseWriter, r *http.Request) {
	var body createCheckInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.createCheck(r.Context(), &createCheckInput{Ref: httpx.Path(r, "ref"), Body: body})
	writeCheckOutput(w, r, http.StatusCreated, output, err)
}

func (h *Handler) handleGetCheck(w http.ResponseWriter, r *http.Request) {
	output, err := h.getCheck(r.Context(), &getCheckInput{Ref: httpx.Path(r, "ref"), CheckID: httpx.Path(r, "check_id")})
	writeCheckOutput(w, r, http.StatusOK, output, err)
}

func (h *Handler) handleUpdateCheck(w http.ResponseWriter, r *http.Request) {
	var body updateCheckInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.updateCheck(r.Context(), &updateCheckInput{Ref: httpx.Path(r, "ref"), CheckID: httpx.Path(r, "check_id"), Body: body})
	writeCheckOutput(w, r, http.StatusOK, output, err)
}

func (h *Handler) handleDeleteCheck(w http.ResponseWriter, r *http.Request) {
	_, err := h.deleteCheck(r.Context(), &deleteCheckInput{Ref: httpx.Path(r, "ref"), CheckID: httpx.Path(r, "check_id")})
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteNoContent(w)
}

func writeCheckOutput(w http.ResponseWriter, r *http.Request, status int, output *checkOutput, err error) {
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, status, output.Body)
}

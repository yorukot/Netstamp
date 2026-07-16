package probe

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	appprobe "github.com/yorukot/netstamp/internal/controller/application/probe"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

type Handler struct {
	service       *appprobe.Service
	verifier      appauth.SessionManager
	cookieName    string
	tokenVerifier httpmiddleware.APITokenVerifier
}

func NewHandler(service *appprobe.Service, verifier appauth.SessionManager, cookieName string, tokenVerifiers ...httpmiddleware.APITokenVerifier) *Handler {
	var tokenVerifier httpmiddleware.APITokenVerifier
	if len(tokenVerifiers) > 0 {
		tokenVerifier = tokenVerifiers[0]
	}
	return &Handler{
		service:       service,
		verifier:      verifier,
		cookieName:    cookieName,
		tokenVerifier: tokenVerifier,
	}
}

func (h *Handler) RegisterRoutes(api chi.Router) {
	api.Group(func(r chi.Router) {
		r.Use(httpmiddleware.RequireUserAuth(h.verifier, h.tokenVerifier, h.cookieName))

		r.With(httpmiddleware.RequireScope(identity.ScopeProbesRead)).Get("/projects/{ref}/probes", h.handleListProbes)
		r.With(httpmiddleware.RequireScope(identity.ScopeProbesWrite)).Post("/projects/{ref}/probes", h.handleCreateProbe)
		r.With(httpmiddleware.RequireScope(identity.ScopeProbesRead)).Get("/projects/{ref}/probes/{probe_id}", h.handleGetProbe)
		r.With(httpmiddleware.RequireScope(identity.ScopeProbesWrite)).Patch("/projects/{ref}/probes/{probe_id}", h.handleUpdateProbe)
		r.With(httpmiddleware.RequireScope(identity.ScopeProbesWrite)).Delete("/projects/{ref}/probes/{probe_id}", h.handleDeleteProbe)
		r.With(httpmiddleware.RequireScope(identity.ScopeProbesWrite)).Post("/projects/{ref}/probes/{probe_id}/secret-rotations", h.handleRotateSecret)
	})
}

func (h *Handler) handleListProbes(w http.ResponseWriter, r *http.Request) {
	output, err := h.listProbes(r.Context(), &listProbesInput{Ref: httpx.Path(r, "ref")})
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func (h *Handler) handleCreateProbe(w http.ResponseWriter, r *http.Request) {
	var body createProbeInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.createProbe(r.Context(), &createProbeInput{Ref: httpx.Path(r, "ref"), Body: body})
	writeCreateProbeOutput(w, r, output, err)
}

func (h *Handler) handleGetProbe(w http.ResponseWriter, r *http.Request) {
	output, err := h.getProbe(r.Context(), newProbeRefInput(r))
	writeProbeOutput(w, r, output, err)
}

func (h *Handler) handleUpdateProbe(w http.ResponseWriter, r *http.Request) {
	var body updateProbeInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.updateProbe(r.Context(), &updateProbeInput{
		Ref:     httpx.Path(r, "ref"),
		ProbeID: httpx.Path(r, "probe_id"),
		Body:    body,
	})
	writeProbeOutput(w, r, output, err)
}

func (h *Handler) handleDeleteProbe(w http.ResponseWriter, r *http.Request) {
	_, err := h.deleteProbe(r.Context(), newProbeRefInput(r))
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) handleRotateSecret(w http.ResponseWriter, r *http.Request) {
	output, err := h.rotateSecret(r.Context(), newProbeRefInput(r))
	writeRotateSecretOutput(w, r, output, err)
}

func newProbeRefInput(r *http.Request) *probeRefInput {
	return &probeRefInput{
		Ref:     httpx.Path(r, "ref"),
		ProbeID: httpx.Path(r, "probe_id"),
	}
}

func writeCreateProbeOutput(w http.ResponseWriter, r *http.Request, output *createProbeOutput, err error) {
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, output.Body)
}

func writeProbeOutput(w http.ResponseWriter, r *http.Request, output *probeOutput, err error) {
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func writeRotateSecretOutput(w http.ResponseWriter, r *http.Request, output *rotateSecretOutput, err error) {
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

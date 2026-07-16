package label

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	applabel "github.com/yorukot/netstamp/internal/controller/application/label"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

type Handler struct {
	service       *applabel.Service
	verifier      appauth.SessionManager
	cookieName    string
	tokenVerifier httpmiddleware.APITokenVerifier
}

func NewHandler(service *applabel.Service, verifier appauth.SessionManager, cookieName string, tokenVerifiers ...httpmiddleware.APITokenVerifier) *Handler {
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

		r.With(httpmiddleware.RequireScope(identity.ScopeLabelsRead)).Get("/projects/{ref}/labels", h.handleListLabels)
		r.With(httpmiddleware.RequireScope(identity.ScopeLabelsWrite)).Post("/projects/{ref}/labels", h.handleCreateLabel)
		r.With(httpmiddleware.RequireScope(identity.ScopeLabelsWrite)).Patch("/projects/{ref}/labels/{label_id}", h.handleUpdateLabel)
		r.With(httpmiddleware.RequireScope(identity.ScopeLabelsWrite)).Delete("/projects/{ref}/labels/{label_id}", h.handleDeleteLabel)
	})
}

func (h *Handler) handleListLabels(w http.ResponseWriter, r *http.Request) {
	output, err := h.listLabels(r.Context(), &listLabelsInput{Ref: httpx.Path(r, "ref")})
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func (h *Handler) handleCreateLabel(w http.ResponseWriter, r *http.Request) {
	var body createLabelInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.createLabel(r.Context(), &createLabelInput{Ref: httpx.Path(r, "ref"), Body: body})
	writeLabelOutput(w, r, http.StatusCreated, output, err)
}

func (h *Handler) handleUpdateLabel(w http.ResponseWriter, r *http.Request) {
	var body updateLabelInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.updateLabel(r.Context(), &updateLabelInput{Ref: httpx.Path(r, "ref"), LabelID: httpx.Path(r, "label_id"), Body: body})
	writeLabelOutput(w, r, http.StatusOK, output, err)
}

func (h *Handler) handleDeleteLabel(w http.ResponseWriter, r *http.Request) {
	_, err := h.deleteLabel(r.Context(), &labelRefInput{Ref: httpx.Path(r, "ref"), LabelID: httpx.Path(r, "label_id")})
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteNoContent(w)
}

func writeLabelOutput(w http.ResponseWriter, r *http.Request, status int, output *labelOutput, err error) {
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, status, output.Body)
}

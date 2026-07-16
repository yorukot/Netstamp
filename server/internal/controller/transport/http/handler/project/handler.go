package project

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	appproject "github.com/yorukot/netstamp/internal/controller/application/project"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

type Handler struct {
	service       *appproject.Service
	verifier      appauth.SessionManager
	cookieName    string
	tokenVerifier httpmiddleware.APITokenVerifier
}

func NewHandler(service *appproject.Service, verifier appauth.SessionManager, cookieName string, tokenVerifiers ...httpmiddleware.APITokenVerifier) *Handler {
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

		r.With(httpmiddleware.RequireScope(identity.ScopeProjectsWrite)).Post("/projects", h.handleCreateProject)
		r.With(httpmiddleware.RequireScope(identity.ScopeProjectsRead)).Get("/projects", h.handleListProjects)
		r.With(httpmiddleware.RequireScope(identity.ScopeProjectsRead)).Get("/projects/{ref}", h.handleGetProject)
		r.With(httpmiddleware.RequireScope(identity.ScopeProjectsWrite)).Patch("/projects/{ref}", h.handleUpdateProject)
		r.With(httpmiddleware.RequireScope(identity.ScopeProjectsWrite)).Delete("/projects/{ref}", h.handleDeleteProject)
	})
	api.Group(func(r chi.Router) {
		r.Use(httpmiddleware.RequireAuth(h.verifier, h.cookieName))
		r.Get("/projects/{ref}/members", h.handleListMembers)
		r.Patch("/projects/{ref}/members/{user_id}", h.handleUpdateMemberRole)
		r.Delete("/projects/{ref}/members/{user_id}", h.handleRemoveMember)
		r.Post("/projects/{ref}/invites", h.handleCreateInvite)
		r.Get("/projects/{ref}/invites", h.handleListProjectInvites)
		r.Delete("/projects/{ref}/invites/{invite_id}", h.handleCancelInvite)
		r.Get("/me/project-invites", h.handleListUserInvites)
		r.Post("/me/project-invites/{invite_id}/accept", h.handleAcceptInvite)
		r.Post("/me/project-invites/{invite_id}/reject", h.handleRejectInvite)
	})
}

func (h *Handler) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var body createProjectInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.createProject(r.Context(), &createProjectInput{Body: body})
	writeProjectOutput(w, r, http.StatusCreated, output, err)
}

func (h *Handler) handleListProjects(w http.ResponseWriter, r *http.Request) {
	output, err := h.listProjects(r.Context(), &listProjectsInput{})
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func (h *Handler) handleGetProject(w http.ResponseWriter, r *http.Request) {
	output, err := h.getProject(r.Context(), &projectRefInput{Ref: httpx.Path(r, "ref")})
	writeProjectOutput(w, r, http.StatusOK, output, err)
}

func (h *Handler) handleUpdateProject(w http.ResponseWriter, r *http.Request) {
	var body updateProjectInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.updateProject(r.Context(), &updateProjectInput{Ref: httpx.Path(r, "ref"), Body: body})
	writeProjectOutput(w, r, http.StatusOK, output, err)
}

func (h *Handler) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	_, err := h.deleteProject(r.Context(), &projectRefInput{Ref: httpx.Path(r, "ref")})
	writeNoContent(w, r, err)
}

func (h *Handler) handleListMembers(w http.ResponseWriter, r *http.Request) {
	output, err := h.listMembers(r.Context(), &projectRefInput{Ref: httpx.Path(r, "ref")})
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func (h *Handler) handleCreateInvite(w http.ResponseWriter, r *http.Request) {
	var body createInviteInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.createInvite(r.Context(), &createInviteInput{Ref: httpx.Path(r, "ref"), Body: body})
	writeInviteOutput(w, r, http.StatusCreated, output, err)
}

func (h *Handler) handleUpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	var body updateMemberRoleInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.updateMemberRole(r.Context(), &updateMemberRoleInput{Ref: httpx.Path(r, "ref"), UserID: httpx.Path(r, "user_id"), Body: body})
	writeMemberOutput(w, r, http.StatusOK, output, err)
}

func (h *Handler) handleRemoveMember(w http.ResponseWriter, r *http.Request) {
	_, err := h.removeMember(r.Context(), &removeMemberInput{Ref: httpx.Path(r, "ref"), UserID: httpx.Path(r, "user_id")})
	writeNoContent(w, r, err)
}

func writeProjectOutput(w http.ResponseWriter, r *http.Request, status int, output *projectOutput, err error) {
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, status, output.Body)
}

func writeMemberOutput(w http.ResponseWriter, r *http.Request, status int, output *memberOutput, err error) {
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, status, output.Body)
}

func writeInviteOutput(w http.ResponseWriter, r *http.Request, status int, output *inviteOutput, err error) {
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, status, output.Body)
}

func writeNoContent(w http.ResponseWriter, r *http.Request, err error) {
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteNoContent(w)
}

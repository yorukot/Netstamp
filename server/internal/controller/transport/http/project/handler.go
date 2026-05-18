package project

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	appproject "github.com/yorukot/netstamp/internal/controller/application/project"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
)

type Handler struct {
	service  *appproject.Service
	verifier appauth.TokenVerifier
}

func NewHandler(service *appproject.Service, verifier appauth.TokenVerifier) *Handler {
	return &Handler{
		service:  service,
		verifier: verifier,
	}
}

func (h *Handler) RegisterRoutes(api chi.Router) {
	api.With(httpmiddleware.RequireAuth(h.verifier)).Post("/projects", h.handleCreateProject)
	api.With(httpmiddleware.RequireAuth(h.verifier)).Get("/projects", h.handleListProjects)
	api.With(httpmiddleware.RequireAuth(h.verifier)).Get("/projects/{ref}", h.handleGetProject)
	api.With(httpmiddleware.RequireAuth(h.verifier)).Patch("/projects/{ref}", h.handleUpdateProject)
	api.With(httpmiddleware.RequireAuth(h.verifier)).Delete("/projects/{ref}", h.handleDeleteProject)
	api.With(httpmiddleware.RequireAuth(h.verifier)).Get("/projects/{ref}/members", h.handleListMembers)
	api.With(httpmiddleware.RequireAuth(h.verifier)).Post("/projects/{ref}/members", h.handleAddMember)
	api.With(httpmiddleware.RequireAuth(h.verifier)).Patch("/projects/{ref}/members/{user_id}", h.handleUpdateMemberRole)
	api.With(httpmiddleware.RequireAuth(h.verifier)).Delete("/projects/{ref}/members/{user_id}", h.handleRemoveMember)
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

func (h *Handler) handleAddMember(w http.ResponseWriter, r *http.Request) {
	var body addMemberInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.addMember(r.Context(), &addMemberInput{Ref: httpx.Path(r, "ref"), Body: body})
	writeMemberOutput(w, r, http.StatusCreated, output, err)
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

func writeNoContent(w http.ResponseWriter, r *http.Request, err error) {
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteNoContent(w)
}

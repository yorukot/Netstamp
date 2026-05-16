package project

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	appproject "github.com/yorukot/netstamp/internal/controller/application/project"
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

func (h *Handler) RegisterRoutes(api huma.API) {
	authMiddleware := httpmiddleware.RequireAuth(h.verifier)
	security := []map[string][]string{{httpmiddleware.SessionCookieSecurityScheme: {}}}
	middlewares := huma.Middlewares{authMiddleware}

	huma.Register(api, huma.Operation{
		OperationID:   "createProject",
		Method:        http.MethodPost,
		Path:          "/projects",
		DefaultStatus: http.StatusCreated,
		Summary:       "Create project",
		Tags:          []string{"Projects"},
		Security:      security,
		Middlewares:   middlewares,
		Errors:        []int{http.StatusUnauthorized, http.StatusConflict, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.createProject)

	huma.Register(api, huma.Operation{
		OperationID: "listProjects",
		Method:      http.MethodGet,
		Path:        "/projects",
		Summary:     "List projects",
		Tags:        []string{"Projects"},
		Security:    security,
		Middlewares: middlewares,
		Errors:      []int{http.StatusUnauthorized, http.StatusInternalServerError},
	}, h.listProjects)

	huma.Register(api, huma.Operation{
		OperationID: "getProject",
		Method:      http.MethodGet,
		Path:        "/projects/{ref}",
		Summary:     "Get project",
		Tags:        []string{"Projects"},
		Security:    security,
		Middlewares: middlewares,
		Errors:      []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusInternalServerError},
	}, h.getProject)

	huma.Register(api, huma.Operation{
		OperationID: "updateProject",
		Method:      http.MethodPatch,
		Path:        "/projects/{ref}",
		Summary:     "Update project",
		Tags:        []string{"Projects"},
		Security:    security,
		Middlewares: middlewares,
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.updateProject)

	huma.Register(api, huma.Operation{
		OperationID:   "deleteProject",
		Method:        http.MethodDelete,
		Path:          "/projects/{ref}",
		DefaultStatus: http.StatusNoContent,
		Summary:       "Delete project",
		Tags:          []string{"Projects"},
		Security:      security,
		Middlewares:   middlewares,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusInternalServerError},
	}, h.deleteProject)

	huma.Register(api, huma.Operation{
		OperationID: "listProjectMembers",
		Method:      http.MethodGet,
		Path:        "/projects/{ref}/members",
		Summary:     "List project members",
		Tags:        []string{"Project Members"},
		Security:    security,
		Middlewares: middlewares,
		Errors:      []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusInternalServerError},
	}, h.listMembers)

	huma.Register(api, huma.Operation{
		OperationID:   "addProjectMember",
		Method:        http.MethodPost,
		Path:          "/projects/{ref}/members",
		DefaultStatus: http.StatusCreated,
		Summary:       "Add project member",
		Tags:          []string{"Project Members"},
		Security:      security,
		Middlewares:   middlewares,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.addMember)

	huma.Register(api, huma.Operation{
		OperationID: "updateProjectMemberRole",
		Method:      http.MethodPatch,
		Path:        "/projects/{ref}/members/{user_id}",
		Summary:     "Update project member role",
		Tags:        []string{"Project Members"},
		Security:    security,
		Middlewares: middlewares,
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.updateMemberRole)

	huma.Register(api, huma.Operation{
		OperationID:   "removeProjectMember",
		Method:        http.MethodDelete,
		Path:          "/projects/{ref}/members/{user_id}",
		DefaultStatus: http.StatusNoContent,
		Summary:       "Remove project member",
		Tags:          []string{"Project Members"},
		Security:      security,
		Middlewares:   middlewares,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.removeMember)
}

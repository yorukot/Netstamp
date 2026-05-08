package auth

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	appauth "github.com/yorukot/netstamp/internal/application/auth"
	httpmiddleware "github.com/yorukot/netstamp/internal/transport/http/middleware"
)

type Handler struct {
	service  *appauth.Service
	verifier appauth.TokenVerifier
}

func NewHandler(service *appauth.Service, verifier appauth.TokenVerifier) *Handler {
	return &Handler{
		service:  service,
		verifier: verifier,
	}
}

func (h *Handler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID:   "registerUser",
		Method:        http.MethodPost,
		Path:          "/auth/register",
		DefaultStatus: http.StatusCreated,
		Summary:       "Register user",
		Tags:          []string{"Auth"},
		Errors:        []int{http.StatusUnprocessableEntity, http.StatusConflict, http.StatusInternalServerError},
	}, h.register)

	huma.Register(api, huma.Operation{
		OperationID: "loginUser",
		Method:      http.MethodPost,
		Path:        "/auth/login",
		Summary:     "Login user",
		Tags:        []string{"Auth"},
		Errors:      []int{http.StatusUnauthorized, http.StatusInternalServerError},
	}, h.login)

	huma.Register(api, huma.Operation{
		OperationID: "getCurrentUser",
		Method:      http.MethodGet,
		Path:        "/auth/me",
		Summary:     "Get current user",
		Tags:        []string{"Auth"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{
			httpmiddleware.RequireAuth(h.verifier),
		},
		Errors: []int{http.StatusUnauthorized, http.StatusInternalServerError},
	}, h.me)
}

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
		Description:   "Create a user account with a normalized email address, display name, and password. On success, returns the created user and a bearer access token for immediate API access.",
		Tags:          []string{"Auth"},
		Errors:        []int{http.StatusUnprocessableEntity, http.StatusConflict, http.StatusInternalServerError},
	}, h.register)

	huma.Register(api, huma.Operation{
		OperationID: "loginUser",
		Method:      http.MethodPost,
		Path:        "/auth/login",
		Summary:     "Login user",
		Description: "Verify an email and password pair, then return the authenticated user and a bearer access token. Invalid credentials always return the same unauthorized response so callers cannot distinguish an unknown email from a password mismatch.",
		Tags:        []string{"Auth"},
		Errors:      []int{http.StatusUnauthorized, http.StatusInternalServerError},
	}, h.login)

	huma.Register(api, huma.Operation{
		OperationID: "getCurrentUser",
		Method:      http.MethodGet,
		Path:        "/auth/me",
		Summary:     "Get current user",
		Description: "Return the user identity embedded in a valid bearer access token. The request must include an Authorization header using the Bearer scheme.",
		Tags:        []string{"Auth"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{
			httpmiddleware.RequireAuth(h.verifier),
		},
		Errors: []int{http.StatusUnauthorized, http.StatusInternalServerError},
	}, h.me)
}

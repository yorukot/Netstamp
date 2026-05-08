package probe

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	appauth "github.com/yorukot/netstamp/internal/application/auth"
	appprobe "github.com/yorukot/netstamp/internal/application/probe"
	httpmiddleware "github.com/yorukot/netstamp/internal/transport/http/middleware"
)

type Handler struct {
	service  *appprobe.Service
	verifier appauth.TokenVerifier
}

func NewHandler(service *appprobe.Service, verifier appauth.TokenVerifier) *Handler {
	return &Handler{
		service:  service,
		verifier: verifier,
	}
}

func (h *Handler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID:   "createProjectProbe",
		Method:        http.MethodPost,
		Path:          "/projects/{ref}/probes",
		DefaultStatus: http.StatusCreated,
		Summary:       "Create project probe",
		Tags:          []string{"Probes"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		Middlewares:   huma.Middlewares{httpmiddleware.RequireAuth(h.verifier)},
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.createProbe)
}

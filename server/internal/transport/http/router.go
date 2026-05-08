package httpserver

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"

	appauth "github.com/yorukot/netstamp/internal/application/auth"
	appprobe "github.com/yorukot/netstamp/internal/application/probe"
	appproject "github.com/yorukot/netstamp/internal/application/project"
	"github.com/yorukot/netstamp/internal/observability/httptrace"
	authhttp "github.com/yorukot/netstamp/internal/transport/http/auth"
	httpmiddleware "github.com/yorukot/netstamp/internal/transport/http/middleware"
	probehttp "github.com/yorukot/netstamp/internal/transport/http/probe"
	projecthttp "github.com/yorukot/netstamp/internal/transport/http/project"
)

type Dependencies struct {
	Log            *zap.Logger
	APIVersion     string
	BackendBaseURL string
	AuthService    *appauth.Service
	AuthVerifier   appauth.TokenVerifier
	ProbeService   *appprobe.Service
	ProjectService *appproject.Service
	ReadinessCheck func(context.Context) error
	RequestTimeout time.Duration
}

func NewRouter(dep Dependencies) http.Handler {
	if dep.Log == nil {
		dep.Log = zap.NewNop()
	}

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(otelhttp.NewMiddleware("http.server",
		otelhttp.WithSpanNameFormatter(httptrace.RequestSpanName),
	))
	r.Use(httpmiddleware.ZapRecoverer(dep.Log))
	r.Use(chimw.Timeout(dep.RequestTimeout))
	r.Use(httpmiddleware.ZapRequestLogger(dep.Log))

	r.Route(dep.basePath(), func(apiRouter chi.Router) {
		api := humachi.New(apiRouter, newHumaConfig(dep))
		registerAPIRoutes(api, dep)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	return r
}

func NewOpenAPI(dep Dependencies) *huma.OpenAPI {
	api := humachi.New(chi.NewRouter(), newHumaConfig(dep))
	registerAPIRoutes(api, dep)
	return api.OpenAPI()
}

func registerAPIRoutes(api huma.API, dep Dependencies) {
	registerSystemRoutes(api, dep.ReadinessCheck)

	if dep.AuthService != nil {
		authhttp.NewHandler(dep.AuthService, dep.AuthVerifier).RegisterRoutes(api)
	}
	if dep.ProjectService != nil {
		projecthttp.NewHandler(dep.ProjectService, dep.AuthVerifier).RegisterRoutes(api)
	}
	if dep.ProbeService != nil {
		probehttp.NewHandler(dep.ProbeService, dep.AuthVerifier).RegisterRoutes(api)
	}
}

func newHumaConfig(dep Dependencies) huma.Config {
	config := huma.DefaultConfig("Netstamp API", dep.APIVersion)
	config.Info.Description = "Controller HTTP API for Netstamp."
	config.Servers = []*huma.Server{{URL: dep.serverURL()}}
	if config.Components.SecuritySchemes == nil {
		config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{}
	}
	config.Components.SecuritySchemes["bearerAuth"] = &huma.SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
	}
	return config
}

func (d *Dependencies) basePath() string {
	return "/api/" + d.APIVersion
}

func (d *Dependencies) serverURL() string {
	backendBaseURL := strings.TrimRight(strings.TrimSpace(d.BackendBaseURL), "/")
	if backendBaseURL == "" {
		return d.basePath()
	}
	return backendBaseURL + d.basePath()
}

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
	appcheck "github.com/yorukot/netstamp/internal/application/check"
	applabel "github.com/yorukot/netstamp/internal/application/label"
	appprobe "github.com/yorukot/netstamp/internal/application/probe"
	appproberuntime "github.com/yorukot/netstamp/internal/application/proberuntime"
	appproject "github.com/yorukot/netstamp/internal/application/project"
	httptracing "github.com/yorukot/netstamp/internal/observability/httptrace"
	authhttp "github.com/yorukot/netstamp/internal/transport/http/auth"
	checkhttp "github.com/yorukot/netstamp/internal/transport/http/check"
	labelhttp "github.com/yorukot/netstamp/internal/transport/http/label"
	httpmiddleware "github.com/yorukot/netstamp/internal/transport/http/middleware"
	probehttp "github.com/yorukot/netstamp/internal/transport/http/probe"
	proberuntimehttp "github.com/yorukot/netstamp/internal/transport/http/proberuntime"
	projecthttp "github.com/yorukot/netstamp/internal/transport/http/project"
)

type Dependencies struct {
	Log            *zap.Logger
	APIVersion     string
	BackendBaseURL string
	AuthService    *appauth.Service
	AuthVerifier   appauth.TokenVerifier
	CheckService   *appcheck.Service
	LabelService   *applabel.Service
	ProbeService   *appprobe.Service
	ProbeRuntime   *appproberuntime.Service
	ProjectService *appproject.Service
	ReadinessCheck func(context.Context) error
	RequestTimeout time.Duration
	MetricsHandler http.Handler
}

func NewRouter(dep Dependencies) http.Handler {
	if dep.Log == nil {
		dep.Log = zap.NewNop()
	}

	apiRouter := newAPIRouter(dep)
	if dep.MetricsHandler == nil {
		return apiRouter
	}

	return routeMetrics(apiRouter, dep.MetricsHandler)
}

func newAPIRouter(dep Dependencies) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(otelhttp.NewMiddleware("http.server",
		otelhttp.WithSpanNameFormatter(httptracing.RequestSpanName),
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

func routeMetrics(apiRouter http.Handler, metricsHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			metricsHandler.ServeHTTP(w, r)
			return
		}

		apiRouter.ServeHTTP(w, r)
	})
}

func registerAPIRoutes(api huma.API, dep Dependencies) {
	registerSystemRoutes(api, dep.ReadinessCheck)

	authhttp.NewHandler(dep.AuthService, dep.AuthVerifier).RegisterRoutes(api)
	projecthttp.NewHandler(dep.ProjectService, dep.AuthVerifier).RegisterRoutes(api)
	labelhttp.NewHandler(dep.LabelService, dep.AuthVerifier).RegisterRoutes(api)
	checkhttp.NewHandler(dep.CheckService, dep.AuthVerifier).RegisterRoutes(api)
	probehttp.NewHandler(dep.ProbeService, dep.AuthVerifier).RegisterRoutes(api)
	proberuntimehttp.NewHandler(dep.ProbeRuntime).RegisterRoutes(api)
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
	config.Components.SecuritySchemes["probeAuth"] = &huma.SecurityScheme{
		Type:        "apiKey",
		In:          "header",
		Name:        "Authorization",
		Description: "Probe runtime credential header using the format `Probe <secret>`.",
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

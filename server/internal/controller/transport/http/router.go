package httpserver

import (
	"context"
	"net/http"
	"net/netip"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"

	appassignment "github.com/yorukot/netstamp/internal/controller/application/assignment"
	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	appcheck "github.com/yorukot/netstamp/internal/controller/application/check"
	applabel "github.com/yorukot/netstamp/internal/controller/application/label"
	appprobe "github.com/yorukot/netstamp/internal/controller/application/probe"
	appproberuntime "github.com/yorukot/netstamp/internal/controller/application/proberuntime"
	appproject "github.com/yorukot/netstamp/internal/controller/application/project"
	appresult "github.com/yorukot/netstamp/internal/controller/application/result"
	appuser "github.com/yorukot/netstamp/internal/controller/application/user"
	"github.com/yorukot/netstamp/internal/controller/transport/http/clientip"
	assignmenthttp "github.com/yorukot/netstamp/internal/controller/transport/http/handler/assignment"
	authhttp "github.com/yorukot/netstamp/internal/controller/transport/http/handler/auth"
	checkhttp "github.com/yorukot/netstamp/internal/controller/transport/http/handler/check"
	installhttp "github.com/yorukot/netstamp/internal/controller/transport/http/handler/install"
	labelhttp "github.com/yorukot/netstamp/internal/controller/transport/http/handler/label"
	probehttp "github.com/yorukot/netstamp/internal/controller/transport/http/handler/probe"
	proberuntimehttp "github.com/yorukot/netstamp/internal/controller/transport/http/handler/proberuntime"
	projecthttp "github.com/yorukot/netstamp/internal/controller/transport/http/handler/project"
	resulthttp "github.com/yorukot/netstamp/internal/controller/transport/http/handler/result"
	userhttp "github.com/yorukot/netstamp/internal/controller/transport/http/handler/user"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	"github.com/yorukot/netstamp/internal/controller/transport/http/openapi"
	httptracing "github.com/yorukot/netstamp/internal/platform/observability/httptrace"
)

type Dependencies struct {
	Log               *zap.Logger
	APIVersion        string
	BackendBaseURL    string
	AuthService       *appauth.Service
	AuthVerifier      appauth.TokenVerifier
	AuthCookieSecure  bool
	UserService       *appuser.Service
	AssignmentService *appassignment.Service
	CheckService      *appcheck.Service
	LabelService      *applabel.Service
	ProbeService      *appprobe.Service
	ProbeRuntime      *appproberuntime.Service
	ProjectService    *appproject.Service
	ResultService     *appresult.Service
	ReadinessCheck    func(context.Context) error
	RequestTimeout    time.Duration
	MetricsHandler    http.Handler
	AgentBinaryDir    string
	TrustedProxies    []netip.Prefix
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
	r.Use(clientip.Middleware(dep.TrustedProxies))
	r.Use(chimw.StripSlashes)
	r.Use(otelhttp.NewMiddleware("http.server",
		otelhttp.WithSpanNameFormatter(httptracing.RequestSpanName),
	))
	r.Use(httpmiddleware.ZapRecoverer(dep.Log))
	r.Use(chimw.Timeout(dep.RequestTimeout))
	r.Use(httpmiddleware.ZapRequestLogger(dep.Log))

	r.Route(dep.basePath(), func(apiRouter chi.Router) {
		registerAPIRoutes(apiRouter, dep)
	})

	r.NotFound(writeNotFound)
	r.MethodNotAllowed(writeMethodNotAllowed)

	return r
}

func routeMetrics(apiRouter, metricsHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			metricsHandler.ServeHTTP(w, r)
			return
		}

		apiRouter.ServeHTTP(w, r)
	})
}

func registerAPIRoutes(api chi.Router, dep Dependencies) {
	registerSystemRoutes(api, dep.ReadinessCheck)
	registerOpenAPIRoutes(api, dep)

	installhttp.NewHandler(dep.AgentBinaryDir).RegisterRoutes(api)

	authhttp.NewHandler(dep.AuthService, dep.AuthVerifier, dep.AuthCookieSecure).RegisterRoutes(api)
	userhttp.NewHandler(dep.UserService, dep.AuthVerifier).RegisterRoutes(api)
	projecthttp.NewHandler(dep.ProjectService, dep.AuthVerifier).RegisterRoutes(api)
	assignmenthttp.NewHandler(dep.AssignmentService, dep.AuthVerifier).RegisterRoutes(api)
	labelhttp.NewHandler(dep.LabelService, dep.AuthVerifier).RegisterRoutes(api)
	checkhttp.NewHandler(dep.CheckService, dep.AuthVerifier).RegisterRoutes(api)
	probehttp.NewHandler(dep.ProbeService, dep.AuthVerifier).RegisterRoutes(api)
	resulthttp.NewHandler(dep.ResultService, dep.AuthVerifier).RegisterRoutes(api)
	proberuntimehttp.NewHandler(dep.ProbeRuntime).RegisterRoutes(api)
}

func registerOpenAPIRoutes(api chi.Router, dep Dependencies) {
	api.Get("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		data, err := openapi.Spec(dep.APIVersion, dep.BackendBaseURL)
		if err != nil {
			httpmiddleware.WriteProblem(w, r, http.StatusInternalServerError, "openapi unavailable")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(data); err != nil {
			return
		}
	})

	api.Get("/docs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(openapi.ScalarHTML(dep.basePath() + "/openapi.json")); err != nil {
			return
		}
	})
}

func (d *Dependencies) basePath() string {
	return "/api/" + d.APIVersion
}

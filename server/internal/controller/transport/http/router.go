package httpserver

import (
	"context"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"

	appadmin "github.com/yorukot/netstamp/internal/controller/application/admin"
	appalert "github.com/yorukot/netstamp/internal/controller/application/alert"
	appassignment "github.com/yorukot/netstamp/internal/controller/application/assignment"
	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	appcheck "github.com/yorukot/netstamp/internal/controller/application/check"
	applabel "github.com/yorukot/netstamp/internal/controller/application/label"
	appprobe "github.com/yorukot/netstamp/internal/controller/application/probe"
	appproberuntime "github.com/yorukot/netstamp/internal/controller/application/proberuntime"
	appproject "github.com/yorukot/netstamp/internal/controller/application/project"
	apppublicstatus "github.com/yorukot/netstamp/internal/controller/application/publicstatus"
	appresult "github.com/yorukot/netstamp/internal/controller/application/result"
	appuser "github.com/yorukot/netstamp/internal/controller/application/user"
	"github.com/yorukot/netstamp/internal/controller/transport/http/clientip"
	adminhttp "github.com/yorukot/netstamp/internal/controller/transport/http/handler/admin"
	alerthttp "github.com/yorukot/netstamp/internal/controller/transport/http/handler/alert"
	assignmenthttp "github.com/yorukot/netstamp/internal/controller/transport/http/handler/assignment"
	authhttp "github.com/yorukot/netstamp/internal/controller/transport/http/handler/auth"
	checkhttp "github.com/yorukot/netstamp/internal/controller/transport/http/handler/check"
	installhttp "github.com/yorukot/netstamp/internal/controller/transport/http/handler/install"
	labelhttp "github.com/yorukot/netstamp/internal/controller/transport/http/handler/label"
	probehttp "github.com/yorukot/netstamp/internal/controller/transport/http/handler/probe"
	proberuntimehttp "github.com/yorukot/netstamp/internal/controller/transport/http/handler/proberuntime"
	projecthttp "github.com/yorukot/netstamp/internal/controller/transport/http/handler/project"
	publicstatushttp "github.com/yorukot/netstamp/internal/controller/transport/http/handler/publicstatus"
	resulthttp "github.com/yorukot/netstamp/internal/controller/transport/http/handler/result"
	userhttp "github.com/yorukot/netstamp/internal/controller/transport/http/handler/user"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	"github.com/yorukot/netstamp/internal/controller/transport/http/openapi"
	httptracing "github.com/yorukot/netstamp/internal/platform/observability/httptrace"
)

type Dependencies struct {
	Log                         *zap.Logger
	APIVersion                  string
	DemoMode                    bool
	BackendBaseURL              string
	PublicWebBaseURL            string
	WebDir                      string
	AuthService                 *appauth.Service
	AuthVerifier                appauth.SessionManager
	AdminService                *appadmin.Service
	AuthCookieName              string
	AuthCookieSecure            bool
	AuthRegistrationDisabled    bool
	AuthPasswordResetRateWindow time.Duration
	AuthPasswordResetIPLimit    int32
	AuthPasswordResetEmailLimit int32
	UserService                 *appuser.Service
	AlertService                *appalert.Service
	AssignmentService           *appassignment.Service
	CheckService                *appcheck.Service
	LabelService                *applabel.Service
	ProbeService                *appprobe.Service
	ProbeRuntime                *appproberuntime.Service
	ProjectService              *appproject.Service
	PublicStatusService         *apppublicstatus.Service
	ResultService               *appresult.Service
	ReadinessCheck              func(context.Context) error
	RequestTimeout              time.Duration
	MetricsHandler              http.Handler
	AgentBinaryDir              string
	TrustedProxies              []netip.Prefix
}

func NewRouter(dep Dependencies) http.Handler {
	if dep.Log == nil {
		dep.Log = zap.NewNop()
	}

	apiRouter := newAPIRouter(dep)
	handler := routeFrontend(apiRouter, dep)
	if dep.MetricsHandler == nil {
		return handler
	}

	return routeMetrics(handler, dep.MetricsHandler)
}

func newAPIRouter(dep Dependencies) http.Handler {
	r := chi.NewRouter()
	dep.AuthCookieName = effectiveAuthCookieName(dep.AuthCookieName)
	r.Use(chimw.RequestID)
	r.Use(clientip.Middleware(dep.TrustedProxies))
	r.Use(chimw.StripSlashes)
	r.Use(otelhttp.NewMiddleware("http.server",
		otelhttp.WithSpanNameFormatter(httptracing.RequestSpanName),
	))
	r.Use(httpmiddleware.ZapRecoverer(dep.Log))
	r.Use(chimw.Timeout(dep.RequestTimeout))
	r.Use(httpmiddleware.ZapRequestLogger(dep.Log))

	basePath := dep.basePath()
	r.Route(basePath, func(apiRouter chi.Router) {
		if dep.DemoMode {
			apiRouter.Use(httpmiddleware.ReadOnly(
				basePath+"/auth/login",
				basePath+"/auth/logout",
			))
		}
		apiRouter.Use(httpmiddleware.CSRF(httpmiddleware.CSRFConfig{
			Verifier:         dep.AuthVerifier,
			CookieName:       dep.AuthCookieName,
			BasePath:         basePath,
			BackendBaseURL:   dep.BackendBaseURL,
			PublicWebBaseURL: dep.PublicWebBaseURL,
		}))
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

	installHandler := installhttp.NewHandler(dep.AgentBinaryDir, dep.BackendBaseURL, dep.basePath())
	if dep.AdminService != nil {
		installHandler = installhttp.NewHandler(dep.AgentBinaryDir, dep.BackendBaseURL, dep.basePath(), dep.AdminService)
	}
	installHandler.RegisterRoutes(api)

	authhttp.NewHandler(dep.AuthService, dep.AuthVerifier, dep.AdminService, dep.AuthCookieName, dep.AuthCookieSecure, !dep.AuthRegistrationDisabled).
		ConfigurePasswordReset(dep.PublicWebBaseURL, authhttp.NewPasswordResetRateLimiter(authhttp.PasswordResetRateLimitConfig{
			Window:     dep.AuthPasswordResetRateWindow,
			IPLimit:    dep.AuthPasswordResetIPLimit,
			EmailLimit: dep.AuthPasswordResetEmailLimit,
		})).
		RegisterRoutes(api)
	adminhttp.NewHandler(dep.AdminService, dep.AuthVerifier, dep.AuthCookieName).RegisterRoutes(api)
	userhttp.NewHandler(dep.UserService, dep.AuthVerifier, dep.AuthCookieName).RegisterRoutes(api)
	projecthttp.NewHandler(dep.ProjectService, dep.AuthVerifier, dep.AuthCookieName).RegisterRoutes(api)
	alerthttp.NewHandler(dep.AlertService, dep.AuthVerifier, dep.AuthCookieName, dep.AdminService).RegisterRoutes(api)
	assignmenthttp.NewHandler(dep.AssignmentService, dep.AuthVerifier, dep.AuthCookieName).RegisterRoutes(api)
	labelhttp.NewHandler(dep.LabelService, dep.AuthVerifier, dep.AuthCookieName).RegisterRoutes(api)
	checkhttp.NewHandler(dep.CheckService, dep.AuthVerifier, dep.AuthCookieName).RegisterRoutes(api)
	probehttp.NewHandler(dep.ProbeService, dep.AuthVerifier, dep.AuthCookieName).RegisterRoutes(api)
	publicstatushttp.NewHandler(dep.PublicStatusService, dep.AuthVerifier, dep.AuthCookieName).RegisterRoutes(api)
	resulthttp.NewHandler(dep.ResultService, dep.AuthVerifier, dep.AuthCookieName).RegisterRoutes(api)
	proberuntimehttp.NewHandler(dep.ProbeRuntime).RegisterRoutes(api)
}

func effectiveAuthCookieName(name string) string {
	if strings.TrimSpace(name) != "" {
		return name
	}
	return httpmiddleware.LocalSessionCookieName
}

func registerOpenAPIRoutes(api chi.Router, dep Dependencies) {
	api.Get("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		data, err := openapi.Spec(dep.APIVersion, effectiveBackendBaseURL(r.Context(), dep))
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

func effectiveBackendBaseURL(ctx context.Context, dep Dependencies) string {
	if dep.AdminService == nil {
		return dep.BackendBaseURL
	}
	value, err := dep.AdminService.BackendBaseURL(ctx)
	if err != nil {
		return dep.BackendBaseURL
	}
	value = strings.TrimRight(strings.TrimSpace(value), "/")
	if value == "" {
		return dep.BackendBaseURL
	}
	return value
}

func (d *Dependencies) basePath() string {
	return "/api/" + d.APIVersion
}

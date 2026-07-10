package middleware

import (
	"net/http"
	"net/url"
	"strings"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

const csrfHeaderName = "X-Csrf-Token"

type CSRFConfig struct {
	Verifier         appauth.SessionManager
	CookieName       string
	BasePath         string
	BackendBaseURL   string
	PublicWebBaseURL string
}

func CSRF(cfg CSRFConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if csrfSafeMethod(r.Method) || csrfSkippedPath(r.URL.Path, cfg.BasePath) {
				next.ServeHTTP(w, r)
				return
			}
			if cfg.Verifier == nil {
				WriteProblem(w, r, http.StatusInternalServerError, "csrf verifier unavailable")
				return
			}

			rawSessionToken, ok := sessionCookie(r, cfg.CookieName)
			if !ok {
				WriteProblemCode(w, r, http.StatusUnauthorized, httpx.CodeAuthMissingSession, "missing auth cookie")
				return
			}
			claims, err := cfg.Verifier.VerifySession(r.Context(), rawSessionToken)
			if err != nil {
				WriteProblemCode(w, r, http.StatusUnauthorized, httpx.CodeAuthInvalidSession, "invalid auth cookie")
				return
			}
			if !validCSRFOrigin(r, cfg) {
				WriteProblemCode(w, r, http.StatusForbidden, httpx.CodeAuthInvalidCSRF, "invalid csrf origin")
				return
			}
			if err := cfg.Verifier.VerifyCSRFToken(r.Context(), claims.SessionID, r.Header.Get(csrfHeaderName)); err != nil {
				WriteProblemCode(w, r, http.StatusForbidden, httpx.CodeAuthInvalidCSRF, "invalid csrf token")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func csrfSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}

func csrfSkippedPath(path, basePath string) bool {
	path = strings.TrimSuffix(path, "/")
	prefix := strings.TrimSuffix(basePath, "/")
	relative := strings.TrimPrefix(path, prefix)
	switch relative {
	case "/auth/register", "/auth/login", "/auth/password-resets", "/auth/email-verifications":
		return true
	default:
		return strings.HasPrefix(relative, "/runtime/")
	}
}

func validCSRFOrigin(r *http.Request, cfg CSRFConfig) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = r.Header.Get("Referer")
	}
	if origin == "" {
		return true
	}

	originURL, err := url.Parse(origin)
	if err != nil || originURL.Scheme == "" || originURL.Host == "" {
		return false
	}

	for _, allowed := range allowedCSRFOrigins(r, cfg) {
		if sameOrigin(originURL, allowed) {
			return true
		}
	}
	return false
}

func allowedCSRFOrigins(r *http.Request, cfg CSRFConfig) []*url.URL {
	values := []string{cfg.BackendBaseURL, cfg.PublicWebBaseURL}
	if r.Host != "" {
		scheme := "https"
		if r.TLS == nil {
			scheme = "http"
		}
		values = append(values, scheme+"://"+r.Host)
	}

	origins := make([]*url.URL, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		parsed, err := url.Parse(value)
		if err == nil && parsed.Scheme != "" && parsed.Host != "" {
			origins = append(origins, parsed)
		}
	}
	return origins
}

func sameOrigin(origin, allowed *url.URL) bool {
	return strings.EqualFold(origin.Scheme, allowed.Scheme) && strings.EqualFold(origin.Host, allowed.Host)
}

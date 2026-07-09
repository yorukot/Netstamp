package middleware

import (
	"net/http"

	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

func ReadOnly(allowedUnsafePaths ...string) func(http.Handler) http.Handler {
	allowedUnsafe := make(map[string]struct{}, len(allowedUnsafePaths))
	for _, path := range allowedUnsafePaths {
		allowedUnsafe[path] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isSafeMethod(r.Method) {
				next.ServeHTTP(w, r)
				return
			}
			if r.Method == http.MethodPost {
				if _, ok := allowedUnsafe[r.URL.Path]; ok {
					next.ServeHTTP(w, r)
					return
				}
			}

			WriteProblemCode(w, r, http.StatusForbidden, httpx.CodeReadOnly, "demo is read-only")
		})
	}
}

func isSafeMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}

package middleware

import "net/http"

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

			WriteProblem(w, r, http.StatusForbidden, "demo is read-only")
		})
	}
}

func isSafeMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}

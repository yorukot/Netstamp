package middleware

import (
	"context"
	"net/http"

	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/yorukot/netstamp/internal/logger"
)

func ZapRecoverer(root *zap.Logger) func(http.Handler) http.Handler {
	if root == nil {
		root = zap.NewNop()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCtx := r.Context()
			defer func(ctx context.Context) {
				if recovered := recover(); recovered != nil {
					requestID := chimw.GetReqID(ctx)
					if requestID == "" {
						requestID = r.Header.Get("X-Request-ID")
					}

					fields := []zap.Field{
						zap.String("request_id", requestID),
						zap.String("http.request.method", r.Method),
						zap.String("url.path", r.URL.Path),
						zap.String("client.address", clientAddress(r.RemoteAddr)),
						zap.String("user_agent.original", r.UserAgent()),
						zap.Any("panic", recovered),
						zap.Stack("stacktrace"),
					}
					fields = append(fields, logger.TraceFields(ctx)...)

					root.Error("http_panic_recovered", fields...)

					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}(requestCtx)

			next.ServeHTTP(w, r)
		})
	}
}

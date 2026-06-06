package middleware

import (
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/yorukot/netstamp/internal/controller/logger"
	"github.com/yorukot/netstamp/internal/controller/transport/http/clientip"
)

func ZapRequestLogger(root *zap.Logger) func(http.Handler) http.Handler {
	if root == nil {
		root = zap.NewNop()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := chimw.GetReqID(r.Context())
			if requestID == "" {
				requestID = r.Header.Get("X-Request-ID")
			}

			wrapped := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
			reqLog := root.With(
				zap.String("request_id", requestID),
				zap.String("http.request.method", r.Method),
				zap.String("url.path", r.URL.Path),
				zap.String("client.address", clientAddress(r)),
				zap.String("user_agent.original", r.UserAgent()),
			)
			if traceFields := logger.TraceFields(r.Context()); len(traceFields) > 0 {
				reqLog = reqLog.With(traceFields...)
			}

			ctx := logger.WithContext(r.Context(), reqLog)
			next.ServeHTTP(wrapped, r.WithContext(ctx))

			status := wrapped.Status()
			if status == 0 {
				status = http.StatusOK
			}

			route := ""
			if routeCtx := chi.RouteContext(ctx); routeCtx != nil {
				route = routeCtx.RoutePattern()
			}

			fields := []zap.Field{
				zap.String("http.route", route),
				zap.Int("http.response.status_code", status),
				zap.Int("http.bytes_written", wrapped.BytesWritten()),
				zap.Float64("duration_ms", float64(time.Since(start).Microseconds())/1000),
			}

			switch {
			case status >= http.StatusInternalServerError:
				reqLog.Error("http_request", fields...)
			case status >= http.StatusBadRequest:
				reqLog.Warn("http_request", fields...)
			default:
				reqLog.Info("http_request", fields...)
			}
		})
	}
}

func clientAddress(r *http.Request) string {
	if addr, ok := clientip.FromContext(r.Context()); ok {
		return addr.String()
	}

	if ip := chimw.GetClientIP(r.Context()); ip != "" {
		return ip
	}

	remoteAddr := r.RemoteAddr
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}

	return host
}

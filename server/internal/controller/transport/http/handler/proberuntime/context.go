package proberuntime

import (
	"context"
	"errors"
	"net/http"
	"strings"

	appproberuntime "github.com/yorukot/netstamp/internal/controller/application/proberuntime"
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

type runtimeAuthContextKey struct{}

func requireRuntimeAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secret, ok := probeSecret(r.Header.Get("Authorization"))
		if !ok {
			writeRuntimeProblem(w, r, http.StatusUnauthorized, "missing probe credential")
			return
		}

		auth := appproberuntime.RuntimeAuthInput{
			ProbeID:    httpx.Path(r, "probe_id"),
			Credential: secret,
		}
		next.ServeHTTP(w, r.WithContext(withRuntimeAuth(r.Context(), auth)))
	})
}

func withRuntimeAuth(ctx context.Context, auth appproberuntime.RuntimeAuthInput) context.Context {
	return context.WithValue(ctx, runtimeAuthContextKey{}, auth)
}

func runtimeAuthFromContext(ctx context.Context) (appproberuntime.RuntimeAuthInput, bool) {
	auth, ok := ctx.Value(runtimeAuthContextKey{}).(appproberuntime.RuntimeAuthInput)
	return auth, ok
}

func requireRuntimeAuthInput(ctx context.Context) (appproberuntime.RuntimeAuthInput, error) {
	auth, ok := runtimeAuthFromContext(ctx)
	if !ok {
		return appproberuntime.RuntimeAuthInput{}, httpx.InternalServerError("probe runtime auth unavailable")
	}

	return auth, nil
}

func probeSecret(header string) (string, bool) {
	scheme, secret, ok := strings.Cut(strings.TrimSpace(header), " ")
	if !ok || !strings.EqualFold(scheme, "Probe") || strings.TrimSpace(secret) == "" {
		return "", false
	}

	return strings.TrimSpace(secret), true
}

func writeRuntimeProblem(w http.ResponseWriter, r *http.Request, status int, detail string) {
	if status == http.StatusUnauthorized {
		w.Header().Set("WWW-Authenticate", "Probe")
	}
	httpx.WriteProblem(w, r, httpx.NewError(status, detail))
}

func mapRuntimeError(err error, fallback string) error {
	switch {
	case errors.Is(err, domainprobe.ErrInvalidCredential):
		return httpx.Unauthorized("invalid probe credential")
	case errors.Is(err, domainprobe.ErrProbeDisabled):
		return httpx.Forbidden("probe disabled")
	case errors.Is(err, domainprobe.ErrProbeNotFound):
		return httpx.NotFound("probe not found")
	case errors.Is(err, appproberuntime.ErrInvalidInput):
		return invalidRuntimeInputError(err)
	case errors.Is(err, domainping.ErrInvalidResult), errors.Is(err, domaintraceroute.ErrInvalidResult):
		return invalidRuntimeInputError(err)
	default:
		return httpx.InternalServerError(fallback)
	}
}

func invalidRuntimeInputError(err error) error {
	fieldErrors, ok := appvalidation.FieldErrors(err)
	if !ok {
		return httpx.UnprocessableEntity("invalid probe runtime input")
	}

	details := make([]httpx.ErrorDetail, 0, len(fieldErrors))
	for _, fieldErr := range fieldErrors {
		details = append(details, httpx.ErrorDetail{
			Message:  fieldErr.Message,
			Location: runtimeErrorLocation(fieldErr.Field),
			Value:    fieldErr.Value,
		})
	}

	return httpx.UnprocessableEntity("invalid probe runtime input", details...)
}

func runtimeErrorLocation(field string) string {
	switch field {
	case "":
		return "body"
	case "probeId":
		return "path.probe_id"
	default:
		return "body." + field
	}
}

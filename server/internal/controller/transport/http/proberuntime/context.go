package proberuntime

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	appproberuntime "github.com/yorukot/netstamp/internal/controller/application/proberuntime"
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

type runtimeAuthContextKey struct{}

func requireRuntimeAuth(ctx huma.Context, next func(huma.Context)) {
	secret, ok := probeSecret(ctx.Header("Authorization"))
	if !ok {
		writeRuntimeProblem(ctx, http.StatusUnauthorized, "missing probe credential")
		return
	}

	auth := appproberuntime.RuntimeAuthInput{
		ProbeID:    ctx.Param("probe_id"),
		Credential: secret,
	}
	next(huma.WithContext(ctx, withRuntimeAuth(ctx.Context(), auth)))
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
		return appproberuntime.RuntimeAuthInput{}, huma.Error500InternalServerError("probe runtime auth unavailable")
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

func writeRuntimeProblem(ctx huma.Context, status int, detail string) {
	if status == http.StatusUnauthorized {
		ctx.SetHeader("WWW-Authenticate", "Probe")
	}
	ctx.SetHeader("Content-Type", "application/problem+json")
	ctx.SetStatus(status)

	if err := json.NewEncoder(ctx.BodyWriter()).Encode(&huma.ErrorModel{
		Status: status,
		Title:  http.StatusText(status),
		Detail: detail,
	}); err != nil {
		return
	}
}

func mapRuntimeError(err error, fallback string) error {
	switch {
	case errors.Is(err, domainprobe.ErrInvalidCredential):
		return huma.Error401Unauthorized("invalid probe credential")
	case errors.Is(err, domainprobe.ErrProbeDisabled):
		return huma.Error403Forbidden("probe disabled")
	case errors.Is(err, domainprobe.ErrProbeNotFound):
		return huma.Error404NotFound("probe not found")
	case errors.Is(err, appproberuntime.ErrInvalidInput):
		return invalidRuntimeInputError(err)
	case errors.Is(err, domainping.ErrInvalidResult), errors.Is(err, domaintraceroute.ErrInvalidResult):
		return invalidRuntimeInputError(err)
	default:
		return huma.Error500InternalServerError(fallback)
	}
}

func invalidRuntimeInputError(err error) error {
	fieldErrors, ok := appvalidation.FieldErrors(err)
	if !ok {
		return huma.Error422UnprocessableEntity("invalid probe runtime input")
	}

	details := make([]error, 0, len(fieldErrors))
	for _, fieldErr := range fieldErrors {
		details = append(details, &huma.ErrorDetail{
			Message:  fieldErr.Message,
			Location: runtimeErrorLocation(fieldErr.Field),
			Value:    fieldErr.Value,
		})
	}

	return huma.Error422UnprocessableEntity("invalid probe runtime input", details...)
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

package proberuntime

import (
	"errors"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	appproberuntime "github.com/yorukot/netstamp/internal/application/proberuntime"
)

func runtimeAuthInput(probeID, header string) (appproberuntime.RuntimeAuthInput, error) {
	secret, err := probeSecret(header)
	if err != nil {
		return appproberuntime.RuntimeAuthInput{}, err
	}

	return appproberuntime.RuntimeAuthInput{
		ProbeID:    probeID,
		Credential: secret,
	}, nil
}

func probeSecret(header string) (string, error) {
	scheme, secret, ok := strings.Cut(strings.TrimSpace(header), " ")
	if !ok || !strings.EqualFold(scheme, "Probe") || strings.TrimSpace(secret) == "" {
		return "", huma.Error401Unauthorized("missing probe credential")
	}

	return strings.TrimSpace(secret), nil
}

func mapRuntimeError(err error, fallback string) error {
	switch {
	case errors.Is(err, appproberuntime.ErrInvalidCredential):
		return huma.Error401Unauthorized("invalid probe credential")
	case errors.Is(err, appproberuntime.ErrProbeDisabled):
		return huma.Error403Forbidden("probe disabled")
	case errors.Is(err, appproberuntime.ErrProbeNotFound):
		return huma.Error404NotFound("probe not found")
	case errors.Is(err, appproberuntime.ErrResultConflict):
		return huma.Error409Conflict("probe result conflicts with assignment")
	case errors.Is(err, appproberuntime.ErrUnsupportedResult):
		return huma.Error422UnprocessableEntity("unsupported result type")
	case errors.Is(err, appproberuntime.ErrInvalidInput), errors.Is(err, appproberuntime.ErrInvalidResult):
		return huma.Error422UnprocessableEntity("invalid probe runtime input")
	default:
		return huma.Error500InternalServerError(fallback)
	}
}

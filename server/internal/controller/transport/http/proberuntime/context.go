package proberuntime

import (
	"errors"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	appproberuntime "github.com/yorukot/netstamp/internal/controller/application/proberuntime"
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
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
	case errors.Is(err, domainprobe.ErrInvalidCredential):
		return huma.Error401Unauthorized("invalid probe credential")
	case errors.Is(err, domainprobe.ErrProbeDisabled):
		return huma.Error403Forbidden("probe disabled")
	case errors.Is(err, domainprobe.ErrProbeNotFound):
		return huma.Error404NotFound("probe not found")
	case errors.Is(err, appproberuntime.ErrResultConflict):
		return huma.Error409Conflict("probe result conflicts with assignment")
	case errors.Is(err, appproberuntime.ErrUnsupportedResult):
		return invalidRuntimeInputError(err)
	case errors.Is(err, appproberuntime.ErrInvalidInput), errors.Is(err, domainping.ErrInvalidResult):
		return invalidRuntimeInputError(err)
	default:
		return huma.Error500InternalServerError(fallback)
	}
}

func invalidRuntimeInputError(err error) error {
	fieldErrors, ok := appvalidation.FieldErrors(err)
	if !ok {
		if errors.Is(err, appproberuntime.ErrUnsupportedResult) {
			return huma.Error422UnprocessableEntity("unsupported result type")
		}
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

	if errors.Is(err, appproberuntime.ErrUnsupportedResult) {
		return huma.Error422UnprocessableEntity("unsupported result type", details...)
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

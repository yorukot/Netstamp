package proberuntime

import (
	"errors"
)

var (
	ErrInvalidInput = errors.New("probe runtime input invalid")

	errSecretVerifierMissing       = errors.New("probe secret verifier is not configured")
	errPingRepositoryMissing       = errors.New("ping result repository is not configured")
	errTracerouteRepositoryMissing = errors.New("traceroute result repository is not configured")
)

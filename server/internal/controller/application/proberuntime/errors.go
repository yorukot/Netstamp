package proberuntime

import (
	"errors"

	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

var (
	ErrInvalidInput      = domainprobe.ErrInvalidInput
	ErrProbeNotFound     = domainprobe.ErrProbeNotFound
	ErrProbeDisabled     = domainprobe.ErrProbeDisabled
	ErrInvalidCredential = domainprobe.ErrInvalidCredential
	ErrInvalidResult     = domainping.ErrInvalidResult
	ErrResultConflict    = errors.New("probe result conflicts with assignment")
	ErrUnsupportedResult = errors.New("probe result type unsupported")

	errSecretVerifierMissing = errors.New("probe secret verifier is not configured")
)

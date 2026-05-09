package proberuntime

import (
	"context"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

type ProbeRepository interface {
	GetActiveProbeCredential(ctx context.Context, probeID string) (domainprobe.Credential, error)
	UpdateProbeStatus(ctx context.Context, input domainprobe.UpdateStatusInput) (domainprobe.Status, error)
	ListAssignments(ctx context.Context, probeID string) ([]domaincheck.Assignment, error)
}

type PingResultRepository interface {
	CreatePingResults(ctx context.Context, inputs []domainping.ResultStorageInput) error
}

type SecretVerifier interface {
	VerifyProbeSecret(secret, expectedHash string) bool
}

package probe

import (
	"context"

	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

type Repository interface {
	GetProjectIDForUser(ctx context.Context, projectRef string, userID string) (string, error)
	CreateProbe(ctx context.Context, input domainprobe.CreateProbeStorageInput) (domainprobe.Probe, error)
}

type SecretGenerator interface {
	GenerateProbeSecret() (plaintext string, hash string, err error)
}

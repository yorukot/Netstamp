package apitoken

import (
	"context"
	"time"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

type Repository interface {
	Create(ctx context.Context, token identity.APIToken, maxActive int, now time.Time) (identity.APIToken, error)
	ListForUser(ctx context.Context, userID string) ([]identity.APIToken, error)
	GetActiveByHash(ctx context.Context, tokenHash []byte, now time.Time) (identity.APIToken, error)
	Touch(ctx context.Context, tokenID string, lastUsedAt, touchBefore time.Time) error
	RevokeForUser(ctx context.Context, userID, tokenID, reason string, revokedAt time.Time) error
	RevokeForUserAll(ctx context.Context, userID, reason string, revokedAt time.Time) error
}

type TokenManager interface {
	Generate() (rawToken, hint string, err error)
	Hash(rawToken string) []byte
}

type EventRecorder interface {
	RecordAPITokenEvent(ctx context.Context, event Event)
}

type Event struct {
	Name    string
	Outcome string
	Reason  string
	UserID  string
	TokenID string
	Err     error
}

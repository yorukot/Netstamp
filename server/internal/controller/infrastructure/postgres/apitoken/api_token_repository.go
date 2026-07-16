package apitoken

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

var tracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/apitoken")

type Repository struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: sqlc.New(pool), pool: pool}
}

func (r *Repository) Create(ctx context.Context, input identity.APIToken, maxActive int, now time.Time) (identity.APIToken, error) {
	ctx, span := postgres.StartUserDBSpan(ctx, tracer, "postgres.api_tokens.insert", "INSERT", "INSERT api_token")
	defer span.End()
	userID, err := postgres.ParseUUID(input.UserID, identity.ErrUserNotFound)
	if err != nil {
		return identity.APIToken{}, err
	}
	var created identity.APIToken
	err = pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		q := r.queries.WithTx(tx)
		if _, lockErr := q.LockUserForAPITokenCreate(ctx, userID); lockErr != nil {
			return lockErr
		}
		count, countErr := q.CountActiveAPITokensForUser(ctx, sqlc.CountActiveAPITokensForUserParams{UserID: userID, NowAt: now})
		if countErr != nil {
			return countErr
		}
		if count >= int64(maxActive) {
			return identity.ErrAPITokenLimitReached
		}
		row, createErr := q.CreateAPIToken(ctx, sqlc.CreateAPITokenParams{UserID: userID, Name: input.Name, TokenHash: input.TokenHash, TokenHint: input.TokenHint, Scopes: scopeStrings(input.Scopes), CreatedAt: input.CreatedAt, ExpiresAt: input.ExpiresAt})
		if createErr != nil {
			return createErr
		}
		created = mapToken(row)
		return nil
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return identity.APIToken{}, err
	}
	return created, nil
}

func (r *Repository) ListForUser(ctx context.Context, userIDValue string) ([]identity.APIToken, error) {
	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return nil, err
	}
	rows, err := postgres.Queries(ctx, r.queries).ListAPITokensForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]identity.APIToken, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapToken(row))
	}
	return result, nil
}

func (r *Repository) GetActiveByHash(ctx context.Context, tokenHash []byte, now time.Time) (identity.APIToken, error) {
	row, err := postgres.Queries(ctx, r.queries).GetActiveAPITokenByHash(ctx, sqlc.GetActiveAPITokenByHashParams{TokenHash: tokenHash, NowAt: now})
	if errors.Is(err, pgx.ErrNoRows) {
		return identity.APIToken{}, identity.ErrAPITokenNotFound
	}
	if err != nil {
		return identity.APIToken{}, err
	}
	return mapToken(row), nil
}

func (r *Repository) Touch(ctx context.Context, tokenID string, lastUsedAt, touchBefore time.Time) error {
	id, err := postgres.ParseUUID(tokenID, identity.ErrAPITokenNotFound)
	if err != nil {
		return err
	}
	return postgres.Queries(ctx, r.queries).TouchAPIToken(ctx, sqlc.TouchAPITokenParams{ID: id, LastUsedAt: &lastUsedAt, TouchBefore: &touchBefore})
}

func (r *Repository) RevokeForUser(ctx context.Context, userIDValue, tokenID, reason string, revokedAt time.Time) error {
	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return err
	}
	id, err := postgres.ParseUUID(tokenID, identity.ErrAPITokenNotFound)
	if err != nil {
		return err
	}
	_, err = postgres.Queries(ctx, r.queries).RevokeAPITokenForUser(ctx, sqlc.RevokeAPITokenForUserParams{ID: id, UserID: userID, RevokedAt: &revokedAt, RevokedReason: &reason})
	if errors.Is(err, pgx.ErrNoRows) {
		return identity.ErrAPITokenNotFound
	}
	return err
}

func (r *Repository) RevokeForUserAll(ctx context.Context, userIDValue, reason string, revokedAt time.Time) error {
	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return err
	}
	return postgres.Queries(ctx, r.queries).RevokeAPITokensForUser(ctx, sqlc.RevokeAPITokensForUserParams{UserID: userID, RevokedAt: &revokedAt, RevokedReason: &reason})
}

func mapToken(row sqlc.ApiToken) identity.APIToken {
	scopes := make([]identity.APITokenScope, 0, len(row.Scopes))
	for _, scope := range row.Scopes {
		scopes = append(scopes, identity.APITokenScope(scope))
	}
	return identity.APIToken{ID: row.ID.String(), UserID: row.UserID.String(), Name: row.Name, TokenHash: row.TokenHash, TokenHint: row.TokenHint, Scopes: scopes, CreatedAt: row.CreatedAt, LastUsedAt: row.LastUsedAt, ExpiresAt: row.ExpiresAt, RevokedAt: row.RevokedAt, RevokedReason: row.RevokedReason}
}

func scopeStrings(scopes []identity.APITokenScope) []string {
	values := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		values = append(values, string(scope))
	}
	return values
}

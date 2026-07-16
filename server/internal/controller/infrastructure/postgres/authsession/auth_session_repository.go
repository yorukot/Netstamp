package authsession

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

var authSessionTracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/authsession")

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: sqlc.New(pool)}
}

func (r *Repository) CreateSession(ctx context.Context, input identity.AuthSession) (identity.AuthSession, error) {
	ctx, span := postgres.StartUserDBSpan(ctx, authSessionTracer, "postgres.auth_sessions.insert", "INSERT", "INSERT auth_sessions")
	defer span.End()

	userID, err := postgres.ParseUUID(input.UserID, identity.ErrUserNotFound)
	if err != nil {
		return identity.AuthSession{}, err
	}

	row, err := postgres.Queries(ctx, r.queries).CreateAuthSession(ctx, sqlc.CreateAuthSessionParams{
		UserID:               userID,
		TokenHash:            input.TokenHash,
		CsrfTokenHash:        input.CSRFTokenHash,
		UserAgent:            input.UserAgent,
		CreatedAt:            input.CreatedAt,
		LastUsedAt:           input.LastUsedAt,
		IdleExpiresAt:        input.IdleExpiresAt,
		AbsoluteExpiresAt:    input.AbsoluteExpiresAt,
		AuthenticatedAt:      input.AuthenticatedAt,
		AuthenticationMethod: input.AuthenticationMethod,
		IdentityID:           parseOptionalUUID(input.IdentityID),
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return identity.AuthSession{}, err
	}

	return mapAuthSession(row), nil
}

func (r *Repository) UpdateSessionAuthentication(ctx context.Context, sessionID string, authenticatedAt time.Time, method string, identityID *string) error {
	id, err := postgres.ParseUUID(sessionID, identity.ErrSessionNotFound)
	if err != nil {
		return err
	}
	rows, err := postgres.Queries(ctx, r.queries).UpdateAuthSessionAuthentication(ctx, sqlc.UpdateAuthSessionAuthenticationParams{
		ID:                   id,
		AuthenticatedAt:      authenticatedAt,
		AuthenticationMethod: method,
		IdentityID:           parseOptionalUUID(identityID),
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return identity.ErrSessionNotFound
	}
	return nil
}

func (r *Repository) GetActiveSessionByTokenHash(ctx context.Context, tokenHash []byte, now time.Time) (identity.AuthSession, error) {
	ctx, span := postgres.StartUserDBSpan(ctx, authSessionTracer, "postgres.auth_sessions.select_active_by_hash", "SELECT", "SELECT active auth_session")
	defer span.End()

	row, err := postgres.Queries(ctx, r.queries).GetActiveAuthSessionByTokenHash(ctx, sqlc.GetActiveAuthSessionByTokenHashParams{
		TokenHash: tokenHash,
		NowAt:     now,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return identity.AuthSession{}, identity.ErrSessionNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return identity.AuthSession{}, err
	}

	return mapAuthSession(row), nil
}

func (r *Repository) GetActiveSessionByID(ctx context.Context, sessionID string, now time.Time) (identity.AuthSession, error) {
	ctx, span := postgres.StartUserDBSpan(ctx, authSessionTracer, "postgres.auth_sessions.select_active_by_id", "SELECT", "SELECT active auth_session")
	defer span.End()

	id, err := postgres.ParseUUID(sessionID, identity.ErrSessionNotFound)
	if err != nil {
		return identity.AuthSession{}, err
	}

	row, err := postgres.Queries(ctx, r.queries).GetActiveAuthSessionByID(ctx, sqlc.GetActiveAuthSessionByIDParams{
		ID:    id,
		NowAt: now,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return identity.AuthSession{}, identity.ErrSessionNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return identity.AuthSession{}, err
	}

	return mapAuthSession(row), nil
}

func (r *Repository) UpdateCSRFTokenHash(ctx context.Context, sessionID string, csrfTokenHash []byte, now time.Time) error {
	ctx, span := postgres.StartUserDBSpan(ctx, authSessionTracer, "postgres.auth_sessions.update_csrf", "UPDATE", "UPDATE auth_session csrf")
	defer span.End()

	id, err := postgres.ParseUUID(sessionID, identity.ErrSessionNotFound)
	if err != nil {
		return err
	}

	if err := postgres.Queries(ctx, r.queries).UpdateAuthSessionCSRFTokenHash(ctx, sqlc.UpdateAuthSessionCSRFTokenHashParams{
		ID:            id,
		CsrfTokenHash: csrfTokenHash,
		NowAt:         now,
	}); err != nil {
		postgres.RecordDBSpanError(span, err)
		return err
	}
	return nil
}

func (r *Repository) TouchSession(ctx context.Context, sessionID string, lastUsedAt, idleExpiresAt time.Time) error {
	ctx, span := postgres.StartUserDBSpan(ctx, authSessionTracer, "postgres.auth_sessions.touch", "UPDATE", "UPDATE auth_session touch")
	defer span.End()

	id, err := postgres.ParseUUID(sessionID, identity.ErrSessionNotFound)
	if err != nil {
		return err
	}

	if err := postgres.Queries(ctx, r.queries).TouchAuthSession(ctx, sqlc.TouchAuthSessionParams{
		ID:            id,
		LastUsedAt:    lastUsedAt,
		IdleExpiresAt: idleExpiresAt,
	}); err != nil {
		postgres.RecordDBSpanError(span, err)
		return err
	}
	return nil
}

func (r *Repository) RevokeSessionByTokenHash(ctx context.Context, tokenHash []byte, revokedAt time.Time, reason string) error {
	ctx, span := postgres.StartUserDBSpan(ctx, authSessionTracer, "postgres.auth_sessions.revoke_by_hash", "UPDATE", "UPDATE auth_session revoke")
	defer span.End()

	if err := postgres.Queries(ctx, r.queries).RevokeAuthSessionByTokenHash(ctx, sqlc.RevokeAuthSessionByTokenHashParams{
		TokenHash:     tokenHash,
		RevokedAt:     &revokedAt,
		RevokedReason: &reason,
	}); err != nil {
		postgres.RecordDBSpanError(span, err)
		return err
	}
	return nil
}

func (r *Repository) ListActiveSessionsForUser(ctx context.Context, userIDValue string, now time.Time) ([]identity.AuthSession, error) {
	ctx, span := postgres.StartUserDBSpan(ctx, authSessionTracer, "postgres.auth_sessions.list_active_for_user", "SELECT", "SELECT active user auth_sessions")
	defer span.End()

	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return nil, err
	}

	rows, err := postgres.Queries(ctx, r.queries).ListActiveAuthSessionsForUser(ctx, sqlc.ListActiveAuthSessionsForUserParams{
		UserID: userID,
		NowAt:  now,
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}

	sessions := make([]identity.AuthSession, 0, len(rows))
	for _, row := range rows {
		sessions = append(sessions, mapAuthSession(row))
	}
	return sessions, nil
}

func (r *Repository) RevokeSessionByIDForUser(ctx context.Context, userIDValue, sessionID string, revokedAt time.Time, reason string) error {
	ctx, span := postgres.StartUserDBSpan(ctx, authSessionTracer, "postgres.auth_sessions.revoke_by_id_for_user", "UPDATE", "UPDATE user auth_session revoke")
	defer span.End()

	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return err
	}
	id, err := postgres.ParseUUID(sessionID, identity.ErrSessionNotFound)
	if err != nil {
		return err
	}

	if _, err := postgres.Queries(ctx, r.queries).RevokeAuthSessionByIDForUser(ctx, sqlc.RevokeAuthSessionByIDForUserParams{
		ID:            id,
		UserID:        userID,
		RevokedAt:     &revokedAt,
		RevokedReason: &reason,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return identity.ErrSessionNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return err
	}
	return nil
}

func (r *Repository) RevokeSessionsForUser(ctx context.Context, userIDValue string, revokedAt time.Time, reason string) error {
	ctx, span := postgres.StartUserDBSpan(ctx, authSessionTracer, "postgres.auth_sessions.revoke_for_user", "UPDATE", "UPDATE user auth_sessions revoke")
	defer span.End()

	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return err
	}

	if err := postgres.Queries(ctx, r.queries).RevokeAuthSessionsForUser(ctx, sqlc.RevokeAuthSessionsForUserParams{
		UserID:        userID,
		RevokedAt:     &revokedAt,
		RevokedReason: &reason,
	}); err != nil {
		postgres.RecordDBSpanError(span, err)
		return err
	}
	return nil
}

func (r *Repository) RevokeSessionsForUserExcept(ctx context.Context, userIDValue, sessionIDValue string, revokedAt time.Time, reason string) error {
	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return err
	}
	sessionID, err := postgres.ParseUUID(sessionIDValue, identity.ErrSessionNotFound)
	if err != nil {
		return err
	}
	return postgres.Queries(ctx, r.queries).RevokeAuthSessionsForUserExcept(ctx, sqlc.RevokeAuthSessionsForUserExceptParams{
		UserID: userID, ExcludedSessionID: sessionID, RevokedAt: &revokedAt, RevokedReason: &reason,
	})
}

func mapAuthSession(row sqlc.AuthSession) identity.AuthSession {
	var identityID *string
	if row.IdentityID != nil {
		value := row.IdentityID.String()
		identityID = &value
	}
	return identity.AuthSession{
		ID:                   row.ID.String(),
		UserID:               row.UserID.String(),
		TokenHash:            row.TokenHash,
		CSRFTokenHash:        row.CsrfTokenHash,
		UserAgent:            row.UserAgent,
		AuthenticatedAt:      row.AuthenticatedAt,
		AuthenticationMethod: row.AuthenticationMethod,
		IdentityID:           identityID,
		CreatedAt:            row.CreatedAt,
		LastUsedAt:           row.LastUsedAt,
		IdleExpiresAt:        row.IdleExpiresAt,
		AbsoluteExpiresAt:    row.AbsoluteExpiresAt,
		RevokedAt:            row.RevokedAt,
		RevokedReason:        row.RevokedReason,
	}
}

func parseOptionalUUID(value *string) *uuid.UUID {
	if value == nil || *value == "" {
		return nil
	}
	parsed, err := uuid.Parse(*value)
	if err != nil {
		return nil
	}
	return &parsed
}

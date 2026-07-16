package pguser

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

func (r *UserRepository) CreateUserIdentity(ctx context.Context, input identity.UserIdentity) (identity.UserIdentity, error) {
	userID, err := postgres.ParseUUID(input.UserID, identity.ErrUserNotFound)
	if err != nil {
		return identity.UserIdentity{}, err
	}
	row, err := postgres.Queries(ctx, r.queries).CreateUserIdentity(ctx, sqlc.CreateUserIdentityParams{
		UserID: userID, Provider: input.Provider, Issuer: input.Issuer, Subject: input.Subject,
		Email: input.Email, EmailVerified: input.EmailVerified, DisplayName: input.DisplayName,
		CreatedAt: input.CreatedAt, LastLoginAt: input.LastLoginAt,
	})
	if err != nil {
		if postgres.IsUniqueViolation(err, "uq_user_identities_issuer_subject") || postgres.IsUniqueViolation(err, "uq_user_identities_user_issuer") {
			return identity.UserIdentity{}, identity.ErrIdentityConflict
		}
		return identity.UserIdentity{}, err
	}
	return mapUserIdentity(row), nil
}

func (r *UserRepository) GetUserIdentityByIssuerSubject(ctx context.Context, issuer, subject string) (identity.UserIdentity, error) {
	row, err := postgres.Queries(ctx, r.queries).GetUserIdentityByIssuerSubject(ctx, sqlc.GetUserIdentityByIssuerSubjectParams{Issuer: issuer, Subject: subject})
	if errors.Is(err, pgx.ErrNoRows) {
		return identity.UserIdentity{}, identity.ErrIdentityNotFound
	}
	if err != nil {
		return identity.UserIdentity{}, err
	}
	return mapUserIdentity(row), nil
}

func (r *UserRepository) GetUserIdentityByIDForUser(ctx context.Context, identityIDValue, userIDValue string) (identity.UserIdentity, error) {
	identityID, err := postgres.ParseUUID(identityIDValue, identity.ErrIdentityNotFound)
	if err != nil {
		return identity.UserIdentity{}, err
	}
	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return identity.UserIdentity{}, err
	}
	row, err := postgres.Queries(ctx, r.queries).GetUserIdentityByIDForUser(ctx, sqlc.GetUserIdentityByIDForUserParams{ID: identityID, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return identity.UserIdentity{}, identity.ErrIdentityNotFound
	}
	if err != nil {
		return identity.UserIdentity{}, err
	}
	return mapUserIdentity(row), nil
}

func (r *UserRepository) ListUserIdentities(ctx context.Context, userIDValue string) ([]identity.UserIdentity, error) {
	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return nil, err
	}
	rows, err := postgres.Queries(ctx, r.queries).ListUserIdentities(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]identity.UserIdentity, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapUserIdentity(row))
	}
	return result, nil
}

func (r *UserRepository) TouchUserIdentityLogin(ctx context.Context, input identity.UserIdentity, at time.Time) (identity.UserIdentity, error) {
	id, err := postgres.ParseUUID(input.ID, identity.ErrIdentityNotFound)
	if err != nil {
		return identity.UserIdentity{}, err
	}
	row, err := postgres.Queries(ctx, r.queries).TouchUserIdentityLogin(ctx, sqlc.TouchUserIdentityLoginParams{
		ID: id, Email: input.Email, EmailVerified: input.EmailVerified, DisplayName: input.DisplayName, LastLoginAt: &at,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return identity.UserIdentity{}, identity.ErrIdentityNotFound
	}
	if err != nil {
		return identity.UserIdentity{}, err
	}
	return mapUserIdentity(row), nil
}

func (r *UserRepository) DeleteUserIdentity(ctx context.Context, userIDValue, identityIDValue string) error {
	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return err
	}
	identityID, err := postgres.ParseUUID(identityIDValue, identity.ErrIdentityNotFound)
	if err != nil {
		return err
	}
	_, err = postgres.Queries(ctx, r.queries).DeleteUserIdentityForUser(ctx, sqlc.DeleteUserIdentityForUserParams{ID: identityID, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return identity.ErrIdentityNotFound
	}
	return err
}

func (r *UserRepository) CountUserAuthenticationMethods(ctx context.Context, userIDValue string) (bool, int64, error) {
	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return false, 0, err
	}
	row, err := postgres.Queries(ctx, r.queries).CountUserAuthenticationMethods(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, 0, identity.ErrUserNotFound
	}
	return row.HasPassword, row.IdentityCount, err
}

func (r *UserRepository) CreateOIDCUser(ctx context.Context, email, displayName, issuer, subject string, now time.Time) (identity.User, identity.UserIdentity, error) {
	row, err := postgres.Queries(ctx, r.queries).CreateOIDCUser(ctx, sqlc.CreateOIDCUserParams{
		Email: email, DisplayName: displayName, EmailVerifiedAt: &now, Issuer: issuer, Subject: subject, CreatedAt: now,
	})
	if err != nil {
		if postgres.IsUniqueViolation(err, "uq_users_email") {
			return identity.User{}, identity.UserIdentity{}, identity.ErrEmailAlreadyExists
		}
		if postgres.IsUniqueViolation(err, "uq_user_identities_issuer_subject") {
			return identity.User{}, identity.UserIdentity{}, identity.ErrIdentityConflict
		}
		return identity.User{}, identity.UserIdentity{}, err
	}
	user := identity.User{ID: row.ID.String(), Email: row.Email, DisplayName: row.DisplayName, EmailVerifiedAt: row.EmailVerifiedAt, DisabledAt: row.DisabledAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
	identityValue, err := r.GetUserIdentityByIDForUser(ctx, row.IdentityID.String(), user.ID)
	return user, identityValue, err
}

func (r *UserRepository) CreateOIDCAuthFlow(ctx context.Context, input identity.OIDCAuthFlow) (identity.OIDCAuthFlow, error) {
	var sessionID *uuid.UUID
	if input.SessionID != nil {
		parsed, err := uuid.Parse(*input.SessionID)
		if err != nil {
			return identity.OIDCAuthFlow{}, identity.ErrSessionNotFound
		}
		sessionID = &parsed
	}
	row, err := postgres.Queries(ctx, r.queries).CreateOIDCAuthFlow(ctx, sqlc.CreateOIDCAuthFlowParams{
		StateHash: input.StateHash, BrowserTokenHash: input.BrowserTokenHash, Nonce: input.Nonce,
		PkceVerifier: input.PKCEVerifier, Intent: input.Intent, SessionID: sessionID,
		ReturnTo: input.ReturnTo, CreatedAt: input.CreatedAt, ExpiresAt: input.ExpiresAt,
	})
	if err != nil {
		return identity.OIDCAuthFlow{}, err
	}
	return mapOIDCAuthFlow(row), nil
}

func (r *UserRepository) ConsumeOIDCAuthFlow(ctx context.Context, stateHash, browserTokenHash []byte, now time.Time) (identity.OIDCAuthFlow, error) {
	row, err := postgres.Queries(ctx, r.queries).ConsumeOIDCAuthFlow(ctx, sqlc.ConsumeOIDCAuthFlowParams{StateHash: stateHash, BrowserTokenHash: browserTokenHash, UsedAt: &now})
	if errors.Is(err, pgx.ErrNoRows) {
		return identity.OIDCAuthFlow{}, identity.ErrOIDCFlowNotFound
	}
	if err != nil {
		return identity.OIDCAuthFlow{}, err
	}
	return mapOIDCAuthFlow(row), nil
}

func (r *UserRepository) DeleteExpiredOIDCAuthFlows(ctx context.Context, now time.Time) error {
	return postgres.Queries(ctx, r.queries).DeleteExpiredOIDCAuthFlows(ctx, now)
}

func mapUserIdentity(row sqlc.UserIdentity) identity.UserIdentity {
	return identity.UserIdentity{ID: row.ID.String(), UserID: row.UserID.String(), Provider: row.Provider, Issuer: row.Issuer, Subject: row.Subject, Email: row.Email, EmailVerified: row.EmailVerified, DisplayName: row.DisplayName, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt, LastLoginAt: row.LastLoginAt}
}

func mapOIDCAuthFlow(row sqlc.OidcAuthFlow) identity.OIDCAuthFlow {
	var sessionID *string
	if row.SessionID != nil {
		value := row.SessionID.String()
		sessionID = &value
	}
	return identity.OIDCAuthFlow{ID: row.ID.String(), StateHash: row.StateHash, BrowserTokenHash: row.BrowserTokenHash, Nonce: row.Nonce, PKCEVerifier: row.PkceVerifier, Intent: row.Intent, SessionID: sessionID, ReturnTo: row.ReturnTo, CreatedAt: row.CreatedAt, ExpiresAt: row.ExpiresAt, UsedAt: row.UsedAt}
}

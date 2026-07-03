package pguser

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/trace"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

type UserRepository struct {
	queries *sqlc.Queries
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		queries: sqlc.New(pool),
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, input identity.User) (identity.User, error) {
	ctx, span := postgres.StartUserDBSpan(ctx, pguserTracer, "postgres.users.insert", "INSERT", "INSERT users")
	defer span.End()

	row, err := postgres.Queries(ctx, r.queries).CreateUser(ctx, sqlc.CreateUserParams{
		Email:        input.Email,
		DisplayName:  input.DisplayName,
		PasswordHash: input.PasswordHash,
	})
	if err != nil {
		if postgres.IsUniqueViolation(err, "uq_users_email") {
			return identity.User{}, fmt.Errorf("email already exists: %w", identity.ErrEmailAlreadyExists)
		}
		postgres.RecordDBSpanError(span, err)
		return identity.User{}, err
	}

	return identity.User{
		ID:          row.ID.String(),
		Email:       row.Email,
		DisplayName: row.DisplayName,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, userIDValue string) (identity.User, error) {
	ctx, span := postgres.StartUserDBSpan(ctx, pguserTracer, "postgres.users.select_by_id", "SELECT", "SELECT users by id")
	defer span.End()

	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return identity.User{}, err
	}

	row, err := postgres.Queries(ctx, r.queries).GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return identity.User{}, identity.ErrUserNotFound
		}

		postgres.RecordDBSpanError(span, err)
		return identity.User{}, err
	}

	return mapUser(row), nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (identity.User, error) {
	ctx, span := postgres.StartUserDBSpan(ctx, pguserTracer, "postgres.users.select_by_email", "SELECT", "SELECT users by email")
	defer span.End()

	row, err := postgres.Queries(ctx, r.queries).GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return identity.User{}, identity.ErrUserNotFound
		}

		postgres.RecordDBSpanError(span, err)
		return identity.User{}, err
	}

	return mapUser(row), nil
}

func (r *UserRepository) UpdateUserDisplayName(ctx context.Context, input identity.User) (identity.User, error) {
	ctx, span := postgres.StartUserDBSpan(ctx, pguserTracer, "postgres.users.update_display_name", "UPDATE", "UPDATE users display name")
	defer span.End()

	userID, err := postgres.ParseUUID(input.ID, identity.ErrUserNotFound)
	if err != nil {
		return identity.User{}, err
	}

	row, err := postgres.Queries(ctx, r.queries).UpdateUserDisplayName(ctx, sqlc.UpdateUserDisplayNameParams{
		ID:          userID,
		DisplayName: input.DisplayName,
	})
	if err != nil {
		return identity.User{}, r.mapUpdateError(span, err)
	}

	return mapUser(row), nil
}

func (r *UserRepository) UpdateUserEmail(ctx context.Context, input identity.User) (identity.User, error) {
	ctx, span := postgres.StartUserDBSpan(ctx, pguserTracer, "postgres.users.update_email", "UPDATE", "UPDATE users email")
	defer span.End()

	userID, err := postgres.ParseUUID(input.ID, identity.ErrUserNotFound)
	if err != nil {
		return identity.User{}, err
	}

	row, err := postgres.Queries(ctx, r.queries).UpdateUserEmail(ctx, sqlc.UpdateUserEmailParams{
		ID:    userID,
		Email: input.Email,
	})
	if err != nil {
		return identity.User{}, r.mapUpdateError(span, err)
	}

	return mapUser(row), nil
}

func (r *UserRepository) UpdateUserPasswordHash(ctx context.Context, input identity.User) (identity.User, error) {
	ctx, span := postgres.StartUserDBSpan(ctx, pguserTracer, "postgres.users.update_password_hash", "UPDATE", "UPDATE users password hash")
	defer span.End()

	userID, err := postgres.ParseUUID(input.ID, identity.ErrUserNotFound)
	if err != nil {
		return identity.User{}, err
	}

	row, err := postgres.Queries(ctx, r.queries).UpdateUserPasswordHash(ctx, sqlc.UpdateUserPasswordHashParams{
		ID:           userID,
		PasswordHash: input.PasswordHash,
	})
	if err != nil {
		return identity.User{}, r.mapUpdateError(span, err)
	}

	return mapUser(row), nil
}

func (r *UserRepository) CreatePasswordResetToken(ctx context.Context, input identity.PasswordResetToken) (identity.PasswordResetToken, error) {
	ctx, span := postgres.StartUserDBSpan(ctx, pguserTracer, "postgres.password_reset_tokens.insert", "INSERT", "INSERT password reset token")
	defer span.End()

	userID, err := postgres.ParseUUID(input.UserID, identity.ErrUserNotFound)
	if err != nil {
		return identity.PasswordResetToken{}, err
	}

	row, err := postgres.Queries(ctx, r.queries).CreatePasswordResetToken(ctx, sqlc.CreatePasswordResetTokenParams{
		UserID:    userID,
		TokenHash: input.TokenHash,
		ExpiresAt: input.ExpiresAt,
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return identity.PasswordResetToken{}, err
	}

	return mapPasswordResetToken(row), nil
}

func (r *UserRepository) InvalidateActivePasswordResetTokens(ctx context.Context, userIDValue string, usedAt time.Time) error {
	ctx, span := postgres.StartUserDBSpan(ctx, pguserTracer, "postgres.password_reset_tokens.invalidate_active", "UPDATE", "UPDATE active password reset tokens")
	defer span.End()

	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return err
	}

	if err := postgres.Queries(ctx, r.queries).InvalidateActivePasswordResetTokens(ctx, sqlc.InvalidateActivePasswordResetTokensParams{
		UserID: userID,
		UsedAt: &usedAt,
	}); err != nil {
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
}

func (r *UserRepository) GetActivePasswordResetTokenByHash(ctx context.Context, tokenHash string, now time.Time) (identity.PasswordResetToken, error) {
	ctx, span := postgres.StartUserDBSpan(ctx, pguserTracer, "postgres.password_reset_tokens.select_active", "SELECT", "SELECT active password reset token")
	defer span.End()

	token, err := postgres.Queries(ctx, r.queries).GetActivePasswordResetTokenByHash(ctx, sqlc.GetActivePasswordResetTokenByHashParams{
		TokenHash: tokenHash,
		NowAt:     now,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return identity.PasswordResetToken{}, identity.ErrResetTokenNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return identity.PasswordResetToken{}, err
	}

	return mapPasswordResetToken(token), nil
}

func (r *UserRepository) MarkPasswordResetTokenUsed(ctx context.Context, tokenIDValue string, usedAt time.Time) error {
	ctx, span := postgres.StartUserDBSpan(ctx, pguserTracer, "postgres.password_reset_tokens.mark_used", "UPDATE", "UPDATE password reset token used")
	defer span.End()

	tokenID, err := postgres.ParseUUID(tokenIDValue, identity.ErrResetTokenNotFound)
	if err != nil {
		return err
	}

	if err := postgres.Queries(ctx, r.queries).MarkPasswordResetTokenUsed(ctx, sqlc.MarkPasswordResetTokenUsedParams{
		ID:     tokenID,
		UsedAt: &usedAt,
	}); err != nil {
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
}

func (r *UserRepository) mapUpdateError(span trace.Span, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return identity.ErrUserNotFound
	}
	if postgres.IsUniqueViolation(err, "uq_users_email") {
		return fmt.Errorf("email already exists: %w", identity.ErrEmailAlreadyExists)
	}
	postgres.RecordDBSpanError(span, err)
	return err
}

func mapUser(row sqlc.User) identity.User {
	return identity.User{
		ID:           row.ID.String(),
		Email:        row.Email,
		DisplayName:  row.DisplayName,
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func mapPasswordResetToken(row sqlc.PasswordResetToken) identity.PasswordResetToken {
	return identity.PasswordResetToken{
		ID:        row.ID.String(),
		UserID:    row.UserID.String(),
		TokenHash: row.TokenHash,
		ExpiresAt: row.ExpiresAt,
		UsedAt:    row.UsedAt,
		CreatedAt: row.CreatedAt,
	}
}

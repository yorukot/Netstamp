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
	tx      *postgres.Transactor
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		queries: sqlc.New(pool),
		tx:      postgres.NewTransactor(pool),
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, input identity.User) (identity.User, error) {
	ctx, span := postgres.StartUserDBSpan(ctx, pguserTracer, "postgres.users.insert", "INSERT", "INSERT users")
	defer span.End()

	row, err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{
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

	row, err := r.queries.GetUserByID(ctx, userID)
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

	row, err := r.queries.GetUserByEmail(ctx, email)
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

	row, err := r.queries.UpdateUserDisplayName(ctx, sqlc.UpdateUserDisplayNameParams{
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

	row, err := r.queries.UpdateUserEmail(ctx, sqlc.UpdateUserEmailParams{
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

	row, err := r.queries.UpdateUserPasswordHash(ctx, sqlc.UpdateUserPasswordHashParams{
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

	row, err := r.queries.CreatePasswordResetToken(ctx, sqlc.CreatePasswordResetTokenParams{
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

	if err := r.queries.InvalidateActivePasswordResetTokens(ctx, sqlc.InvalidateActivePasswordResetTokensParams{
		UserID: userID,
		UsedAt: &usedAt,
	}); err != nil {
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
}

func (r *UserRepository) ResetPasswordWithToken(ctx context.Context, tokenHash, passwordHash string, usedAt time.Time) (identity.User, error) {
	ctx, span := postgres.StartUserDBSpan(ctx, pguserTracer, "postgres.password_reset_tokens.consume", "UPDATE", "UPDATE password from reset token")
	defer span.End()

	var user identity.User
	err := r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		token, err := q.GetActivePasswordResetTokenByHash(ctx, sqlc.GetActivePasswordResetTokenByHashParams{
			TokenHash: tokenHash,
			NowAt:     usedAt,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return identity.ErrResetTokenNotFound
			}
			return err
		}

		row, err := q.UpdateUserPasswordHash(ctx, sqlc.UpdateUserPasswordHashParams{
			ID:           token.UserID,
			PasswordHash: passwordHash,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return identity.ErrUserNotFound
			}
			return err
		}

		if err := q.MarkPasswordResetTokenUsed(ctx, sqlc.MarkPasswordResetTokenUsedParams{
			ID:     token.ID,
			UsedAt: &usedAt,
		}); err != nil {
			return err
		}

		user = mapUser(row)
		return nil
	})
	if err != nil {
		if !errors.Is(err, identity.ErrResetTokenNotFound) && !errors.Is(err, identity.ErrUserNotFound) {
			postgres.RecordDBSpanError(span, err)
		}
		return identity.User{}, err
	}

	return user, nil
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

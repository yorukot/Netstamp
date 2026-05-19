package pguser

import (
	"context"
	"errors"
	"fmt"

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
	return &UserRepository{queries: sqlc.New(pool)}
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
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
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
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}
}

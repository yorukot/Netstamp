package pguser

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yorukot/netstamp/internal/domain/identity"
	"github.com/yorukot/netstamp/internal/infrastructure/postgres"
	"github.com/yorukot/netstamp/internal/infrastructure/postgres/sqlc"
)

type UserRepository struct {
	queries *sqlc.Queries
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{queries: sqlc.New(pool)}
}

func (r *UserRepository) CreateUser(ctx context.Context, input identity.CreateUserInput) (identity.User, error) {
	ctx, span := postgres.StartUserDBSpan(ctx, pguserTracer, "postgres.users.insert", "INSERT", "INSERT users")
	defer span.End()

	row, err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        input.Email,
		DisplayName:  &input.DisplayName,
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
		IsActive:    row.IsActive,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}, nil
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

	return identity.User{
		ID:           row.ID.String(),
		Email:        row.Email,
		DisplayName:  row.DisplayName,
		PasswordHash: row.PasswordHash,
		IsActive:     row.IsActive,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}, nil
}

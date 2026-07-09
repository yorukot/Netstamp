package pguser

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
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
		Email:           input.Email,
		DisplayName:     input.DisplayName,
		PasswordHash:    input.PasswordHash,
		EmailVerifiedAt: input.EmailVerifiedAt,
	})
	if err != nil {
		if postgres.IsUniqueViolation(err, "uq_users_email") {
			return identity.User{}, fmt.Errorf("email already exists: %w", identity.ErrEmailAlreadyExists)
		}
		postgres.RecordDBSpanError(span, err)
		return identity.User{}, err
	}

	return mapCreateUser(row), nil
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

	return mapGetUserByID(row), nil
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

	return mapGetUserByEmail(row), nil
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

	return mapUpdateUserDisplayName(row), nil
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

	return mapUpdateUserEmail(row), nil
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

	return mapUpdateUserPasswordHash(row), nil
}

func (r *UserRepository) DisableUser(ctx context.Context, userIDValue string) (identity.User, error) {
	ctx, span := postgres.StartUserDBSpan(ctx, pguserTracer, "postgres.users.disable", "UPDATE", "DISABLE user")
	defer span.End()

	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return identity.User{}, err
	}

	now := time.Now().UTC()
	row, err := postgres.Queries(ctx, r.queries).DisableUser(ctx, sqlc.DisableUserParams{
		ID:         userID,
		DisabledAt: &now,
	})
	if err != nil {
		return identity.User{}, r.mapUpdateError(span, err)
	}

	return mapDisableUser(row), nil
}

//nolint:dupl // Password reset and email verification tokens intentionally keep parallel repository flows.
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

//nolint:dupl // Password reset and email verification tokens intentionally keep parallel repository flows.
func (r *UserRepository) CreateEmailVerificationToken(ctx context.Context, input identity.EmailVerificationToken) (identity.EmailVerificationToken, error) {
	ctx, span := postgres.StartUserDBSpan(ctx, pguserTracer, "postgres.email_verification_tokens.insert", "INSERT", "INSERT email verification token")
	defer span.End()

	userID, err := postgres.ParseUUID(input.UserID, identity.ErrUserNotFound)
	if err != nil {
		return identity.EmailVerificationToken{}, err
	}

	row, err := postgres.Queries(ctx, r.queries).CreateEmailVerificationToken(ctx, sqlc.CreateEmailVerificationTokenParams{
		UserID:    userID,
		TokenHash: input.TokenHash,
		ExpiresAt: input.ExpiresAt,
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return identity.EmailVerificationToken{}, err
	}

	return mapEmailVerificationToken(row), nil
}

func (r *UserRepository) InvalidateActiveEmailVerificationTokens(ctx context.Context, userIDValue string, usedAt time.Time) error {
	ctx, span := postgres.StartUserDBSpan(ctx, pguserTracer, "postgres.email_verification_tokens.invalidate_active", "UPDATE", "UPDATE active email verification tokens")
	defer span.End()

	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return err
	}

	if err := postgres.Queries(ctx, r.queries).InvalidateActiveEmailVerificationTokens(ctx, sqlc.InvalidateActiveEmailVerificationTokensParams{
		UserID: userID,
		UsedAt: &usedAt,
	}); err != nil {
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
}

func (r *UserRepository) GetActiveEmailVerificationTokenByHash(ctx context.Context, tokenHash string, now time.Time) (identity.EmailVerificationToken, error) {
	ctx, span := postgres.StartUserDBSpan(ctx, pguserTracer, "postgres.email_verification_tokens.select_active", "SELECT", "SELECT active email verification token")
	defer span.End()

	token, err := postgres.Queries(ctx, r.queries).GetActiveEmailVerificationTokenByHash(ctx, sqlc.GetActiveEmailVerificationTokenByHashParams{
		TokenHash: tokenHash,
		NowAt:     now,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return identity.EmailVerificationToken{}, identity.ErrEmailVerificationTokenNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return identity.EmailVerificationToken{}, err
	}

	return mapEmailVerificationToken(token), nil
}

func (r *UserRepository) MarkEmailVerificationTokenUsed(ctx context.Context, tokenIDValue string, usedAt time.Time) error {
	ctx, span := postgres.StartUserDBSpan(ctx, pguserTracer, "postgres.email_verification_tokens.mark_used", "UPDATE", "UPDATE email verification token used")
	defer span.End()

	tokenID, err := postgres.ParseUUID(tokenIDValue, identity.ErrEmailVerificationTokenNotFound)
	if err != nil {
		return err
	}

	if err := postgres.Queries(ctx, r.queries).MarkEmailVerificationTokenUsed(ctx, sqlc.MarkEmailVerificationTokenUsedParams{
		ID:     tokenID,
		UsedAt: &usedAt,
	}); err != nil {
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
}

func (r *UserRepository) MarkUserEmailVerified(ctx context.Context, userIDValue string, verifiedAt time.Time) (identity.User, error) {
	ctx, span := postgres.StartUserDBSpan(ctx, pguserTracer, "postgres.users.mark_email_verified", "UPDATE", "UPDATE users email verified")
	defer span.End()

	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return identity.User{}, err
	}

	row, err := postgres.Queries(ctx, r.queries).MarkUserEmailVerified(ctx, sqlc.MarkUserEmailVerifiedParams{
		ID:         userID,
		VerifiedAt: &verifiedAt,
	})
	if err != nil {
		return identity.User{}, r.mapUpdateError(span, err)
	}

	return mapMarkUserEmailVerified(row), nil
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

func mapCreateUser(row sqlc.CreateUserRow) identity.User {
	return mapUserFields(row.ID, row.Email, row.PasswordHash, row.DisplayName, row.EmailVerifiedAt, row.DisabledAt, row.IsSystemAdmin, row.CreatedAt, row.UpdatedAt)
}

func mapGetUserByID(row sqlc.GetUserByIDRow) identity.User {
	return mapUserFields(row.ID, row.Email, row.PasswordHash, row.DisplayName, row.EmailVerifiedAt, row.DisabledAt, row.IsSystemAdmin, row.CreatedAt, row.UpdatedAt)
}

func mapGetUserByEmail(row sqlc.GetUserByEmailRow) identity.User {
	return mapUserFields(row.ID, row.Email, row.PasswordHash, row.DisplayName, row.EmailVerifiedAt, row.DisabledAt, row.IsSystemAdmin, row.CreatedAt, row.UpdatedAt)
}

func mapUpdateUserDisplayName(row sqlc.UpdateUserDisplayNameRow) identity.User {
	return mapUserFields(row.ID, row.Email, row.PasswordHash, row.DisplayName, row.EmailVerifiedAt, row.DisabledAt, row.IsSystemAdmin, row.CreatedAt, row.UpdatedAt)
}

func mapUpdateUserEmail(row sqlc.UpdateUserEmailRow) identity.User {
	return mapUserFields(row.ID, row.Email, row.PasswordHash, row.DisplayName, row.EmailVerifiedAt, row.DisabledAt, row.IsSystemAdmin, row.CreatedAt, row.UpdatedAt)
}

func mapUpdateUserPasswordHash(row sqlc.UpdateUserPasswordHashRow) identity.User {
	return mapUserFields(row.ID, row.Email, row.PasswordHash, row.DisplayName, row.EmailVerifiedAt, row.DisabledAt, row.IsSystemAdmin, row.CreatedAt, row.UpdatedAt)
}

func mapMarkUserEmailVerified(row sqlc.MarkUserEmailVerifiedRow) identity.User {
	return mapUserFields(row.ID, row.Email, row.PasswordHash, row.DisplayName, row.EmailVerifiedAt, row.DisabledAt, row.IsSystemAdmin, row.CreatedAt, row.UpdatedAt)
}

func mapDisableUser(row sqlc.DisableUserRow) identity.User {
	return mapUserFields(row.ID, row.Email, row.PasswordHash, row.DisplayName, row.EmailVerifiedAt, row.DisabledAt, row.IsSystemAdmin, row.CreatedAt, row.UpdatedAt)
}

func mapUserFields(id uuid.UUID, email, passwordHash, displayName string, emailVerifiedAt, disabledAt *time.Time, isSystemAdmin bool, createdAt, updatedAt time.Time) identity.User {
	return identity.User{
		ID:              id.String(),
		Email:           email,
		DisplayName:     displayName,
		PasswordHash:    passwordHash,
		EmailVerifiedAt: emailVerifiedAt,
		DisabledAt:      disabledAt,
		IsSystemAdmin:   isSystemAdmin,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
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

func mapEmailVerificationToken(row sqlc.EmailVerificationToken) identity.EmailVerificationToken {
	return identity.EmailVerificationToken{
		ID:        row.ID.String(),
		UserID:    row.UserID.String(),
		TokenHash: row.TokenHash,
		ExpiresAt: row.ExpiresAt,
		UsedAt:    row.UsedAt,
		CreatedAt: row.CreatedAt,
	}
}

package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

func TestRequestPasswordResetCreatesTokenAndSendsEmail(t *testing.T) {
	ctx := context.Background()
	user := identity.User{ID: "11111111-1111-1111-1111-111111111111", Email: "user@example.com", DisplayName: "Jane Doe"}
	users := &passwordResetUserRepo{user: user}
	resets := &passwordResetRepo{}
	mailer := &passwordResetMailer{}
	service := newPasswordResetTestService(users, resets, mailer)

	err := service.RequestPasswordReset(ctx, RequestPasswordResetInput{
		Email:        " USER@example.com ",
		ResetBaseURL: "https://app.example.com",
	})
	if err != nil {
		t.Fatalf("RequestPasswordReset returned error: %v", err)
	}

	if len(resets.created) != 1 {
		t.Fatalf("expected one reset token, got %d", len(resets.created))
	}
	if resets.created[0].UserID != user.ID || resets.created[0].TokenHash != "hash:raw-token" {
		t.Fatalf("unexpected reset token: %#v", resets.created[0])
	}
	if len(mailer.sent) != 1 {
		t.Fatalf("expected one reset email, got %d", len(mailer.sent))
	}
	if mailer.sent[0].To != user.Email || mailer.sent[0].ResetURL != "https://app.example.com/reset-password?token=raw-token" {
		t.Fatalf("unexpected reset email: %#v", mailer.sent[0])
	}
}

func TestRequestPasswordResetDoesNotRevealMissingEmail(t *testing.T) {
	ctx := context.Background()
	users := &passwordResetUserRepo{err: identity.ErrUserNotFound}
	resets := &passwordResetRepo{}
	mailer := &passwordResetMailer{}
	service := newPasswordResetTestService(users, resets, mailer)

	err := service.RequestPasswordReset(ctx, RequestPasswordResetInput{
		Email:        "missing@example.com",
		ResetBaseURL: "https://app.example.com",
	})
	if err != nil {
		t.Fatalf("RequestPasswordReset returned error: %v", err)
	}
	if len(resets.created) != 0 {
		t.Fatalf("expected no reset tokens, got %d", len(resets.created))
	}
	if len(mailer.sent) != 0 {
		t.Fatalf("expected no reset emails, got %d", len(mailer.sent))
	}
}

func TestRequestPasswordResetDoesNotRevealMailerFailure(t *testing.T) {
	ctx := context.Background()
	user := identity.User{ID: "11111111-1111-1111-1111-111111111111", Email: "user@example.com", DisplayName: "Jane Doe"}
	users := &passwordResetUserRepo{user: user}
	resets := &passwordResetRepo{}
	mailer := &passwordResetMailer{err: errors.New("smtp unavailable")}
	service := newPasswordResetTestService(users, resets, mailer)

	err := service.RequestPasswordReset(ctx, RequestPasswordResetInput{
		Email:        "user@example.com",
		ResetBaseURL: "https://app.example.com",
	})
	if err != nil {
		t.Fatalf("RequestPasswordReset returned error: %v", err)
	}
	if len(resets.created) != 1 {
		t.Fatalf("expected token creation before mail failure, got %d", len(resets.created))
	}
}

func TestConfirmPasswordResetHashesPasswordAndConsumesToken(t *testing.T) {
	ctx := context.Background()
	user := identity.User{ID: "11111111-1111-1111-1111-111111111111", Email: "user@example.com", DisplayName: "Jane Doe"}
	resets := &passwordResetRepo{resetUser: user}
	service := newPasswordResetTestService(&passwordResetUserRepo{user: user}, resets, &passwordResetMailer{})

	err := service.ConfirmPasswordReset(ctx, ConfirmPasswordResetInput{
		Token:       "raw-token",
		NewPassword: "correct-horse-battery-staple",
	})
	if err != nil {
		t.Fatalf("ConfirmPasswordReset returned error: %v", err)
	}
	if resets.consumedTokenHash != "hash:raw-token" {
		t.Fatalf("expected hashed token, got %q", resets.consumedTokenHash)
	}
	if resets.consumedPasswordHash != "hashed:correct-horse-battery-staple" {
		t.Fatalf("expected hashed password, got %q", resets.consumedPasswordHash)
	}
}

func TestConfirmPasswordResetRejectsInvalidToken(t *testing.T) {
	ctx := context.Background()
	resets := &passwordResetRepo{resetErr: identity.ErrResetTokenNotFound}
	service := newPasswordResetTestService(&passwordResetUserRepo{}, resets, &passwordResetMailer{})

	err := service.ConfirmPasswordReset(ctx, ConfirmPasswordResetInput{
		Token:       "raw-token",
		NewPassword: "correct-horse-battery-staple",
	})
	if !errors.Is(err, ErrResetTokenInvalid) {
		t.Fatalf("expected reset token invalid error, got %v", err)
	}
}

func newPasswordResetTestService(users *passwordResetUserRepo, resets *passwordResetRepo, mailer *passwordResetMailer) *Service {
	service := NewService(users, passwordResetHasher{}, nil, nil)
	service.now = func() time.Time { return time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC) }
	service.ConfigurePasswordReset(resets, passwordResetTokenManager{}, mailer, PasswordResetConfig{TokenTTL: 30 * time.Minute})
	return service
}

type passwordResetUserRepo struct {
	user identity.User
	err  error
}

func (r *passwordResetUserRepo) CreateUser(context.Context, identity.User) (identity.User, error) {
	return identity.User{}, nil
}

func (r *passwordResetUserRepo) GetUserByID(context.Context, string) (identity.User, error) {
	return r.user, r.err
}

func (r *passwordResetUserRepo) GetUserByEmail(context.Context, string) (identity.User, error) {
	return r.user, r.err
}

type passwordResetRepo struct {
	created              []identity.PasswordResetToken
	resetUser            identity.User
	resetErr             error
	consumedTokenHash    string
	consumedPasswordHash string
}

func (r *passwordResetRepo) CreatePasswordResetToken(_ context.Context, input identity.PasswordResetToken) (identity.PasswordResetToken, error) {
	r.created = append(r.created, input)
	input.ID = "22222222-2222-2222-2222-222222222222"
	return input, nil
}

func (r *passwordResetRepo) InvalidateActivePasswordResetTokens(context.Context, string, time.Time) error {
	return nil
}

func (r *passwordResetRepo) ResetPasswordWithToken(_ context.Context, tokenHash, passwordHash string, _ time.Time) (identity.User, error) {
	r.consumedTokenHash = tokenHash
	r.consumedPasswordHash = passwordHash
	if r.resetErr != nil {
		return identity.User{}, r.resetErr
	}
	return r.resetUser, nil
}

type passwordResetTokenManager struct{}

func (passwordResetTokenManager) Generate(context.Context) (string, error) {
	return "raw-token", nil
}

func (passwordResetTokenManager) Hash(value string) string {
	return "hash:" + value
}

type passwordResetMailer struct {
	sent []identity.PasswordResetEmail
	err  error
}

func (m *passwordResetMailer) SendPasswordReset(_ context.Context, input identity.PasswordResetEmail) error {
	m.sent = append(m.sent, input)
	return m.err
}

type passwordResetHasher struct{}

func (passwordResetHasher) Hash(_ context.Context, password string) (string, error) {
	return "hashed:" + password, nil
}

func (passwordResetHasher) Compare(context.Context, string, string) error {
	return nil
}

package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

func TestRegisterWithEmailVerificationCreatesTokenAndRequiresVerification(t *testing.T) {
	ctx := context.Background()
	users := &emailVerificationUserRepo{}
	verifications := &emailVerificationRepo{}
	mailer := &emailVerificationMailer{}
	service := newEmailVerificationTestService(users, verifications, mailer)

	result, err := service.Register(ctx, RegisterInput{
		Email:                    " USER@example.com ",
		DisplayName:              "Jane Doe",
		Password:                 "correct-horse-battery-staple",
		RequireEmailVerification: true,
		EmailVerificationBaseURL: "https://app.example.com",
	})
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if !result.EmailVerificationRequired {
		t.Fatal("expected email verification requirement")
	}
	if result.SessionToken != "" {
		t.Fatalf("expected no session token before verification, got %q", result.SessionToken)
	}
	if len(verifications.created) != 1 {
		t.Fatalf("expected one verification token, got %d", len(verifications.created))
	}
	if verifications.created[0].UserID != "11111111-1111-1111-1111-111111111111" || verifications.created[0].TokenHash != "hash:raw-email-token" {
		t.Fatalf("unexpected verification token: %#v", verifications.created[0])
	}
	if len(mailer.sent) != 1 {
		t.Fatalf("expected one verification email, got %d", len(mailer.sent))
	}
	if mailer.sent[0].To != "user@example.com" || mailer.sent[0].VerificationURL != "https://app.example.com/verify-email?token=raw-email-token" {
		t.Fatalf("unexpected verification email: %#v", mailer.sent[0])
	}
}

func TestRegisterFirstSystemAdminBypassesEmailVerificationLockout(t *testing.T) {
	ctx := context.Background()
	users := &emailVerificationUserRepo{}
	verifications := &emailVerificationRepo{}
	mailer := &emailVerificationMailer{}
	service := newEmailVerificationTestService(users, verifications, mailer)
	service.ConfigureSystemAdmin(emailVerificationSystemAdminRepo{grant: true})

	result, err := service.Register(ctx, RegisterInput{
		Email:                    "admin@example.com",
		DisplayName:              "Admin",
		Password:                 "correct-horse-battery-staple",
		RequireEmailVerification: true,
		EmailVerificationBaseURL: "https://app.example.com",
	})
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if result.EmailVerificationRequired {
		t.Fatal("expected first system admin to be signed in immediately")
	}
	if !result.IsSystemAdmin {
		t.Fatal("expected first system admin grant")
	}
	if result.SessionToken != "issued-session-token" {
		t.Fatalf("expected issued session token, got %q", result.SessionToken)
	}
	if verifications.verifiedUserID != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("expected user email to be marked verified, got %q", verifications.verifiedUserID)
	}
	if len(verifications.created) != 0 {
		t.Fatalf("expected no verification token for first admin, got %d", len(verifications.created))
	}
	if len(mailer.sent) != 0 {
		t.Fatalf("expected no verification email for first admin, got %d", len(mailer.sent))
	}
}

func TestLoginRejectsUnverifiedEmail(t *testing.T) {
	ctx := context.Background()
	users := &emailVerificationUserRepo{
		user: identity.User{
			ID:           "11111111-1111-1111-1111-111111111111",
			Email:        "user@example.com",
			DisplayName:  "Jane Doe",
			PasswordHash: "hashed:correct-horse-battery-staple",
			HasPassword:  true,
		},
	}
	service := newEmailVerificationTestService(users, &emailVerificationRepo{}, &emailVerificationMailer{})

	_, err := service.Login(ctx, LoginInput{
		Email:                    "user@example.com",
		Password:                 "correct-horse-battery-staple",
		RequireEmailVerification: true,
	})
	if !errors.Is(err, ErrEmailVerificationRequired) {
		t.Fatalf("expected email verification required error, got %v", err)
	}
}

func TestLoginAllowsUnverifiedEmailWhenVerificationIsDisabled(t *testing.T) {
	ctx := context.Background()
	users := &emailVerificationUserRepo{
		user: identity.User{
			ID:           "11111111-1111-1111-1111-111111111111",
			Email:        "user@example.com",
			DisplayName:  "Jane Doe",
			PasswordHash: "hashed:correct-horse-battery-staple",
			HasPassword:  true,
		},
	}
	service := newEmailVerificationTestService(users, &emailVerificationRepo{}, &emailVerificationMailer{})

	result, err := service.Login(ctx, LoginInput{
		Email:    "user@example.com",
		Password: "correct-horse-battery-staple",
	})
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if result.SessionToken != "issued-session-token" {
		t.Fatalf("expected issued session token, got %q", result.SessionToken)
	}
	if result.EmailVerified {
		t.Fatal("expected result to preserve unverified email state")
	}
}

func TestLoginRejectsPasswordlessAccount(t *testing.T) {
	ctx := context.Background()
	users := &emailVerificationUserRepo{user: identity.User{
		ID: "11111111-1111-1111-1111-111111111111", Email: "sso@example.com", DisplayName: "SSO User",
	}}
	service := newEmailVerificationTestService(users, &emailVerificationRepo{}, &emailVerificationMailer{})

	_, err := service.Login(ctx, LoginInput{Email: "sso@example.com", Password: "irrelevant-password"})
	if !errors.Is(err, ErrCredentialsInvalid) {
		t.Fatalf("expected passwordless account to reject password login, got %v", err)
	}
}

func TestConfirmEmailVerificationMarksUserAndConsumesToken(t *testing.T) {
	ctx := context.Background()
	verifications := &emailVerificationRepo{
		token: identity.EmailVerificationToken{
			ID:        "22222222-2222-2222-2222-222222222222",
			UserID:    "11111111-1111-1111-1111-111111111111",
			TokenHash: "hash:raw-email-token",
		},
	}
	service := newEmailVerificationTestService(&emailVerificationUserRepo{}, verifications, &emailVerificationMailer{})

	err := service.ConfirmEmailVerification(ctx, ConfirmEmailVerificationInput{Token: "raw-email-token"})
	if err != nil {
		t.Fatalf("ConfirmEmailVerification returned error: %v", err)
	}

	if verifications.consumedTokenHash != "hash:raw-email-token" {
		t.Fatalf("expected hashed token lookup, got %q", verifications.consumedTokenHash)
	}
	if verifications.verifiedUserID != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("expected verified user id, got %q", verifications.verifiedUserID)
	}
	if verifications.usedTokenID != "22222222-2222-2222-2222-222222222222" {
		t.Fatalf("expected consumed token id, got %q", verifications.usedTokenID)
	}
}

func TestConfirmEmailVerificationRejectsInvalidToken(t *testing.T) {
	ctx := context.Background()
	verifications := &emailVerificationRepo{tokenErr: identity.ErrEmailVerificationTokenNotFound}
	service := newEmailVerificationTestService(&emailVerificationUserRepo{}, verifications, &emailVerificationMailer{})

	err := service.ConfirmEmailVerification(ctx, ConfirmEmailVerificationInput{Token: "raw-email-token"})
	if !errors.Is(err, ErrEmailVerificationTokenInvalid) {
		t.Fatalf("expected invalid email verification token error, got %v", err)
	}
}

func newEmailVerificationTestService(users *emailVerificationUserRepo, verifications *emailVerificationRepo, mailer *emailVerificationMailer) *Service {
	service := NewService(users, emailVerificationHasher{}, emailVerificationSessionManager{}, nil)
	service.now = func() time.Time { return time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC) }
	service.ConfigureEmailVerification(verifications, emailVerificationTokenManager{}, mailer, EmailVerificationConfig{TokenTTL: 24 * time.Hour})
	return service
}

type emailVerificationUserRepo struct {
	user identity.User
	err  error
}

func (r *emailVerificationUserRepo) CreateUser(_ context.Context, input identity.User) (identity.User, error) {
	if r.err != nil {
		return identity.User{}, r.err
	}
	input.ID = "11111111-1111-1111-1111-111111111111"
	return input, nil
}

func (r *emailVerificationUserRepo) GetUserByID(context.Context, string) (identity.User, error) {
	return r.user, r.err
}

func (r *emailVerificationUserRepo) GetUserByEmail(context.Context, string) (identity.User, error) {
	return r.user, r.err
}

func (r *emailVerificationUserRepo) UpdateUserPasswordHash(context.Context, identity.User) (identity.User, error) {
	return r.user, r.err
}

type emailVerificationRepo struct {
	created           []identity.EmailVerificationToken
	token             identity.EmailVerificationToken
	tokenErr          error
	consumedTokenHash string
	usedTokenID       string
	verifiedUserID    string
}

func (r *emailVerificationRepo) CreateEmailVerificationToken(_ context.Context, input identity.EmailVerificationToken) (identity.EmailVerificationToken, error) {
	r.created = append(r.created, input)
	input.ID = "22222222-2222-2222-2222-222222222222"
	return input, nil
}

func (r *emailVerificationRepo) InvalidateActiveEmailVerificationTokens(context.Context, string, time.Time) error {
	return nil
}

func (r *emailVerificationRepo) GetActiveEmailVerificationTokenByHash(_ context.Context, tokenHash string, _ time.Time) (identity.EmailVerificationToken, error) {
	r.consumedTokenHash = tokenHash
	if r.tokenErr != nil {
		return identity.EmailVerificationToken{}, r.tokenErr
	}
	return r.token, nil
}

func (r *emailVerificationRepo) MarkEmailVerificationTokenUsed(_ context.Context, tokenID string, _ time.Time) error {
	r.usedTokenID = tokenID
	return nil
}

func (r *emailVerificationRepo) MarkUserEmailVerified(_ context.Context, userID string, verifiedAt time.Time) (identity.User, error) {
	r.verifiedUserID = userID
	return identity.User{
		ID:              userID,
		Email:           "user@example.com",
		DisplayName:     "Jane Doe",
		EmailVerifiedAt: &verifiedAt,
	}, nil
}

type emailVerificationTokenManager struct{}

func (emailVerificationTokenManager) Generate(context.Context) (string, error) {
	return "raw-email-token", nil
}

func (emailVerificationTokenManager) Hash(value string) string {
	return "hash:" + value
}

type emailVerificationMailer struct {
	sent []identity.EmailVerificationEmail
	err  error
}

func (m *emailVerificationMailer) SendEmailVerification(_ context.Context, input identity.EmailVerificationEmail) error {
	m.sent = append(m.sent, input)
	return m.err
}

type emailVerificationHasher struct{}

func (emailVerificationHasher) Hash(_ context.Context, password string) (string, error) {
	return "hashed:" + password, nil
}

func (emailVerificationHasher) Compare(context.Context, string, string) error {
	return nil
}

type emailVerificationSessionManager struct{}

func (emailVerificationSessionManager) CreateSession(context.Context, CreateSessionInput) (identity.CreatedSession, error) {
	return identity.CreatedSession{
		RawToken:  "issued-session-token",
		ExpiresIn: 3600,
	}, nil
}

func (emailVerificationSessionManager) VerifySession(context.Context, string) (identity.SessionClaims, error) {
	return identity.SessionClaims{}, nil
}

func (emailVerificationSessionManager) CreateCSRFToken(context.Context, string) (string, error) {
	return "csrf-token", nil
}

func (emailVerificationSessionManager) VerifyCSRFToken(context.Context, string, string) error {
	return nil
}

func (emailVerificationSessionManager) RevokeSession(context.Context, string, string) error {
	return nil
}

func (emailVerificationSessionManager) ListUserSessions(context.Context, string) ([]identity.AuthSession, error) {
	return nil, nil
}

func (emailVerificationSessionManager) RevokeUserSession(context.Context, string, string, string) error {
	return nil
}

func (emailVerificationSessionManager) RevokeUserSessions(context.Context, string, string) error {
	return nil
}

type emailVerificationSystemAdminRepo struct {
	grant bool
}

func (r emailVerificationSystemAdminRepo) GrantFirstSystemAdminIfNone(context.Context, string) (bool, error) {
	return r.grant, nil
}

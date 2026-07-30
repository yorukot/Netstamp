package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

func TestRegisterRejectsDisabledAccountCreationBeforeHashing(t *testing.T) {
	users := &instancePolicyUserRepository{}
	hasher := &instancePolicyHasher{}
	service := NewService(users, hasher, nil, nil)
	service.ConfigureInstancePolicy(instancePolicyFake{
		accountCreationEnabled:    false,
		credentialChangesEnabled:  true,
		emailVerificationRequired: false,
	})

	_, err := service.Register(context.Background(), RegisterInput{
		Email: "person@example.com", DisplayName: "Person", Password: "correct-horse-battery-staple",
	})
	if !errors.Is(err, ErrAccountCreationDisabled) {
		t.Fatalf("expected account creation disabled, got %v", err)
	}
	if hasher.hashCalls != 0 || users.createCalls != 0 {
		t.Fatalf("disabled account creation must not hash or persist: hash=%d create=%d", hasher.hashCalls, users.createCalls)
	}
}

func TestRegisterFailsClosedWhenAccountPolicyCannotBeRead(t *testing.T) {
	policyErr := errors.New("settings unavailable")
	users := &instancePolicyUserRepository{}
	hasher := &instancePolicyHasher{}
	service := NewService(users, hasher, nil, nil)
	service.ConfigureInstancePolicy(instancePolicyFake{
		accountCreationEnabled:   true,
		credentialChangesEnabled: true,
		accountCreationErr:       policyErr,
	})

	_, err := service.Register(context.Background(), RegisterInput{
		Email: "person@example.com", DisplayName: "Person", Password: "correct-horse-battery-staple",
	})
	if !errors.Is(err, ErrInstancePolicyUnavailable) || !errors.Is(err, policyErr) {
		t.Fatalf("expected wrapped policy failure, got %v", err)
	}
	if hasher.hashCalls != 0 || users.createCalls != 0 {
		t.Fatalf("policy failure must not hash or persist: hash=%d create=%d", hasher.hashCalls, users.createCalls)
	}
}

func TestLoginUsesEmailVerificationPolicyInsteadOfCallerFlag(t *testing.T) {
	users := &instancePolicyUserRepository{user: identity.User{
		ID: "11111111-1111-1111-1111-111111111111", Email: "person@example.com",
		PasswordHash: "hash", HasPassword: true,
	}}
	service := NewService(users, &instancePolicyHasher{}, nil, nil)
	service.ConfigureInstancePolicy(instancePolicyFake{
		accountCreationEnabled:    true,
		credentialChangesEnabled:  true,
		emailVerificationRequired: true,
	})

	_, err := service.Login(context.Background(), LoginInput{
		Email: "person@example.com", Password: "correct-horse-battery-staple", RequireEmailVerification: false,
	})
	if !errors.Is(err, ErrEmailVerificationRequired) {
		t.Fatalf("expected policy-required email verification, got %v", err)
	}
}

func TestPasswordResetRequestIsGenericWhenCredentialChangesDisabled(t *testing.T) {
	user := identity.User{
		ID: "11111111-1111-1111-1111-111111111111", Email: "person@example.com",
		DisplayName: "Person", HasPassword: true,
	}
	resets := &passwordResetRepo{}
	mailer := &passwordResetMailer{}
	service := newPasswordResetTestService(&passwordResetUserRepo{user: user}, resets, mailer)
	service.ConfigureInstancePolicy(instancePolicyFake{
		accountCreationEnabled:    true,
		credentialChangesEnabled:  false,
		emailVerificationRequired: false,
	})

	err := service.RequestPasswordReset(context.Background(), RequestPasswordResetInput{
		Email: user.Email, ResetBaseURL: "https://app.example.com",
	})
	if err != nil {
		t.Fatalf("disabled reset request must keep a generic success response, got %v", err)
	}
	if len(resets.created) != 0 || len(mailer.sent) != 0 {
		t.Fatalf("disabled reset request must not create a token or send mail: tokens=%d mail=%d", len(resets.created), len(mailer.sent))
	}
}

func TestPasswordResetRequestFailsClosedWithoutRevealingPolicyFailure(t *testing.T) {
	policyErr := errors.New("settings unavailable")
	user := identity.User{
		ID: "11111111-1111-1111-1111-111111111111", Email: "person@example.com",
		DisplayName: "Person", HasPassword: true,
	}
	resets := &passwordResetRepo{}
	mailer := &passwordResetMailer{}
	service := newPasswordResetTestService(&passwordResetUserRepo{user: user}, resets, mailer)
	service.ConfigureInstancePolicy(instancePolicyFake{
		accountCreationEnabled:   true,
		credentialChangesEnabled: true,
		credentialChangesErr:     policyErr,
	})

	err := service.RequestPasswordReset(context.Background(), RequestPasswordResetInput{
		Email: user.Email, ResetBaseURL: "https://app.example.com",
	})
	if err != nil {
		t.Fatalf("reset request policy failure must remain generic, got %v", err)
	}
	if len(resets.created) != 0 || len(mailer.sent) != 0 {
		t.Fatalf("policy failure must not create a token or send mail: tokens=%d mail=%d", len(resets.created), len(mailer.sent))
	}
}

func TestPasswordResetConfirmationRejectsDisabledCredentialChanges(t *testing.T) {
	resets := &passwordResetRepo{resetUser: identity.User{ID: "11111111-1111-1111-1111-111111111111"}}
	service := newPasswordResetTestService(&passwordResetUserRepo{}, resets, &passwordResetMailer{})
	service.ConfigureInstancePolicy(instancePolicyFake{
		accountCreationEnabled:    true,
		credentialChangesEnabled:  false,
		emailVerificationRequired: false,
	})

	err := service.ConfirmPasswordReset(context.Background(), ConfirmPasswordResetInput{
		Token: "raw-token", NewPassword: "correct-horse-battery-staple",
	})
	if !errors.Is(err, ErrCredentialChangesDisabled) {
		t.Fatalf("expected credential changes disabled, got %v", err)
	}
	if resets.consumedTokenHash != "" {
		t.Fatalf("disabled confirmation must not consume or inspect the reset token, got %q", resets.consumedTokenHash)
	}
}

func TestExternalJITRespectsAccountCreationPolicyAtCallback(t *testing.T) {
	now := time.Date(2026, time.July, 29, 12, 0, 0, 0, time.UTC)
	repo := &externalAuthRepositoryFake{
		flow: identity.ExternalAuthFlow{
			Provider: identity.AuthenticationMethodOIDC, Intent: ExternalAuthIntentLogin,
			CreatedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Minute),
		},
		identityErr: identity.ErrIdentityNotFound,
	}
	client := &externalAuthClientFake{claims: ExternalIdentityClaims{
		Issuer: "https://idp.example.com", Subject: "subject-1", Email: "person@example.com", EmailVerified: true,
	}}
	service := newExternalAuthTestService(repo, client, ExternalProviderConfig{
		ID: identity.AuthenticationMethodOIDC, JITEnabled: true,
	})
	service.now = func() time.Time { return now }
	service.ConfigureInstancePolicy(instancePolicyFake{
		accountCreationEnabled:    false,
		credentialChangesEnabled:  true,
		emailVerificationRequired: false,
	})

	_, err := service.CompleteExternalAuth(context.Background(), CompleteExternalAuthInput{
		Provider: identity.AuthenticationMethodOIDC, Code: "code", State: "state", BrowserToken: "browser",
	})
	if !errors.Is(err, ErrAccountCreationDisabled) {
		t.Fatalf("expected JIT account creation to be disabled, got %v", err)
	}
	if repo.createUserCalls != 0 {
		t.Fatalf("disabled JIT must not create a user, got %d calls", repo.createUserCalls)
	}
}

func TestExistingExternalIdentityLoginIgnoresAccountCreationSwitch(t *testing.T) {
	now := time.Date(2026, time.July, 29, 12, 0, 0, 0, time.UTC)
	userID := "11111111-1111-1111-1111-111111111111"
	identityID := "22222222-2222-2222-2222-222222222222"
	users := &externalAuthUserRepositoryFake{userByID: identity.User{
		ID: userID, Email: "person@example.com", DisplayName: "Person",
	}}
	repo := &externalAuthRepositoryFake{linkedIdentity: identity.UserIdentity{
		ID: identityID, UserID: userID, Provider: identity.AuthenticationMethodOIDC,
	}}
	service := NewService(users, passwordResetHasher{}, emailVerificationSessionManager{}, nil)
	service.ConfigureExternalAuth(repo, &externalAuthTokenManagerFake{}, ExternalAuthConfig{},
		ExternalProviderRegistration{
			Config: ExternalProviderConfig{ID: identity.AuthenticationMethodOIDC, JITEnabled: true},
			Client: &externalAuthClientFake{},
		},
	)
	service.ConfigureInstancePolicy(instancePolicyFake{
		accountCreationEnabled:    false,
		credentialChangesEnabled:  true,
		emailVerificationRequired: false,
	})

	access, err := service.completeExternalAuthLogin(context.Background(),
		service.externalProviders[identity.AuthenticationMethodOIDC],
		ExternalIdentityClaims{Issuer: "https://idp.example.com", Subject: "subject-1"},
		"test-agent",
		now,
	)
	if err != nil {
		t.Fatalf("existing linked identity should still log in: %v", err)
	}
	if access.UserID != userID || repo.createUserCalls != 0 {
		t.Fatalf("unexpected linked login result: access=%#v createCalls=%d", access, repo.createUserCalls)
	}
}

func TestExternalLinkRechecksCredentialPolicyAtCallback(t *testing.T) {
	now := time.Date(2026, time.July, 29, 12, 0, 0, 0, time.UTC)
	sessionID := "11111111-1111-1111-1111-111111111111"
	repo := &externalAuthRepositoryFake{identityErr: identity.ErrIdentityNotFound}
	service := newExternalAuthTestService(repo, &externalAuthClientFake{}, ExternalProviderConfig{
		ID: identity.AuthenticationMethodGitHub,
	})
	service.recentAuth = &recentAuthenticationFake{session: identity.AuthSession{
		ID: sessionID, UserID: "22222222-2222-2222-2222-222222222222",
	}}
	service.ConfigureInstancePolicy(instancePolicyFake{
		accountCreationEnabled:    true,
		credentialChangesEnabled:  false,
		emailVerificationRequired: false,
	})

	err := service.completeExternalAuthLink(
		context.Background(),
		service.externalProviders[identity.AuthenticationMethodGitHub],
		identity.ExternalAuthFlow{SessionID: &sessionID},
		ExternalIdentityClaims{Issuer: "https://github.com", Subject: "123"},
		now,
	)
	if !errors.Is(err, ErrCredentialChangesDisabled) {
		t.Fatalf("expected callback policy rejection, got %v", err)
	}
	if repo.createIdentityCalls != 0 {
		t.Fatalf("disabled identity linking must not persist an identity, got %d calls", repo.createIdentityCalls)
	}
}

type instancePolicyFake struct {
	accountCreationEnabled    bool
	credentialChangesEnabled  bool
	emailVerificationRequired bool
	accountCreationErr        error
	credentialChangesErr      error
	emailVerificationErr      error
}

func (p instancePolicyFake) AccountCreationEnabled(context.Context) (bool, error) {
	return p.accountCreationEnabled, p.accountCreationErr
}

func (p instancePolicyFake) CredentialChangesEnabled(context.Context) (bool, error) {
	return p.credentialChangesEnabled, p.credentialChangesErr
}

func (p instancePolicyFake) EmailVerificationRequired(context.Context) (bool, error) {
	return p.emailVerificationRequired, p.emailVerificationErr
}

type instancePolicyUserRepository struct {
	user        identity.User
	createCalls int
}

func (r *instancePolicyUserRepository) CreateUser(_ context.Context, input identity.User) (identity.User, error) {
	r.createCalls++
	return input, nil
}

func (r *instancePolicyUserRepository) GetUserByID(context.Context, string) (identity.User, error) {
	if r.user.ID == "" {
		return identity.User{}, identity.ErrUserNotFound
	}
	return r.user, nil
}

func (r *instancePolicyUserRepository) GetUserByEmail(context.Context, string) (identity.User, error) {
	if r.user.ID == "" {
		return identity.User{}, identity.ErrUserNotFound
	}
	return r.user, nil
}

func (r *instancePolicyUserRepository) UpdateUserPasswordHash(context.Context, identity.User) (identity.User, error) {
	return identity.User{}, nil
}

type instancePolicyHasher struct {
	hashCalls int
}

func (h *instancePolicyHasher) Hash(_ context.Context, password string) (string, error) {
	h.hashCalls++
	return "hashed:" + password, nil
}

func (*instancePolicyHasher) Compare(context.Context, string, string) error {
	return nil
}

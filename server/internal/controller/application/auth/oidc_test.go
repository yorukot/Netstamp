package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

func TestStartOIDCPersistsBoundFlowAndNormalizesReturnPath(t *testing.T) {
	now := time.Date(2026, time.July, 16, 10, 0, 0, 0, time.UTC)
	repo := &oidcRepositoryFake{}
	client := &oidcClientFake{authorizationURL: "https://idp.example.com/authorize"}
	service := NewService(&oidcUserRepositoryFake{}, passwordResetHasher{}, nil, nil)
	service.now = func() time.Time { return now }
	service.ConfigureOIDC(repo, client, &oidcTokenManagerFake{tokens: []string{"state", "browser", "nonce", "pkce"}}, OIDCConfig{
		Enabled: true, FlowTTL: 10 * time.Minute,
	})

	result, err := service.StartOIDC(context.Background(), StartOIDCInput{Intent: OIDCIntentLogin, ReturnTo: "//evil.example.com"})
	if err != nil {
		t.Fatalf("start OIDC: %v", err)
	}
	if result.AuthorizationURL != client.authorizationURL || result.BrowserToken != "browser" {
		t.Fatalf("unexpected start result: %#v", result)
	}
	if string(repo.createdFlow.StateHash) != "hash:state" || string(repo.createdFlow.BrowserTokenHash) != "hash:browser" {
		t.Fatalf("flow tokens were not persisted as hashes: %#v", repo.createdFlow)
	}
	if repo.createdFlow.ReturnTo != "/" || repo.createdFlow.Nonce != "nonce" || repo.createdFlow.PKCEVerifier != "pkce" {
		t.Fatalf("unexpected stored flow: %#v", repo.createdFlow)
	}
	if client.forceReauthentication {
		t.Fatal("login flow should not force provider reauthentication")
	}
}

func TestCompleteOIDCDoesNotAutoLinkAnExistingEmail(t *testing.T) {
	now := time.Date(2026, time.July, 16, 10, 0, 0, 0, time.UTC)
	users := &oidcUserRepositoryFake{userByEmail: identity.User{
		ID: "11111111-1111-1111-1111-111111111111", Email: "person@example.com",
	}}
	repo := &oidcRepositoryFake{
		flow:        identity.OIDCAuthFlow{Intent: OIDCIntentLogin, ReturnTo: "/", CreatedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Minute)},
		identityErr: identity.ErrIdentityNotFound,
	}
	client := &oidcClientFake{claims: OIDCClaims{
		Issuer: "https://idp.example.com", Subject: "subject-1", Email: "person@example.com", EmailVerified: true,
	}}
	service := NewService(users, passwordResetHasher{}, nil, nil)
	service.now = func() time.Time { return now }
	service.ConfigureOIDC(repo, client, &oidcTokenManagerFake{}, OIDCConfig{Enabled: true, JITEnabled: true})

	_, err := service.CompleteOIDC(context.Background(), CompleteOIDCInput{Code: "code", State: "state", BrowserToken: "browser"})
	if !errors.Is(err, ErrIdentityConflict) {
		t.Fatalf("expected identity conflict for an existing email, got %v", err)
	}
	if repo.createOIDCUserCalls != 0 {
		t.Fatalf("expected no automatic account link or JIT user, got %d create calls", repo.createOIDCUserCalls)
	}
}

func TestCompleteOIDCSudoRejectsIdentityLinkedToAnotherUser(t *testing.T) {
	now := time.Date(2026, time.July, 16, 10, 0, 0, 0, time.UTC)
	sessionID := "22222222-2222-2222-2222-222222222222"
	recent := &recentAuthenticationFake{session: identity.AuthSession{
		ID: sessionID, UserID: "11111111-1111-1111-1111-111111111111",
	}}
	repo := &oidcRepositoryFake{
		flow:           identity.OIDCAuthFlow{Intent: OIDCIntentSudo, SessionID: &sessionID, ReturnTo: "/settings", CreatedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Minute)},
		linkedIdentity: identity.UserIdentity{ID: "33333333-3333-3333-3333-333333333333", UserID: "44444444-4444-4444-4444-444444444444"},
	}
	client := &oidcClientFake{claims: OIDCClaims{
		Issuer: "https://idp.example.com", Subject: "subject-1", AuthTime: now,
	}}
	service := NewService(&oidcUserRepositoryFake{}, passwordResetHasher{}, nil, nil)
	service.now = func() time.Time { return now }
	service.recentAuth = recent
	service.ConfigureOIDC(repo, client, &oidcTokenManagerFake{}, OIDCConfig{Enabled: true})

	_, err := service.CompleteOIDC(context.Background(), CompleteOIDCInput{Code: "code", State: "state", BrowserToken: "browser"})
	if !errors.Is(err, ErrIdentityConflict) {
		t.Fatalf("expected cross-account identity conflict, got %v", err)
	}
	if recent.elevated {
		t.Fatal("cross-account identity must not elevate the session")
	}
}

func TestRecentOIDCAuthenticationTimeUsesProviderTime(t *testing.T) {
	now := time.Date(2026, time.July, 16, 10, 5, 0, 0, time.UTC)
	flowCreatedAt := now.Add(-time.Minute)
	sessionCreatedAt := now.Add(-time.Hour)
	providerAuthTime := now.Add(-30 * time.Second)

	authenticatedAt, ok := recentOIDCAuthenticationTime(providerAuthTime, flowCreatedAt, sessionCreatedAt, now, time.Minute)
	if !ok || !authenticatedAt.Equal(providerAuthTime) {
		t.Fatalf("expected provider authentication time %s, got %s (ok=%t)", providerAuthTime, authenticatedAt, ok)
	}
	if _, ok := recentOIDCAuthenticationTime(flowCreatedAt.Add(-2*time.Minute), flowCreatedAt, sessionCreatedAt, now, time.Minute); ok {
		t.Fatal("expected stale provider authentication to be rejected")
	}
}

func TestNormalizeReturnToAllowsRelativePathsAndRejectsRedirects(t *testing.T) {
	if got := normalizeReturnTo("/settings?tab=login-methods"); got != "/settings?tab=login-methods" {
		t.Fatalf("expected relative settings path, got %q", got)
	}
	for _, unsafe := range []string{"//evil.example.com", "/\\evil.example.com", "/settings\r\nLocation: https://evil.example.com"} {
		if got := normalizeReturnTo(unsafe); got != "/" {
			t.Fatalf("expected unsafe return path %q to normalize to root, got %q", unsafe, got)
		}
	}
}

type oidcUserRepositoryFake struct {
	userByEmail identity.User
	emailErr    error
}

func (*oidcUserRepositoryFake) CreateUser(context.Context, identity.User) (identity.User, error) {
	return identity.User{}, nil
}

func (*oidcUserRepositoryFake) GetUserByID(context.Context, string) (identity.User, error) {
	return identity.User{}, identity.ErrUserNotFound
}

func (r *oidcUserRepositoryFake) GetUserByEmail(context.Context, string) (identity.User, error) {
	if r.emailErr != nil {
		return identity.User{}, r.emailErr
	}
	if r.userByEmail.ID == "" {
		return identity.User{}, identity.ErrUserNotFound
	}
	return r.userByEmail, nil
}

func (*oidcUserRepositoryFake) UpdateUserPasswordHash(context.Context, identity.User) (identity.User, error) {
	return identity.User{}, nil
}

type oidcRepositoryFake struct {
	createdFlow         identity.OIDCAuthFlow
	flow                identity.OIDCAuthFlow
	linkedIdentity      identity.UserIdentity
	identityErr         error
	createOIDCUserCalls int
}

func (r *oidcRepositoryFake) CreateOIDCUser(context.Context, string, string, string, string, time.Time) (identity.User, identity.UserIdentity, error) {
	r.createOIDCUserCalls++
	return identity.User{}, identity.UserIdentity{}, nil
}

func (*oidcRepositoryFake) CreateUserIdentity(_ context.Context, input identity.UserIdentity) (identity.UserIdentity, error) {
	return input, nil
}

func (r *oidcRepositoryFake) GetUserIdentityByIssuerSubject(context.Context, string, string) (identity.UserIdentity, error) {
	if r.identityErr != nil {
		return identity.UserIdentity{}, r.identityErr
	}
	return r.linkedIdentity, nil
}

func (*oidcRepositoryFake) GetUserIdentityByIDForUser(context.Context, string, string) (identity.UserIdentity, error) {
	return identity.UserIdentity{}, identity.ErrIdentityNotFound
}

func (*oidcRepositoryFake) ListUserIdentities(context.Context, string) ([]identity.UserIdentity, error) {
	return nil, nil
}

func (*oidcRepositoryFake) TouchUserIdentityLogin(_ context.Context, input identity.UserIdentity, _ time.Time) (identity.UserIdentity, error) {
	return input, nil
}

func (r *oidcRepositoryFake) CreateOIDCAuthFlow(_ context.Context, input identity.OIDCAuthFlow) (identity.OIDCAuthFlow, error) {
	r.createdFlow = input
	return input, nil
}

func (r *oidcRepositoryFake) ConsumeOIDCAuthFlow(context.Context, []byte, []byte, time.Time) (identity.OIDCAuthFlow, error) {
	return r.flow, nil
}

func (*oidcRepositoryFake) DeleteExpiredOIDCAuthFlows(context.Context, time.Time) error { return nil }

type oidcClientFake struct {
	authorizationURL      string
	claims                OIDCClaims
	forceReauthentication bool
}

func (c *oidcClientFake) AuthorizationURL(_ context.Context, _, _, _ string, forceReauthentication bool) (string, error) {
	c.forceReauthentication = forceReauthentication
	return c.authorizationURL, nil
}

func (c *oidcClientFake) Exchange(context.Context, string, string, string) (OIDCClaims, error) {
	return c.claims, nil
}

type oidcTokenManagerFake struct {
	tokens []string
	next   int
}

func (m *oidcTokenManagerFake) Generate(context.Context) (string, error) {
	if m.next >= len(m.tokens) {
		return "unused", nil
	}
	token := m.tokens[m.next]
	m.next++
	return token, nil
}

func (*oidcTokenManagerFake) Hash(value string) string { return "hash:" + value }

type recentAuthenticationFake struct {
	session  identity.AuthSession
	elevated bool
}

func (*recentAuthenticationFake) SudoStatus(context.Context, string) (identity.SudoStatus, error) {
	return identity.SudoStatus{}, nil
}

func (r *recentAuthenticationFake) ElevateSession(context.Context, string, string, *string, time.Time) error {
	r.elevated = true
	return nil
}

func (r *recentAuthenticationFake) GetSession(context.Context, string) (identity.AuthSession, error) {
	return r.session, nil
}

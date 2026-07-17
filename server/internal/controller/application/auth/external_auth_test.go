package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

func TestStartExternalAuthPersistsProviderBoundFlow(t *testing.T) {
	now := time.Date(2026, time.July, 16, 10, 0, 0, 0, time.UTC)
	repo := &externalAuthRepositoryFake{}
	client := &externalAuthClientFake{authorizationURL: "https://idp.example.com/authorize"}
	service := newExternalAuthTestService(repo, client, ExternalProviderConfig{ID: identity.AuthenticationMethodOIDC})
	service.now = func() time.Time { return now }

	result, err := service.StartExternalAuth(context.Background(), StartExternalAuthInput{
		Provider: identity.AuthenticationMethodOIDC, Intent: ExternalAuthIntentLogin, ReturnTo: "//evil.example.com",
	})
	if err != nil {
		t.Fatalf("start external auth: %v", err)
	}
	if result.AuthorizationURL != client.authorizationURL || result.BrowserToken != "browser" {
		t.Fatalf("unexpected start result: %#v", result)
	}
	if repo.createdFlow.Provider != identity.AuthenticationMethodOIDC || string(repo.createdFlow.StateHash) != "hash:state" || string(repo.createdFlow.BrowserTokenHash) != "hash:browser" {
		t.Fatalf("flow was not provider-bound with hashed tokens: %#v", repo.createdFlow)
	}
	if repo.createdFlow.ReturnTo != "/" || repo.createdFlow.Nonce != "nonce" || repo.createdFlow.PKCEVerifier != "pkce" {
		t.Fatalf("unexpected stored flow: %#v", repo.createdFlow)
	}
	if client.intent != ExternalAuthIntentLogin {
		t.Fatalf("unexpected provider intent %q", client.intent)
	}
}

func TestStartExternalAuthAllowsGitHubSudo(t *testing.T) {
	repo := &externalAuthRepositoryFake{}
	client := &externalAuthClientFake{authorizationURL: "https://github.com/login/oauth/authorize"}
	service := newExternalAuthTestService(repo, client, ExternalProviderConfig{ID: identity.AuthenticationMethodGitHub, SudoCapable: true})
	service.recentAuth = &recentAuthenticationFake{}

	result, err := service.StartExternalAuth(context.Background(), StartExternalAuthInput{
		Provider: identity.AuthenticationMethodGitHub, Intent: ExternalAuthIntentSudo, SessionID: "session-id",
	})
	if err != nil {
		t.Fatalf("start GitHub sudo: %v", err)
	}
	if result.AuthorizationURL != client.authorizationURL || client.intent != ExternalAuthIntentSudo || repo.createdFlow.SessionID == nil || *repo.createdFlow.SessionID != "session-id" {
		t.Fatalf("unexpected GitHub sudo flow: result=%#v flow=%#v intent=%q", result, repo.createdFlow, client.intent)
	}
}

func TestSudoStatusIncludesGitHubProvider(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"
	users := &externalAuthUserRepositoryFake{userByID: identity.User{ID: userID}}
	repo := &externalAuthRepositoryFake{identities: []identity.UserIdentity{
		{Provider: identity.AuthenticationMethodGitHub},
		{Provider: identity.AuthenticationMethodGoogle},
	}}
	service := NewService(users, passwordResetHasher{}, nil, nil)
	service.recentAuth = &recentAuthenticationFake{}
	service.ConfigureExternalAuth(repo, &externalAuthTokenManagerFake{}, ExternalAuthConfig{},
		ExternalProviderRegistration{Config: ExternalProviderConfig{ID: identity.AuthenticationMethodGitHub, SudoCapable: true}, Client: &externalAuthClientFake{}},
		ExternalProviderRegistration{Config: ExternalProviderConfig{ID: identity.AuthenticationMethodGoogle, SudoCapable: true}, Client: &externalAuthClientFake{}},
	)

	status, err := service.SudoStatus(context.Background(), userID, "session-id")
	if err != nil {
		t.Fatalf("get sudo status: %v", err)
	}
	if len(status.Methods) != 2 || status.Methods[0] != identity.AuthenticationMethodGitHub || status.Methods[1] != identity.AuthenticationMethodGoogle {
		t.Fatalf("expected linked GitHub and Google sudo methods, got %#v", status.Methods)
	}
}

func TestCompleteExternalAuthDoesNotAutoLinkExistingEmail(t *testing.T) {
	now := time.Date(2026, time.July, 16, 10, 0, 0, 0, time.UTC)
	users := &externalAuthUserRepositoryFake{userByEmail: identity.User{
		ID: "11111111-1111-1111-1111-111111111111", Email: "person@example.com",
	}}
	repo := &externalAuthRepositoryFake{
		flow:        identity.ExternalAuthFlow{Provider: identity.AuthenticationMethodOIDC, Intent: ExternalAuthIntentLogin, ReturnTo: "/", CreatedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Minute)},
		identityErr: identity.ErrIdentityNotFound,
	}
	client := &externalAuthClientFake{claims: ExternalIdentityClaims{
		Issuer: "https://idp.example.com", Subject: "subject-1", Email: "person@example.com", EmailVerified: true,
	}}
	service := NewService(users, passwordResetHasher{}, nil, nil)
	service.now = func() time.Time { return now }
	service.ConfigureExternalAuth(repo, &externalAuthTokenManagerFake{}, ExternalAuthConfig{}, ExternalProviderRegistration{
		Config: ExternalProviderConfig{ID: identity.AuthenticationMethodOIDC, JITEnabled: true}, Client: client,
	})

	_, err := service.CompleteExternalAuth(context.Background(), CompleteExternalAuthInput{
		Provider: identity.AuthenticationMethodOIDC, Code: "code", State: "state", BrowserToken: "browser",
	})
	if !errors.Is(err, ErrIdentityConflict) {
		t.Fatalf("expected identity conflict for an existing email, got %v", err)
	}
	if repo.createUserCalls != 0 {
		t.Fatalf("expected no automatic account link or JIT user, got %d create calls", repo.createUserCalls)
	}
}

func TestCompleteExternalAuthSudoRejectsIdentityLinkedToAnotherUser(t *testing.T) {
	now := time.Date(2026, time.July, 16, 10, 0, 0, 0, time.UTC)
	sessionID := "22222222-2222-2222-2222-222222222222"
	recent := &recentAuthenticationFake{session: identity.AuthSession{
		ID: sessionID, UserID: "11111111-1111-1111-1111-111111111111",
	}}
	repo := &externalAuthRepositoryFake{
		flow:           identity.ExternalAuthFlow{Provider: identity.AuthenticationMethodGoogle, Intent: ExternalAuthIntentSudo, SessionID: &sessionID, ReturnTo: "/settings", CreatedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Minute)},
		linkedIdentity: identity.UserIdentity{ID: "33333333-3333-3333-3333-333333333333", UserID: "44444444-4444-4444-4444-444444444444"},
	}
	client := &externalAuthClientFake{claims: ExternalIdentityClaims{
		Issuer: "https://accounts.google.com", Subject: "subject-1", AuthTime: now,
	}}
	service := newExternalAuthTestService(repo, client, ExternalProviderConfig{ID: identity.AuthenticationMethodGoogle, SudoCapable: true})
	service.now = func() time.Time { return now }
	service.recentAuth = recent

	_, err := service.CompleteExternalAuth(context.Background(), CompleteExternalAuthInput{
		Provider: identity.AuthenticationMethodGoogle, Code: "code", State: "state", BrowserToken: "browser",
	})
	if !errors.Is(err, ErrIdentityConflict) {
		t.Fatalf("expected cross-account identity conflict, got %v", err)
	}
	if recent.elevated {
		t.Fatal("cross-account identity must not elevate the session")
	}
}

func TestCompleteGitHubSudoUsesVerifiedFlowCompletionTime(t *testing.T) {
	now := time.Date(2026, time.July, 16, 10, 0, 0, 0, time.UTC)
	sessionID := "22222222-2222-2222-2222-222222222222"
	userID := "11111111-1111-1111-1111-111111111111"
	identityID := "33333333-3333-3333-3333-333333333333"
	recent := &recentAuthenticationFake{session: identity.AuthSession{
		ID: sessionID, UserID: userID, CreatedAt: now.Add(-time.Hour),
	}}
	repo := &externalAuthRepositoryFake{
		flow: identity.ExternalAuthFlow{
			Provider: identity.AuthenticationMethodGitHub, Intent: ExternalAuthIntentSudo, SessionID: &sessionID,
			ReturnTo: "/settings?reauth=set-password", CreatedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Minute),
		},
		linkedIdentity: identity.UserIdentity{ID: identityID, UserID: userID, Provider: identity.AuthenticationMethodGitHub},
	}
	client := &externalAuthClientFake{claims: ExternalIdentityClaims{
		Issuer: "https://github.com", Subject: "1234567",
	}}
	service := newExternalAuthTestService(repo, client, ExternalProviderConfig{ID: identity.AuthenticationMethodGitHub, SudoCapable: true})
	service.now = func() time.Time { return now }
	service.recentAuth = recent

	result, err := service.CompleteExternalAuth(context.Background(), CompleteExternalAuthInput{
		Provider: identity.AuthenticationMethodGitHub, Code: "code", State: "state", BrowserToken: "browser",
	})
	if err != nil {
		t.Fatalf("complete GitHub sudo: %v", err)
	}
	if result.ReturnTo != "/settings?reauth=set-password" {
		t.Fatalf("expected password flow return path, got %q", result.ReturnTo)
	}
	if !recent.elevated || recent.elevatedMethod != identity.AuthenticationMethodGitHub || !recent.elevatedAt.Equal(now) {
		t.Fatalf("unexpected GitHub sudo elevation: elevated=%t method=%q at=%s", recent.elevated, recent.elevatedMethod, recent.elevatedAt)
	}
}

func TestCompleteExternalAuthSudoErrorPreservesReturnPath(t *testing.T) {
	now := time.Date(2026, time.July, 16, 10, 0, 0, 0, time.UTC)
	sessionID := "22222222-2222-2222-2222-222222222222"
	repo := &externalAuthRepositoryFake{flow: identity.ExternalAuthFlow{
		Provider: identity.AuthenticationMethodGoogle, Intent: ExternalAuthIntentSudo, SessionID: &sessionID,
		ReturnTo: "/settings?reauth=set-password", CreatedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Minute),
	}}
	service := newExternalAuthTestService(repo, &externalAuthClientFake{}, ExternalProviderConfig{
		ID: identity.AuthenticationMethodGoogle, SudoCapable: true,
	})
	service.now = func() time.Time { return now }

	result, err := service.CompleteExternalAuth(context.Background(), CompleteExternalAuthInput{
		Provider: identity.AuthenticationMethodGoogle, State: "state", BrowserToken: "browser",
	})
	if !errors.Is(err, ErrExternalAuthCallbackInvalid) {
		t.Fatalf("expected invalid callback after provider cancellation, got %v", err)
	}
	if result.Intent != ExternalAuthIntentSudo || result.ReturnTo != "/settings?reauth=set-password" {
		t.Fatalf("expected sudo return path to survive callback failure, got %#v", result)
	}
}

func TestRecentExternalAuthenticationTimeUsesProviderTime(t *testing.T) {
	now := time.Date(2026, time.July, 16, 10, 5, 0, 0, time.UTC)
	flowCreatedAt := now.Add(-time.Minute)
	sessionCreatedAt := now.Add(-time.Hour)
	providerAuthTime := now.Add(-30 * time.Second)

	authenticatedAt, ok := recentExternalAuthenticationTime(providerAuthTime, flowCreatedAt, sessionCreatedAt, now, time.Minute)
	if !ok || !authenticatedAt.Equal(providerAuthTime) {
		t.Fatalf("expected provider authentication time %s, got %s (ok=%t)", providerAuthTime, authenticatedAt, ok)
	}
	if _, ok := recentExternalAuthenticationTime(flowCreatedAt.Add(-2*time.Minute), flowCreatedAt, sessionCreatedAt, now, time.Minute); ok {
		t.Fatal("expected stale provider authentication to be rejected")
	}
}

func TestExternalAuthenticationFlowTimeRejectsFlowPredatingSession(t *testing.T) {
	now := time.Date(2026, time.July, 16, 10, 5, 0, 0, time.UTC)
	if validExternalAuthenticationFlowTime(now.Add(-2*time.Minute), now, now, time.Minute) {
		t.Fatal("expected an OAuth flow predating the current session to be rejected")
	}
	if !validExternalAuthenticationFlowTime(now.Add(-time.Second), now, now, time.Minute) {
		t.Fatal("expected a new session-bound OAuth flow to be accepted within clock skew")
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

func newExternalAuthTestService(repo *externalAuthRepositoryFake, client ExternalAuthClient, config ExternalProviderConfig) *Service {
	service := NewService(&externalAuthUserRepositoryFake{}, passwordResetHasher{}, nil, nil)
	service.ConfigureExternalAuth(repo, &externalAuthTokenManagerFake{tokens: []string{"state", "browser", "nonce", "pkce"}}, ExternalAuthConfig{
		FlowTTL: 10 * time.Minute, AuthTimeSkew: time.Minute,
	}, ExternalProviderRegistration{Config: config, Client: client})
	return service
}

type externalAuthUserRepositoryFake struct {
	userByID    identity.User
	userByEmail identity.User
	emailErr    error
}

func (*externalAuthUserRepositoryFake) CreateUser(context.Context, identity.User) (identity.User, error) {
	return identity.User{}, nil
}

func (r *externalAuthUserRepositoryFake) GetUserByID(context.Context, string) (identity.User, error) {
	if r.userByID.ID == "" {
		return identity.User{}, identity.ErrUserNotFound
	}
	return r.userByID, nil
}

func (r *externalAuthUserRepositoryFake) GetUserByEmail(context.Context, string) (identity.User, error) {
	if r.emailErr != nil {
		return identity.User{}, r.emailErr
	}
	if r.userByEmail.ID == "" {
		return identity.User{}, identity.ErrUserNotFound
	}
	return r.userByEmail, nil
}

func (*externalAuthUserRepositoryFake) UpdateUserPasswordHash(context.Context, identity.User) (identity.User, error) {
	return identity.User{}, nil
}

type externalAuthRepositoryFake struct {
	createdFlow     identity.ExternalAuthFlow
	flow            identity.ExternalAuthFlow
	linkedIdentity  identity.UserIdentity
	identities      []identity.UserIdentity
	identityErr     error
	createUserCalls int
}

func (r *externalAuthRepositoryFake) CreateExternalAuthUser(context.Context, string, string, identity.UserIdentity, time.Time) (identity.User, identity.UserIdentity, error) {
	r.createUserCalls++
	return identity.User{}, identity.UserIdentity{}, nil
}

func (*externalAuthRepositoryFake) CreateUserIdentity(_ context.Context, input identity.UserIdentity) (identity.UserIdentity, error) {
	return input, nil
}

func (r *externalAuthRepositoryFake) GetUserIdentityByIssuerSubject(context.Context, string, string, string) (identity.UserIdentity, error) {
	if r.identityErr != nil {
		return identity.UserIdentity{}, r.identityErr
	}
	return r.linkedIdentity, nil
}

func (r *externalAuthRepositoryFake) GetUserIdentityByIDForUser(context.Context, string, string) (identity.UserIdentity, error) {
	if r.linkedIdentity.ID == "" {
		return identity.UserIdentity{}, identity.ErrIdentityNotFound
	}
	return r.linkedIdentity, nil
}

func (r *externalAuthRepositoryFake) ListUserIdentities(context.Context, string) ([]identity.UserIdentity, error) {
	return r.identities, nil
}

func (*externalAuthRepositoryFake) TouchUserIdentityLogin(_ context.Context, input identity.UserIdentity, _ time.Time) (identity.UserIdentity, error) {
	return input, nil
}

func (r *externalAuthRepositoryFake) CreateExternalAuthFlow(_ context.Context, input identity.ExternalAuthFlow) (identity.ExternalAuthFlow, error) {
	r.createdFlow = input
	return input, nil
}

func (r *externalAuthRepositoryFake) ConsumeExternalAuthFlow(context.Context, string, []byte, []byte, time.Time) (identity.ExternalAuthFlow, error) {
	return r.flow, nil
}

func (*externalAuthRepositoryFake) DeleteExpiredExternalAuthFlows(context.Context, time.Time) error {
	return nil
}

type externalAuthClientFake struct {
	authorizationURL string
	claims           ExternalIdentityClaims
	intent           string
}

func (c *externalAuthClientFake) AuthorizationURL(_ context.Context, _, _, _, intent string) (string, error) {
	c.intent = intent
	return c.authorizationURL, nil
}

func (c *externalAuthClientFake) Exchange(context.Context, string, string, string) (ExternalIdentityClaims, error) {
	return c.claims, nil
}

type externalAuthTokenManagerFake struct {
	tokens []string
	next   int
}

func (m *externalAuthTokenManagerFake) Generate(context.Context) (string, error) {
	if m.next >= len(m.tokens) {
		return "unused", nil
	}
	token := m.tokens[m.next]
	m.next++
	return token, nil
}
func (*externalAuthTokenManagerFake) Hash(value string) string { return "hash:" + value }

type recentAuthenticationFake struct {
	session        identity.AuthSession
	elevated       bool
	elevatedAt     time.Time
	elevatedMethod string
}

func (*recentAuthenticationFake) SudoStatus(context.Context, string) (identity.SudoStatus, error) {
	return identity.SudoStatus{Active: true}, nil
}

func (r *recentAuthenticationFake) ElevateSession(_ context.Context, _, method string, _ *string, authenticatedAt time.Time) error {
	r.elevated = true
	r.elevatedMethod = method
	r.elevatedAt = authenticatedAt
	return nil
}

func (r *recentAuthenticationFake) GetSession(context.Context, string) (identity.AuthSession, error) {
	return r.session, nil
}

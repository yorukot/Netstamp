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

func TestSudoStatusUsesDynamicProviderSource(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"
	users := &externalAuthUserRepositoryFake{userByID: identity.User{ID: userID}}
	repo := &externalAuthRepositoryFake{identities: []identity.UserIdentity{
		{Provider: identity.AuthenticationMethodGitHub},
	}}
	source := &externalProviderSourceFake{
		ids: []string{identity.AuthenticationMethodGitHub},
		registrations: map[string]ExternalProviderRegistration{
			identity.AuthenticationMethodGitHub: {
				Config: ExternalProviderConfig{ID: identity.AuthenticationMethodGitHub, SudoCapable: true},
				Client: &externalAuthClientFake{},
			},
		},
	}
	service := NewService(users, passwordResetHasher{}, nil, nil)
	service.recentAuth = &recentAuthenticationFake{}
	service.ConfigureExternalAuth(repo, &externalAuthTokenManagerFake{}, ExternalAuthConfig{})
	service.ConfigureExternalAuthProviderSource(source)

	status, err := service.SudoStatus(context.Background(), userID, "session-id")
	if err != nil {
		t.Fatalf("get dynamic sudo status: %v", err)
	}
	if len(status.Methods) != 1 || status.Methods[0] != identity.AuthenticationMethodGitHub {
		t.Fatalf("expected dynamic GitHub sudo method, got %#v", status.Methods)
	}
}

func TestSudoStatusIsolatesBrokenDynamicProvider(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"
	users := &externalAuthUserRepositoryFake{userByID: identity.User{ID: userID}}
	repo := &externalAuthRepositoryFake{identities: []identity.UserIdentity{
		{Provider: identity.AuthenticationMethodOIDC},
		{Provider: identity.AuthenticationMethodGoogle},
	}}
	source := &externalProviderSourceFake{
		ids: []string{identity.AuthenticationMethodOIDC, identity.AuthenticationMethodGoogle},
		registrations: map[string]ExternalProviderRegistration{
			identity.AuthenticationMethodGoogle: {
				Config: ExternalProviderConfig{ID: identity.AuthenticationMethodGoogle, SudoCapable: true},
				Client: &externalAuthClientFake{},
			},
		},
		errs: map[string]error{identity.AuthenticationMethodOIDC: errors.New("corrupt OIDC secret")},
	}
	service := NewService(users, passwordResetHasher{}, nil, nil)
	service.recentAuth = &recentAuthenticationFake{}
	service.ConfigureExternalAuth(repo, &externalAuthTokenManagerFake{}, ExternalAuthConfig{})
	service.ConfigureExternalAuthProviderSource(source)

	status, err := service.SudoStatus(context.Background(), userID, "session-id")
	if err != nil {
		t.Fatalf("get sudo status with one broken provider: %v", err)
	}
	if len(status.Methods) != 1 || status.Methods[0] != identity.AuthenticationMethodGoogle {
		t.Fatalf("expected only healthy Google sudo method, got %#v", status.Methods)
	}
}

func TestStartExternalAuthResolvesOnlyRequestedDynamicProvider(t *testing.T) {
	repo := &externalAuthRepositoryFake{}
	source := &externalProviderSourceFake{
		ids: []string{
			identity.AuthenticationMethodGoogle,
			identity.AuthenticationMethodGitHub,
			identity.AuthenticationMethodOIDC,
		},
		registrations: map[string]ExternalProviderRegistration{
			identity.AuthenticationMethodGitHub: {
				Config: ExternalProviderConfig{ID: identity.AuthenticationMethodGitHub},
				Client: &externalAuthClientFake{authorizationURL: "https://github.com/login/oauth/authorize"},
			},
		},
		errs: map[string]error{
			identity.AuthenticationMethodGoogle: errors.New("corrupt Google secret"),
			identity.AuthenticationMethodOIDC:   errors.New("corrupt OIDC secret"),
		},
	}
	service := NewService(&externalAuthUserRepositoryFake{}, passwordResetHasher{}, nil, nil)
	service.ConfigureExternalAuth(repo, &externalAuthTokenManagerFake{
		tokens: []string{"state", "browser", "nonce", "pkce"},
	}, ExternalAuthConfig{})
	service.ConfigureExternalAuthProviderSource(source)

	if _, err := service.StartExternalAuth(context.Background(), StartExternalAuthInput{
		Provider: identity.AuthenticationMethodGitHub,
	}); err != nil {
		t.Fatalf("start healthy dynamic GitHub provider: %v", err)
	}
	if len(source.resolved) != 1 || source.resolved[0] != identity.AuthenticationMethodGitHub {
		t.Fatalf("expected only GitHub resolution, got %#v", source.resolved)
	}
}

func TestDynamicExternalProviderReauthenticationElevatesExternalLoginSession(t *testing.T) {
	tests := []struct {
		provider string
		issuer   string
	}{
		{provider: identity.AuthenticationMethodOIDC, issuer: "https://idp.example.com"},
		{provider: identity.AuthenticationMethodGoogle, issuer: "https://accounts.google.com"},
		{provider: identity.AuthenticationMethodGitHub, issuer: "https://github.com"},
	}

	for _, test := range tests {
		t.Run(test.provider, func(t *testing.T) {
			now := time.Date(2026, time.July, 29, 10, 0, 0, 0, time.UTC)
			userID := "11111111-1111-1111-1111-111111111111"
			sessionID := "22222222-2222-2222-2222-222222222222"
			identityID := "33333333-3333-3333-3333-333333333333"
			users := &externalAuthUserRepositoryFake{userByID: identity.User{ID: userID}}
			repo := &externalAuthRepositoryFake{
				identities: []identity.UserIdentity{{
					ID: identityID, UserID: userID, Provider: test.provider,
				}},
				linkedIdentity: identity.UserIdentity{
					ID: identityID, UserID: userID, Provider: test.provider,
				},
			}
			client := &externalAuthClientFake{
				authorizationURL: "https://provider.example.com/authorize",
				claims: ExternalIdentityClaims{
					Issuer: test.issuer, Subject: "provider-subject", AuthTime: now,
				},
			}
			source := &externalProviderSourceFake{
				ids: []string{test.provider},
				registrations: map[string]ExternalProviderRegistration{
					test.provider: {
						Config: ExternalProviderConfig{
							ID: test.provider, DisplayName: test.provider, SudoCapable: true,
						},
						Client: client,
					},
				},
			}
			recent := &statefulRecentAuthenticationFake{session: identity.AuthSession{
				ID: sessionID, UserID: userID, AuthenticationMethod: test.provider,
				SudoEligible: false, IdentityID: &identityID, CreatedAt: now.Add(-time.Hour),
			}}
			service := NewService(users, passwordResetHasher{}, nil, nil)
			service.now = func() time.Time { return now }
			service.recentAuth = recent
			service.ConfigureExternalAuth(
				repo,
				&externalAuthTokenManagerFake{tokens: []string{"state", "browser", "nonce", "pkce"}},
				ExternalAuthConfig{FlowTTL: 10 * time.Minute, AuthTimeSkew: time.Minute},
			)
			service.ConfigureExternalAuthProviderSource(source)

			status, err := service.SudoStatus(context.Background(), userID, sessionID)
			if err != nil {
				t.Fatalf("get sudo status before reauthentication: %v", err)
			}
			if status.Active || len(status.Methods) != 1 || status.Methods[0] != test.provider {
				t.Fatalf("expected inactive session with linked dynamic %s method, got %#v", test.provider, status)
			}
			err = service.RequireSudo(context.Background(), sessionID)
			if !errors.Is(err, ErrSudoRequired) {
				t.Fatalf("expected sensitive operation to require reauthentication, got %v", err)
			}

			start, err := service.StartExternalAuth(context.Background(), StartExternalAuthInput{
				Provider: test.provider, Intent: ExternalAuthIntentSudo, SessionID: sessionID, ReturnTo: "/admin",
			})
			if err != nil {
				t.Fatalf("start dynamic %s reauthentication: %v", test.provider, err)
			}
			if start.AuthorizationURL != client.authorizationURL {
				t.Fatalf("unexpected authorization URL %q", start.AuthorizationURL)
			}
			repo.flow = repo.createdFlow

			result, err := service.CompleteExternalAuth(context.Background(), CompleteExternalAuthInput{
				Provider: test.provider, Code: "code", State: "state", BrowserToken: "browser",
			})
			if err != nil {
				t.Fatalf("complete dynamic %s reauthentication: %v", test.provider, err)
			}
			if result.Intent != ExternalAuthIntentSudo || result.ReturnTo != "/admin" || result.Access != nil {
				t.Fatalf("unexpected reauthentication result: %#v", result)
			}
			if err := service.RequireSudo(context.Background(), sessionID); err != nil {
				t.Fatalf("expected sensitive operation after reauthentication to pass: %v", err)
			}
			if !recent.elevated ||
				recent.elevatedMethod != test.provider ||
				recent.elevatedIdentityID == nil ||
				*recent.elevatedIdentityID != identityID ||
				!recent.elevatedAt.Equal(now) {
				t.Fatalf("unexpected session elevation: %#v", recent)
			}
			if len(source.resolved) != 3 {
				t.Fatalf("expected status, start, and callback to resolve DB runtime provider, got %#v", source.resolved)
			}
			for _, provider := range source.resolved {
				if provider != test.provider {
					t.Fatalf("expected only %s resolution, got %#v", test.provider, source.resolved)
				}
			}
		})
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
	createdFlow         identity.ExternalAuthFlow
	flow                identity.ExternalAuthFlow
	linkedIdentity      identity.UserIdentity
	identities          []identity.UserIdentity
	identityErr         error
	createUserCalls     int
	createIdentityCalls int
}

func (r *externalAuthRepositoryFake) CreateExternalAuthUser(context.Context, string, string, identity.UserIdentity, time.Time) (identity.User, identity.UserIdentity, error) {
	r.createUserCalls++
	return identity.User{}, identity.UserIdentity{}, nil
}

func (r *externalAuthRepositoryFake) CreateUserIdentity(_ context.Context, input identity.UserIdentity) (identity.UserIdentity, error) {
	r.createIdentityCalls++
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

type externalProviderSourceFake struct {
	ids           []string
	registrations map[string]ExternalProviderRegistration
	errs          map[string]error
	resolved      []string
}

func (s *externalProviderSourceFake) ExternalProviderIDs() []string {
	return s.ids
}

func (s *externalProviderSourceFake) ExternalProviderRegistration(_ context.Context, provider string) (ExternalProviderRegistration, error) {
	s.resolved = append(s.resolved, provider)
	if err := s.errs[provider]; err != nil {
		return ExternalProviderRegistration{}, err
	}
	return s.registrations[provider], nil
}

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

type statefulRecentAuthenticationFake struct {
	session            identity.AuthSession
	elevated           bool
	elevatedAt         time.Time
	elevatedMethod     string
	elevatedIdentityID *string
}

func (r *statefulRecentAuthenticationFake) SudoStatus(context.Context, string) (identity.SudoStatus, error) {
	return identity.SudoStatus{Active: r.elevated}, nil
}

func (r *statefulRecentAuthenticationFake) ElevateSession(
	_ context.Context,
	sessionID string,
	method string,
	identityID *string,
	authenticatedAt time.Time,
) error {
	if sessionID != r.session.ID {
		return ErrSessionInvalid
	}
	r.elevated = true
	r.elevatedAt = authenticatedAt
	r.elevatedMethod = method
	r.elevatedIdentityID = identityID
	return nil
}

func (r *statefulRecentAuthenticationFake) GetSession(
	_ context.Context,
	sessionID string,
) (identity.AuthSession, error) {
	if sessionID != r.session.ID {
		return identity.AuthSession{}, ErrSessionInvalid
	}
	return r.session, nil
}

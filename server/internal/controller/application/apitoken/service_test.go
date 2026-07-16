package apitoken

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

func TestCreateStoresOnlyHashedTokenWithNormalizedScopes(t *testing.T) {
	now := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	repo := &apiTokenTestRepository{}
	manager := &apiTokenTestManager{raw: "nst_pat_one-time-secret", hint: "e-secret"}
	service := NewService(repo, manager, nil)
	service.now = func() time.Time { return now }

	output, err := service.Create(context.Background(), CreateInput{
		CurrentUserID: " user-1 ",
		Name:          " CI deploy ",
		Scopes:        []string{string(identity.ScopeLabelsRead), string(identity.ScopeChecksRead)},
		ExpiresAt:     now.Add(90 * 24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	if output.RawToken != manager.raw {
		t.Fatalf("expected one-time token %q, got %q", manager.raw, output.RawToken)
	}
	if repo.created.Name != "CI deploy" || repo.created.UserID != "user-1" {
		t.Fatalf("unexpected normalized token: %#v", repo.created)
	}
	if string(repo.created.TokenHash) != "digest:"+manager.raw {
		t.Fatalf("expected token digest, got %q", repo.created.TokenHash)
	}
	if len(repo.created.Scopes) != 2 || repo.created.Scopes[0] != identity.ScopeChecksRead || repo.created.Scopes[1] != identity.ScopeLabelsRead {
		t.Fatalf("expected sorted scopes, got %#v", repo.created.Scopes)
	}
	if repo.maxActive != MaxActiveTokens || !repo.now.Equal(now) {
		t.Fatalf("expected max %d at %s, got %d at %s", MaxActiveTokens, now, repo.maxActive, repo.now)
	}
}

func TestCreateRejectsExpiryPastMaximumBeforeGeneratingToken(t *testing.T) {
	now := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	repo := &apiTokenTestRepository{}
	manager := &apiTokenTestManager{raw: "should-not-be-generated", hint: "enerated"}
	service := NewService(repo, manager, nil)
	service.now = func() time.Time { return now }

	_, err := service.Create(context.Background(), CreateInput{
		CurrentUserID: "user-1",
		Name:          "too long",
		Scopes:        []string{string(identity.ScopeProjectsRead)},
		ExpiresAt:     now.Add(MaxTokenTTL + time.Nanosecond),
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
	if manager.generated || repo.createCalled {
		t.Fatal("invalid input must not generate or persist a token")
	}
}

func TestCreateDoesNotRepeatPasswordVerificationAfterSudo(t *testing.T) {
	now := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	manager := &apiTokenTestManager{raw: "should-not-be-generated", hint: "enerated"}
	service := NewService(&apiTokenTestRepository{}, manager, nil)
	service.now = func() time.Time { return now }

	_, err := service.Create(context.Background(), CreateInput{
		CurrentUserID: "user-1",
		Name:          "CI",
		Scopes:        []string{string(identity.ScopeProjectsRead)},
		ExpiresAt:     now.Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("expected token creation after transport sudo check, got %v", err)
	}
	if !manager.generated {
		t.Fatal("token should be generated without another password comparison")
	}
}

func TestVerifyReturnsPrincipalAndThrottlesLastUsedWrites(t *testing.T) {
	now := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	repo := &apiTokenTestRepository{active: identity.APIToken{
		ID: "token-1", UserID: "user-1", Scopes: []identity.APITokenScope{identity.ScopeProjectsRead},
	}}
	service := NewService(repo, &apiTokenTestManager{}, nil)
	service.now = func() time.Time { return now }

	principal, err := service.Verify(context.Background(), "nst_pat_secret")
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}
	if principal.TokenID != "token-1" || principal.UserID != "user-1" || !principal.HasScope(identity.ScopeProjectsRead) {
		t.Fatalf("unexpected principal: %#v", principal)
	}
	if string(repo.gotHash) != "digest:nst_pat_secret" {
		t.Fatalf("unexpected lookup digest %q", repo.gotHash)
	}
	if repo.touchCount != 1 || !repo.touchedAt.Equal(now) || !repo.touchBefore.Equal(now.Add(-touchInterval)) {
		t.Fatalf("unexpected touch: count=%d at=%s before=%s", repo.touchCount, repo.touchedAt, repo.touchBefore)
	}

	recent := now.Add(-time.Minute)
	repo.active.LastUsedAt = &recent
	if _, err := service.Verify(context.Background(), "nst_pat_secret"); err != nil {
		t.Fatalf("verify recently used token: %v", err)
	}
	if repo.touchCount != 1 {
		t.Fatalf("expected throttled last-used write, got %d touches", repo.touchCount)
	}
}

func TestVerifyDistinguishesUnknownTokenFromRepositoryFailure(t *testing.T) {
	notFoundRepo := &apiTokenTestRepository{getErr: identity.ErrAPITokenNotFound}
	service := NewService(notFoundRepo, &apiTokenTestManager{}, nil)
	if _, err := service.Verify(context.Background(), "unknown"); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("expected invalid token, got %v", err)
	}

	repositoryErr := errors.New("database unavailable")
	failingRepo := &apiTokenTestRepository{getErr: repositoryErr}
	service = NewService(failingRepo, &apiTokenTestManager{}, nil)
	if _, err := service.Verify(context.Background(), "unknown"); !errors.Is(err, repositoryErr) {
		t.Fatalf("expected repository failure, got %v", err)
	}
}

type apiTokenTestRepository struct {
	created      identity.APIToken
	active       identity.APIToken
	createErr    error
	getErr       error
	gotHash      []byte
	maxActive    int
	now          time.Time
	createCalled bool
	touchCount   int
	touchedAt    time.Time
	touchBefore  time.Time
}

func (r *apiTokenTestRepository) Create(_ context.Context, token identity.APIToken, maxActive int, now time.Time) (identity.APIToken, error) {
	r.createCalled = true
	r.created = token
	r.maxActive = maxActive
	r.now = now
	if r.createErr != nil {
		return identity.APIToken{}, r.createErr
	}
	token.ID = "token-1"
	return token, nil
}

func (r *apiTokenTestRepository) ListForUser(context.Context, string) ([]identity.APIToken, error) {
	return nil, nil
}

func (r *apiTokenTestRepository) GetActiveByHash(_ context.Context, hash []byte, _ time.Time) (identity.APIToken, error) {
	r.gotHash = append([]byte(nil), hash...)
	if r.getErr != nil {
		return identity.APIToken{}, r.getErr
	}
	return r.active, nil
}

func (r *apiTokenTestRepository) Touch(_ context.Context, _ string, lastUsedAt, touchBefore time.Time) error {
	r.touchCount++
	r.touchedAt = lastUsedAt
	r.touchBefore = touchBefore
	return nil
}

func (r *apiTokenTestRepository) RevokeForUser(context.Context, string, string, string, time.Time) error {
	return nil
}

func (r *apiTokenTestRepository) RevokeForUserAll(context.Context, string, string, time.Time) error {
	return nil
}

type apiTokenTestManager struct {
	raw       string
	hint      string
	generated bool
}

func (m *apiTokenTestManager) Generate() (string, string, error) {
	m.generated = true
	return m.raw, m.hint, nil
}

func (*apiTokenTestManager) Hash(raw string) []byte { return []byte("digest:" + raw) }

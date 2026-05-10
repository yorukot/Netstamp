package middleware

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

func TestRequireAuthRejectsMissingBearerToken(t *testing.T) {
	_, api := humatest.New(t)
	verifier := &recordingTokenVerifier{}
	registerClaimsRoute(t, api, verifier)

	res := api.Get("/me")

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", res.Code)
	}
	if verifier.gotToken != "" {
		t.Fatalf("expected verifier not to be called, got token %q", verifier.gotToken)
	}
	if got := res.Header().Get("WWW-Authenticate"); got != "Bearer" {
		t.Fatalf("expected WWW-Authenticate Bearer, got %q", got)
	}
}

func TestRequireAuthRejectsInvalidAccessToken(t *testing.T) {
	_, api := humatest.New(t)
	verifier := &recordingTokenVerifier{err: identity.ErrAccessTokenInvalid}
	registerClaimsRoute(t, api, verifier)

	res := api.Get("/me", "Authorization: Bearer bad-token")

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", res.Code)
	}
	if verifier.gotToken != "bad-token" {
		t.Fatalf("expected verifier token %q, got %q", "bad-token", verifier.gotToken)
	}
}

func TestRequireAuthStoresClaimsInContext(t *testing.T) {
	_, api := humatest.New(t)
	verifier := &recordingTokenVerifier{
		claims: identity.AccessTokenClaims{
			Subject: "user-1",
			Email:   "user@example.com",
		},
	}
	registerClaimsRoute(t, api, verifier)

	res := api.Get("/me", "Authorization: bearer good-token")

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}
	if verifier.gotToken != "good-token" {
		t.Fatalf("expected verifier token %q, got %q", "good-token", verifier.gotToken)
	}
}

func registerClaimsRoute(t *testing.T, api huma.API, verifier appauth.TokenVerifier) {
	t.Helper()

	huma.Register(api, huma.Operation{
		Method: http.MethodGet,
		Path:   "/me",
		Middlewares: huma.Middlewares{
			RequireAuth(verifier),
		},
	}, func(ctx context.Context, _ *struct{}) (*claimsRouteOutput, error) {
		claims, ok := AccessTokenClaimsFromContext(ctx)
		if !ok {
			return nil, errors.New("missing claims")
		}
		if claims.Subject == "" || claims.Email == "" {
			return nil, errors.New("empty claims")
		}
		return &claimsRouteOutput{
			Body: claimsRouteBody{
				Subject: claims.Subject,
				Email:   claims.Email,
			},
		}, nil
	})
}

type claimsRouteOutput struct {
	Body claimsRouteBody
}

type claimsRouteBody struct {
	Subject string `json:"subject"`
	Email   string `json:"email"`
}

type recordingTokenVerifier struct {
	claims   identity.AccessTokenClaims
	err      error
	gotToken string
}

func (v *recordingTokenVerifier) VerifyAccessToken(_ context.Context, value string) (identity.AccessTokenClaims, error) {
	v.gotToken = value
	if v.err != nil {
		return identity.AccessTokenClaims{}, v.err
	}
	return v.claims, nil
}

package userhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

func TestChangeCurrentUserPasswordReturnsForbiddenWhenCredentialChangesDisabled(t *testing.T) {
	router := chi.NewRouter()
	NewHandler(nil, &userHandlerTokenVerifier{
		claims: identity.AccessTokenClaims{
			Subject: "11111111-1111-1111-1111-111111111111",
			Email:   "user@example.com",
		},
	}, false).RegisterRoutes(router)

	req := httptest.NewRequest(http.MethodPost, "/users/me/password-change", strings.NewReader(`{"currentPassword":"old-password","newPassword":"new-password"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "netstamp_session=valid-token")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", res.Code)
	}

	var body httpx.ProblemDetails
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Detail != "credential changes are disabled" {
		t.Fatalf("expected disabled credential changes detail, got %q", body.Detail)
	}
}

type userHandlerTokenVerifier struct {
	claims identity.AccessTokenClaims
}

func (v *userHandlerTokenVerifier) VerifyAccessToken(context.Context, string) (identity.AccessTokenClaims, error) {
	return v.claims, nil
}

var _ appauth.TokenVerifier = (*userHandlerTokenVerifier)(nil)

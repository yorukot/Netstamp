package security

import (
	"context"
	"testing"
	"time"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

func TestJWTIssuerRoundTripsClaims(t *testing.T) {
	issuer := NewJWTIssuer("secret", time.Hour)

	token, err := issuer.IssueAccessToken(context.Background(), identity.AccessTokenClaims{
		Subject: "user-1",
		Email:   "user@example.com",
	})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	claims, err := issuer.VerifyAccessToken(context.Background(), token.Value)
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}
	if claims.Subject != "user-1" {
		t.Fatalf("expected subject, got %q", claims.Subject)
	}
	if claims.Email != "user@example.com" {
		t.Fatalf("expected email, got %q", claims.Email)
	}
}

package security

import (
	"context"
	"testing"
	"time"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

func TestJWTIssuerRoundTripsDisplayName(t *testing.T) {
	issuer := NewJWTIssuer("secret", time.Hour)

	displayName := "Example User"
	token, err := issuer.IssueAccessToken(context.Background(), identity.AccessTokenInput{
		Subject:     "user-1",
		Email:       "user@example.com",
		DisplayName: &displayName,
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
	if claims.DisplayName == nil || *claims.DisplayName != "Example User" {
		t.Fatalf("expected display name, got %#v", claims.DisplayName)
	}
}

package app

import (
	"context"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

type accountStatusTokenVerifier struct {
	inner appauth.TokenVerifier
	users appauth.UserRepository
}

func (v accountStatusTokenVerifier) VerifyAccessToken(ctx context.Context, value string) (identity.AccessTokenClaims, error) {
	claims, err := v.inner.VerifyAccessToken(ctx, value)
	if err != nil {
		return identity.AccessTokenClaims{}, err
	}

	user, err := v.users.GetUserByID(ctx, claims.Subject)
	if err != nil || user.DisabledAt != nil {
		return identity.AccessTokenClaims{}, appauth.ErrAccessTokenInvalid
	}

	return claims, nil
}

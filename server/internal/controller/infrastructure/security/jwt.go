package security

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

type JWTIssuer struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

type accessTokenClaims struct {
	Email       string  `json:"email"`
	DisplayName *string `json:"displayName,omitempty"`
	jwt.RegisteredClaims
}

func NewJWTIssuer(secret string, ttl time.Duration) *JWTIssuer {
	return &JWTIssuer{
		secret: []byte(secret),
		ttl:    ttl,
		now:    time.Now,
	}
}

func (i *JWTIssuer) IssueAccessToken(ctx context.Context, input identity.AccessTokenInput) (identity.IssuedToken, error) {
	if err := ctx.Err(); err != nil {
		return identity.IssuedToken{}, err
	}

	now := i.now().UTC()
	expiresAt := now.Add(i.ttl)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims{
		Email:       input.Email,
		DisplayName: input.DisplayName,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   input.Subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	})

	value, err := token.SignedString(i.secret)
	if err != nil {
		return identity.IssuedToken{}, err
	}

	return identity.IssuedToken{
		Value:     value,
		TokenType: "Bearer",
		ExpiresIn: int(i.ttl.Seconds()),
	}, nil
}

func (i *JWTIssuer) VerifyAccessToken(ctx context.Context, value string) (identity.AccessTokenClaims, error) {
	if err := ctx.Err(); err != nil {
		return identity.AccessTokenClaims{}, err
	}

	var claims accessTokenClaims
	token, err := jwt.ParseWithClaims(value, &claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, identity.ErrAccessTokenInvalid
		}
		return i.secret, nil
	})
	if err != nil {
		return identity.AccessTokenClaims{}, errors.Join(identity.ErrAccessTokenInvalid, err)
	}
	if token == nil || !token.Valid || claims.Subject == "" || claims.Email == "" {
		return identity.AccessTokenClaims{}, identity.ErrAccessTokenInvalid
	}

	return identity.AccessTokenClaims{
		Subject:     claims.Subject,
		Email:       claims.Email,
		DisplayName: claims.DisplayName,
	}, nil
}

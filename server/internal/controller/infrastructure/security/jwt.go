package security

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"

	authapp "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

const (
	jwtIssuer   = "netstamp"
	jwtAudience = "netstamp"
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

func (i *JWTIssuer) IssueAccessToken(ctx context.Context, input identity.AccessTokenClaims) (identity.IssuedToken, error) {
	if err := ctx.Err(); err != nil {
		return identity.IssuedToken{}, err
	}

	now := i.now().UTC()
	expiresAt := now.Add(i.ttl)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims{
		Email: input.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jwtIssuer,
			Audience:  jwt.ClaimStrings{jwtAudience},
			Subject:   input.Subject,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	})

	value, err := token.SignedString(i.secret)
	if err != nil {
		return identity.IssuedToken{}, err
	}

	return identity.IssuedToken{
		Value:     value,
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
			return nil, authapp.ErrAccessTokenInvalid
		}
		return i.secret, nil
	})
	if err != nil {
		return identity.AccessTokenClaims{}, errors.Join(authapp.ErrAccessTokenInvalid, err)
	}
	if token == nil || !token.Valid || claims.Subject == "" || claims.Email == "" {
		return identity.AccessTokenClaims{}, authapp.ErrAccessTokenInvalid
	}
	if claims.Issuer != jwtIssuer {
		return identity.AccessTokenClaims{}, authapp.ErrAccessTokenInvalid
	}
	if !claims.VerifyAudience(jwtAudience, true) {
		return identity.AccessTokenClaims{}, authapp.ErrAccessTokenInvalid
	}

	return identity.AccessTokenClaims{
		Subject: claims.Subject,
		Email:   claims.Email,
	}, nil
}

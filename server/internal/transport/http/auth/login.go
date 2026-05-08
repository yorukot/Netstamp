package auth

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	appauth "github.com/yorukot/netstamp/internal/application/auth"
)

func (h *Handler) login(ctx context.Context, input *loginInput) (*loginOutput, error) {
	result, err := h.service.Login(ctx, appauth.LoginInput{
		Email:    input.Body.Email,
		Password: input.Body.Password,
	})

	if err != nil {
		switch {
		case errors.Is(err, appauth.ErrCredentialsInvalid):
			return nil, huma.Error401Unauthorized("invalid email or password")
		default:
			return nil, huma.Error500InternalServerError("login failed")
		}
	}

	return &loginOutput{
		Body: loginOutputBody{
			User: userResponse{
				ID:          result.UserID,
				Email:       result.Email,
				DisplayName: result.DisplayName,
			},
			TokenType:   result.TokenType,
			AccessToken: result.AccessToken,
			ExpiresIn:   result.ExpiresIn,
		},
	}, nil
}

type loginInput struct {
	Body loginInputBody
}

type loginOutput struct {
	Body loginOutputBody
}

type loginInputBody struct {
	Email    string `json:"email" format:"email" maxLength:"254" required:"true" doc:"Email address used to sign in. It is normalized before lookup." example:"user@example.com"`
	Password string `json:"password" minLength:"1" maxLength:"128" required:"true" writeOnly:"true" doc:"Plain-text password to verify. It is never returned by the API." example:"correct-horse-battery-staple"`
}

type loginOutputBody struct {
	User        userResponse `json:"user" doc:"Authenticated user."`
	TokenType   string       `json:"tokenType" example:"Bearer" doc:"Token scheme to use in the Authorization header."`
	AccessToken string       `json:"accessToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.example.signature" doc:"JWT access token issued for the authenticated user."`
	ExpiresIn   int          `json:"expiresIn" example:"43200" doc:"Access token lifetime in seconds."`
}

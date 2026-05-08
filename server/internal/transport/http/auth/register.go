package auth

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"

	appauth "github.com/yorukot/netstamp/internal/application/auth"
)

func (h *Handler) register(ctx context.Context, input *registerInput) (*registerOutput, error) {
	result, err := h.service.Register(ctx, appauth.RegisterInput{
		Email:       input.Body.Email,
		DisplayName: input.Body.DisplayName,
		Password:    input.Body.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, appauth.ErrDisplayNameRequired):
			return nil, huma.Error422UnprocessableEntity("display name is required")
		case errors.Is(err, appauth.ErrDisplayNameTooLong):
			return nil, huma.Error422UnprocessableEntity("display name is too long")
		case errors.Is(err, appauth.ErrEmailAlreadyExists):
			return nil, huma.Error409Conflict("email already exists")
		default:
			return nil, huma.Error500InternalServerError("register user failed")
		}
	}

	return &registerOutput{
		Body: registerOutputBody{
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

type registerInput struct {
	Body registerInputBody
}

type registerOutput struct {
	Body registerOutputBody
}

type registerInputBody struct {
	Email       string `json:"email" format:"email" maxLength:"254" required:"true" doc:"Email address used to sign in." example:"user@example.com"`
	DisplayName string `json:"displayName" minLength:"1" maxLength:"100" required:"true" doc:"Name shown in the app." example:"Jane Doe"`
	Password    string `json:"password" minLength:"8" maxLength:"128" required:"true" writeOnly:"true" doc:"Plain-text password. It is stored only as an Argon2id hash." example:"correct-horse-battery-staple"`
}

type registerOutputBody struct {
	User        userResponse `json:"user" doc:"Created user."`
	TokenType   string       `json:"tokenType" example:"Bearer" doc:"Token scheme to use in the Authorization header."`
	AccessToken string       `json:"accessToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.example.signature" doc:"JWT access token issued for the created user."`
	ExpiresIn   int          `json:"expiresIn" example:"43200" doc:"Access token lifetime in seconds."`
}

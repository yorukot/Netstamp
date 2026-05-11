package auth

import (
	"context"

	"github.com/danielgtaylor/huma/v2"

	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
)

func (h *Handler) me(ctx context.Context, _ *meInput) (*meOutput, error) {
	claims, ok := httpmiddleware.AccessTokenClaimsFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing bearer token")
	}

	// TODO: We should lookup for the user detail and return it back
	
	return &meOutput{
		Body: meOutputBody{
			Authenticated: true,
			User: userResponse{
				ID:          claims.Subject,
				Email:       claims.Email,
			},
		},
	}, nil
}

type meInput struct{}

type meOutput struct {
	Body meOutputBody
}

type meOutputBody struct {
	Authenticated bool         `json:"authenticated" example:"true" doc:"Always true when the bearer token is valid."`
	User          userResponse `json:"user" doc:"User identity from the verified access token claims."`
}

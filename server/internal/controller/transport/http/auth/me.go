package auth

import (
	"context"

	"github.com/danielgtaylor/huma/v2"

	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
)

func (h *Handler) me(ctx context.Context, _ *meInput) (*meOutput, error) {
	claims, _ := httpmiddleware.AccessTokenClaimsFromContext(ctx)

	user, err := h.service.GetCurrentUser(ctx, claims.Email)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to fetch user")
	}

	return &meOutput{
		Body: meOutputBody{
			Authenticated: true,
			User: userResponse{
				ID:          user.ID,
				Email:       user.Email,
				DisplayName: user.DisplayName,
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
	User          userResponse `json:"user" doc:"Current user fetched live from the database."`
}

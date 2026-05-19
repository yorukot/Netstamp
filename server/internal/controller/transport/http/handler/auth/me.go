package auth

import (
	"context"

	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
)

func (h *Handler) me(ctx context.Context, _ *meInput) (*meOutput, error) {
	claims, _ := httpmiddleware.AccessTokenClaimsFromContext(ctx)

	user, err := h.service.GetCurrentUser(ctx, claims.Email)
	if err != nil {
		return nil, httpx.InternalServerError("failed to fetch user")
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
	Authenticated bool         `json:"authenticated"`
	User          userResponse `json:"user"`
}

package auth

import (
	"context"
	"errors"

	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

func (h *Handler) me(ctx context.Context, _ *meInput) (*meOutput, error) {
	claims, _ := httpmiddleware.AccessTokenClaimsFromContext(ctx)

	user, err := h.service.GetCurrentUser(ctx, claims.Subject)
	if err != nil {
		if errors.Is(err, identity.ErrUserNotFound) {
			return nil, httpx.Unauthorized("invalid session")
		}
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

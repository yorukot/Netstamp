package userhttp

import (
	"context"

	appuser "github.com/yorukot/netstamp/internal/controller/application/user"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

func (h *Handler) changeCurrentUserPassword(ctx context.Context, input *changeCurrentUserPasswordInput) error {
	if !h.credentialChangesEnabled {
		return httpx.Forbidden("credential changes are disabled")
	}

	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return err
	}

	err = h.service.ChangeCurrentUserPassword(ctx, appuser.ChangeCurrentUserPasswordInput{
		CurrentUserID:   currentUserID,
		CurrentPassword: input.Body.CurrentPassword,
		NewPassword:     input.Body.NewPassword,
	})
	if err != nil {
		return mapUserError(err, "change current user password failed")
	}

	return nil
}

type changeCurrentUserPasswordInput struct {
	Body changeCurrentUserPasswordInputBody
}

type changeCurrentUserPasswordInputBody struct {
	CurrentPassword string `json:"currentPassword,omitempty"`
	NewPassword     string `json:"newPassword,omitempty"`
}

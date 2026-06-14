package userhttp

import (
	"context"

	appuser "github.com/yorukot/netstamp/internal/controller/application/user"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

func (h *Handler) changeCurrentUserEmail(ctx context.Context, input *changeCurrentUserEmailInput) (*userOutput, error) {
	if !h.credentialChangesEnabled {
		return nil, httpx.Forbidden("credential changes are disabled")
	}

	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	output, err := h.service.ChangeCurrentUserEmail(ctx, appuser.ChangeCurrentUserEmailInput{
		CurrentUserID: currentUserID,
		NewEmail:      input.Body.NewEmail,
		Password:      input.Body.Password,
	})
	if err != nil {
		return nil, mapUserError(err, "change current user email failed")
	}

	return newUserOutput(output.User), nil
}

type changeCurrentUserEmailInput struct {
	Body changeCurrentUserEmailInputBody
}

type changeCurrentUserEmailInputBody struct {
	NewEmail string `json:"newEmail,omitempty"`
	Password string `json:"password,omitempty"`
}

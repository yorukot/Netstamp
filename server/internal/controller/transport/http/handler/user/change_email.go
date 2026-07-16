package userhttp

import (
	"context"

	appuser "github.com/yorukot/netstamp/internal/controller/application/user"
)

func (h *Handler) changeCurrentUserEmail(ctx context.Context, input *changeCurrentUserEmailInput) (*userOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	output, err := h.service.ChangeCurrentUserEmail(ctx, appuser.ChangeCurrentUserEmailInput{
		CurrentUserID: currentUserID,
		NewEmail:      input.Body.NewEmail,
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
}

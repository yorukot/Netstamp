package userhttp

import (
	"context"

	appuser "github.com/yorukot/netstamp/internal/controller/application/user"
)

func (h *Handler) updateCurrentUser(ctx context.Context, input *updateCurrentUserInput) (*userOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	output, err := h.service.UpdateCurrentUser(ctx, appuser.UpdateCurrentUserInput{
		CurrentUserID: currentUserID,
		DisplayName:   input.Body.DisplayName,
	})
	if err != nil {
		return nil, mapUserError(err, "update current user failed")
	}

	return newUserOutput(output.User), nil
}

type updateCurrentUserInput struct {
	Body updateCurrentUserInputBody
}

type updateCurrentUserInputBody struct {
	DisplayName *string `json:"displayName,omitempty"`
}

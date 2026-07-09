package userhttp

import (
	"context"

	appuser "github.com/yorukot/netstamp/internal/controller/application/user"
)

func (h *Handler) deactivateCurrentUser(ctx context.Context) error {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return err
	}

	if err := h.service.DeactivateCurrentUser(ctx, appuser.DeactivateCurrentUserInput{CurrentUserID: currentUserID}); err != nil {
		return mapUserError(err, "deactivate current user failed")
	}
	return nil
}

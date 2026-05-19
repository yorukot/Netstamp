package project

import (
	"context"

	appproject "github.com/yorukot/netstamp/internal/controller/application/project"
)

func (h *Handler) removeMember(ctx context.Context, input *removeMemberInput) (*removeMemberOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.service.RemoveMember(ctx, appproject.RemoveMemberInput{
		CurrentUserID: currentUserID,
		ProjectRef:    input.Ref,
		UserID:        input.UserID,
	}); err != nil {
		return nil, mapProjectError(err, "remove project member failed")
	}

	return &removeMemberOutput{}, nil
}

type removeMemberInput struct {
	Ref    string
	UserID string
}

type removeMemberOutput struct{}

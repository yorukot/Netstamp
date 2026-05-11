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
	Ref    string `path:"ref" minLength:"1" maxLength:"64" pattern:"^[a-z0-9-]+$" patternDescription:"lowercase letters, numbers, and dashes" doc:"Project UUID or slug." example:"engineering"`
	UserID string `path:"user_id" format:"uuid" doc:"User ID." example:"00000000-0000-0000-0000-000000000000"`
}

type removeMemberOutput struct{}

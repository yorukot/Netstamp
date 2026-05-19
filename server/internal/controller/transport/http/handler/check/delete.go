package check

import (
	"context"

	appcheck "github.com/yorukot/netstamp/internal/controller/application/check"
)

func (h *Handler) deleteCheck(ctx context.Context, input *deleteCheckInput) (*deleteCheckOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.service.DeleteCheck(ctx, appcheck.GetCheckInput{
		CurrentUserID: currentUserID,
		ProjectRef:    input.Ref,
		CheckID:       input.CheckID,
	}); err != nil {
		return nil, mapCheckError(err, "delete check failed")
	}

	return &deleteCheckOutput{}, nil
}

type deleteCheckInput struct {
	Ref     string
	CheckID string
}

type deleteCheckOutput struct{}

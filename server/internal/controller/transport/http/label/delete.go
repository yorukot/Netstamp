package label

import (
	"context"

	applabel "github.com/yorukot/netstamp/internal/controller/application/label"
)

func (h *Handler) deleteLabel(ctx context.Context, input *labelRefInput) (*deleteLabelOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.service.DeleteLabel(ctx, applabel.DeleteLabelInput{
		CurrentUserID: currentUserID,
		ProjectRef:    input.Ref,
		LabelID:       input.LabelID,
	}); err != nil {
		return nil, mapLabelError(err, "delete label failed")
	}

	return &deleteLabelOutput{}, nil
}

type labelRefInput struct {
	Ref     string
	LabelID string
}

type deleteLabelOutput struct{}

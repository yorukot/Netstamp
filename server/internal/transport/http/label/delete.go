package label

import (
	"context"

	applabel "github.com/yorukot/netstamp/internal/application/label"
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
	Ref     string `path:"ref" minLength:"1" maxLength:"100" doc:"Project UUID or slug." example:"engineering"`
	LabelID string `path:"label_id" format:"uuid" doc:"Label ID."`
}

type deleteLabelOutput struct{}

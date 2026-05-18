package label

import (
	"context"

	applabel "github.com/yorukot/netstamp/internal/controller/application/label"
)

func (h *Handler) updateLabel(ctx context.Context, input *updateLabelInput) (*labelOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	label, err := h.service.UpdateLabel(ctx, applabel.UpdateLabelInput{
		CurrentUserID: currentUserID,
		ProjectRef:    input.Ref,
		LabelID:       input.LabelID,
		Key:           input.Body.Key,
		Value:         input.Body.Value,
	})
	if err != nil {
		return nil, mapLabelError(err, "update label failed")
	}

	return &labelOutput{Body: labelOutputBody{Label: label}}, nil
}

type updateLabelInput struct {
	Ref     string
	LabelID string
	Body    updateLabelInputBody
}

type updateLabelInputBody struct {
	Key   *string `json:"key,omitempty"`
	Value *string `json:"value,omitempty"`
}

package label

import (
	"context"

	applabel "github.com/yorukot/netstamp/internal/controller/application/label"
)

func (h *Handler) createLabel(ctx context.Context, input *createLabelInput) (*labelOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	label, err := h.service.CreateLabel(ctx, applabel.CreateLabelInput{
		CurrentUserID: currentUserID,
		ProjectRef:    input.Ref,
		Key:           input.Body.Key,
		Value:         input.Body.Value,
	})
	if err != nil {
		return nil, mapLabelError(err, "create label failed")
	}

	return &labelOutput{Body: labelOutputBody{Label: label}}, nil
}

type createLabelInput struct {
	Ref  string `path:"ref" minLength:"1" maxLength:"64" pattern:"^[a-z0-9-]+$" patternDescription:"lowercase letters, numbers, and dashes" doc:"Project UUID or slug." example:"engineering"`
	Body createLabelInputBody
}

type createLabelInputBody struct {
	Key   string `json:"key,omitempty" maxLength:"64" doc:"Label key." example:"region"`
	Value string `json:"value,omitempty" maxLength:"64" doc:"Label value." example:"tokyo"`
}

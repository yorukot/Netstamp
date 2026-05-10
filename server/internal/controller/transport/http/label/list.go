package label

import (
	"context"

	applabel "github.com/yorukot/netstamp/internal/controller/application/label"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
)

func (h *Handler) listLabels(ctx context.Context, input *listLabelsInput) (*listLabelsOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	labels, err := h.service.ListLabels(ctx, applabel.ListLabelsInput{
		CurrentUserID: currentUserID,
		ProjectRef:    input.Ref,
	})
	if err != nil {
		return nil, mapLabelError(err, "list labels failed")
	}

	return &listLabelsOutput{Body: listLabelsOutputBody{Labels: newLabelResponses(labels)}}, nil
}

type listLabelsInput struct {
	Ref string `path:"ref" doc:"Project UUID or slug." example:"engineering"`
}

type listLabelsOutput struct {
	Body listLabelsOutputBody
}

type listLabelsOutputBody struct {
	Labels []labelResponse `json:"labels"`
}

func newLabelResponses(labels []domainlabel.Label) []labelResponse {
	responses := make([]labelResponse, 0, len(labels))
	for _, label := range labels {
		responses = append(responses, newLabelResponse(label))
	}

	return responses
}

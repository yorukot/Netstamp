package assignment

import (
	"context"
	"encoding/json"

	appassignment "github.com/yorukot/netstamp/internal/controller/application/assignment"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

func (h *Handler) previewSelector(ctx context.Context, input *previewSelectorInput) (*previewSelectorOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	output, err := h.service.PreviewSelector(ctx, appassignment.PreviewSelectorInput{
		CurrentUserID: currentUserID,
		ProjectRef:    input.Ref,
		Selector:      input.Body.Selector,
	})
	if err != nil {
		return nil, mapAssignmentError(err, "preview selector failed")
	}

	return &previewSelectorOutput{Body: previewSelectorOutputBody{
		Selector:     output.Selector,
		MatchedCount: output.MatchedCount,
		Probes:       output.Probes,
	}}, nil
}

type previewSelectorInput struct {
	Ref  string
	Body previewSelectorInputBody
}

type previewSelectorInputBody struct {
	Selector json.RawMessage `json:"selector,omitempty"`
}

type previewSelectorOutput struct {
	Body previewSelectorOutputBody
}

type previewSelectorOutputBody struct {
	Selector     json.RawMessage     `json:"selector"`
	MatchedCount int32               `json:"matchedCount"`
	Probes       []domainprobe.Probe `json:"probes"`
}

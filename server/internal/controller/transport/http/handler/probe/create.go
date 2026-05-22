package probe

import (
	"context"

	appprobe "github.com/yorukot/netstamp/internal/controller/application/probe"
)

func (h *Handler) createProbe(ctx context.Context, input *createProbeInput) (*createProbeOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	output, err := h.service.CreateProbe(ctx, appprobe.CreateProbeInput{
		CurrentUserID: currentUserID,
		ProjectRef:    input.Ref,
		Name:          input.Body.Name,
		Enabled:       input.Body.Enabled,
		LocationName:  input.Body.LocationName,
		Latitude:      input.Body.Latitude,
		Longitude:     input.Body.Longitude,
		LabelIDs:      input.Body.LabelIDs,
	})
	if err != nil {
		return nil, mapProbeError(err, "create probe failed")
	}

	return &createProbeOutput{
		Body: createProbeOutputBody{
			Probe:  output.Probe,
			Secret: output.Secret,
		},
	}, nil
}

type createProbeInput struct {
	Ref  string
	Body createProbeInputBody
}

type createProbeInputBody struct {
	Name         string   `json:"name,omitempty"`
	Enabled      *bool    `json:"enabled,omitempty"`
	LocationName *string  `json:"locationName,omitempty"`
	Latitude     *float64 `json:"latitude,omitempty"`
	Longitude    *float64 `json:"longitude,omitempty"`
	LabelIDs     []string `json:"labelIds,omitempty"`
}

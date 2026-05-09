package probe

import (
	"context"

	appprobe "github.com/yorukot/netstamp/internal/application/probe"
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
		City:          input.Body.City,
		Latitude:      input.Body.Latitude,
		Longitude:     input.Body.Longitude,
		LabelIDs:      input.Body.LabelIDs,
	})
	if err != nil {
		return nil, mapProbeError(err, "create probe failed")
	}

	return &createProbeOutput{
		Body: createProbeOutputBody{
			Probe:  newProbeResponse(output.Probe),
			Secret: output.Secret,
		},
	}, nil
}

type createProbeInput struct {
	Ref  string `path:"ref" doc:"Project UUID or slug." example:"engineering"`
	Body createProbeInputBody
}

type createProbeInputBody struct {
	Name      string   `json:"name,omitempty" doc:"Probe display name." example:"tokyo-vps-1"`
	Enabled   *bool    `json:"enabled,omitempty" doc:"Whether the probe is enabled. Defaults to true when omitted." example:"true"`
	City      *string  `json:"city,omitempty" doc:"Location code stored in the city field." example:"JP-13"`
	Latitude  *float64 `json:"latitude,omitempty" doc:"Probe latitude. Must be provided with longitude." example:"35.6762"`
	Longitude *float64 `json:"longitude,omitempty" doc:"Probe longitude. Must be provided with latitude." example:"139.6503"`
	LabelIDs  []string `json:"labelIds,omitempty" doc:"Existing project label IDs to attach to the probe."`
}

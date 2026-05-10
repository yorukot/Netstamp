package probe

import (
	"context"

	appprobe "github.com/yorukot/netstamp/internal/application/probe"
)

func (h *Handler) updateProbe(ctx context.Context, input *updateProbeInput) (*probeOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	probe, err := h.service.UpdateProbe(ctx, appprobe.UpdateProbeInput{
		CurrentUserID: currentUserID,
		ProjectRef:    input.Ref,
		ProbeID:       input.ProbeID,
		Name:          input.Body.Name,
		Enabled:       input.Body.Enabled,
		City:          input.Body.City,
		Latitude:      input.Body.Latitude,
		Longitude:     input.Body.Longitude,
		LabelIDs:      input.Body.LabelIDs,
	})
	if err != nil {
		return nil, mapProbeError(err, "update probe failed")
	}

	return &probeOutput{Body: probeOutputBody{Probe: newProbeResponse(probe)}}, nil
}

type updateProbeInput struct {
	Ref     string `path:"ref" doc:"Project UUID or slug." example:"engineering"`
	ProbeID string `path:"probe_id" doc:"Probe ID." example:"33333333-3333-3333-3333-333333333333"`
	Body    updateProbeInputBody
}

type updateProbeInputBody struct {
	Name      *string   `json:"name,omitempty" doc:"Probe display name." example:"tokyo-vps-1"`
	Enabled   *bool     `json:"enabled,omitempty" doc:"Whether the probe is enabled." example:"true"`
	City      *string   `json:"city,omitempty" doc:"Location code stored in the city field." example:"JP-13"`
	Latitude  *float64  `json:"latitude,omitempty" doc:"Probe latitude. Must be provided with longitude." example:"35.6762"`
	Longitude *float64  `json:"longitude,omitempty" doc:"Probe longitude. Must be provided with latitude." example:"139.6503"`
	LabelIDs  *[]string `json:"labelIds,omitempty" doc:"Existing project label IDs to attach to the probe. Replaces the full set when provided."`
}

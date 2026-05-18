package probe

import (
	"context"

	appprobe "github.com/yorukot/netstamp/internal/controller/application/probe"
)

func (h *Handler) updateProbe(ctx context.Context, input *updateProbeInput) (*probeOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	probe, err := h.service.UpdateProbe(ctx, appprobe.UpdateProbeInput{
		CurrentUserID:   currentUserID,
		ProjectRef:      input.Ref,
		ProbeID:         input.ProbeID,
		Name:            input.Body.Name,
		Enabled:         input.Body.Enabled,
		SubdivisionCode: input.Body.SubdivisionCode,
		Latitude:        input.Body.Latitude,
		Longitude:       input.Body.Longitude,
		LabelIDs:        input.Body.LabelIDs,
	})
	if err != nil {
		return nil, mapProbeError(err, "update probe failed")
	}

	return &probeOutput{Body: probeOutputBody{Probe: probe}}, nil
}

type updateProbeInput struct {
	Ref     string
	ProbeID string
	Body    updateProbeInputBody
}

type updateProbeInputBody struct {
	Name            *string   `json:"name,omitempty"`
	Enabled         *bool     `json:"enabled,omitempty"`
	SubdivisionCode *string   `json:"subdivisionCode,omitempty"`
	Latitude        *float64  `json:"latitude,omitempty"`
	Longitude       *float64  `json:"longitude,omitempty"`
	LabelIDs        *[]string `json:"labelIds,omitempty"`
}

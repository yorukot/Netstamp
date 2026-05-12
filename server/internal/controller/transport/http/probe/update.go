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
	Ref     string `path:"ref" doc:"Project UUID or slug." example:"engineering"`
	ProbeID string `path:"probe_id" minLength:"1" format:"uuid" doc:"Probe ID." example:"33333333-3333-3333-3333-333333333333"`
	Body    updateProbeInputBody
}

type updateProbeInputBody struct {
	Name            *string   `json:"name,omitempty" minLength:"1" maxLength:"128" doc:"Probe display name." example:"tokyo-vps-1"`
	Enabled         *bool     `json:"enabled,omitempty" doc:"Whether the probe is enabled. Omitted values leave the current setting unchanged." example:"true"`
	SubdivisionCode *string   `json:"subdivisionCode,omitempty" pattern:"^[A-Z]{2}-[A-Z0-9]{1,3}$" patternDescription:"value must be ISO 3166-2 subdivision code" doc:"ISO 3166-2 subdivision code, e.g. JP-12 for Chiba Prefecture, Japan"`
	Latitude        *float64  `json:"latitude,omitempty" minimum:"-90" maximum:"90" doc:"Probe latitude. Must be provided with longitude." example:"35.6762"`
	Longitude       *float64  `json:"longitude,omitempty" minimum:"-180" maximum:"180" doc:"Probe longitude. Must be provided with latitude." example:"139.6503"`
	LabelIDs        *[]string `json:"labelIds,omitempty" doc:"Existing project label IDs to attach to the probe."`
}

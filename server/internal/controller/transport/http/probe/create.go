package probe

import (
	"context"

	appprobe "github.com/yorukot/netstamp/internal/controller/application/proberegistry"
)

func (h *Handler) createProbe(ctx context.Context, input *createProbeInput) (*createProbeOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	output, err := h.service.CreateProbe(ctx, appprobe.CreateProbeInput{
		CurrentUserID:   currentUserID,
		ProjectRef:      input.Ref,
		Name:            input.Body.Name,
		Enabled:         input.Body.Enabled,
		SubdivisionCode: input.Body.SubdivisionCode,
		Latitude:        input.Body.Latitude,
		Longitude:       input.Body.Longitude,
		LabelIDs:        input.Body.LabelIDs,
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
	Ref  string `path:"ref" doc:"Project UUID or slug." example:"engineering"`
	Body createProbeInputBody
}

type createProbeInputBody struct {
	Name            string   `json:"name,omitempty" minLength:"1" maxLength:"128" doc:"Probe display name." example:"tokyo-vps-1"`
	Enabled         *bool    `json:"enabled,omitempty" default:"true" doc:"Whether the probe is enabled. Defaults to true when omitted." example:"true"`
	SubdivisionCode *string  `json:"subdivisionCode" pattern:"^[A-Z]{2}-[A-Z0-9]{1,3}$" patternDescription:"value must be ISO 3166-2 subdivision code" doc:"ISO 3166-2 subdivision code, e.g. JP-12 for Chiba Prefecture, Japan"`
	Latitude        *float64 `json:"latitude,omitempty" default:"0" minimum:"-90" maximum:"90" doc:"Probe latitude. Must be provided with longitude." example:"35.6762"`
	Longitude       *float64 `json:"longitude,omitempty" default:"0" minimum:"-180" maximum:"180" doc:"Probe longitude. Must be provided with latitude." example:"139.6503"`
	LabelIDs        []string `json:"labelIds,omitempty" doc:"Existing project label IDs to attach to the probe."`
}

package probe

import (
	"context"

	appprobe "github.com/yorukot/netstamp/internal/controller/application/proberegistry"
)

func (h *Handler) getProbe(ctx context.Context, input *probeRefInput) (*probeOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	probe, err := h.service.GetProbe(ctx, appprobe.GetProbeInput{
		CurrentUserID: currentUserID,
		ProjectRef:    input.Ref,
		ProbeID:       input.ProbeID,
	})
	if err != nil {
		return nil, mapProbeError(err, "get probe failed")
	}

	return &probeOutput{Body: probeOutputBody{Probe: newProbeResponse(probe)}}, nil
}

type probeRefInput struct {
	Ref     string `path:"ref" doc:"Project UUID or slug." example:"engineering"`
	ProbeID string `path:"probe_id" doc:"Probe ID." example:"33333333-3333-3333-3333-333333333333"`
}

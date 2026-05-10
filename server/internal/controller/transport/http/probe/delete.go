package probe

import (
	"context"

	appprobe "github.com/yorukot/netstamp/internal/controller/application/proberegistry"
)

func (h *Handler) deleteProbe(ctx context.Context, input *probeRefInput) (*deleteProbeOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.service.DeleteProbe(ctx, appprobe.DeleteProbeInput{
		CurrentUserID: currentUserID,
		ProjectRef:    input.Ref,
		ProbeID:       input.ProbeID,
	}); err != nil {
		return nil, mapProbeError(err, "delete probe failed")
	}

	return &deleteProbeOutput{}, nil
}

type deleteProbeOutput struct{}

package probe

import (
	"context"

	appprobe "github.com/yorukot/netstamp/internal/application/probe"
)

func (h *Handler) rotateSecret(ctx context.Context, input *probeRefInput) (*rotateSecretOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	output, err := h.service.RotateProbeSecret(ctx, appprobe.RotateProbeSecretInput{
		CurrentUserID: currentUserID,
		ProjectRef:    input.Ref,
		ProbeID:       input.ProbeID,
	})
	if err != nil {
		return nil, mapProbeError(err, "rotate probe secret failed")
	}

	return &rotateSecretOutput{
		Body: rotateSecretOutputBody{
			Probe:  newProbeResponse(output.Probe),
			Secret: output.Secret,
		},
	}, nil
}

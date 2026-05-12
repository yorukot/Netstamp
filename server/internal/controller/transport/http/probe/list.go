package probe

import (
	"context"

	appprobe "github.com/yorukot/netstamp/internal/controller/application/probe"
)

func (h *Handler) listProbes(ctx context.Context, input *listProbesInput) (*listProbesOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	probes, err := h.service.ListProbes(ctx, appprobe.ListProbesInput{
		CurrentUserID: currentUserID,
		ProjectRef:    input.Ref,
	})
	if err != nil {
		return nil, mapProbeError(err, "list probes failed")
	}

	return &listProbesOutput{Body: listProbesOutputBody{Probes: probes}}, nil
}

type listProbesInput struct {
	Ref string `path:"ref" doc:"Project UUID or slug." example:"engineering"`
}

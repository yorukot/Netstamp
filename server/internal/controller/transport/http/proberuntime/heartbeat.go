package proberuntime

import (
	"context"
	"time"
)

func (h *Handler) heartbeat(ctx context.Context, input *heartbeatInput) (*heartbeatOutput, error) {
	auth, err := requireRuntimeAuthInput(ctx)
	if err != nil {
		return nil, err
	}
	runtimeInput := newRuntimeStatusInput(auth, input.Body)
	output, err := h.service.Heartbeat(ctx, runtimeInput)
	if err != nil {
		return nil, mapRuntimeError(err, "update probe runtime status failed")
	}

	return &heartbeatOutput{Body: heartbeatOutputBody{ServerTime: output.ServerTime}}, nil
}

type heartbeatInput struct {
	ProbeID string `path:"probe_id" doc:"Probe ID."`
	Body    runtimeStatusBody
}

type heartbeatOutput struct {
	Body heartbeatOutputBody
}

type heartbeatOutputBody struct {
	ServerTime time.Time `json:"serverTime"`
}

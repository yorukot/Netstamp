package proberuntime

import (
	"context"
	"time"
)

func (h *Handler) hello(ctx context.Context, input *helloInput) (*helloOutput, error) {
	auth, err := requireRuntimeAuthInput(ctx)
	if err != nil {
		return nil, err
	}
	runtimeInput := newRuntimeStatusInput(auth, input.Body)
	output, err := h.service.Hello(ctx, runtimeInput)
	if err != nil {
		return nil, mapRuntimeError(err, "start probe runtime session failed")
	}

	return &helloOutput{Body: helloOutputBody{
		ServerTime:                    output.ServerTime,
		HeartbeatIntervalSeconds:      output.HeartbeatIntervalSeconds,
		AssignmentPollIntervalSeconds: output.AssignmentPollIntervalSeconds,
	}}, nil
}

type helloInput struct {
	ProbeID string `path:"probe_id" doc:"Probe ID."`
	Body    runtimeStatusBody
}

type helloOutput struct {
	Body helloOutputBody
}

type helloOutputBody struct {
	ServerTime                    time.Time `json:"serverTime"`
	HeartbeatIntervalSeconds      int32     `json:"heartbeatIntervalSeconds"`
	AssignmentPollIntervalSeconds int32     `json:"assignmentPollIntervalSeconds"`
}

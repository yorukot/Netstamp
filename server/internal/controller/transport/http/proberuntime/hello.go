package proberuntime

import (
	"context"
	"time"

	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

func (h *Handler) hello(ctx context.Context, input *helloInput) (*helloOutput, error) {
	auth, err := requireRuntimeAuthInput(ctx)
	if err != nil {
		return nil, err
	}
	output, err := h.service.Hello(ctx, auth)
	if err != nil {
		return nil, mapRuntimeError(err, "start probe runtime session failed")
	}

	return &helloOutput{Body: helloOutputBody{
		ServerTime:                   output.ServerTime,
		MinimumSupportedAgentVersion: output.MinimumSupportedAgentVersion,
		Config:                       output.Config,
	}}, nil
}

type helloInput struct {
	ProbeID string `path:"probe_id" doc:"Probe ID."`
}

type helloOutput struct {
	Body helloOutputBody
}

type helloOutputBody struct {
	ServerTime                   time.Time                 `json:"serverTime"`
	MinimumSupportedAgentVersion string                    `json:"minimumSupportedAgentVersion"`
	Config                       domainprobe.RuntimeConfig `json:"config"`
}

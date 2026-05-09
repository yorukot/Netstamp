package proberuntime

import (
	"context"
	"net/netip"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"

	appproberuntime "github.com/yorukot/netstamp/internal/application/proberuntime"
)

func (h *Handler) hello(ctx context.Context, input *runtimeStatusInput) (*helloOutput, error) {
	runtimeInput, err := newRuntimeStatusInput(input.ProbeID, input.Authorization, input.Body)
	if err != nil {
		return nil, err
	}
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

func (h *Handler) heartbeat(ctx context.Context, input *runtimeStatusInput) (*heartbeatOutput, error) {
	runtimeInput, err := newRuntimeStatusInput(input.ProbeID, input.Authorization, input.Body)
	if err != nil {
		return nil, err
	}
	output, err := h.service.Heartbeat(ctx, runtimeInput)
	if err != nil {
		return nil, mapRuntimeError(err, "update probe runtime status failed")
	}

	return &heartbeatOutput{Body: heartbeatOutputBody{ServerTime: output.ServerTime}}, nil
}

type runtimeStatusInput struct {
	ProbeID       string `path:"probe_id" doc:"Probe ID."`
	Authorization string `header:"Authorization" doc:"Probe authorization header. Use 'Probe <secret>'."`
	Body          runtimeStatusInputBody
}

type runtimeStatusInputBody struct {
	AgentVersion *string  `json:"agentVersion,omitempty" doc:"Probe agent version." example:"netstamp-probe/0.1.0"`
	PublicV4     *string  `json:"publicV4,omitempty" doc:"Observed public IPv4 address." example:"203.0.113.10"`
	PublicV6     *string  `json:"publicV6,omitempty" doc:"Observed public IPv6 address." example:"2001:db8::10"`
	Addrs        []string `json:"addrs,omitempty" doc:"Probe local or observed interface addresses."`
}

type helloOutput struct {
	Body helloOutputBody
}

type helloOutputBody struct {
	ServerTime                    time.Time `json:"serverTime"`
	HeartbeatIntervalSeconds      int32     `json:"heartbeatIntervalSeconds"`
	AssignmentPollIntervalSeconds int32     `json:"assignmentPollIntervalSeconds"`
}

type heartbeatOutput struct {
	Body heartbeatOutputBody
}

type heartbeatOutputBody struct {
	ServerTime time.Time `json:"serverTime"`
}

func newRuntimeStatusInput(probeID, header string, body runtimeStatusInputBody) (appproberuntime.RuntimeStatusInput, error) {
	auth, err := runtimeAuthInput(probeID, header)
	if err != nil {
		return appproberuntime.RuntimeStatusInput{}, err
	}
	publicV4, err := parseOptionalIP(body.PublicV4, true)
	if err != nil {
		return appproberuntime.RuntimeStatusInput{}, err
	}
	publicV6, err := parseOptionalIP(body.PublicV6, false)
	if err != nil {
		return appproberuntime.RuntimeStatusInput{}, err
	}
	addrs, err := parseAddrs(body.Addrs)
	if err != nil {
		return appproberuntime.RuntimeStatusInput{}, err
	}

	return appproberuntime.RuntimeStatusInput{
		RuntimeAuthInput: auth,
		AgentVersion:     body.AgentVersion,
		PublicV4:         publicV4,
		PublicV6:         publicV6,
		Addrs:            addrs,
	}, nil
}

func parseOptionalIP(value *string, requireIPv4 bool) (*netip.Addr, error) {
	if value == nil {
		return nil, nil //nolint:nilnil // Omitted IP fields are represented as nil.
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, huma.Error422UnprocessableEntity("invalid IP address")
	}
	addr, err := netip.ParseAddr(trimmed)
	if err != nil {
		return nil, huma.Error422UnprocessableEntity("invalid IP address")
	}
	if requireIPv4 && !addr.Is4() {
		return nil, huma.Error422UnprocessableEntity("publicV4 must be an IPv4 address")
	}
	if !requireIPv4 && !addr.Is6() {
		return nil, huma.Error422UnprocessableEntity("publicV6 must be an IPv6 address")
	}

	return &addr, nil
}

func parseAddrs(values []string) ([]netip.Addr, error) {
	addrs := make([]netip.Addr, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return nil, huma.Error422UnprocessableEntity("invalid probe address")
		}
		addr, err := netip.ParseAddr(trimmed)
		if err != nil {
			return nil, huma.Error422UnprocessableEntity("invalid probe address")
		}
		addrs = append(addrs, addr)
	}

	return addrs, nil
}

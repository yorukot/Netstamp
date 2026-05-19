package proberuntime

import (
	"net/netip"

	appproberuntime "github.com/yorukot/netstamp/internal/controller/application/proberuntime"
)

type runtimeStatusBody struct {
	AgentVersion *string      `json:"agentVersion,omitempty"`
	PublicV4     *netip.Addr  `json:"publicV4,omitempty"`
	PublicV6     *netip.Addr  `json:"publicV6,omitempty"`
	AS           *string      `json:"as,omitempty"`
	Addrs        []netip.Addr `json:"addrs,omitempty"`
}

func newRuntimeStatusInput(auth appproberuntime.RuntimeAuthInput, body runtimeStatusBody) appproberuntime.RuntimeStatusInput {
	return appproberuntime.RuntimeStatusInput{
		RuntimeAuthInput: auth,
		AgentVersion:     body.AgentVersion,
		PublicV4:         body.PublicV4,
		PublicV6:         body.PublicV6,
		AS:               body.AS,
		Addrs:            append([]netip.Addr(nil), body.Addrs...),
	}
}

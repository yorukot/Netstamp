package proberuntime

import (
	"net/netip"

	appproberuntime "github.com/yorukot/netstamp/internal/controller/application/proberuntime"
)

type runtimeStatusBody struct {
	AgentVersion *string      `json:"agentVersion,omitempty" doc:"Probe agent version." example:"netstamp-probe/0.1.0"`
	PublicV4     *netip.Addr  `json:"publicV4,omitempty" doc:"Observed public IPv4 address." example:"203.0.113.10"`
	PublicV6     *netip.Addr  `json:"publicV6,omitempty" doc:"Observed public IPv6 address." example:"2001:db8::10"`
	AS           *string      `json:"as,omitempty" doc:"Observed autonomous system." example:"AS15169 Google LLC"`
	Addrs        []netip.Addr `json:"addrs,omitempty" doc:"Probe local or observed interface addresses."`
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

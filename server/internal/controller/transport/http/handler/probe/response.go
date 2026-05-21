package probe

import (
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

// CreateProbeOutput is the response body for the create probe endpoint.
type createProbeOutput struct {
	Body createProbeOutputBody
}

type createProbeOutputBody struct {
	Probe  domainprobe.Probe `json:"probe"`
	Secret string            `json:"secret"`
}

// ListProbesOutput is the response body for the list probes endpoint.
type listProbesOutput struct {
	Body listProbesOutputBody
}

type listProbesOutputBody struct {
	Probes []domainprobe.Probe `json:"probes"`
}

// ProbeOutput is the response body for the probe endpoint.
type probeOutput struct {
	Body probeOutputBody
}

type probeOutputBody struct {
	Probe domainprobe.Probe `json:"probe"`
}

// RotateSecretOutput is the response body for the rotate secret endpoint.
type rotateSecretOutput struct {
	Body rotateSecretOutputBody
}

type rotateSecretOutputBody struct {
	Probe  domainprobe.Probe `json:"probe"`
	Secret string            `json:"secret"`
}

package probe

import domainprobe "github.com/yorukot/netstamp/internal/domain/probe"

type CreateProbeInput struct {
	CurrentUserID   string
	ProjectRef      string
	Name            string
	Enabled         *bool
	SubdivisionCode *string
	Latitude        *float64
	Longitude       *float64
	LabelIDs        []string
}

type CreateProbeOutput struct {
	Probe  domainprobe.Probe
	Secret string
}

type ListProbesInput struct {
	CurrentUserID string
	ProjectRef    string
}

type TargetProbeInput struct {
	CurrentUserID string
	ProjectRef    string
	ProbeID       string
}

type UpdateProbeInput struct {
	CurrentUserID   string
	ProjectRef      string
	ProbeID         string
	Name            *string
	Enabled         *bool
	SubdivisionCode *string
	Latitude        *float64
	Longitude       *float64
	LabelIDs        *[]string
}

type RotateProbeSecretOutput struct {
	Probe  domainprobe.Probe
	Secret string
}

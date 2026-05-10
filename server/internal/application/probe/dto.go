package probe

import domainprobe "github.com/yorukot/netstamp/internal/domain/probe"

type CreateProbeInput struct {
	CurrentUserID string
	ProjectRef    string
	Name          string
	Enabled       *bool
	City          *string
	Latitude      *float64
	Longitude     *float64
	LabelIDs      []string
}

type CreateProbeOutput struct {
	Probe  domainprobe.Probe
	Secret string //nolint:gosec // The plaintext probe secret is returned once to the creator.
}

type ListProbesInput struct {
	CurrentUserID string
	ProjectRef    string
}

type GetProbeInput struct {
	CurrentUserID string
	ProjectRef    string
	ProbeID       string
}

type UpdateProbeInput struct {
	CurrentUserID string
	ProjectRef    string
	ProbeID       string
	Name          *string
	Enabled       *bool
	City          *string
	Latitude      *float64
	Longitude     *float64
	LabelIDs      *[]string
}

type DeleteProbeInput struct {
	CurrentUserID string
	ProjectRef    string
	ProbeID       string
}

type RotateProbeSecretInput struct {
	CurrentUserID string
	ProjectRef    string
	ProbeID       string
}

type RotateProbeSecretOutput struct {
	Probe  domainprobe.Probe
	Secret string //nolint:gosec // The plaintext probe secret is returned once after rotation.
}

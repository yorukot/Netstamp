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
	Secret string
}

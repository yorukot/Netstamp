package probe

import (
	"time"

	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

type createProbeOutput struct {
	Body createProbeOutputBody
}

type createProbeOutputBody struct {
	Probe  probeResponse `json:"probe"`
	Secret string        `json:"secret"`
}

type probeResponse struct {
	ID        string          `json:"id" format:"uuid"`
	ProjectID string          `json:"projectId" format:"uuid"`
	Name      string          `json:"name"`
	Enabled   bool            `json:"enabled"`
	City      *string         `json:"city"`
	Latitude  *float64        `json:"latitude"`
	Longitude *float64        `json:"longitude"`
	Labels    []labelResponse `json:"labels"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

type labelResponse struct {
	ID    string `json:"id" format:"uuid"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

func newProbeResponse(probe domainprobe.Probe) probeResponse {
	labels := make([]labelResponse, 0, len(probe.Labels))
	for _, label := range probe.Labels {
		labels = append(labels, labelResponse{
			ID:    label.ID,
			Key:   label.Key,
			Value: label.Value,
		})
	}

	return probeResponse{
		ID:        probe.ID,
		ProjectID: probe.ProjectID,
		Name:      probe.Name,
		Enabled:   probe.Enabled,
		City:      probe.City,
		Latitude:  probe.Latitude,
		Longitude: probe.Longitude,
		Labels:    labels,
		CreatedAt: probe.CreatedAt,
		UpdatedAt: probe.UpdatedAt,
	}
}

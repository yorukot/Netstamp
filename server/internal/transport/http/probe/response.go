package probe

import (
	"time"

	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

type createProbeOutput struct {
	Body createProbeOutputBody
}

type listProbesOutput struct {
	Body listProbesOutputBody
}

type listProbesOutputBody struct {
	Probes []probeResponse `json:"probes"`
}

type probeOutput struct {
	Body probeOutputBody
}

type probeOutputBody struct {
	Probe probeResponse `json:"probe"`
}

type createProbeOutputBody struct {
	Probe  probeResponse `json:"probe"`
	Secret string        `json:"secret"` //nolint:gosec // The plaintext probe secret is returned once to the creator.
}

type rotateSecretOutput struct {
	Body rotateSecretOutputBody
}

type rotateSecretOutputBody struct {
	Probe  probeResponse `json:"probe"`
	Secret string        `json:"secret"` //nolint:gosec // The plaintext probe secret is returned once after rotation.
}

type probeResponse struct {
	ID        string               `json:"id" format:"uuid"`
	ProjectID string               `json:"projectId" format:"uuid"`
	Name      string               `json:"name"`
	Enabled   bool                 `json:"enabled"`
	City      *string              `json:"city"`
	Latitude  *float64             `json:"latitude"`
	Longitude *float64             `json:"longitude"`
	Labels    []probeLabelResponse `json:"labels"`
	Status    *probeStatusResponse `json:"status,omitempty"`
	CreatedAt time.Time            `json:"createdAt"`
	UpdatedAt time.Time            `json:"updatedAt"`
}

type probeStatusResponse struct {
	State        string     `json:"state"`
	LastSeenAt   *time.Time `json:"lastSeenAt"`
	AgentVersion *string    `json:"agentVersion"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type probeLabelResponse struct {
	ID    string `json:"id" format:"uuid"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

func newProbeResponses(probes []domainprobe.Probe) []probeResponse {
	responses := make([]probeResponse, 0, len(probes))
	for _, probe := range probes {
		responses = append(responses, newProbeResponse(probe))
	}

	return responses
}

func newProbeResponse(probe domainprobe.Probe) probeResponse {
	labels := make([]probeLabelResponse, 0, len(probe.Labels))
	for _, label := range probe.Labels {
		labels = append(labels, probeLabelResponse{
			ID:    label.ID,
			Key:   label.Key,
			Value: label.Value,
		})
	}

	var status *probeStatusResponse
	if probe.Status != nil {
		status = &probeStatusResponse{
			State:        string(probe.Status.State),
			LastSeenAt:   probe.Status.LastSeenAt,
			AgentVersion: probe.Status.AgentVersion,
			UpdatedAt:    probe.Status.UpdatedAt,
		}
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
		Status:    status,
		CreatedAt: probe.CreatedAt,
		UpdatedAt: probe.UpdatedAt,
	}
}

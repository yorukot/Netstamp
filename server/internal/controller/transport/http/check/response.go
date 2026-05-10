package check

import (
	"encoding/json"
	"time"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
)

type checkOutput struct {
	Body checkOutputBody
}

type checkOutputBody struct {
	Check checkResponse `json:"check"`
}

type checkResponse struct {
	ID              string               `json:"id" format:"uuid"`
	ProjectID       string               `json:"projectId" format:"uuid"`
	Name            string               `json:"name"`
	Type            string               `json:"type" enum:"ping"`
	Target          string               `json:"target"`
	Selector        map[string]any       `json:"selector"`
	Description     *string              `json:"description"`
	IntervalSeconds int32                `json:"intervalSeconds"`
	PacketCount     int32                `json:"packetCount"`
	PacketSizeBytes int32                `json:"packetSizeBytes"`
	TimeoutMs       int32                `json:"timeoutMs"`
	IPFamily        *string              `json:"ipFamily,omitempty" enum:"inet,inet6"`
	Labels          []checkLabelResponse `json:"labels"`
	CreatedAt       time.Time            `json:"createdAt"`
	UpdatedAt       time.Time            `json:"updatedAt"`
}

type checkLabelResponse struct {
	ID    string `json:"id" format:"uuid"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

func newCheckResponse(check domaincheck.Check) checkResponse {
	labels := make([]checkLabelResponse, 0, len(check.Labels))
	for _, label := range check.Labels {
		labels = append(labels, checkLabelResponse{
			ID:    label.ID,
			Key:   label.Key,
			Value: label.Value,
		})
	}

	return checkResponse{
		ID:              check.ID,
		ProjectID:       check.ProjectID,
		Name:            check.Name,
		Type:            string(check.Type),
		Target:          check.Target,
		Selector:        selectorObject(check.Selector),
		Description:     check.Description,
		IntervalSeconds: check.IntervalSeconds,
		PacketCount:     check.PingConfig.PacketCount,
		PacketSizeBytes: check.PingConfig.PacketSizeBytes,
		TimeoutMs:       check.PingConfig.TimeoutMs,
		IPFamily:        ipFamilyString(check.PingConfig.IPFamily),
		Labels:          labels,
		CreatedAt:       check.CreatedAt,
		UpdatedAt:       check.UpdatedAt,
	}
}

func selectorObject(raw json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}

	var selector map[string]any
	if err := json.Unmarshal(raw, &selector); err != nil || selector == nil {
		return map[string]any{}
	}

	return selector
}

func ipFamilyString(ipFamily *domainnetwork.IPFamily) *string {
	if ipFamily == nil {
		return nil
	}

	value := string(*ipFamily)
	return &value
}

package result

import "time"

type LatestResultType string

const (
	LatestResultTypePing       LatestResultType = "ping"
	LatestResultTypeTCP        LatestResultType = "tcp"
	LatestResultTypeTraceroute LatestResultType = "traceroute"
)

type LatestResult struct {
	Type            LatestResultType `json:"type"`
	ProbeID         string           `json:"probeId"`
	CheckID         string           `json:"checkId"`
	LatestStartedAt time.Time        `json:"latestStartedAt"`
	LatestStatus    string           `json:"latestStatus"`
}

type LatestResultQuery struct {
	ProjectID string
	ProbeID   string
	CheckID   string
	Type      *LatestResultType
}

type LatestResultList struct {
	Results []LatestResult
}

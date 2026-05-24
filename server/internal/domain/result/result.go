package result

import "time"

type MeasurementType string

const (
	MeasurementTypePing       MeasurementType = "ping"
	MeasurementTypeTCP        MeasurementType = "tcp"
	MeasurementTypeTraceroute MeasurementType = "traceroute"
)

type Measurement struct {
	Type         MeasurementType `json:"type"`
	StartedAt    time.Time       `json:"startedAt"`
	FinishedAt   time.Time       `json:"finishedAt"`
	ProbeID      string          `json:"probeId"`
	CheckID      string          `json:"checkId"`
	Status       string          `json:"status"`
	DurationMs   int32           `json:"durationMs"`
	LatencyMs    *float64        `json:"latencyMs,omitempty"`
	LossPercent  *float64        `json:"lossPercent,omitempty"`
	Metadata     *string         `json:"metadata,omitempty"`
	ErrorCode    *string         `json:"errorCode,omitempty"`
	ErrorMessage *string         `json:"errorMessage,omitempty"`
}

type MeasurementQuery struct {
	ProjectID string
	ProbeID   string
	CheckID   string
	Type      *MeasurementType
	Status    *string
	From      time.Time
	To        time.Time
	Limit     int32
	Cursor    *time.Time
}

type MeasurementResult struct {
	Measurements []Measurement
	NextCursor   *time.Time
}

package result

import "time"

type PingMetric string

const (
	PingMetricRTTAvgMS    PingMetric = "rttAvgMs"
	PingMetricLossPercent PingMetric = "lossPercent"
	PingMetricSuccessRate PingMetric = "successRate"
)

type QueryPingSeriesInput struct {
	CurrentUserID string
	ProjectRef    string
	ProbeID       string
	CheckID       string
	FromMs        *int64
	ToMs          *int64
	Metric        string
	MaxDataPoints *int32
	Now           time.Time
}

type PingSeriesOutput struct {
	Series []Series
	Query  QueryMetadata
}

type Series struct {
	Name   string
	Labels SeriesLabels
	Unit   string
	Points []SeriesPoint
}

type SeriesLabels struct {
	ProbeID   string
	CheckID   string
	CheckType string
}

type SeriesPoint struct {
	TimestampMs int64
	Value       float64
}

type QueryMetadata struct {
	FromMs        int64
	ToMs          int64
	MaxDataPoints int32
	Resolution    string
	TotalPoints   int64
}

type normalizedQueryPingSeriesInput struct {
	currentUserID string
	projectRef    string
	probeID       string
	checkID       string
	from          time.Time
	to            time.Time
	metric        PingMetric
	maxDataPoints int32
}

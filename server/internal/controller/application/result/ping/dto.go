package ping

import (
	"time"

	resultshared "github.com/yorukot/netstamp/internal/controller/application/result/shared"
)

type SeriesKey string

const (
	SeriesLatencyAvg  SeriesKey = "latency_avg"
	SeriesLatencyMin  SeriesKey = "latency_min"
	SeriesLatencyMax  SeriesKey = "latency_max"
	SeriesLossPercent SeriesKey = "loss_percent"
)

type QuerySeriesInput struct {
	CurrentUserID string
	ProjectRef    string
	ProbeID       string
	CheckID       string
	FromMs        *int64
	ToMs          *int64
	Series        string
	MaxDataPoints *int32
	Now           time.Time
}

type QueryInsightInput struct {
	CurrentUserID string
	ProjectRef    string
	ProbeID       string
	CheckID       string
	FromMs        *int64
	ToMs          *int64
	MaxDataPoints *int32
	Now           time.Time
}

type SeriesOutput struct {
	Series map[string]Series
	Meta   resultshared.QueryMetadata
}

type InsightOutput struct {
	Summary InsightSummary
	Meta    resultshared.QueryMetadata
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

type InsightSummary struct {
	AverageRttMs *float64
	MaxRttMs     *float64
	LossPercent  *float64
	SuccessRate  *float64
	Samples      int64
}

type normalizedSeriesInput struct {
	base resultshared.QueryBase

	series        []SeriesKey
	maxDataPoints int32
}

type normalizedInsightInput struct {
	base resultshared.QueryBase

	maxDataPoints int32
}

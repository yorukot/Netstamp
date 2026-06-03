package tcp

import (
	"time"

	resultshared "github.com/yorukot/netstamp/internal/controller/application/result/shared"
)

type SeriesKey string

const (
	SeriesConnectAvg     SeriesKey = "connect_avg"
	SeriesConnectMin     SeriesKey = "connect_min"
	SeriesConnectMax     SeriesKey = "connect_max"
	SeriesFailurePercent SeriesKey = "failure_percent"
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
	AverageConnectMs *float64
	MaxConnectMs     *float64
	FailurePercent   *float64
	SuccessRate      *float64
	Samples          int64
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

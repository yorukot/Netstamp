package httpcheck

import (
	"time"

	resultshared "github.com/yorukot/netstamp/internal/controller/application/result/shared"
	domainhttp "github.com/yorukot/netstamp/internal/domain/httpcheck"
)

type SeriesKey string

const (
	SeriesDNSAvg         SeriesKey = "dns_avg"
	SeriesConnectAvg     SeriesKey = "connect_avg"
	SeriesTLSAvg         SeriesKey = "tls_avg"
	SeriesTTFBAvg        SeriesKey = "ttfb_avg"
	SeriesTotalAvg       SeriesKey = "total_avg"
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

type QueryLatestInput struct {
	CurrentUserID string
	ProjectRef    string
	ProbeID       string
	CheckID       string
}

type SeriesOutput struct {
	Series map[string]Series
	Meta   resultshared.QueryMetadata
}

type InsightOutput struct {
	Summary InsightSummary
	Meta    resultshared.QueryMetadata
}

type LatestResultsOutput struct {
	Results []LatestResult
}

type LatestResult struct {
	ProbeID string
	CheckID string
	Result  domainhttp.Result
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
	AverageTotalMs           *float64
	MaxTotalMs               *float64
	AverageTTFBMs            *float64
	MaxTTFBMs                *float64
	FailurePercent           *float64
	SuccessRate              *float64
	CertificateDaysRemaining *float64
	Samples                  int64
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

type normalizedLatestInput struct {
	currentUserID string
	projectRef    string
	probeID       string
	checkID       string
}

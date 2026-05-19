package result

import (
	"net/netip"
	"time"
)

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

type QueryTracerouteRunsInput struct {
	CurrentUserID string
	ProjectRef    string
	ProbeID       string
	CheckID       string
	FromMs        *int64
	ToMs          *int64
	Limit         *int32
	CursorMs      *int64
	Now           time.Time
}

type QueryMeasurementsInput struct {
	CurrentUserID string
	ProjectRef    string
	ProbeID       string
	CheckID       string
	Type          string
	Status        string
	FromMs        *int64
	ToMs          *int64
	Limit         *int32
	CursorMs      *int64
	Now           time.Time
}

type PingSeriesOutput struct {
	Series []Series
	Query  QueryMetadata
}

type TracerouteRunsOutput struct {
	Runs  []TracerouteRun
	Query TracerouteRunsQueryMetadata
}

type MeasurementsOutput struct {
	Measurements []Measurement
	Query        MeasurementQueryMetadata
}

type Measurement struct {
	Type         string
	StartedAt    time.Time
	FinishedAt   time.Time
	ProbeID      string
	CheckID      string
	Status       string
	DurationMs   int32
	LatencyMs    *float64
	LossPercent  *float64
	Metadata     *string
	ErrorCode    *string
	ErrorMessage *string
}

type TracerouteRun struct {
	StartedAt          time.Time
	FinishedAt         time.Time
	DurationMs         int32
	Status             string
	ResolvedIP         *netip.Addr
	IPFamily           *string
	DestinationReached bool
	HopCount           int32
	ErrorCode          *string
	ErrorMessage       *string
	Hops               []TracerouteHop
}

type TracerouteHop struct {
	HopIndex      int32
	Address       *netip.Addr
	Hostname      *string
	SentCount     int32
	ReceivedCount int32
	LossPercent   float64
	RttMinMs      *float64
	RttAvgMs      *float64
	RttMedianMs   *float64
	RttMaxMs      *float64
	RttStddevMs   *float64
	RttSamplesMs  []float64
	ErrorCode     *string
	ErrorMessage  *string
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

type TracerouteRunsQueryMetadata struct {
	FromMs     int64
	ToMs       int64
	Limit      int32
	NextCursor *int64
}

type MeasurementQueryMetadata struct {
	FromMs     int64
	ToMs       int64
	Limit      int32
	NextCursor *int64
}

type normalizedQueryPingSeriesInput struct {
	normalizedQueryBase

	metric        PingMetric
	maxDataPoints int32
}

type normalizedQueryBase struct {
	currentUserID string
	projectRef    string
	probeID       string
	checkID       string
	from          time.Time
	to            time.Time
}

type normalizedQueryTracerouteRunsInput struct {
	normalizedQueryBase

	limit  int32
	cursor *time.Time
}

type normalizedQueryMeasurementsInput struct {
	currentUserID string
	projectRef    string
	probeID       string
	checkID       string
	resultType    *string
	status        *string
	from          time.Time
	to            time.Time
	limit         int32
	cursor        *time.Time
}

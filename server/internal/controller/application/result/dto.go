package result

import (
	"net/netip"
	"time"
)

type PingSeriesKey string

const (
	PingSeriesLatencyAvg  PingSeriesKey = "latency_avg"
	PingSeriesLatencyMin  PingSeriesKey = "latency_min"
	PingSeriesLatencyMax  PingSeriesKey = "latency_max"
	PingSeriesLossPercent PingSeriesKey = "loss_percent"
)

type QueryPingSeriesInput struct {
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

type QueryPingInsightInput struct {
	CurrentUserID string
	ProjectRef    string
	ProbeID       string
	CheckID       string
	FromMs        *int64
	ToMs          *int64
	MaxDataPoints *int32
	Now           time.Time
}

type QueryTCPInsightInput struct {
	CurrentUserID string
	ProjectRef    string
	ProbeID       string
	CheckID       string
	FromMs        *int64
	ToMs          *int64
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

type QueryTracerouteInsightInput struct {
	CurrentUserID string
	ProjectRef    string
	ProbeID       string
	CheckID       string
	FromMs        *int64
	ToMs          *int64
	MaxDataPoints *int32
	Now           time.Time
}

type QueryTracerouteTopologyInput struct {
	CurrentUserID string
	ProjectRef    string
	ProbeID       string
	CheckID       string
	FromMs        *int64
	ToMs          *int64
	Limit         *int32
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
	Series map[string]Series
	Meta   QueryMetadata
}

type PingInsightOutput struct {
	Summary PingInsightSummary
	Meta    QueryMetadata
}

type TCPInsightOutput struct {
	Buckets []TCPInsightBucket
	Summary TCPInsightSummary
	Query   QueryMetadata
}

type TracerouteRunsOutput struct {
	Runs  []TracerouteRun
	Query TracerouteRunsQueryMetadata
}

type TracerouteInsightOutput struct {
	Points []TracerouteInsightPoint
	Query  TracerouteInsightQueryMetadata
}

type TracerouteTopologyOutput struct {
	Nodes []TracerouteTopologyNode
	Edges []TracerouteTopologyEdge
	Query TracerouteTopologyQueryMetadata
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

type TracerouteInsightPoint struct {
	TimestampMs        int64
	BucketFromMs       int64
	BucketToMs         int64
	RunStartedAt       *time.Time
	ResultCount        int64
	FinalRttAvgMs      *float64
	FinalLossPercent   *float64
	HasLoss            bool
	HasRouteChange     bool
	DestinationReached bool
}

type TracerouteTopologyNode struct {
	ID          string
	Kind        string
	Label       string
	Address     *netip.Addr
	Hostname    *string
	ProbeID     *string
	CheckID     *string
	Target      *string
	HopIndex    *int32
	SeenCount   int32
	AvgRttMs    *float64
	LossPercent *float64
}

type TracerouteTopologyEdge struct {
	ID          string
	Source      string
	Target      string
	SeenCount   int32
	AvgRttMs    *float64
	LossPercent *float64
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

type PingInsightSummary struct {
	AverageRttMs *float64
	MaxRttMs     *float64
	LossPercent  *float64
	SuccessRate  *float64
	Samples      int64
}

type TCPInsightBucket struct {
	TimestampMs     int64
	ResultCount     int64
	DurationAvgMs   *float64
	ConnectMinMs    *float64
	ConnectAvgMs    *float64
	ConnectMedianMs *float64
	ConnectMaxMs    *float64
	ConnectStddevMs *float64
	SuccessRate     *float64
	TimeoutCount    int64
	ErrorCount      int64
}

type TCPInsightSummary struct {
	TotalResults      int64
	SuccessfulCount   int64
	TimeoutCount      int64
	ErrorCount        int64
	AvgConnectMs      *float64
	MedianConnectMs   *float64
	MaxConnectMs      *float64
	P95ConnectMs      *float64
	P99ConnectMs      *float64
	LatestStatus      *string
	LatestStartedAtMs *int64
	LatestConnectMs   *float64
	LatestResolvedIP  *netip.Addr
}

type QueryMetadata struct {
	FromMs        int64
	ToMs          int64
	MaxDataPoints int32
	Source        string
	Resolution    string
	TotalPoints   int64
}

type TracerouteRunsQueryMetadata struct {
	FromMs     int64
	ToMs       int64
	Limit      int32
	NextCursor *int64
}

type TracerouteInsightQueryMetadata struct {
	FromMs        int64
	ToMs          int64
	MaxDataPoints int32
	Resolution    string
	TotalRuns     int64
}

type TracerouteTopologyQueryMetadata struct {
	FromMs int64
	ToMs   int64
	Limit  int32
}

type MeasurementQueryMetadata struct {
	FromMs     int64
	ToMs       int64
	Limit      int32
	NextCursor *int64
}

type normalizedQueryPingSeriesInput struct {
	normalizedQueryBase

	series        []PingSeriesKey
	maxDataPoints int32
}

type normalizedQueryPingInsightInput struct {
	normalizedQueryBase

	maxDataPoints int32
}

type normalizedQueryTCPInsightInput struct {
	normalizedQueryBase

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

type normalizedQueryTracerouteInsightInput struct {
	normalizedQueryBase

	maxDataPoints int32
}

type normalizedQueryTracerouteTopologyInput struct {
	currentUserID string
	projectRef    string
	probeID       string
	checkID       string
	from          time.Time
	to            time.Time
	limit         int32
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

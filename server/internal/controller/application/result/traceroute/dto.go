package traceroute

import (
	"net/netip"
	"time"

	resultshared "github.com/yorukot/netstamp/internal/controller/application/result/shared"
)

type QueryRunsInput struct {
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

type QueryTopologyInput struct {
	CurrentUserID string
	ProjectRef    string
	ProbeID       string
	CheckID       string
	FromMs        *int64
	ToMs          *int64
	Limit         *int32
	Now           time.Time
}

type RunsOutput struct {
	Runs  []Run
	Query RunsQueryMetadata
}

type InsightOutput struct {
	Points []InsightPoint
	Query  InsightQueryMetadata
}

type TopologyOutput struct {
	Nodes []TopologyNode
	Edges []TopologyEdge
	Query TopologyQueryMetadata
}

type Run struct {
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
	Hops               []Hop
}

type Hop struct {
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

type InsightPoint struct {
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

type TopologyNode struct {
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

type TopologyEdge struct {
	ID          string
	Source      string
	Target      string
	SeenCount   int32
	AvgRttMs    *float64
	LossPercent *float64
}

type RunsQueryMetadata struct {
	FromMs     int64
	ToMs       int64
	Limit      int32
	NextCursor *int64
}

type InsightQueryMetadata struct {
	FromMs        int64
	ToMs          int64
	MaxDataPoints int32
	Resolution    string
	TotalRuns     int64
}

type TopologyQueryMetadata struct {
	FromMs int64
	ToMs   int64
	Limit  int32
}

type normalizedRunsInput struct {
	base resultshared.QueryBase

	limit  int32
	cursor *time.Time
}

type normalizedInsightInput struct {
	base resultshared.QueryBase

	maxDataPoints int32
}

type normalizedTopologyInput struct {
	currentUserID string
	projectRef    string
	probeID       string
	checkID       string
	from          time.Time
	to            time.Time
	limit         int32
}

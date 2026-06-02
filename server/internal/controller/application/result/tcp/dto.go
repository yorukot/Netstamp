package tcp

import (
	"net/netip"
	"time"

	resultshared "github.com/yorukot/netstamp/internal/controller/application/result/shared"
)

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

type InsightOutput struct {
	Buckets []InsightBucket
	Summary InsightSummary
	Query   resultshared.QueryMetadata
}

type InsightBucket struct {
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

type InsightSummary struct {
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

type normalizedInsightInput struct {
	base resultshared.QueryBase

	maxDataPoints int32
}

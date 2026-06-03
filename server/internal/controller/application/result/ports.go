package result

import (
	"context"

	"github.com/yorukot/netstamp/internal/controller/application/result/measurement"
	"github.com/yorukot/netstamp/internal/controller/application/result/ping"
	resultshared "github.com/yorukot/netstamp/internal/controller/application/result/shared"
	"github.com/yorukot/netstamp/internal/controller/application/result/tcp"
	"github.com/yorukot/netstamp/internal/controller/application/result/traceroute"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainresult "github.com/yorukot/netstamp/internal/domain/result"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

type PingSeriesRepository interface {
	CountPingSeriesPoints(ctx context.Context, input domainping.SeriesPointCountQuery) (int64, error)
	ListPingSeries(ctx context.Context, input domainping.SeriesReadQuery) (map[string]domainping.SeriesData, error)
	GetPingInsightSummary(ctx context.Context, input domainping.InsightSummaryQuery) (domainping.InsightSummary, error)
}

type TCPInsightRepository interface {
	CountTCPSeriesPoints(ctx context.Context, input domaintcp.SeriesPointCountQuery) (int64, error)
	ListTCPSeries(ctx context.Context, input domaintcp.SeriesReadQuery) (map[string]domaintcp.SeriesData, error)
	GetTCPInsightSummary(ctx context.Context, input domaintcp.InsightSummaryQuery) (domaintcp.InsightSummary, error)
}

type TracerouteRunsRepository interface {
	ListTracerouteRuns(ctx context.Context, input domaintraceroute.RunQuery) (domaintraceroute.RunResult, error)
	ListTracerouteInsight(ctx context.Context, input domaintraceroute.InsightQuery) (domaintraceroute.InsightResult, error)
	ListTracerouteTopologyRuns(ctx context.Context, input domaintraceroute.TopologyQuery) (domaintraceroute.TopologyRunResult, error)
}

type MeasurementRepository interface {
	ListMeasurements(ctx context.Context, input domainresult.MeasurementQuery) (domainresult.MeasurementResult, error)
}

type ProjectAccess interface {
	resultshared.ProjectAccess
}

var _ ping.SeriesRepository = (PingSeriesRepository)(nil)
var _ tcp.InsightRepository = (TCPInsightRepository)(nil)
var _ traceroute.RunsRepository = (TracerouteRunsRepository)(nil)
var _ measurement.Repository = (MeasurementRepository)(nil)

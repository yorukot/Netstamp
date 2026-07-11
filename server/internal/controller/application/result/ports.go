package result

import (
	"context"

	apphttp "github.com/yorukot/netstamp/internal/controller/application/result/httpcheck"
	"github.com/yorukot/netstamp/internal/controller/application/result/latest"
	"github.com/yorukot/netstamp/internal/controller/application/result/ping"
	resultshared "github.com/yorukot/netstamp/internal/controller/application/result/shared"
	"github.com/yorukot/netstamp/internal/controller/application/result/tcp"
	"github.com/yorukot/netstamp/internal/controller/application/result/traceroute"
	domainhttp "github.com/yorukot/netstamp/internal/domain/httpcheck"
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

type HTTPInsightRepository interface {
	CountHTTPSeriesPoints(ctx context.Context, input domainhttp.SeriesPointCountQuery) (int64, error)
	ListHTTPSeries(ctx context.Context, input domainhttp.SeriesReadQuery) (map[string]domainhttp.SeriesData, error)
	GetHTTPInsightSummary(ctx context.Context, input domainhttp.InsightSummaryQuery) (domainhttp.InsightSummary, error)
}

type TracerouteRunsRepository interface {
	ListTracerouteRuns(ctx context.Context, input domaintraceroute.RunQuery) (domaintraceroute.RunResult, error)
	ListTracerouteInsight(ctx context.Context, input domaintraceroute.InsightQuery) (domaintraceroute.InsightResult, error)
	ListTracerouteTopologyRuns(ctx context.Context, input domaintraceroute.TopologyQuery) (domaintraceroute.TopologyRunResult, error)
}

type LatestRepository interface {
	ListLatestResults(ctx context.Context, input domainresult.LatestResultQuery) (domainresult.LatestResultList, error)
}

type ProjectAccess interface {
	resultshared.ProjectAccess
}

var (
	_ ping.SeriesRepository     = PingSeriesRepository(nil)
	_ tcp.InsightRepository     = TCPInsightRepository(nil)
	_ apphttp.Repository        = HTTPInsightRepository(nil)
	_ traceroute.RunsRepository = TracerouteRunsRepository(nil)
	_ latest.Repository         = LatestRepository(nil)
)

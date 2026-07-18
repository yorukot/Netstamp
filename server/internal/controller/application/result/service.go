package result

import (
	"context"

	apphttp "github.com/yorukot/netstamp/internal/controller/application/result/httpcheck"
	"github.com/yorukot/netstamp/internal/controller/application/result/latest"
	"github.com/yorukot/netstamp/internal/controller/application/result/ping"
	"github.com/yorukot/netstamp/internal/controller/application/result/tcp"
	"github.com/yorukot/netstamp/internal/controller/application/result/traceroute"
)

type Service struct {
	pings         *ping.Service
	tcps          *tcp.Service
	traceroutes   *traceroute.Service
	latestResults *latest.Service
	httpResults   *apphttp.Service
}

func NewService(pings PingSeriesRepository, tcps TCPInsightRepository, traceroutes TracerouteRunsRepository, latestResults LatestRepository, projectAccess ProjectAccess) *Service {
	return NewServiceWithHTTP(pings, tcps, nil, traceroutes, latestResults, projectAccess)
}

func NewServiceWithHTTP(pings PingSeriesRepository, tcps TCPInsightRepository, httpResults HTTPInsightRepository, traceroutes TracerouteRunsRepository, latestResults LatestRepository, projectAccess ProjectAccess) *Service {
	return &Service{
		pings:         ping.NewService(pings, projectAccess),
		tcps:          tcp.NewService(tcps, projectAccess),
		traceroutes:   traceroute.NewService(traceroutes, projectAccess),
		latestResults: latest.NewService(latestResults, projectAccess),
		httpResults:   apphttp.NewService(httpResults, projectAccess),
	}
}

func (s *Service) QueryHTTPSeries(ctx context.Context, input QueryHTTPSeriesInput) (HTTPSeriesOutput, error) {
	return s.httpResults.QuerySeries(ctx, input)
}

func (s *Service) QueryHTTPInsight(ctx context.Context, input QueryHTTPInsightInput) (HTTPInsightOutput, error) {
	return s.httpResults.QueryInsight(ctx, input)
}

func (s *Service) QueryLatestHTTPResults(ctx context.Context, input QueryLatestHTTPResultsInput) (LatestHTTPResultsOutput, error) {
	return s.httpResults.QueryLatest(ctx, input)
}

func (s *Service) QueryPingSeries(ctx context.Context, input QueryPingSeriesInput) (PingSeriesOutput, error) {
	return s.pings.QuerySeries(ctx, input)
}

func (s *Service) QueryPingInsight(ctx context.Context, input QueryPingInsightInput) (PingInsightOutput, error) {
	return s.pings.QueryInsight(ctx, input)
}

func (s *Service) QueryTCPSeries(ctx context.Context, input QueryTCPSeriesInput) (TCPSeriesOutput, error) {
	return s.tcps.QuerySeries(ctx, input)
}

func (s *Service) QueryTCPInsight(ctx context.Context, input QueryTCPInsightInput) (TCPInsightOutput, error) {
	return s.tcps.QueryInsight(ctx, input)
}

func (s *Service) QueryTracerouteRuns(ctx context.Context, input QueryTracerouteRunsInput) (TracerouteRunsOutput, error) {
	return s.traceroutes.QueryRuns(ctx, input)
}

func (s *Service) QueryTracerouteInsight(ctx context.Context, input QueryTracerouteInsightInput) (TracerouteInsightOutput, error) {
	return s.traceroutes.QueryInsight(ctx, input)
}

func (s *Service) QueryTracerouteTopology(ctx context.Context, input QueryTracerouteTopologyInput) (TracerouteTopologyOutput, error) {
	return s.traceroutes.QueryTopology(ctx, input)
}

func (s *Service) QueryLatestResults(ctx context.Context, input QueryLatestResultsInput) (LatestResultsOutput, error) {
	return s.latestResults.Query(ctx, input)
}

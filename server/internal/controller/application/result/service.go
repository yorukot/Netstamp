package result

import (
	"context"

	"github.com/yorukot/netstamp/internal/controller/application/result/measurement"
	"github.com/yorukot/netstamp/internal/controller/application/result/ping"
	"github.com/yorukot/netstamp/internal/controller/application/result/tcp"
	"github.com/yorukot/netstamp/internal/controller/application/result/traceroute"
)

type Service struct {
	pings        *ping.Service
	tcps         *tcp.Service
	traceroutes  *traceroute.Service
	measurements *measurement.Service
}

func NewService(pings PingSeriesRepository, tcps TCPInsightRepository, traceroutes TracerouteRunsRepository, measurements MeasurementRepository, projectAccess ProjectAccess) *Service {
	return &Service{
		pings:        ping.NewService(pings, projectAccess),
		tcps:         tcp.NewService(tcps, projectAccess),
		traceroutes:  traceroute.NewService(traceroutes, projectAccess),
		measurements: measurement.NewService(measurements, projectAccess),
	}
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

func (s *Service) QueryMeasurements(ctx context.Context, input QueryMeasurementsInput) (MeasurementsOutput, error) {
	return s.measurements.Query(ctx, input)
}

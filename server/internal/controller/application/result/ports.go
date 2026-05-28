package result

import (
	"context"

	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainresult "github.com/yorukot/netstamp/internal/domain/result"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

type PingSeriesRepository interface {
	ListPingSeries(ctx context.Context, input domainping.SeriesQuery) (domainping.SeriesResult, error)
	ListPingInsight(ctx context.Context, input domainping.InsightQuery) (domainping.InsightResult, error)
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
	GetProjectForUser(ctx context.Context, projectRef, userID string) (domainproject.Project, error)
}

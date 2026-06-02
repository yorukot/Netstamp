package traceroute

import (
	"context"

	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

type RunsRepository interface {
	ListTracerouteRuns(ctx context.Context, input domaintraceroute.RunQuery) (domaintraceroute.RunResult, error)
	ListTracerouteInsight(ctx context.Context, input domaintraceroute.InsightQuery) (domaintraceroute.InsightResult, error)
	ListTracerouteTopologyRuns(ctx context.Context, input domaintraceroute.TopologyQuery) (domaintraceroute.TopologyRunResult, error)
}

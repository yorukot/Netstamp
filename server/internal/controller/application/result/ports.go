package result

import (
	"context"

	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

type PingSeriesRepository interface {
	ListPingSeries(ctx context.Context, input domainping.SeriesQuery) (domainping.SeriesResult, error)
}

type TracerouteRunsRepository interface {
	ListTracerouteRuns(ctx context.Context, input domaintraceroute.RunQuery) (domaintraceroute.RunResult, error)
}

type ProjectAccess interface {
	GetProjectForUser(ctx context.Context, projectRef, userID string) (domainproject.Project, error)
}

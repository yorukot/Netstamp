package result

import (
	"context"

	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type PingSeriesRepository interface {
	ListPingSeries(ctx context.Context, input domainping.SeriesQuery) (domainping.SeriesResult, error)
}

type ProjectAccess interface {
	GetProjectForUser(ctx context.Context, projectRef, userID string) (domainproject.Project, error)
}

package ping

import (
	"context"

	domainping "github.com/yorukot/netstamp/internal/domain/ping"
)

type SeriesRepository interface {
	CountPingSeriesPoints(ctx context.Context, input domainping.SeriesPointCountQuery) (int64, error)
	ListPingSeries(ctx context.Context, input domainping.SeriesReadQuery) (map[string]domainping.SeriesData, error)
	GetPingInsightSummary(ctx context.Context, input domainping.InsightSummaryQuery) (domainping.InsightSummary, error)
}

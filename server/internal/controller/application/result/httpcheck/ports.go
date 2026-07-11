package httpcheck

import (
	"context"

	domainhttp "github.com/yorukot/netstamp/internal/domain/httpcheck"
)

type Repository interface {
	CountHTTPSeriesPoints(ctx context.Context, input domainhttp.SeriesPointCountQuery) (int64, error)
	ListHTTPSeries(ctx context.Context, input domainhttp.SeriesReadQuery) (map[string]domainhttp.SeriesData, error)
	GetHTTPInsightSummary(ctx context.Context, input domainhttp.InsightSummaryQuery) (domainhttp.InsightSummary, error)
}

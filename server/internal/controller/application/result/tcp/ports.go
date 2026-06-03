package tcp

import (
	"context"

	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
)

type InsightRepository interface {
	CountTCPSeriesPoints(ctx context.Context, input domaintcp.SeriesPointCountQuery) (int64, error)
	ListTCPSeries(ctx context.Context, input domaintcp.SeriesReadQuery) (map[string]domaintcp.SeriesData, error)
	GetTCPInsightSummary(ctx context.Context, input domaintcp.InsightSummaryQuery) (domaintcp.InsightSummary, error)
}

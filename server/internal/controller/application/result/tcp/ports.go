package tcp

import (
	"context"

	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
)

type InsightRepository interface {
	ListTCPInsight(ctx context.Context, input domaintcp.InsightQuery) (domaintcp.InsightResult, error)
}

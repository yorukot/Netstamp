package latest

import (
	"context"

	domainresult "github.com/yorukot/netstamp/internal/domain/result"
)

type Repository interface {
	ListLatestResults(ctx context.Context, input domainresult.LatestResultQuery) (domainresult.LatestResultList, error)
}

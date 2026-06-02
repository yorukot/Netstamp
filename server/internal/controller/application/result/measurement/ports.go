package measurement

import (
	"context"

	domainresult "github.com/yorukot/netstamp/internal/domain/result"
)

type Repository interface {
	ListMeasurements(ctx context.Context, input domainresult.MeasurementQuery) (domainresult.MeasurementResult, error)
}

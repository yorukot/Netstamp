package workerloop

import (
	"context"
	"time"
)

type FailureLogger func(error)

func Run(ctx context.Context, enabled bool, interval time.Duration, runOnce func(context.Context) error, logFailure FailureLogger) error {
	if !enabled {
		<-ctx.Done()
		return nil
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if err := runOnce(ctx); err != nil && logFailure != nil {
			logFailure(err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

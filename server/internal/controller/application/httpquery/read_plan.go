package httpquery

import (
	"time"

	domainhttp "github.com/yorukot/netstamp/internal/domain/httpcheck"
)

const rawRetentionWindow = 3 * 24 * time.Hour

func SelectReadPlan(rawPoints int64, from, now time.Time, maxDataPoints int32) domainhttp.SeriesReadPlan {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if from.UTC().Before(now.UTC().Add(-rawRetentionWindow)) {
		return domainhttp.SeriesReadPlan{Mode: domainhttp.SeriesReadModeRollup, Source: domainhttp.SeriesSourceAggregate, Resolution: domainhttp.SeriesResolutionOneMinute, TotalPoints: rawPoints}
	}
	if rawPoints <= int64(maxDataPoints) {
		return domainhttp.SeriesReadPlan{Mode: domainhttp.SeriesReadModeRaw, Source: domainhttp.SeriesSourceRaw, Resolution: domainhttp.SeriesResolutionRaw, TotalPoints: rawPoints}
	}
	return domainhttp.SeriesReadPlan{Mode: domainhttp.SeriesReadModeBucket, Source: domainhttp.SeriesSourceRaw, Resolution: domainhttp.SeriesResolutionBucket, TotalPoints: rawPoints}
}

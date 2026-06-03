package pingquery

import (
	"time"

	domainping "github.com/yorukot/netstamp/internal/domain/ping"
)

const RawRetentionWindow = 3 * 24 * time.Hour

func SelectReadPlan(rawPoints int64, from, now time.Time, maxDataPoints int32) domainping.SeriesReadPlan {
	if useRollup(from, now) {
		return domainping.SeriesReadPlan{
			Mode:        domainping.SeriesReadModeRollup,
			Source:      domainping.SeriesSourceAggregate,
			Resolution:  domainping.SeriesResolutionOneMinute,
			TotalPoints: rawPoints,
		}
	}
	if rawPoints <= int64(maxDataPoints) {
		return domainping.SeriesReadPlan{
			Mode:        domainping.SeriesReadModeRaw,
			Source:      domainping.SeriesSourceRaw,
			Resolution:  domainping.SeriesResolutionRaw,
			TotalPoints: rawPoints,
		}
	}

	return domainping.SeriesReadPlan{
		Mode:        domainping.SeriesReadModeBucket,
		Source:      domainping.SeriesSourceRaw,
		Resolution:  domainping.SeriesResolutionBucket,
		TotalPoints: rawPoints,
	}
}

func useRollup(from, now time.Time) bool {
	now = now.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return from.UTC().Before(now.Add(-RawRetentionWindow))
}

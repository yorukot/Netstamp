package tcpquery

import (
	"time"

	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
)

const RawRetentionWindow = 3 * 24 * time.Hour

func SelectReadPlan(rawPoints int64, from, now time.Time, maxDataPoints int32) domaintcp.SeriesReadPlan {
	if useRollup(from, now) {
		return domaintcp.SeriesReadPlan{
			Mode:        domaintcp.SeriesReadModeRollup,
			Source:      domaintcp.SeriesSourceAggregate,
			Resolution:  domaintcp.SeriesResolutionOneMinute,
			TotalPoints: rawPoints,
		}
	}
	if rawPoints <= int64(maxDataPoints) {
		return domaintcp.SeriesReadPlan{
			Mode:        domaintcp.SeriesReadModeRaw,
			Source:      domaintcp.SeriesSourceRaw,
			Resolution:  domaintcp.SeriesResolutionRaw,
			TotalPoints: rawPoints,
		}
	}

	return domaintcp.SeriesReadPlan{
		Mode:        domaintcp.SeriesReadModeBucket,
		Source:      domaintcp.SeriesSourceRaw,
		Resolution:  domaintcp.SeriesResolutionBucket,
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

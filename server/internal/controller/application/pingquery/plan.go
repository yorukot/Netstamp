package pingquery

import domainping "github.com/yorukot/netstamp/internal/domain/ping"

func SelectReadPlan(counts domainping.SeriesPointCounts, maxDataPoints int32) domainping.SeriesReadPlan {
	if useRollup(counts.Raw, counts.Rollup) {
		return domainping.SeriesReadPlan{
			Mode:        domainping.SeriesReadModeRollup,
			Source:      domainping.SeriesSourceAggregate,
			Resolution:  domainping.SeriesResolutionOneMinute,
			TotalPoints: counts.Rollup,
		}
	}
	if counts.Raw <= int64(maxDataPoints) {
		return domainping.SeriesReadPlan{
			Mode:        domainping.SeriesReadModeRaw,
			Source:      domainping.SeriesSourceRaw,
			Resolution:  domainping.SeriesResolutionRaw,
			TotalPoints: counts.Raw,
		}
	}

	return domainping.SeriesReadPlan{
		Mode:        domainping.SeriesReadModeBucket,
		Source:      domainping.SeriesSourceRaw,
		Resolution:  domainping.SeriesResolutionBucket,
		TotalPoints: counts.Raw,
	}
}

func useRollup(rawPoints, rollupPoints int64) bool {
	return rollupPoints > rawPoints
}

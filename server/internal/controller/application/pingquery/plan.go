package pingquery

import domainping "github.com/yorukot/netstamp/internal/domain/ping"

func SelectReadPlan(rawPoints, rollupPoints int64, maxDataPoints int32) domainping.SeriesReadPlan {
	if useRollup(rawPoints, rollupPoints) {
		return domainping.SeriesReadPlan{
			Mode:        domainping.SeriesReadModeRollup,
			Source:      domainping.SeriesSourceAggregate,
			Resolution:  domainping.SeriesResolutionOneMinute,
			TotalPoints: rollupPoints,
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

func useRollup(rawPoints, rollupPoints int64) bool {
	return rollupPoints > rawPoints
}

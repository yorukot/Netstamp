package pgping

import (
	"time"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
)

func sqlcPingStatus(value domainping.Status) sqlc.PingStatus {
	return sqlc.PingStatus(value)
}

func sqlcIPFamily(value *domainnetwork.IPFamily) *sqlc.IpFamily {
	if value == nil {
		return nil
	}

	ipFamily := sqlc.IpFamily(*value)
	return &ipFamily
}

func latencyAvgRawSeriesRows(rows []sqlc.ListPingLatencyAvgRawSeriesRow) []domainping.SeriesPoint {
	return pingSeriesRows(rows, func(row sqlc.ListPingLatencyAvgRawSeriesRow) (int64, float64) {
		return row.BucketMs, row.Value
	})
}

func latencyMinRawSeriesRows(rows []sqlc.ListPingLatencyMinRawSeriesRow) []domainping.SeriesPoint {
	return pingSeriesRows(rows, func(row sqlc.ListPingLatencyMinRawSeriesRow) (int64, float64) {
		return row.BucketMs, row.Value
	})
}

func latencyMaxRawSeriesRows(rows []sqlc.ListPingLatencyMaxRawSeriesRow) []domainping.SeriesPoint {
	return pingSeriesRows(rows, func(row sqlc.ListPingLatencyMaxRawSeriesRow) (int64, float64) {
		return row.BucketMs, row.Value
	})
}

func lossPercentRawSeriesRows(rows []sqlc.ListPingLossPercentRawSeriesRow) []domainping.SeriesPoint {
	return pingSeriesRows(rows, func(row sqlc.ListPingLossPercentRawSeriesRow) (int64, float64) {
		return row.BucketMs, row.Value
	})
}

func latencyAvgBucketSeriesRows(rows []sqlc.ListPingLatencyAvgBucketSeriesRow) []domainping.SeriesPoint {
	return pingSeriesRows(rows, func(row sqlc.ListPingLatencyAvgBucketSeriesRow) (int64, float64) {
		return row.BucketMs, row.Value
	})
}

func latencyMinBucketSeriesRows(rows []sqlc.ListPingLatencyMinBucketSeriesRow) []domainping.SeriesPoint {
	return pingSeriesRows(rows, func(row sqlc.ListPingLatencyMinBucketSeriesRow) (int64, float64) {
		return row.BucketMs, row.Value
	})
}

func latencyMaxBucketSeriesRows(rows []sqlc.ListPingLatencyMaxBucketSeriesRow) []domainping.SeriesPoint {
	return pingSeriesRows(rows, func(row sqlc.ListPingLatencyMaxBucketSeriesRow) (int64, float64) {
		return row.BucketMs, row.Value
	})
}

func lossPercentBucketSeriesRows(rows []sqlc.ListPingLossPercentBucketSeriesRow) []domainping.SeriesPoint {
	return pingSeriesRows(rows, func(row sqlc.ListPingLossPercentBucketSeriesRow) (int64, float64) {
		return row.BucketMs, row.Value
	})
}

func latencyAvgRollupSeriesRows(rows []sqlc.ListPingLatencyAvgRollupSeriesRow) []domainping.SeriesPoint {
	return pingSeriesRows(rows, func(row sqlc.ListPingLatencyAvgRollupSeriesRow) (int64, float64) {
		return row.BucketMs, row.Value
	})
}

func latencyMinRollupSeriesRows(rows []sqlc.ListPingLatencyMinRollupSeriesRow) []domainping.SeriesPoint {
	return pingSeriesRows(rows, func(row sqlc.ListPingLatencyMinRollupSeriesRow) (int64, float64) {
		return row.BucketMs, row.Value
	})
}

func latencyMaxRollupSeriesRows(rows []sqlc.ListPingLatencyMaxRollupSeriesRow) []domainping.SeriesPoint {
	return pingSeriesRows(rows, func(row sqlc.ListPingLatencyMaxRollupSeriesRow) (int64, float64) {
		return row.BucketMs, row.Value
	})
}

func lossPercentRollupSeriesRows(rows []sqlc.ListPingLossPercentRollupSeriesRow) []domainping.SeriesPoint {
	return pingSeriesRows(rows, func(row sqlc.ListPingLossPercentRollupSeriesRow) (int64, float64) {
		return row.BucketMs, row.Value
	})
}

func pingSeriesRows[T any](rows []T, values func(T) (int64, float64)) []domainping.SeriesPoint {
	points := make([]domainping.SeriesPoint, 0, len(rows))
	for _, row := range rows {
		timestampMs, value := values(row)
		points = append(points, domainping.SeriesPoint{
			Timestamp: time.UnixMilli(timestampMs).UTC(),
			Value:     value,
		})
	}
	return points
}

func pingInsightSummary(row sqlc.GetPingInsightSummaryRow) domainping.InsightSummary {
	return domainping.InsightSummary{
		AverageRttMs: floatPtrIf(row.RttValueCount, row.AverageRttMs),
		MaxRttMs:     floatPtrIf(row.RttValueCount, row.MaxRttMs),
		LossPercent:  floatPtrIf(row.TotalResults, row.LossPercent),
		SuccessRate:  floatPtrIf(row.TotalResults, row.SuccessRate),
		Samples:      row.Samples,
	}
}

func pingRollupInsightSummary(row sqlc.GetPingInsightRollupSummaryRow) domainping.InsightSummary {
	return domainping.InsightSummary{
		AverageRttMs: floatPtrIf(row.RttValueCount, row.AverageRttMs),
		MaxRttMs:     floatPtrIf(row.RttValueCount, row.MaxRttMs),
		LossPercent:  floatPtrIf(row.TotalResults, row.LossPercent),
		SuccessRate:  floatPtrIf(row.TotalResults, row.SuccessRate),
		Samples:      row.Samples,
	}
}

func floatPtrIf(count int64, value float64) *float64 {
	if count == 0 {
		return nil
	}

	copied := value
	return &copied
}

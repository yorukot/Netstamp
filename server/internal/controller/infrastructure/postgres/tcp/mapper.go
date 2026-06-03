package pgtcp

import (
	"time"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
)

func sqlcTCPStatus(value domaintcp.Status) sqlc.TcpStatus {
	return sqlc.TcpStatus(value)
}

func sqlcIPFamily(value *domainnetwork.IPFamily) *sqlc.IpFamily {
	if value == nil {
		return nil
	}

	ipFamily := sqlc.IpFamily(*value)
	return &ipFamily
}

func connectAvgRawSeriesRows(rows []sqlc.ListTCPConnectAvgRawSeriesRow) []domaintcp.SeriesPoint {
	return tcpSeriesRows(rows, func(row sqlc.ListTCPConnectAvgRawSeriesRow) (int64, float64) {
		return row.BucketMs, row.Value
	})
}

func connectMinRawSeriesRows(rows []sqlc.ListTCPConnectMinRawSeriesRow) []domaintcp.SeriesPoint {
	return tcpSeriesRows(rows, func(row sqlc.ListTCPConnectMinRawSeriesRow) (int64, float64) {
		return row.BucketMs, row.Value
	})
}

func connectMaxRawSeriesRows(rows []sqlc.ListTCPConnectMaxRawSeriesRow) []domaintcp.SeriesPoint {
	return tcpSeriesRows(rows, func(row sqlc.ListTCPConnectMaxRawSeriesRow) (int64, float64) {
		return row.BucketMs, row.Value
	})
}

func failurePercentRawSeriesRows(rows []sqlc.ListTCPFailurePercentRawSeriesRow) []domaintcp.SeriesPoint {
	return tcpSeriesRows(rows, func(row sqlc.ListTCPFailurePercentRawSeriesRow) (int64, float64) {
		return row.BucketMs, row.Value
	})
}

func connectAvgBucketSeriesRows(rows []sqlc.ListTCPConnectAvgBucketSeriesRow) []domaintcp.SeriesPoint {
	return tcpSeriesRows(rows, func(row sqlc.ListTCPConnectAvgBucketSeriesRow) (int64, float64) {
		return row.BucketMs, row.Value
	})
}

func connectMinBucketSeriesRows(rows []sqlc.ListTCPConnectMinBucketSeriesRow) []domaintcp.SeriesPoint {
	return tcpSeriesRows(rows, func(row sqlc.ListTCPConnectMinBucketSeriesRow) (int64, float64) {
		return row.BucketMs, row.Value
	})
}

func connectMaxBucketSeriesRows(rows []sqlc.ListTCPConnectMaxBucketSeriesRow) []domaintcp.SeriesPoint {
	return tcpSeriesRows(rows, func(row sqlc.ListTCPConnectMaxBucketSeriesRow) (int64, float64) {
		return row.BucketMs, row.Value
	})
}

func failurePercentBucketSeriesRows(rows []sqlc.ListTCPFailurePercentBucketSeriesRow) []domaintcp.SeriesPoint {
	return tcpSeriesRows(rows, func(row sqlc.ListTCPFailurePercentBucketSeriesRow) (int64, float64) {
		return row.BucketMs, row.Value
	})
}

func connectAvgRollupSeriesRows(rows []sqlc.ListTCPConnectAvgRollupSeriesRow) []domaintcp.SeriesPoint {
	return tcpSeriesRows(rows, func(row sqlc.ListTCPConnectAvgRollupSeriesRow) (int64, float64) {
		return row.BucketMs, row.Value
	})
}

func connectMinRollupSeriesRows(rows []sqlc.ListTCPConnectMinRollupSeriesRow) []domaintcp.SeriesPoint {
	return tcpSeriesRows(rows, func(row sqlc.ListTCPConnectMinRollupSeriesRow) (int64, float64) {
		return row.BucketMs, row.Value
	})
}

func connectMaxRollupSeriesRows(rows []sqlc.ListTCPConnectMaxRollupSeriesRow) []domaintcp.SeriesPoint {
	return tcpSeriesRows(rows, func(row sqlc.ListTCPConnectMaxRollupSeriesRow) (int64, float64) {
		return row.BucketMs, row.Value
	})
}

func failurePercentRollupSeriesRows(rows []sqlc.ListTCPFailurePercentRollupSeriesRow) []domaintcp.SeriesPoint {
	return tcpSeriesRows(rows, func(row sqlc.ListTCPFailurePercentRollupSeriesRow) (int64, float64) {
		return row.BucketMs, row.Value
	})
}

func tcpSeriesRows[T any](rows []T, values func(T) (int64, float64)) []domaintcp.SeriesPoint {
	points := make([]domaintcp.SeriesPoint, 0, len(rows))
	for _, row := range rows {
		timestampMs, value := values(row)
		points = append(points, domaintcp.SeriesPoint{
			Timestamp: time.UnixMilli(timestampMs).UTC(),
			Value:     value,
		})
	}
	return points
}

func tcpInsightSummary(row sqlc.GetTCPInsightSummaryRow) domaintcp.InsightSummary {
	return domaintcp.InsightSummary{
		TotalResults:     row.TotalResults,
		AverageConnectMs: floatPtrIf(row.ConnectValueCount, row.AverageConnectMs),
		MaxConnectMs:     floatPtrIf(row.ConnectValueCount, row.MaxConnectMs),
		FailurePercent:   floatPtrIf(row.TotalResults, row.FailurePercent),
		SuccessRate:      floatPtrIf(row.TotalResults, row.SuccessRate),
		Samples:          row.Samples,
	}
}

func tcpInsightRollupSummary(row sqlc.GetTCPInsightRollupSummaryRow) domaintcp.InsightSummary {
	return domaintcp.InsightSummary{
		TotalResults:     row.TotalResults,
		AverageConnectMs: floatPtrIf(row.ConnectValueCount, row.AverageConnectMs),
		MaxConnectMs:     floatPtrIf(row.ConnectValueCount, row.MaxConnectMs),
		FailurePercent:   floatPtrIf(row.TotalResults, row.FailurePercent),
		SuccessRate:      floatPtrIf(row.TotalResults, row.SuccessRate),
		Samples:          row.Samples,
	}
}

func floatPtrIf(count int64, value float64) *float64 {
	if count == 0 {
		return nil
	}

	copied := value
	return &copied
}

import type { SummaryMetric } from "@/features/insight/insightTypes";
import type { PingInsightResponse } from "@/shared/api/types";
import type { PingInsightChartBucket, PingInsightSampleDensityCell } from "@/shared/visualizations/chartOptions";
import { formatCount, formatMs, formatPercent } from "../insightFormatters";

function pingSuccessRate(summary: PingInsightResponse["summary"] | undefined) {
	if (!summary?.totalResults) {
		return undefined;
	}

	return (summary.successfulCount / summary.totalResults) * 100;
}

export function pingSummaryMetrics(data: PingInsightResponse | undefined): SummaryMetric[] {
	const summary = data?.summary;
	const sampleCount = data?.sampleDensity.reduce((total, cell) => total + cell.sampleCount, 0) ?? 0;

	return [
		{ label: "Latest", value: formatMs(summary?.latestRttAvgMs), detail: summary?.latestStatus || "no result" },
		{ label: "Average", value: formatMs(summary?.avgRttMs), detail: "rtt avg" },
		{ label: "P95", value: formatMs(summary?.p95RttMs), detail: "sample percentile" },
		{ label: "P99", value: formatMs(summary?.p99RttMs), detail: "sample percentile" },
		{ label: "Max", value: formatMs(summary?.maxRttMs), detail: "rtt max" },
		{ label: "Loss", value: formatPercent(summary?.avgLossPercent), detail: "average" },
		{ label: "Success", value: formatPercent(pingSuccessRate(summary)), detail: `${formatCount(summary?.successfulCount)}/${formatCount(summary?.totalResults)}` },
		{ label: "Samples", value: formatCount(sampleCount), detail: `${formatCount(summary?.receivedCount)} replies` }
	];
}

export function pingChartBuckets(data: PingInsightResponse | undefined): PingInsightChartBucket[] {
	return (data?.buckets ?? []).map(bucket => ({
		timestampMs: bucket.timestampMs,
		rttMinMs: bucket.rttMinMs,
		rttAvgMs: bucket.rttAvgMs,
		rttMedianMs: bucket.rttMedianMs,
		rttMaxMs: bucket.rttMaxMs,
		rttStddevMs: bucket.rttStddevMs,
		lossPercent: bucket.lossPercent,
		sentCount: bucket.sentCount,
		receivedCount: bucket.receivedCount,
		resultCount: bucket.resultCount
	}));
}

export function pingSampleDensity(data: PingInsightResponse | undefined): PingInsightSampleDensityCell[] {
	return (data?.sampleDensity ?? []).map(cell => ({
		timestampMs: cell.timestampMs,
		rttBucketStartMs: cell.rttBucketStartMs,
		rttBucketEndMs: cell.rttBucketEndMs,
		sampleCount: cell.sampleCount
	}));
}

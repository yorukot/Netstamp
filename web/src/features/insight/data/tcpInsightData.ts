import type { SummaryMetric } from "@/features/insight/insightTypes";
import type { TcpInsightResponse } from "@/shared/api/types";
import type { TcpInsightChartBucket } from "@/shared/visualizations/chartOptions";
import { formatCount, formatMs, formatPercent } from "../insightFormatters";

function tcpSuccessRate(summary: TcpInsightResponse["summary"] | undefined) {
	if (!summary?.totalResults) {
		return undefined;
	}

	return (summary.successfulCount / summary.totalResults) * 100;
}

function tcpFailureRate(summary: TcpInsightResponse["summary"] | undefined) {
	if (!summary?.totalResults) {
		return undefined;
	}

	return ((summary.timeoutCount + summary.errorCount) / summary.totalResults) * 100;
}

export function tcpSummaryMetrics(data: TcpInsightResponse | undefined): SummaryMetric[] {
	const summary = data?.summary;

	return [
		{ label: "Latest", value: formatMs(summary?.latestConnectMs), detail: summary?.latestStatus || "no result" },
		{ label: "Average", value: formatMs(summary?.avgConnectMs), detail: "connect avg" },
		{ label: "P95", value: formatMs(summary?.p95ConnectMs), detail: "connect percentile" },
		{ label: "P99", value: formatMs(summary?.p99ConnectMs), detail: "connect percentile" },
		{ label: "Max", value: formatMs(summary?.maxConnectMs), detail: "connect max" },
		{ label: "Failure", value: formatPercent(tcpFailureRate(summary)), detail: `${formatCount(summary?.timeoutCount)} timeout · ${formatCount(summary?.errorCount)} error` },
		{ label: "Success", value: formatPercent(tcpSuccessRate(summary)), detail: `${formatCount(summary?.successfulCount)}/${formatCount(summary?.totalResults)}` },
		{ label: "Results", value: formatCount(summary?.totalResults), detail: "connect attempts" }
	];
}

export function tcpChartBuckets(data: TcpInsightResponse | undefined): TcpInsightChartBucket[] {
	return (data?.buckets ?? []).map(bucket => ({
		timestampMs: bucket.timestampMs,
		connectMinMs: bucket.connectMinMs,
		connectAvgMs: bucket.connectAvgMs,
		connectMedianMs: bucket.connectMedianMs,
		connectMaxMs: bucket.connectMaxMs,
		connectStddevMs: bucket.connectStddevMs,
		successRate: bucket.successRate,
		resultCount: bucket.resultCount,
		timeoutCount: bucket.timeoutCount,
		errorCount: bucket.errorCount
	}));
}

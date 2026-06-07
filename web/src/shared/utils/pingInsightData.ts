import type { PingInsightResponse, PingSeriesKey, PingSeriesResponse } from "@/shared/api/types";
import type { PingSeriesChartData, PingSeriesPoint } from "@/shared/visualizations/chartOptions";
import { formatCount, formatMs, formatPercent } from "./insightFormatters";
import type { SummaryMetric } from "./insightTypes";

export function pingSummaryMetrics(data: PingInsightResponse | undefined): SummaryMetric[] {
	const summary = data?.summary;

	return [
		{ label: "Average", value: formatMs(summary?.averageRttMs), detail: "rtt avg" },
		{ label: "Max", value: formatMs(summary?.maxRttMs), detail: "rtt max" },
		{ label: "Loss", value: formatPercent(summary?.lossPercent), detail: "average" },
		{ label: "Success", value: formatPercent(summary?.successRate), detail: "successful" },
		{ label: "Samples", value: formatCount(summary?.samples), detail: "replies" }
	];
}

function seriesPoints(data: PingSeriesResponse | undefined, key: PingSeriesKey): PingSeriesPoint[] {
	return (data?.series[key]?.points ?? []).filter((point): point is PingSeriesPoint => {
		const [timestampMs, value] = point;
		return Number.isFinite(timestampMs) && Number.isFinite(value);
	});
}

export function pingSeriesChartData(data: PingSeriesResponse | undefined): PingSeriesChartData {
	return {
		latencyAvg: seriesPoints(data, "latency_avg"),
		latencyMin: seriesPoints(data, "latency_min"),
		latencyMax: seriesPoints(data, "latency_max"),
		lossPercent: seriesPoints(data, "loss_percent")
	};
}

export function hasPingSeriesChartData(data: PingSeriesChartData) {
	return data.latencyAvg.length > 0 || data.latencyMin.length > 0 || data.latencyMax.length > 0 || data.lossPercent.length > 0;
}

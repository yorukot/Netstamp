import type { TcpInsightResponse, TcpSeriesKey, TcpSeriesResponse } from "@/shared/api/types";
import { formatCount, formatMs, formatPercent } from "@/shared/utils/insightFormatters";
import type { SummaryMetric } from "@/shared/utils/insightTypes";
import type { PingSeriesPoint, TcpSeriesChartData } from "@/shared/visualizations/chartOptions";

export function tcpSummaryMetrics(data: TcpInsightResponse | undefined): SummaryMetric[] {
	const summary = data?.summary;

	return [
		{ label: "Average", value: formatMs(summary?.averageConnectMs), detail: "connect avg" },
		{ label: "Max", value: formatMs(summary?.maxConnectMs), detail: "connect max" },
		{ label: "Failure", value: formatPercent(summary?.failurePercent), detail: "timeout + error" },
		{ label: "Success", value: formatPercent(summary?.successRate), detail: "successful" },
		{ label: "Samples", value: formatCount(summary?.samples), detail: "connect attempts" }
	];
}

function seriesPoints(data: TcpSeriesResponse | undefined, key: TcpSeriesKey): PingSeriesPoint[] {
	return (data?.series[key]?.points ?? []).filter((point): point is PingSeriesPoint => {
		const [timestampMs, value] = point;
		return Number.isFinite(timestampMs) && Number.isFinite(value);
	});
}

export function tcpSeriesChartData(data: TcpSeriesResponse | undefined): TcpSeriesChartData {
	return {
		connectAvg: seriesPoints(data, "connect_avg"),
		connectMin: seriesPoints(data, "connect_min"),
		connectMax: seriesPoints(data, "connect_max"),
		failurePercent: seriesPoints(data, "failure_percent")
	};
}

export function hasTcpSeriesChartData(data: TcpSeriesChartData) {
	return data.connectAvg.length > 0 || data.connectMin.length > 0 || data.connectMax.length > 0 || data.failurePercent.length > 0;
}

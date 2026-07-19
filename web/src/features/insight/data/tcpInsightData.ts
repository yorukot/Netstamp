import type { TcpSeriesKey, TcpSeriesResponse } from "@/shared/api/types";
import type { PingSeriesPoint, TcpSeriesChartData } from "@/shared/visualizations/chartOptions";

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

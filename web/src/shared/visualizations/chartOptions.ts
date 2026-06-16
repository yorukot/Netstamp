const axisLabel = { color: "#64748B", fontFamily: "Inter, system-ui, sans-serif", fontSize: 10 };
const splitLine = { lineStyle: { color: "rgba(148,163,184,0.18)" } };
const chartPrimary = "#EA6A1A";
const chartPrimaryBrush = "rgba(234, 106, 26, 0.12)";
const chartPrimaryBrushBorder = "rgba(234, 106, 26, 0.72)";
const chartSecondary = "#2563EB";
const chartWarning = "#B7791F";

export type ChartOption = Record<string, unknown>;
export { barChartOption, lineChartOption } from "./basicChartOptions";

export type PingSeriesPoint = [number, number];

export interface PingSeriesChartData {
	latencyAvg: PingSeriesPoint[];
	latencyMin: PingSeriesPoint[];
	latencyMax: PingSeriesPoint[];
	lossPercent: PingSeriesPoint[];
}

export interface TcpSeriesChartData {
	connectAvg: PingSeriesPoint[];
	connectMin: PingSeriesPoint[];
	connectMax: PingSeriesPoint[];
	failurePercent: PingSeriesPoint[];
}

export interface InsightMultiSeriesLine {
	id: string;
	name: string;
	color: string;
	points: PingSeriesPoint[];
}

interface TooltipParam {
	seriesName?: string;
	value?: unknown;
	marker?: string;
}

interface CustomRenderApi {
	coord: (value: [number, number]) => [number, number];
	value: (dimension: number) => number | string;
}

interface CartesianCoordSys {
	height: number;
	width: number;
	x: number;
	y: number;
}

interface CustomRenderParams {
	coordSys?: CartesianCoordSys;
	dataIndex: number;
}

function timestampLabel(value: number) {
	return new Date(value).toLocaleString([], {
		month: "short",
		day: "2-digit",
		hour: "2-digit",
		minute: "2-digit"
	});
}

function pingSeriesData(points: PingSeriesPoint[]) {
	return points.filter(([, value]) => Number.isFinite(value));
}

function pointMap(points: PingSeriesPoint[]) {
	return new Map(points.filter(([timestampMs, value]) => Number.isFinite(timestampMs) && Number.isFinite(value)).map(([timestampMs, value]) => [timestampMs, value]));
}

function pingSpreadData(data: PingSeriesChartData, key: "base" | "range") {
	const minByTimestamp = pointMap(data.latencyMin);
	const maxByTimestamp = pointMap(data.latencyMax);
	const timestamps = uniqueSortedTimestamps([...minByTimestamp.keys()].filter(timestamp => maxByTimestamp.has(timestamp)));

	return timestamps.flatMap(timestamp => {
		const min = minByTimestamp.get(timestamp);
		const max = maxByTimestamp.get(timestamp);

		if (typeof min !== "number" || typeof max !== "number") {
			return [];
		}

		return [[timestamp, key === "base" ? min : Math.max(max - min, 0)]];
	});
}

function tcpSpreadData(data: TcpSeriesChartData, key: "base" | "range") {
	const minByTimestamp = pointMap(data.connectMin);
	const maxByTimestamp = pointMap(data.connectMax);
	const timestamps = uniqueSortedTimestamps([...minByTimestamp.keys()].filter(timestamp => maxByTimestamp.has(timestamp)));

	return timestamps.flatMap(timestamp => {
		const min = minByTimestamp.get(timestamp);
		const max = maxByTimestamp.get(timestamp);

		if (typeof min !== "number" || typeof max !== "number") {
			return [];
		}

		return [[timestamp, key === "base" ? min : Math.max(max - min, 0)]];
	});
}

function tooltipValue(value: unknown) {
	if (!Array.isArray(value)) {
		return null;
	}

	const [, metric, sampleCount] = value;
	if (typeof metric !== "number") {
		return null;
	}

	return {
		metric,
		sampleCount: typeof sampleCount === "number" ? sampleCount : null,
		rttBucketStartMs: typeof value[3] === "number" ? value[3] : null,
		rttBucketEndMs: typeof value[4] === "number" ? value[4] : null
	};
}

function pingTooltipFormatter(params: unknown) {
	const items = (Array.isArray(params) ? params : [params]) as TooltipParam[];
	const firstValue = tooltipValue(items[0]?.value);
	const timestamp = Array.isArray(items[0]?.value) && typeof items[0]?.value?.[0] === "number" ? items[0].value[0] : null;
	const lines = timestamp ? [`<strong>${timestampLabel(timestamp)}</strong>`] : [];

	for (const item of items) {
		const parsed = tooltipValue(item.value);
		if (!parsed || item.seriesName === "range base" || item.seriesName === "latency spread") {
			continue;
		}

		const suffix = item.seriesName === "loss" ? "%" : "ms";
		lines.push(`${item.marker || ""}${item.seriesName}: ${parsed.metric.toFixed(1)}${suffix}`);
	}

	if (firstValue === null && lines.length === 0) {
		return "";
	}

	return lines.join("<br />");
}

function tcpTooltipFormatter(params: unknown) {
	const items = (Array.isArray(params) ? params : [params]) as TooltipParam[];
	const firstValue = tooltipValue(items[0]?.value);
	const timestamp = Array.isArray(items[0]?.value) && typeof items[0]?.value?.[0] === "number" ? items[0].value[0] : null;
	const lines = timestamp ? [`<strong>${timestampLabel(timestamp)}</strong>`] : [];

	for (const item of items) {
		const parsed = tooltipValue(item.value);
		if (!parsed || item.seriesName === "range base" || item.seriesName === "connect spread") {
			continue;
		}

		const suffix = item.seriesName === "failure" ? "%" : "ms";
		lines.push(`${item.marker || ""}${item.seriesName}: ${parsed.metric.toFixed(1)}${suffix}`);
	}

	if (firstValue === null && lines.length === 0) {
		return "";
	}

	return lines.join("<br />");
}

function multiSeriesTooltipFormatter(unit: string) {
	return (params: unknown) => {
		const items = (Array.isArray(params) ? params : [params]) as TooltipParam[];
		const timestamp = Array.isArray(items[0]?.value) && typeof items[0].value[0] === "number" ? items[0].value[0] : null;
		const lines = timestamp ? [`<strong>${timestampLabel(timestamp)}</strong>`] : [];

		for (const item of items) {
			const parsed = tooltipValue(item.value);
			if (!parsed) {
				continue;
			}

			lines.push(`${item.marker || ""}${item.seriesName}: ${parsed.metric.toFixed(1)}${unit}`);
		}

		return lines.join("<br />");
	};
}

function finiteNumbers(values: Array<number | undefined>) {
	return values.filter((value): value is number => typeof value === "number" && Number.isFinite(value));
}

function roundAxisFloor(value: number) {
	return Math.floor(value * 10) / 10;
}

function roundAxisCeil(value: number) {
	return Math.ceil(value * 10) / 10;
}

function pingRttAxisBounds(data: PingSeriesChartData) {
	const values = [...data.latencyAvg, ...data.latencyMin, ...data.latencyMax].flatMap(([, value]) => finiteNumbers([value])).filter(value => value >= 0);

	if (!values.length) {
		return { min: 0 };
	}

	const minValue = Math.min(...values);
	const maxValue = Math.max(...values);
	const span = Math.max(maxValue - minValue, maxValue * 0.08, 1);
	const padding = Math.max(span * 0.12, 0.5);
	const min = minValue <= 1 ? 0 : Math.max(0, roundAxisFloor(minValue - padding));
	const max = Math.max(min + 1, roundAxisCeil(maxValue + padding));

	return { min, max };
}

function tcpConnectAxisBounds(data: TcpSeriesChartData) {
	const values = [...data.connectAvg, ...data.connectMin, ...data.connectMax].flatMap(([, value]) => finiteNumbers([value])).filter(value => value >= 0);

	if (!values.length) {
		return { min: 0 };
	}

	const minValue = Math.min(...values);
	const maxValue = Math.max(...values);
	const span = Math.max(maxValue - minValue, maxValue * 0.08, 1);
	const padding = Math.max(span * 0.12, 0.5);
	const min = minValue <= 1 ? 0 : Math.max(0, roundAxisFloor(minValue - padding));
	const max = Math.max(min + 1, roundAxisCeil(maxValue + padding));

	return { min, max };
}

function lineAxisBounds(lines: InsightMultiSeriesLine[]) {
	const values = lines.flatMap(line => line.points.map(([, value]) => value)).filter(value => Number.isFinite(value) && value >= 0);

	if (!values.length) {
		return { min: 0 };
	}

	const minValue = Math.min(...values);
	const maxValue = Math.max(...values);
	const span = Math.max(maxValue - minValue, maxValue * 0.08, 1);
	const padding = Math.max(span * 0.12, 0.5);
	const min = minValue <= 1 ? 0 : Math.max(0, roundAxisFloor(minValue - padding));
	const max = Math.max(min + 1, roundAxisCeil(maxValue + padding));

	return { min, max };
}

function uniqueSortedTimestamps(timestamps: number[]) {
	return Array.from(new Set(timestamps.filter(timestamp => Number.isFinite(timestamp)))).sort((a, b) => a - b);
}

function inferTimestampStep(timestamps: number[]) {
	const gaps = timestamps.flatMap((timestamp, index) => {
		const next = timestamps[index + 1];
		return typeof next === "number" && next > timestamp ? [next - timestamp] : [];
	});

	if (!gaps.length) {
		return 60_000;
	}

	return gaps[Math.floor(gaps.length / 2)] || gaps[0] || 60_000;
}

function timeBoundsByTimestamp(timestampsInput: number[]) {
	const timestamps = uniqueSortedTimestamps(timestampsInput);
	const defaultStep = inferTimestampStep(timestamps);
	const bounds = new Map<number, { from: number; to: number }>();

	timestamps.forEach((timestamp, index) => {
		const previous = timestamps[index - 1];
		const next = timestamps[index + 1];
		const leftGap = typeof previous === "number" && timestamp > previous ? timestamp - previous : defaultStep;
		const rightGap = typeof next === "number" && next > timestamp ? next - timestamp : defaultStep;

		bounds.set(timestamp, {
			from: timestamp - leftGap / 2,
			to: timestamp + rightGap / 2
		});
	});

	return bounds;
}

function lossBandData(points: PingSeriesPoint[]) {
	const timeBounds = timeBoundsByTimestamp(points.map(([timestampMs]) => timestampMs));

	return points.flatMap(([timestampMs, lossPercent]) => {
		if (!Number.isFinite(lossPercent) || lossPercent <= 5) {
			return [];
		}

		const bounds = timeBounds.get(timestampMs) ?? { from: timestampMs - 30_000, to: timestampMs + 30_000 };
		return [[timestampMs, lossPercent, bounds.from, bounds.to]];
	});
}

function lossBandColor(lossPercent: number) {
	if (lossPercent >= 50) {
		return "rgba(201, 54, 44, 0.2)";
	}

	if (lossPercent >= 20) {
		return "rgba(201, 54, 44, 0.14)";
	}

	if (lossPercent >= 5) {
		return "rgba(183, 121, 31, 0.12)";
	}

	return "rgba(183, 121, 31, 0.12)";
}

const dataZoomToolbox = {
	show: true,
	left: -1000,
	top: -1000,
	itemSize: 0,
	itemGap: 0,
	feature: {
		dataZoom: {
			show: true,
			xAxisIndex: [0],
			yAxisIndex: false,
			filterMode: "none",
			brushStyle: {
				color: chartPrimaryBrush,
				borderColor: chartPrimaryBrushBorder,
				borderWidth: 1
			}
		}
	}
};

function multiSeriesChartOption(lines: InsightMultiSeriesLine[], axisName: string, unit: string): ChartOption {
	return {
		backgroundColor: "transparent",
		color: lines.map(line => line.color),
		tooltip: {
			trigger: "axis",
			backgroundColor: "rgba(255,255,255,0.98)",
			borderColor: "rgba(100,116,139,0.24)",
			textStyle: { color: "#111827", fontFamily: "Inter, system-ui, sans-serif", fontSize: 11 },
			formatter: multiSeriesTooltipFormatter(unit)
		},
		grid: { top: 18, right: 44, bottom: 30, left: 48 },
		toolbox: dataZoomToolbox,
		xAxis: {
			type: "time",
			axisLabel: {
				...axisLabel,
				formatter: (value: number) => timestampLabel(value)
			},
			axisLine: { lineStyle: { color: "rgba(148,163,184,0.16)" } },
			axisTick: { show: false }
		},
		yAxis: [
			{
				type: "value",
				name: axisName,
				nameTextStyle: { color: "#64748B", fontFamily: "Inter, system-ui, sans-serif", fontSize: 10 },
				axisLabel,
				splitLine,
				axisLine: { show: false },
				axisTick: { show: false },
				...lineAxisBounds(lines)
			}
		],
		series: lines.map(line => ({
			name: line.name,
			type: "line",
			data: pingSeriesData(line.points),
			showSymbol: false,
			smooth: true,
			lineStyle: { width: 2, color: line.color },
			z: 8
		}))
	};
}

function lossBandRenderItem(params: CustomRenderParams, api: CustomRenderApi) {
	const lossPercent = Number(api.value(1));
	const timeStartMs = Number(api.value(2));
	const timeEndMs = Number(api.value(3));
	const coordSys = params.coordSys;

	if (!coordSys || ![lossPercent, timeStartMs, timeEndMs].every(Number.isFinite) || lossPercent <= 0) {
		return undefined;
	}

	const start = api.coord([timeStartMs, 0]);
	const end = api.coord([timeEndMs, 0]);
	const x = Math.min(start[0], end[0]);
	const width = Math.max(1, Math.abs(end[0] - start[0]));

	return {
		type: "rect",
		shape: {
			x,
			y: coordSys.y,
			width,
			height: coordSys.height
		},
		style: {
			fill: lossBandColor(lossPercent)
		}
	};
}

export const insightSeriesPalette = [chartPrimary, chartSecondary, "#0891B2", "#0F766E", "#64748B", "#94A3B8", "#1D4ED8", "#0EA5E9", "#475569", "#334155"];

export function insightSeriesColor(index: number) {
	return insightSeriesPalette[index % insightSeriesPalette.length];
}

export function multiPingInsightChartOption(lines: InsightMultiSeriesLine[]): ChartOption {
	return multiSeriesChartOption(lines, "RTT ms", "ms");
}

export function multiTcpInsightChartOption(lines: InsightMultiSeriesLine[]): ChartOption {
	return multiSeriesChartOption(lines, "connect ms", "ms");
}

export function pingInsightChartOption(data: PingSeriesChartData): ChartOption {
	const lossBands = lossBandData(data.lossPercent);
	const rttAxisBounds = pingRttAxisBounds(data);

	return {
		backgroundColor: "transparent",
		color: [chartPrimary, chartWarning, "#C9362C"],
		tooltip: {
			trigger: "axis",
			backgroundColor: "rgba(255,255,255,0.98)",
			borderColor: "rgba(100,116,139,0.24)",
			textStyle: { color: "#111827", fontFamily: "Inter, system-ui, sans-serif", fontSize: 11 },
			formatter: pingTooltipFormatter
		},
		legend: {
			top: 0,
			right: 0,
			data: ["avg", "loss"],
			textStyle: { color: "#64748B", fontFamily: "Inter, system-ui, sans-serif", fontSize: 10 },
			itemWidth: 12,
			itemHeight: 6
		},
		grid: { top: 34, right: 44, bottom: 30, left: 48 },
		// Toolbox dataZoom must be mounted for ECharts' select brush to work.
		toolbox: {
			show: true,
			left: -1000,
			top: -1000,
			itemSize: 0,
			itemGap: 0,
			feature: {
				dataZoom: {
					show: true,
					xAxisIndex: [0],
					yAxisIndex: false,
					filterMode: "none",
					brushStyle: {
						color: chartPrimaryBrush,
						borderColor: chartPrimaryBrushBorder,
						borderWidth: 1
					}
				}
			}
		},
		xAxis: {
			type: "time",
			axisLabel: {
				...axisLabel,
				formatter: (value: number) => timestampLabel(value)
			},
			axisLine: { lineStyle: { color: "rgba(148,163,184,0.16)" } },
			axisTick: { show: false }
		},
		yAxis: [
			{
				type: "value",
				name: "RTT ms",
				nameTextStyle: { color: "#64748B", fontFamily: "Inter, system-ui, sans-serif", fontSize: 10 },
				axisLabel,
				splitLine,
				axisLine: { show: false },
				axisTick: { show: false },
				...rttAxisBounds
			}
		],
		series: [
			{
				name: "loss",
				type: "custom",
				coordinateSystem: "cartesian2d",
				renderItem: lossBandRenderItem,
				data: lossBands,
				clip: true,
				encode: { x: 0, tooltip: 1 },
				z: 1
			},
			{
				name: "range base",
				type: "line",
				stack: "rtt-spread",
				data: pingSpreadData(data, "base"),
				showSymbol: false,
				lineStyle: { opacity: 0 },
				areaStyle: { opacity: 0 },
				silent: true
			},
			{
				name: "latency spread",
				type: "line",
				stack: "rtt-spread",
				data: pingSpreadData(data, "range"),
				showSymbol: false,
				lineStyle: { opacity: 0 },
				areaStyle: { color: "rgba(196,204,217,0.08)" },
				silent: true
			},
			{
				name: "avg",
				type: "line",
				data: pingSeriesData(data.latencyAvg),
				showSymbol: false,
				smooth: true,
				lineStyle: { width: 2.25, color: chartPrimary },
				z: 8
			}
		]
	};
}

export function tcpInsightChartOption(data: TcpSeriesChartData): ChartOption {
	const failureBands = lossBandData(data.failurePercent);
	const connectAxisBounds = tcpConnectAxisBounds(data);

	return {
		backgroundColor: "transparent",
		color: [chartPrimary, chartWarning, "#C9362C"],
		tooltip: {
			trigger: "axis",
			backgroundColor: "rgba(255,255,255,0.98)",
			borderColor: "rgba(100,116,139,0.24)",
			textStyle: { color: "#111827", fontFamily: "Inter, system-ui, sans-serif", fontSize: 11 },
			formatter: tcpTooltipFormatter
		},
		legend: {
			top: 0,
			right: 0,
			data: ["avg", "failure"],
			textStyle: { color: "#64748B", fontFamily: "Inter, system-ui, sans-serif", fontSize: 10 },
			itemWidth: 12,
			itemHeight: 6
		},
		grid: { top: 34, right: 44, bottom: 30, left: 48 },
		toolbox: {
			show: true,
			left: -1000,
			top: -1000,
			itemSize: 0,
			itemGap: 0,
			feature: {
				dataZoom: {
					show: true,
					xAxisIndex: [0],
					yAxisIndex: false,
					filterMode: "none",
					brushStyle: {
						color: chartPrimaryBrush,
						borderColor: chartPrimaryBrushBorder,
						borderWidth: 1
					}
				}
			}
		},
		xAxis: {
			type: "time",
			axisLabel: {
				...axisLabel,
				formatter: (value: number) => timestampLabel(value)
			},
			axisLine: { lineStyle: { color: "rgba(148,163,184,0.16)" } },
			axisTick: { show: false }
		},
		yAxis: [
			{
				type: "value",
				name: "connect ms",
				nameTextStyle: { color: "#64748B", fontFamily: "Inter, system-ui, sans-serif", fontSize: 10 },
				axisLabel,
				splitLine,
				axisLine: { show: false },
				axisTick: { show: false },
				...connectAxisBounds
			}
		],
		series: [
			{
				name: "failure",
				type: "custom",
				coordinateSystem: "cartesian2d",
				renderItem: lossBandRenderItem,
				data: failureBands,
				clip: true,
				encode: { x: 0, tooltip: 1 },
				z: 1
			},
			{
				name: "range base",
				type: "line",
				stack: "connect-spread",
				data: tcpSpreadData(data, "base"),
				showSymbol: false,
				lineStyle: { opacity: 0 },
				areaStyle: { opacity: 0 },
				silent: true
			},
			{
				name: "connect spread",
				type: "line",
				stack: "connect-spread",
				data: tcpSpreadData(data, "range"),
				showSymbol: false,
				lineStyle: { opacity: 0 },
				areaStyle: { color: "rgba(196,204,217,0.08)" },
				silent: true
			},
			{
				name: "avg",
				type: "line",
				data: pingSeriesData(data.connectAvg),
				showSymbol: false,
				smooth: true,
				lineStyle: { width: 2.25, color: chartPrimary },
				z: 8
			}
		]
	};
}

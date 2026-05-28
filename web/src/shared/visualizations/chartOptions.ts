const axisLabel = { color: "#8C877F", fontFamily: "JetBrains Mono, monospace", fontSize: 10 };
const splitLine = { lineStyle: { color: "rgba(148,163,184,0.12)" } };

export type ChartOption = Record<string, unknown>;

export interface PingInsightChartBucket {
	timestampMs: number;
	rttMinMs?: number;
	rttAvgMs?: number;
	rttMedianMs?: number;
	rttMaxMs?: number;
	rttStddevMs?: number;
	lossPercent?: number;
	sentCount: number;
	receivedCount: number;
	resultCount: number;
}

export interface PingInsightSampleDensityCell {
	timestampMs: number;
	rttBucketStartMs: number;
	rttBucketEndMs: number;
	sampleCount: number;
}

export interface TcpInsightChartBucket {
	timestampMs: number;
	connectMinMs?: number;
	connectAvgMs?: number;
	connectMedianMs?: number;
	connectMaxMs?: number;
	connectStddevMs?: number;
	successRate?: number;
	resultCount: number;
	timeoutCount: number;
	errorCount: number;
}

type PingMetricKey = "rttMinMs" | "rttAvgMs" | "rttMedianMs" | "rttMaxMs" | "rttStddevMs" | "lossPercent";
type PingRttMetricKey = "rttMinMs" | "rttAvgMs" | "rttMedianMs" | "rttMaxMs";
type TcpMetricKey = "connectMinMs" | "connectAvgMs" | "connectMedianMs" | "connectMaxMs" | "connectStddevMs" | "failurePercent";
type TcpConnectMetricKey = "connectMinMs" | "connectAvgMs" | "connectMedianMs" | "connectMaxMs";

interface TooltipParam {
	seriesName?: string;
	value?: unknown;
	marker?: string;
}

interface SmokeRenderApi {
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

const pingRttMetricKeys: PingRttMetricKey[] = ["rttMinMs", "rttAvgMs", "rttMedianMs", "rttMaxMs"];
const tcpConnectMetricKeys: TcpConnectMetricKey[] = ["connectMinMs", "connectAvgMs", "connectMedianMs", "connectMaxMs"];

function timestampLabel(value: number) {
	return new Date(value).toLocaleString([], {
		month: "short",
		day: "2-digit",
		hour: "2-digit",
		minute: "2-digit"
	});
}

function metricData(buckets: PingInsightChartBucket[], key: PingMetricKey) {
	return buckets.flatMap(bucket => {
		const value = bucket[key];

		return typeof value === "number" && Number.isFinite(value) ? [[bucket.timestampMs, value]] : [];
	});
}

function tcpMetricData(buckets: TcpInsightChartBucket[], key: TcpMetricKey) {
	return buckets.flatMap(bucket => {
		const value = key === "failurePercent" ? tcpFailurePercent(bucket) : bucket[key];

		return typeof value === "number" && Number.isFinite(value) ? [[bucket.timestampMs, value]] : [];
	});
}

function spreadData(buckets: PingInsightChartBucket[], key: "base" | "range") {
	return buckets.flatMap(bucket => {
		if (typeof bucket.rttMinMs !== "number" || typeof bucket.rttMaxMs !== "number") {
			return [];
		}

		return [[bucket.timestampMs, key === "base" ? bucket.rttMinMs : Math.max(bucket.rttMaxMs - bucket.rttMinMs, 0)]];
	});
}

function tcpSpreadData(buckets: TcpInsightChartBucket[], key: "base" | "range") {
	return buckets.flatMap(bucket => {
		if (typeof bucket.connectMinMs !== "number" || typeof bucket.connectMaxMs !== "number") {
			return [];
		}

		return [[bucket.timestampMs, key === "base" ? bucket.connectMinMs : Math.max(bucket.connectMaxMs - bucket.connectMinMs, 0)]];
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

		if (item.seriesName === "smoke") {
			const bucket =
				parsed.rttBucketStartMs !== null && parsed.rttBucketEndMs !== null ? `${parsed.rttBucketStartMs.toFixed(1)}-${parsed.rttBucketEndMs.toFixed(1)}ms` : `${parsed.metric.toFixed(1)}ms`;
			const count = parsed.sampleCount !== null ? ` · ${parsed.sampleCount} samples` : "";
			lines.push(`${item.marker || ""}smoke: ${bucket}${count}`);
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

function finiteNumbers(values: Array<number | undefined>) {
	return values.filter((value): value is number => typeof value === "number" && Number.isFinite(value));
}

function roundAxisFloor(value: number) {
	return Math.floor(value * 10) / 10;
}

function roundAxisCeil(value: number) {
	return Math.ceil(value * 10) / 10;
}

function pingRttAxisBounds(buckets: PingInsightChartBucket[], sampleDensity: PingInsightSampleDensityCell[]) {
	const values = [
		...buckets.flatMap(bucket => pingRttMetricKeys.flatMap(key => finiteNumbers([bucket[key]]))),
		...sampleDensity.flatMap(cell => finiteNumbers([cell.rttBucketStartMs, cell.rttBucketEndMs]))
	].filter(value => value >= 0);

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

function tcpConnectAxisBounds(buckets: TcpInsightChartBucket[]) {
	const values = buckets.flatMap(bucket => tcpConnectMetricKeys.flatMap(key => finiteNumbers([bucket[key]]))).filter(value => value >= 0);

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

function smokeData(sampleDensity: PingInsightSampleDensityCell[]) {
	const timeBounds = timeBoundsByTimestamp(sampleDensity.map(cell => cell.timestampMs));

	return sampleDensity.map(cell => {
		const bounds = timeBounds.get(cell.timestampMs) ?? { from: cell.timestampMs - 30_000, to: cell.timestampMs + 30_000 };
		return [cell.timestampMs, (cell.rttBucketStartMs + cell.rttBucketEndMs) / 2, cell.sampleCount, cell.rttBucketStartMs, cell.rttBucketEndMs, bounds.from, bounds.to];
	});
}

function lossBandData(buckets: PingInsightChartBucket[]) {
	const timeBounds = timeBoundsByTimestamp(buckets.map(bucket => bucket.timestampMs));

	return buckets.flatMap(bucket => {
		const lossPercent = bucket.lossPercent;
		if (typeof lossPercent !== "number" || !Number.isFinite(lossPercent) || lossPercent <= 5) {
			return [];
		}

		const bounds = timeBounds.get(bucket.timestampMs) ?? { from: bucket.timestampMs - 30_000, to: bucket.timestampMs + 30_000 };
		return [[bucket.timestampMs, lossPercent, bounds.from, bounds.to]];
	});
}

function tcpFailurePercent(bucket: TcpInsightChartBucket) {
	if (typeof bucket.successRate === "number" && Number.isFinite(bucket.successRate)) {
		return Math.max(0, Math.min(100, 100 - bucket.successRate));
	}

	if (!bucket.resultCount) {
		return undefined;
	}

	return (100 * (bucket.timeoutCount + bucket.errorCount)) / bucket.resultCount;
}

function tcpFailureBandData(buckets: TcpInsightChartBucket[]) {
	const timeBounds = timeBoundsByTimestamp(buckets.map(bucket => bucket.timestampMs));

	return buckets.flatMap(bucket => {
		const failurePercent = tcpFailurePercent(bucket);
		if (typeof failurePercent !== "number" || !Number.isFinite(failurePercent) || failurePercent <= 5) {
			return [];
		}

		const bounds = timeBounds.get(bucket.timestampMs) ?? { from: bucket.timestampMs - 30_000, to: bucket.timestampMs + 30_000 };
		return [[bucket.timestampMs, failurePercent, bounds.from, bounds.to]];
	});
}

function smokeCellColor(sampleCount: number, maxSampleCount: number) {
	const intensity = Math.log1p(Math.max(sampleCount, 0)) / Math.log1p(Math.max(maxSampleCount, 1));
	const opacity = 0.08 + Math.min(1, intensity) * 0.46;
	return `rgba(255, 247, 236, ${opacity.toFixed(3)})`;
}

function smokeRenderItem(maxSampleCount: number) {
	return function renderSmokeCell(_params: CustomRenderParams, api: SmokeRenderApi) {
		const sampleCount = Number(api.value(2));
		const rttStartMs = Number(api.value(3));
		const rttEndMs = Number(api.value(4));
		const timeStartMs = Number(api.value(5));
		const timeEndMs = Number(api.value(6));

		if (![sampleCount, rttStartMs, rttEndMs, timeStartMs, timeEndMs].every(Number.isFinite)) {
			return undefined;
		}

		const start = api.coord([timeStartMs, rttEndMs]);
		const end = api.coord([timeEndMs, rttStartMs]);
		const x = Math.min(start[0], end[0]);
		const y = Math.min(start[1], end[1]);
		const width = Math.max(1, Math.abs(end[0] - start[0]));
		const height = Math.max(1.5, Math.abs(end[1] - start[1]));

		return {
			type: "rect",
			shape: { x, y, width, height },
			style: {
				fill: smokeCellColor(sampleCount, maxSampleCount)
			}
		};
	};
}

function lossBandColor(lossPercent: number) {
	if (lossPercent >= 50) {
		return "rgba(255, 69, 58, 0.34)";
	}

	if (lossPercent >= 20) {
		return "rgba(255, 69, 58, 0.24)";
	}

	if (lossPercent >= 5) {
		return "rgba(255, 122, 26, 0.18)";
	}

	return "rgba(255, 122, 26, 0.18)";
}

function lossBandRenderItem(params: CustomRenderParams, api: SmokeRenderApi) {
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

export function lineChartOption(title: string, values: number[], secondaryValues: number[] = []): ChartOption {
	const labels = values.map((_, index) => `${index * 10}m`);

	return {
		backgroundColor: "transparent",
		color: ["#FF7A1A", "#94A3B8"],
		tooltip: {
			trigger: "axis",
			backgroundColor: "rgba(10,13,18,0.92)",
			borderColor: "rgba(255,255,255,0.14)",
			textStyle: { color: "#F6F2EA", fontFamily: "Inter, sans-serif" }
		},
		grid: { top: 22, right: 12, bottom: 24, left: 34 },
		xAxis: { type: "category", data: labels, boundaryGap: false, axisLabel, axisLine: { lineStyle: { color: "rgba(148,163,184,0.16)" } }, axisTick: { show: false } },
		yAxis: { type: "value", axisLabel, splitLine, axisLine: { show: false }, axisTick: { show: false } },
		series: [
			{
				name: title,
				type: "line",
				data: values,
				smooth: true,
				showSymbol: false,
				lineStyle: { width: 3, shadowBlur: 18, shadowColor: "rgba(255,122,26,0.34)" },
				areaStyle: {
					color: {
						type: "linear",
						x: 0,
						y: 0,
						x2: 0,
						y2: 1,
						colorStops: [
							{ offset: 0, color: "rgba(255,122,26,0.28)" },
							{ offset: 1, color: "rgba(255,122,26,0.0)" }
						]
					}
				}
			},
			secondaryValues.length
				? {
						name: "baseline",
						type: "line",
						data: secondaryValues,
						smooth: true,
						showSymbol: false,
						lineStyle: { width: 1, color: "rgba(148,163,184,0.52)" }
					}
				: null
		].filter(Boolean)
	};
}

export function barChartOption(values: number[], name = "events"): ChartOption {
	return {
		backgroundColor: "transparent",
		color: ["#FF7A1A"],
		tooltip: {
			trigger: "axis",
			backgroundColor: "rgba(10,13,18,0.92)",
			borderColor: "rgba(255,255,255,0.14)",
			textStyle: { color: "#F6F2EA" }
		},
		grid: { top: 22, right: 10, bottom: 22, left: 28 },
		xAxis: { type: "category", data: values.map((_, index) => `${index + 1}`), axisLabel, axisTick: { show: false }, axisLine: { lineStyle: { color: "rgba(148,163,184,0.16)" } } },
		yAxis: { type: "value", axisLabel, splitLine, axisTick: { show: false }, axisLine: { show: false } },
		series: [
			{
				name,
				type: "bar",
				data: values,
				barWidth: "42%",
				itemStyle: {
					borderRadius: [6, 6, 0, 0],
					color: {
						type: "linear",
						x: 0,
						y: 0,
						x2: 0,
						y2: 1,
						colorStops: [
							{ offset: 0, color: "#FF8F3D" },
							{ offset: 1, color: "rgba(255,122,26,0.18)" }
						]
					}
				}
			}
		]
	};
}

export function pingInsightChartOption(buckets: PingInsightChartBucket[], sampleDensity: PingInsightSampleDensityCell[]): ChartOption {
	const maxSampleCount = Math.max(1, ...sampleDensity.map(cell => cell.sampleCount));
	const smokeCells = smokeData(sampleDensity);
	const lossBands = lossBandData(buckets);
	const rttAxisBounds = pingRttAxisBounds(buckets, sampleDensity);

	return {
		backgroundColor: "transparent",
		color: ["#FFF7EC", "#FF7A1A", "#FF453A"],
		tooltip: {
			trigger: "axis",
			backgroundColor: "rgba(2,3,4,0.96)",
			borderColor: "rgba(255,122,26,0.34)",
			textStyle: { color: "#FFF7EC", fontFamily: "JetBrains Mono, monospace", fontSize: 11 },
			formatter: pingTooltipFormatter
		},
		legend: {
			top: 0,
			right: 0,
			data: ["smoke", "median", "loss"],
			textStyle: { color: "#B8B3AA", fontFamily: "JetBrains Mono, monospace", fontSize: 10 },
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
						color: "rgba(255, 122, 26, 0.16)",
						borderColor: "rgba(255, 122, 26, 0.76)",
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
				nameTextStyle: { color: "#77736B", fontFamily: "JetBrains Mono, monospace", fontSize: 10 },
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
				name: "smoke",
				type: "custom",
				coordinateSystem: "cartesian2d",
				renderItem: smokeRenderItem(maxSampleCount),
				data: smokeCells,
				clip: true,
				encode: { x: 0, y: 1 },
				tooltip: { show: false },
				z: 4
			},
			{
				name: "range base",
				type: "line",
				stack: "rtt-spread",
				data: spreadData(buckets, "base"),
				showSymbol: false,
				lineStyle: { opacity: 0 },
				areaStyle: { opacity: 0 },
				silent: true
			},
			{
				name: "latency spread",
				type: "line",
				stack: "rtt-spread",
				data: spreadData(buckets, "range"),
				showSymbol: false,
				lineStyle: { opacity: 0 },
				areaStyle: { color: "rgba(196,204,217,0.08)" },
				silent: true
			},
			{
				name: "median",
				type: "line",
				data: metricData(buckets, "rttMedianMs"),
				showSymbol: false,
				smooth: true,
				lineStyle: { width: 2.25, color: "#FF7A1A", shadowBlur: 12, shadowColor: "rgba(255,122,26,0.36)" },
				z: 8
			}
		]
	};
}

export function tcpInsightChartOption(buckets: TcpInsightChartBucket[]): ChartOption {
	const failureBands = tcpFailureBandData(buckets);
	const connectAxisBounds = tcpConnectAxisBounds(buckets);

	return {
		backgroundColor: "transparent",
		color: ["#FFF7EC", "#FF7A1A", "#FF453A"],
		tooltip: {
			trigger: "axis",
			backgroundColor: "rgba(2,3,4,0.96)",
			borderColor: "rgba(255,122,26,0.34)",
			textStyle: { color: "#FFF7EC", fontFamily: "JetBrains Mono, monospace", fontSize: 11 },
			formatter: tcpTooltipFormatter
		},
		legend: {
			top: 0,
			right: 0,
			data: ["median", "failure"],
			textStyle: { color: "#B8B3AA", fontFamily: "JetBrains Mono, monospace", fontSize: 10 },
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
						color: "rgba(255, 122, 26, 0.16)",
						borderColor: "rgba(255, 122, 26, 0.76)",
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
				nameTextStyle: { color: "#77736B", fontFamily: "JetBrains Mono, monospace", fontSize: 10 },
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
				data: tcpSpreadData(buckets, "base"),
				showSymbol: false,
				lineStyle: { opacity: 0 },
				areaStyle: { opacity: 0 },
				silent: true
			},
			{
				name: "connect spread",
				type: "line",
				stack: "connect-spread",
				data: tcpSpreadData(buckets, "range"),
				showSymbol: false,
				lineStyle: { opacity: 0 },
				areaStyle: { color: "rgba(196,204,217,0.08)" },
				silent: true
			},
			{
				name: "median",
				type: "line",
				data: tcpMetricData(buckets, "connectMedianMs"),
				showSymbol: false,
				smooth: true,
				lineStyle: { width: 2.25, color: "#FF7A1A", shadowBlur: 12, shadowColor: "rgba(255,122,26,0.36)" },
				z: 8
			}
		]
	};
}

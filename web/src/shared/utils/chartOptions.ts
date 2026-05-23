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

type PingMetricKey = "rttMinMs" | "rttAvgMs" | "rttMedianMs" | "rttMaxMs" | "rttStddevMs" | "lossPercent";

interface TooltipParam {
	seriesName?: string;
	value?: unknown;
	marker?: string;
}

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

function spreadData(buckets: PingInsightChartBucket[], key: "base" | "range") {
	return buckets.flatMap(bucket => {
		if (typeof bucket.rttMinMs !== "number" || typeof bucket.rttMaxMs !== "number") {
			return [];
		}

		return [[bucket.timestampMs, key === "base" ? bucket.rttMinMs : Math.max(bucket.rttMaxMs - bucket.rttMinMs, 0)]];
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
		sampleCount: typeof sampleCount === "number" ? sampleCount : null
	};
}

function pingTooltipFormatter(params: unknown) {
	const items = (Array.isArray(params) ? params : [params]) as TooltipParam[];
	const firstValue = tooltipValue(items[0]?.value);
	const timestamp = Array.isArray(items[0]?.value) && typeof items[0]?.value?.[0] === "number" ? items[0].value[0] : null;
	const lines = timestamp ? [`<strong>${timestampLabel(timestamp)}</strong>`] : [];

	for (const item of items) {
		const parsed = tooltipValue(item.value);
		if (!parsed || item.seriesName === "spread base" || item.seriesName === "spread range") {
			continue;
		}

		const suffix = item.seriesName === "loss" ? "%" : item.seriesName === "sample density" ? "ms" : "ms";
		const count = item.seriesName === "sample density" && parsed.sampleCount !== null ? ` · ${parsed.sampleCount} samples` : "";
		lines.push(`${item.marker || ""}${item.seriesName}: ${parsed.metric.toFixed(1)}${suffix}${count}`);
	}

	if (firstValue === null && lines.length === 0) {
		return "";
	}

	return lines.join("<br />");
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
	const densityData = sampleDensity.map(cell => [cell.timestampMs, (cell.rttBucketStartMs + cell.rttBucketEndMs) / 2, cell.sampleCount]);

	return {
		backgroundColor: "transparent",
		color: ["#FF7A1A", "#FFE0C2", "#C4CCD9", "#77736B", "#FF453A"],
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
			textStyle: { color: "#B8B3AA", fontFamily: "JetBrains Mono, monospace", fontSize: 10 },
			itemWidth: 12,
			itemHeight: 6
		},
		grid: { top: 34, right: 44, bottom: 54, left: 48 },
		visualMap: {
			show: false,
			type: "continuous",
			min: 0,
			max: maxSampleCount,
			dimension: 2,
			seriesIndex: 0,
			inRange: {
				color: ["rgba(255,122,26,0.18)", "rgba(255,122,26,0.5)", "#FF7A1A", "#FFF7EC"]
			}
		},
		dataZoom: [
			{ type: "inside", xAxisIndex: [0], filterMode: "none" },
			{
				type: "slider",
				xAxisIndex: [0],
				bottom: 10,
				height: 18,
				borderColor: "rgba(255,255,255,0.16)",
				backgroundColor: "rgba(255,255,255,0.025)",
				fillerColor: "rgba(255,122,26,0.18)",
				handleStyle: { color: "#FF7A1A" },
				textStyle: { color: "#77736B", fontFamily: "JetBrains Mono, monospace" }
			}
		],
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
				min: 0
			},
			{
				type: "value",
				name: "loss %",
				nameTextStyle: { color: "#77736B", fontFamily: "JetBrains Mono, monospace", fontSize: 10 },
				axisLabel,
				splitLine: { show: false },
				axisLine: { show: false },
				axisTick: { show: false },
				min: 0,
				max: 100
			}
		],
		series: [
			{
				name: "sample density",
				type: "scatter",
				data: densityData,
				showSymbol: true,
				symbolSize: (value: number[]) => Math.min(13, 3 + Math.sqrt(value[2] || 1) * 2.2),
				itemStyle: { opacity: 0.82 },
				z: 6
			},
			{
				name: "spread base",
				type: "line",
				stack: "rtt-spread",
				data: spreadData(buckets, "base"),
				showSymbol: false,
				lineStyle: { opacity: 0 },
				areaStyle: { opacity: 0 },
				silent: true
			},
			{
				name: "spread range",
				type: "line",
				stack: "rtt-spread",
				data: spreadData(buckets, "range"),
				showSymbol: false,
				lineStyle: { opacity: 0 },
				areaStyle: { color: "rgba(255,122,26,0.12)" },
				silent: true
			},
			{
				name: "avg",
				type: "line",
				data: metricData(buckets, "rttAvgMs"),
				showSymbol: false,
				smooth: true,
				lineStyle: { width: 2.5, color: "#FF7A1A", shadowBlur: 14, shadowColor: "rgba(255,122,26,0.32)" },
				z: 8
			},
			{
				name: "median",
				type: "line",
				data: metricData(buckets, "rttMedianMs"),
				showSymbol: false,
				smooth: true,
				lineStyle: { width: 1.5, color: "#FFE0C2" },
				z: 7
			},
			{
				name: "min",
				type: "line",
				data: metricData(buckets, "rttMinMs"),
				showSymbol: false,
				lineStyle: { width: 1, color: "rgba(196,204,217,0.42)" }
			},
			{
				name: "max",
				type: "line",
				data: metricData(buckets, "rttMaxMs"),
				showSymbol: false,
				lineStyle: { width: 1, color: "rgba(196,204,217,0.52)" }
			},
			{
				name: "loss",
				type: "bar",
				yAxisIndex: 1,
				data: metricData(buckets, "lossPercent"),
				barWidth: 4,
				itemStyle: { color: "rgba(255,69,58,0.52)" },
				z: 2
			}
		]
	};
}

import type { ChartOption } from "./chartOptions";

const axisLabel = { color: "#64748B", fontFamily: "Inter, system-ui, sans-serif", fontSize: 10 };
const splitLine = { lineStyle: { color: "rgba(148,163,184,0.18)" } };

export function lineChartOption(title: string, values: number[], secondaryValues: number[] = []): ChartOption {
	const labels = values.map((_, index) => `${index * 10}m`);

	return {
		backgroundColor: "transparent",
		color: ["#2563EB", "#94A3B8"],
		tooltip: {
			trigger: "axis",
			backgroundColor: "rgba(255,255,255,0.98)",
			borderColor: "rgba(100,116,139,0.24)",
			textStyle: { color: "#111827", fontFamily: "Inter, system-ui, sans-serif" }
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
				lineStyle: { width: 2.5 },
				areaStyle: {
					color: "rgba(37,99,235,0.12)"
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
		color: ["#2563EB"],
		tooltip: {
			trigger: "axis",
			backgroundColor: "rgba(255,255,255,0.98)",
			borderColor: "rgba(100,116,139,0.24)",
			textStyle: { color: "#111827", fontFamily: "Inter, system-ui, sans-serif" }
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
					color: "#2563EB"
				}
			}
		]
	};
}

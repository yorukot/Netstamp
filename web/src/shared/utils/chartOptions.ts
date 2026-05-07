const axisLabel = { color: "#8C877F", fontFamily: "JetBrains Mono, monospace", fontSize: 10 };
const splitLine = { lineStyle: { color: "rgba(148,163,184,0.12)" } };

export type ChartOption = Record<string, unknown>;

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

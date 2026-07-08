import type { ChartOption } from "./chartOptions";
import { chartAxisLabel, chartTheme, chartTooltipTextStyle } from "./chartTheme";

export function lineChartOption(title: string, values: number[], secondaryValues: number[] = []): ChartOption {
	const labels = values.map((_, index) => `${index * 10}m`);
	const theme = chartTheme();
	const splitLine = { lineStyle: { color: theme.splitLine } };

	return {
		backgroundColor: "transparent",
		color: [theme.primary, theme.secondary],
		tooltip: {
			trigger: "axis",
			backgroundColor: theme.tooltipBackground,
			borderColor: theme.tooltipBorder,
			textStyle: chartTooltipTextStyle()
		},
		grid: { top: 22, right: 12, bottom: 24, left: 34 },
		xAxis: { type: "category", data: labels, boundaryGap: false, axisLabel: chartAxisLabel(), axisLine: { lineStyle: { color: theme.axisLine } }, axisTick: { show: false } },
		yAxis: { type: "value", axisLabel: chartAxisLabel(), splitLine, axisLine: { show: false }, axisTick: { show: false } },
		series: [
			{
				name: title,
				type: "line",
				data: values,
				smooth: true,
				showSymbol: false,
				lineStyle: { width: 2.5 },
				areaStyle: {
					color: theme.primaryFill
				}
			},
			secondaryValues.length
				? {
						name: "baseline",
						type: "line",
						data: secondaryValues,
						smooth: true,
						showSymbol: false,
						lineStyle: { width: 1, color: theme.baselineLine }
					}
				: null
		].filter(Boolean)
	};
}

export function barChartOption(values: number[], name = "events"): ChartOption {
	const theme = chartTheme();
	const splitLine = { lineStyle: { color: theme.splitLine } };

	return {
		backgroundColor: "transparent",
		color: [theme.primary],
		tooltip: {
			trigger: "axis",
			backgroundColor: theme.tooltipBackground,
			borderColor: theme.tooltipBorder,
			textStyle: chartTooltipTextStyle()
		},
		grid: { top: 22, right: 10, bottom: 22, left: 28 },
		xAxis: { type: "category", data: values.map((_, index) => `${index + 1}`), axisLabel: chartAxisLabel(), axisTick: { show: false }, axisLine: { lineStyle: { color: theme.axisLine } } },
		yAxis: { type: "value", axisLabel: chartAxisLabel(), splitLine, axisTick: { show: false }, axisLine: { show: false } },
		series: [
			{
				name,
				type: "bar",
				data: values,
				barWidth: "42%",
				itemStyle: {
					borderRadius: [6, 6, 0, 0],
					color: theme.primary
				}
			}
		]
	};
}

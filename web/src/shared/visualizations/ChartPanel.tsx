import type { ChartOption } from "@/shared/visualizations/chartOptions";
import { useEffect, useRef } from "react";
import styles from "./ChartPanel.module.css";

interface ChartPanelProps {
	option: ChartOption;
	height?: string;
	className?: string;
}

interface ChartInstance {
	dispose: () => void;
	resize: () => void;
	setOption: (option: ChartOption) => void;
}

let chartRuntimePromise: Promise<typeof import("echarts/core")> | null = null;

function loadChartRuntime() {
	chartRuntimePromise ??= Promise.all([import("echarts/core"), import("echarts/charts"), import("echarts/components"), import("echarts/renderers")]).then(
		([echarts, charts, components, renderers]) => {
			echarts.use([
				charts.LineChart,
				charts.BarChart,
				charts.GraphChart,
				charts.ScatterChart,
				components.GridComponent,
				components.TooltipComponent,
				components.LegendComponent,
				components.DataZoomComponent,
				components.VisualMapComponent,
				renderers.CanvasRenderer
			]);

			return echarts;
		}
	);

	return chartRuntimePromise;
}

export function ChartPanel({ option, height = "16rem", className }: ChartPanelProps) {
	const chartRef = useRef<HTMLDivElement | null>(null);
	const instanceRef = useRef<ChartInstance | null>(null);
	const optionRef = useRef(option);

	useEffect(() => {
		optionRef.current = option;
		instanceRef.current?.setOption(option);
	}, [option]);

	useEffect(() => {
		if (!chartRef.current) {
			return undefined;
		}

		let cancelled = false;
		let observer: ResizeObserver | null = null;

		async function initializeChart() {
			const echarts = await loadChartRuntime();

			if (cancelled || !chartRef.current) {
				return;
			}

			const chart = echarts.init(chartRef.current, null, { renderer: "canvas" });
			chart.setOption(optionRef.current);
			instanceRef.current = chart;

			observer = new ResizeObserver(() => chart.resize());
			observer.observe(chartRef.current);
		}

		initializeChart();

		return () => {
			cancelled = true;
			observer?.disconnect();
			instanceRef.current?.dispose();
			instanceRef.current = null;
		};
	}, []);

	return <div ref={chartRef} className={[styles.chart, className].filter(Boolean).join(" ")} style={{ height }} />;
}

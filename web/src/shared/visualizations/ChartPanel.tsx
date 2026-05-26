import type { ChartOption } from "@/shared/visualizations/chartOptions";
import { useEffect, useRef, useState } from "react";
import styles from "./ChartPanel.module.css";

export interface ChartTimeRange {
	from: number;
	to: number;
}

interface ChartPanelProps {
	option: ChartOption;
	height?: string;
	className?: string;
	onTimeRangeSelect?: (range: ChartTimeRange) => void;
	timeRangeBounds?: ChartTimeRange;
	minTimeRangeMs?: number;
}

interface ChartInstance {
	dispatchAction: (payload: Record<string, unknown>) => void;
	dispose: () => void;
	off: (eventName: "datazoom", handler: DataZoomEventHandler) => void;
	on: (eventName: "datazoom", handler: DataZoomEventHandler) => void;
	resize: () => void;
	setOption: (option: ChartOption, settings?: { notMerge?: boolean }) => void;
}

interface DataZoomRangePayload {
	end?: number;
	endValue?: number | string;
	start?: number;
	startValue?: number | string;
}

interface DataZoomEvent extends DataZoomRangePayload {
	batch?: DataZoomRangePayload[];
}

type DataZoomEventHandler = (event: DataZoomEvent) => void;

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
				components.ToolboxComponent,
				components.DataZoomComponent,
				components.VisualMapComponent,
				renderers.CanvasRenderer
			]);

			return echarts;
		}
	);

	return chartRuntimePromise;
}

function clampTimestamp(value: number, bounds: ChartTimeRange | undefined) {
	if (!bounds) {
		return value;
	}

	return Math.min(Math.max(value, bounds.from), bounds.to);
}

function numericTimestamp(value: number | string | undefined) {
	if (typeof value === "number" && Number.isFinite(value)) {
		return value;
	}

	if (typeof value !== "string") {
		return null;
	}

	const timestamp = Number(value);
	if (Number.isFinite(timestamp)) {
		return timestamp;
	}

	const parsed = Date.parse(value);
	return Number.isFinite(parsed) ? parsed : null;
}

function rangeFromPercent(payload: DataZoomRangePayload, bounds: ChartTimeRange | undefined) {
	if (!bounds || typeof payload.start !== "number" || typeof payload.end !== "number") {
		return null;
	}

	const span = bounds.to - bounds.from;
	if (span <= 0) {
		return null;
	}

	return {
		from: bounds.from + (span * payload.start) / 100,
		to: bounds.from + (span * payload.end) / 100
	};
}

function isTimestampRange(range: ChartTimeRange, bounds: ChartTimeRange | undefined) {
	if (!bounds) {
		return range.from > 100_000_000_000 && range.to > 100_000_000_000;
	}

	const tolerance = Math.max(bounds.to - bounds.from, 1) * 0.1;
	return range.to >= bounds.from - tolerance && range.from <= bounds.to + tolerance;
}

function rangeFromDataZoomEvent(event: DataZoomEvent, bounds: ChartTimeRange | undefined, minTimeRangeMs: number) {
	const payloads = event.batch?.length ? event.batch : [event];

	for (const payload of payloads) {
		const startValue = numericTimestamp(payload.startValue);
		const endValue = numericTimestamp(payload.endValue);
		const range = startValue !== null && endValue !== null ? { from: startValue, to: endValue } : rangeFromPercent(payload, bounds);

		if (!range || !isTimestampRange(range, bounds)) {
			continue;
		}

		const from = Math.trunc(clampTimestamp(Math.min(range.from, range.to), bounds));
		const to = Math.trunc(clampTimestamp(Math.max(range.from, range.to), bounds));

		if (to - from >= minTimeRangeMs) {
			return { from, to };
		}
	}

	return null;
}

function setDataZoomSelectActive(chart: ChartInstance, active: boolean) {
	chart.dispatchAction({
		type: "takeGlobalCursor",
		key: "dataZoomSelect",
		dataZoomSelectActive: active
	});
}

export function ChartPanel({ option, height = "16rem", className, onTimeRangeSelect, timeRangeBounds, minTimeRangeMs = 1000 }: ChartPanelProps) {
	const rootRef = useRef<HTMLDivElement | null>(null);
	const chartRef = useRef<HTMLDivElement | null>(null);
	const instanceRef = useRef<ChartInstance | null>(null);
	const optionRef = useRef(option);
	const onTimeRangeSelectRef = useRef(onTimeRangeSelect);
	const timeRangeBoundsRef = useRef(timeRangeBounds);
	const minTimeRangeMsRef = useRef(minTimeRangeMs);
	const [chartReadyKey, setChartReadyKey] = useState(0);
	const isTimeRangeSelectable = Boolean(onTimeRangeSelect);

	useEffect(() => {
		optionRef.current = option;
		instanceRef.current?.setOption(option, { notMerge: true });
		if (instanceRef.current && isTimeRangeSelectable) {
			setDataZoomSelectActive(instanceRef.current, true);
		}
	}, [isTimeRangeSelectable, option]);

	useEffect(() => {
		onTimeRangeSelectRef.current = onTimeRangeSelect;
		timeRangeBoundsRef.current = timeRangeBounds;
		minTimeRangeMsRef.current = minTimeRangeMs;
	}, [minTimeRangeMs, onTimeRangeSelect, timeRangeBounds]);

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

			const chart = echarts.init(chartRef.current, null, { renderer: "canvas" }) as unknown as ChartInstance;
			chart.setOption(optionRef.current, { notMerge: true });
			instanceRef.current = chart;
			setChartReadyKey(key => key + 1);

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

	useEffect(() => {
		const chart = instanceRef.current;
		if (!chart || !isTimeRangeSelectable) {
			return undefined;
		}

		function handleDataZoom(event: DataZoomEvent) {
			const nextRange = rangeFromDataZoomEvent(event, timeRangeBoundsRef.current, minTimeRangeMsRef.current);
			if (!nextRange) {
				return;
			}

			onTimeRangeSelectRef.current?.(nextRange);
		}

		chart.on("datazoom", handleDataZoom);
		setDataZoomSelectActive(chart, true);

		return () => {
			chart.off("datazoom", handleDataZoom);
			setDataZoomSelectActive(chart, false);
		};
	}, [chartReadyKey, isTimeRangeSelectable]);

	return (
		<div ref={rootRef} className={[styles.chart, className].filter(Boolean).join(" ")} data-selectable={Boolean(onTimeRangeSelect) || undefined} style={{ height }}>
			<div ref={chartRef} className={styles.chartSurface} />
		</div>
	);
}

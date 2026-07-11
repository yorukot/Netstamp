import type { CheckDefinition } from "@/features/checks/data/checks";
import type { Probe } from "@/features/probes/data/probes";
import type { HttpInsightResponse, HttpSeriesKey, HttpSeriesResponse } from "@/shared/api/types";
import { formatCount, formatMs, formatPercent } from "@/shared/utils/insightFormatters";
import type { ChartOption } from "@/shared/visualizations/chartOptions";
import { ChartPanel } from "@/shared/visualizations/ChartPanel";
import { chartAxisLabel, chartTheme, chartTooltipTextStyle } from "@/shared/visualizations/chartTheme";
import { BodyCopy, Panel, Spinner } from "@netstamp/ui";
import styles from "./PingInsightPanel.module.css";

function points(data: HttpSeriesResponse | undefined, key: HttpSeriesKey) {
	return data?.series[key]?.points ?? [];
}
function chartOption(data: HttpSeriesResponse | undefined): ChartOption {
	const theme = chartTheme();
	const metrics: Array<[HttpSeriesKey, string]> = [
		["total_avg", "Total"],
		["ttfb_avg", "TTFB"],
		["dns_avg", "DNS"],
		["connect_avg", "Connect"],
		["tls_avg", "TLS"]
	];
	return {
		backgroundColor: "transparent",
		color: [theme.primary, theme.secondary, theme.success, theme.warning, theme.metal],
		tooltip: { trigger: "axis", backgroundColor: theme.tooltipBackground, borderColor: theme.tooltipBorder, textStyle: chartTooltipTextStyle() },
		legend: { textStyle: chartAxisLabel() },
		grid: { top: 42, right: 18, bottom: 48, left: 52 },
		xAxis: { type: "time", axisLabel: chartAxisLabel(), axisLine: { lineStyle: { color: theme.axisLine } } },
		yAxis: [{ type: "value", name: "ms", axisLabel: chartAxisLabel(), splitLine: { lineStyle: { color: theme.splitLine } } }],
		dataZoom: [{ type: "inside" }, { type: "slider", height: 18, bottom: 6 }],
		series: metrics.map(([key, name]) => ({ name, type: "line", showSymbol: false, connectNulls: true, data: points(data, key) }))
	};
}

interface Props {
	selectedProbe: Probe | null;
	selectedTarget: CheckDefinition | null;
	insightData?: HttpInsightResponse;
	seriesData?: HttpSeriesResponse;
	isLoading: boolean;
	isFetching: boolean;
	onSelectTimeWindow: (value: { from: number; to: number }) => void;
}
export function HttpInsightPanel({ selectedProbe, selectedTarget, insightData, seriesData, isLoading, isFetching, onSelectTimeWindow }: Props) {
	if (!selectedProbe || !selectedTarget)
		return (
			<Panel tone="deep" title="No HTTP target selected">
				<BodyCopy>Select a probe and HTTP target.</BodyCopy>
			</Panel>
		);
	if (isLoading && !insightData && !seriesData)
		return (
			<Panel tone="deep" title="HTTP insight">
				<Spinner label="Loading HTTP insight" layout="panel" size="lg" />
			</Panel>
		);
	const summary = insightData?.summary;
	const meta = seriesData?.meta ?? insightData?.meta;
	const metrics = [
		{ label: "Average total", value: formatMs(summary?.averageTotalMs), detail: "request total" },
		{ label: "Max total", value: formatMs(summary?.maxTotalMs), detail: "request total" },
		{ label: "Average TTFB", value: formatMs(summary?.averageTtfbMs), detail: "server wait" },
		{ label: "Failure", value: formatPercent(summary?.failurePercent), detail: "timeout + error" },
		{ label: "Success", value: formatPercent(summary?.successRate), detail: "successful" },
		{ label: "Certificate", value: summary?.certificateDaysRemaining == null ? "-" : `${Math.floor(summary.certificateDaysRemaining)}d`, detail: "days remaining" }
	];
	const hasData = points(seriesData, "total_avg").length > 0;
	return (
		<div className={styles.pingStack}>
			<div className={styles.summaryGrid}>
				{metrics.map(metric => (
					<div className={styles.summaryCell} key={metric.label}>
						<span>{metric.label}</span>
						<strong>{metric.value}</strong>
						<small>{metric.detail}</small>
					</div>
				))}
			</div>
			<Panel tone="deep" title={`${selectedProbe.name} -> ${selectedTarget.target}`}>
				<div className={styles.chartMeta}>
					<span>{isFetching ? "syncing result series" : `${formatCount(meta?.totalPoints)} points`}</span>
					<span>{[meta?.source, meta?.resolution].filter(Boolean).join(" / ")}</span>
				</div>
				{hasData ? (
					<ChartPanel option={chartOption(seriesData)} height="27rem" onTimeRangeSelect={onSelectTimeWindow} timeRangeBounds={meta ? { from: meta.from, to: meta.to } : undefined} />
				) : (
					<div className={styles.emptyState}>No HTTP series points were recorded in this time range.</div>
				)}
			</Panel>
		</div>
	);
}

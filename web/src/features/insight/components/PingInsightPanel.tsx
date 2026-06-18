import type { CheckDefinition } from "@/features/checks/data/checks";
import type { Probe } from "@/features/probes/data/probes";
import type { PingInsightResponse, PingSeriesResponse } from "@/shared/api/types";
import { formatCount } from "@/shared/utils/insightFormatters";
import { hasPingSeriesChartData, pingSeriesChartData, pingSummaryMetrics } from "@/shared/utils/pingInsightData";
import { ChartPanel } from "@/shared/visualizations/ChartPanel";
import { pingInsightChartOption } from "@/shared/visualizations/chartOptions";
import { BodyCopy, LoadingState, Panel } from "@netstamp/ui";
import styles from "./PingInsightPanel.module.css";

interface PingInsightPanelProps {
	selectedProbe: Probe | null;
	selectedTarget: CheckDefinition | null;
	insightData: PingInsightResponse | undefined;
	seriesData: PingSeriesResponse | undefined;
	isInsightLoading: boolean;
	isSeriesLoading: boolean;
	isFetching: boolean;
	onSelectTimeWindow: (timeWindow: { from: number; to: number }) => void;
}

export function PingInsightPanel({ selectedProbe, selectedTarget, insightData, seriesData, isInsightLoading, isSeriesLoading, isFetching, onSelectTimeWindow }: PingInsightPanelProps) {
	if (!selectedProbe || !selectedTarget) {
		return (
			<Panel tone="deep" title="No ping target selected">
				<BodyCopy>Select a probe and ping target to inspect latency, packet loss, and samples.</BodyCopy>
			</Panel>
		);
	}

	if ((isInsightLoading || isSeriesLoading || isFetching) && !insightData && !seriesData) {
		return (
			<Panel tone="deep" title="Loading ping insight">
				<LoadingState label="Loading ping insight" detail="Fetching latency, loss, and sample windows for this probe-target pair." />
			</Panel>
		);
	}

	const chartData = pingSeriesChartData(seriesData);
	const metrics = pingSummaryMetrics(insightData);
	const meta = seriesData?.meta ?? insightData?.meta;
	const totalPoints = meta?.totalPoints ?? 0;
	const hasChartData = hasPingSeriesChartData(chartData);
	const queryWindow = meta ? { from: meta.from, to: meta.to } : undefined;
	const sourceLabel = [meta?.source, meta?.resolution].filter(Boolean).join(" / ") || "pending";

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

			<Panel tone="deep" title={`${selectedProbe.name} → ${selectedTarget.target}`}>
				<div className={styles.chartMeta}>
					<span>{isFetching ? "syncing result series" : `${formatCount(totalPoints)} points`}</span>
					<span>{sourceLabel}</span>
				</div>
				{hasChartData ? (
					<ChartPanel option={pingInsightChartOption(chartData)} height="27rem" onTimeRangeSelect={onSelectTimeWindow} timeRangeBounds={queryWindow} />
				) : isSeriesLoading || isFetching ? (
					<LoadingState label="Loading ping series" detail="Syncing result points for the selected time range." />
				) : (
					<div className={styles.emptyState}>No ping series points were recorded for this probe-target pair in the selected time range.</div>
				)}
			</Panel>
		</div>
	);
}

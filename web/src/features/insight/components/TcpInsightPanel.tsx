import type { CheckDefinition } from "@/features/checks/data/checks";
import { hasTcpSeriesChartData, tcpSeriesChartData, tcpSummaryMetrics } from "@/features/insight/data/tcpInsightData";
import type { Probe } from "@/features/probes/data/probes";
import type { TcpInsightResponse, TcpSeriesResponse } from "@/shared/api/types";
import { BodyCopy } from "@/shared/components/BodyCopy";
import { LoadingState } from "@/shared/components/LoadingState";
import { classNames } from "@/shared/utils/classNames";
import { formatCount } from "@/shared/utils/insightFormatters";
import { ChartPanel } from "@/shared/visualizations/ChartPanel";
import { tcpInsightChartOption } from "@/shared/visualizations/chartOptions";
import { Panel } from "@netstamp/ui";
import styles from "./PingInsightPanel.module.css";

interface TcpInsightPanelProps {
	selectedProbe: Probe | null;
	selectedTarget: CheckDefinition | null;
	insightData: TcpInsightResponse | undefined;
	seriesData: TcpSeriesResponse | undefined;
	isInsightLoading: boolean;
	isSeriesLoading: boolean;
	isFetching: boolean;
	onSelectTimeWindow: (timeWindow: { from: number; to: number }) => void;
}

export function TcpInsightPanel({ selectedProbe, selectedTarget, insightData, seriesData, isInsightLoading, isSeriesLoading, isFetching, onSelectTimeWindow }: TcpInsightPanelProps) {
	if (!selectedProbe || !selectedTarget) {
		return (
			<Panel tone="deep" title="No TCP target selected">
				<BodyCopy>Select a probe and TCP target to inspect connect latency and failure windows.</BodyCopy>
			</Panel>
		);
	}

	if ((isInsightLoading || isSeriesLoading || isFetching) && !insightData && !seriesData) {
		return (
			<Panel tone="deep" title="Loading TCP insight">
				<LoadingState label="Loading TCP insight" detail="Fetching connect latency, failures, and sample windows for this probe-target pair." />
			</Panel>
		);
	}

	const chartData = tcpSeriesChartData(seriesData);
	const metrics = tcpSummaryMetrics(insightData);
	const meta = seriesData?.meta ?? insightData?.meta;
	const totalPoints = meta?.totalPoints ?? 0;
	const hasChartData = hasTcpSeriesChartData(chartData);
	const queryWindow = meta ? { from: meta.from, to: meta.to } : undefined;
	const sourceLabel = [meta?.source, meta?.resolution].filter(Boolean).join(" / ") || "pending";

	return (
		<div className={styles.pingStack}>
			<div className={styles.summaryGrid}>
				{metrics.map(metric => (
					<div className={classNames("ns-cut-frame", styles.summaryCell)} key={metric.label}>
						<span>{metric.label}</span>
						<strong>{metric.value}</strong>
						<small>{metric.detail}</small>
					</div>
				))}
			</div>

			<Panel tone="deep" title={`${selectedProbe.name} -> ${selectedTarget.target}`}>
				<div className={styles.chartMeta}>
					<span>{isFetching ? "syncing result series" : `${formatCount(totalPoints)} points`}</span>
					<span>{sourceLabel}</span>
				</div>
				{hasChartData ? (
					<ChartPanel option={tcpInsightChartOption(chartData)} height="27rem" onTimeRangeSelect={onSelectTimeWindow} timeRangeBounds={queryWindow} />
				) : isSeriesLoading || isFetching ? (
					<LoadingState label="Loading TCP series" detail="Syncing result points for the selected time range." />
				) : (
					<div className={styles.emptyState}>No TCP series points were recorded for this probe-target pair in the selected time range.</div>
				)}
			</Panel>
		</div>
	);
}

import type { CheckDefinition } from "@/features/checks/data/checks";
import { tcpChartBuckets, tcpSummaryMetrics } from "@/features/insight/data/tcpInsightData";
import type { Probe } from "@/features/probes/data/probes";
import type { TcpInsightResponse } from "@/shared/api/types";
import { BodyCopy } from "@/shared/components/BodyCopy";
import { classNames } from "@/shared/utils/classNames";
import { formatCount, formatEpochMs } from "@/shared/utils/insightFormatters";
import { ChartPanel } from "@/shared/visualizations/ChartPanel";
import { tcpInsightChartOption } from "@/shared/visualizations/chartOptions";
import { Panel } from "@netstamp/ui";
import styles from "./PingInsightPanel.module.css";

interface TcpInsightPanelProps {
	selectedProbe: Probe | null;
	selectedTarget: CheckDefinition | null;
	data: TcpInsightResponse | undefined;
	isLoading: boolean;
	isFetching: boolean;
	timeLabel: string;
	onSelectTimeWindow: (timeWindow: { from: number; to: number }) => void;
}

export function TcpInsightPanel({ selectedProbe, selectedTarget, data, isLoading, isFetching, timeLabel, onSelectTimeWindow }: TcpInsightPanelProps) {
	if (!selectedProbe || !selectedTarget) {
		return (
			<Panel tone="deep" title="No TCP target selected">
				<BodyCopy>Select a probe and TCP target to inspect connect latency and failure windows.</BodyCopy>
			</Panel>
		);
	}

	if (isLoading && !data) {
		return (
			<Panel tone="deep" title="Loading TCP insight">
				<BodyCopy>Loading TCP result buckets for this probe-target pair.</BodyCopy>
			</Panel>
		);
	}

	const buckets = tcpChartBuckets(data);
	const metrics = tcpSummaryMetrics(data);
	const totalPoints = data?.query.totalPoints ?? 0;
	const hasChartData = buckets.length > 0;
	const queryWindow = data?.query ? { from: data.query.from, to: data.query.to } : undefined;

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
					<span>{isFetching ? "syncing result buckets" : `${formatCount(totalPoints)} results`}</span>
					<span>latest {formatEpochMs(data?.summary.latestStartedAtMs)}</span>
					<span>{data?.summary.latestResolvedIp || "unresolved"}</span>
				</div>
				{hasChartData ? (
					<ChartPanel option={tcpInsightChartOption(buckets)} height="27rem" onTimeRangeSelect={onSelectTimeWindow} timeRangeBounds={queryWindow} />
				) : (
					<div className={styles.emptyState}>No TCP results were recorded for this probe-target pair in the selected time range.</div>
				)}
			</Panel>
		</div>
	);
}

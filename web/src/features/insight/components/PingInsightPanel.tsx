import type { CheckDefinition } from "@/features/checks/data/checks";
import type { Probe } from "@/features/probes/data/probes";
import type { PingInsightResponse } from "@/shared/api/types";
import { BodyCopy } from "@/shared/components/BodyCopy";
import { classNames } from "@/shared/utils/classNames";
import { formatCount, formatEpochMs } from "@/shared/utils/insightFormatters";
import { pingChartBuckets, pingSampleDensity, pingSummaryMetrics } from "@/shared/utils/pingInsightData";
import { ChartPanel } from "@/shared/visualizations/ChartPanel";
import { pingInsightChartOption } from "@/shared/visualizations/chartOptions";
import { Panel } from "@netstamp/ui";
import styles from "./PingInsightPanel.module.css";

interface PingInsightPanelProps {
	selectedProbe: Probe | null;
	selectedTarget: CheckDefinition | null;
	data: PingInsightResponse | undefined;
	isLoading: boolean;
	isFetching: boolean;
	timeLabel: string;
	onSelectTimeWindow: (timeWindow: { from: number; to: number }) => void;
}

export function PingInsightPanel({ selectedProbe, selectedTarget, data, isLoading, isFetching, timeLabel, onSelectTimeWindow }: PingInsightPanelProps) {
	if (!selectedProbe || !selectedTarget) {
		return (
			<Panel tone="deep" eyebrow="Ping" title="No ping target selected">
				<BodyCopy>Select a probe and ping target to inspect latency spread, packet loss, and sample density.</BodyCopy>
			</Panel>
		);
	}

	if (isLoading && !data) {
		return (
			<Panel tone="deep" eyebrow="Ping" title="Loading ping insight">
				<BodyCopy>Loading ping result buckets for this probe-target pair.</BodyCopy>
			</Panel>
		);
	}

	const buckets = pingChartBuckets(data);
	const density = pingSampleDensity(data);
	const metrics = pingSummaryMetrics(data);
	const totalPoints = data?.query.totalPoints ?? 0;
	const hasChartData = buckets.length > 0 || density.length > 0;
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

			<Panel tone="deep" eyebrow={`${timeLabel} · ${data?.query.resolution || "pending"}`} title={`${selectedProbe.name} → ${selectedTarget.target}`}>
				<div className={styles.chartMeta}>
					<span>{isFetching ? "syncing result buckets" : `${formatCount(totalPoints)} results`}</span>
					<span>latest {formatEpochMs(data?.summary.latestStartedAtMs)}</span>
					<span>{data?.summary.latestResolvedIp || "unresolved"}</span>
				</div>
				{hasChartData ? (
					<ChartPanel option={pingInsightChartOption(buckets, density)} height="27rem" onTimeRangeSelect={onSelectTimeWindow} timeRangeBounds={queryWindow} />
				) : (
					<div className={styles.emptyState}>No ping results were recorded for this probe-target pair in the selected time range.</div>
				)}
			</Panel>
		</div>
	);
}

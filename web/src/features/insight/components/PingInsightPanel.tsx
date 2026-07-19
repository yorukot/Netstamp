import type { CheckDefinition } from "@/features/checks/data/checks";
import type { Probe } from "@/features/probes/data/probes";
import type { PingInsightResponse, PingSeriesResponse } from "@/shared/api/types";
import { formatCount, formatMs, formatPercent } from "@/shared/utils/insightFormatters";
import { hasPingSeriesChartData, pingSeriesChartData } from "@/shared/utils/pingInsightData";
import { ChartPanel } from "@/shared/visualizations/ChartPanel";
import { pingInsightChartOption } from "@/shared/visualizations/chartOptions";
import { BodyCopy, Panel, Spinner } from "@netstamp/ui";
import { useTranslation } from "react-i18next";
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
	const { t } = useTranslation("insight");
	if (!selectedProbe || !selectedTarget) {
		return (
			<Panel tone="deep" title={t("panel.noPingTitle")}>
				<BodyCopy>{t("panel.noPingDescription")}</BodyCopy>
			</Panel>
		);
	}

	if ((isInsightLoading || isSeriesLoading || isFetching) && !insightData && !seriesData) {
		return (
			<Panel tone="deep" title={t("panel.pingTitle")}>
				<Spinner label={t("panel.loadingPing")} layout="panel" size="lg" />
			</Panel>
		);
	}

	const chartData = pingSeriesChartData(seriesData);
	const summary = insightData?.summary;
	const metrics = [
		{ label: t("panel.average"), value: formatMs(summary?.averageRttMs), detail: t("panel.rttAverage") },
		{ label: t("panel.maximum"), value: formatMs(summary?.maxRttMs), detail: t("panel.rttMaximum") },
		{ label: t("panel.loss"), value: formatPercent(summary?.lossPercent), detail: t("panel.average") },
		{ label: t("panel.success"), value: formatPercent(summary?.successRate), detail: t("panel.successful") },
		{ label: t("panel.samples"), value: formatCount(summary?.samples), detail: t("panel.replies") }
	];
	const meta = seriesData?.meta ?? insightData?.meta;
	const totalPoints = meta?.totalPoints ?? 0;
	const hasChartData = hasPingSeriesChartData(chartData);
	const queryWindow = meta ? { from: meta.from, to: meta.to } : undefined;
	const sourceLabel = [meta?.source, meta?.resolution].filter(Boolean).join(" / ") || t("panel.pending");

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
					<span>{isFetching ? t("panel.syncingSeries") : t("panel.points", { count: totalPoints })}</span>
					<span>{sourceLabel}</span>
				</div>
				{hasChartData ? (
					<ChartPanel option={pingInsightChartOption(chartData)} height="27rem" onTimeRangeSelect={onSelectTimeWindow} timeRangeBounds={queryWindow} />
				) : isSeriesLoading || isFetching ? (
					<Spinner label={t("panel.loadingPingSeries")} layout="panel" size="lg" />
				) : (
					<div className={styles.emptyState}>{t("panel.noPingSeries")}</div>
				)}
			</Panel>
		</div>
	);
}

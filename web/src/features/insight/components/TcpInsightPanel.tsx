import type { CheckDefinition } from "@/features/checks/data/checks";
import { hasTcpSeriesChartData, tcpSeriesChartData } from "@/features/insight/data/tcpInsightData";
import type { Probe } from "@/features/probes/data/probes";
import type { TcpInsightResponse, TcpSeriesResponse } from "@/shared/api/types";
import { formatCount, formatMs, formatPercent } from "@/shared/utils/insightFormatters";
import { ChartPanel } from "@/shared/visualizations/ChartPanel";
import { tcpInsightChartOption } from "@/shared/visualizations/chartOptions";
import { BodyCopy, Panel, Spinner } from "@netstamp/ui";
import { useTranslation } from "react-i18next";
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
	const { t } = useTranslation("insight");
	if (!selectedProbe || !selectedTarget) {
		return (
			<Panel tone="deep" title={t("panel.noTcpTitle")}>
				<BodyCopy>{t("panel.noTcpDescription")}</BodyCopy>
			</Panel>
		);
	}

	if ((isInsightLoading || isSeriesLoading || isFetching) && !insightData && !seriesData) {
		return (
			<Panel tone="deep" title={t("panel.tcpTitle")}>
				<Spinner label={t("panel.loadingTcp")} layout="panel" size="lg" />
			</Panel>
		);
	}

	const chartData = tcpSeriesChartData(seriesData);
	const summary = insightData?.summary;
	const metrics = [
		{ label: t("panel.average"), value: formatMs(summary?.averageConnectMs), detail: t("panel.connectAverage") },
		{ label: t("panel.maximum"), value: formatMs(summary?.maxConnectMs), detail: t("panel.connectMaximum") },
		{ label: t("panel.failure"), value: formatPercent(summary?.failurePercent), detail: t("panel.timeoutError") },
		{ label: t("panel.success"), value: formatPercent(summary?.successRate), detail: t("panel.successful") },
		{ label: t("panel.samples"), value: formatCount(summary?.samples), detail: t("panel.connectAttempts") }
	];
	const meta = seriesData?.meta ?? insightData?.meta;
	const totalPoints = meta?.totalPoints ?? 0;
	const hasChartData = hasTcpSeriesChartData(chartData);
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

			<Panel tone="deep" title={`${selectedProbe.name} -> ${selectedTarget.target}`}>
				<div className={styles.chartMeta}>
					<span>{isFetching ? t("panel.syncingSeries") : t("panel.points", { count: totalPoints })}</span>
					<span>{sourceLabel}</span>
				</div>
				{hasChartData ? (
					<ChartPanel option={tcpInsightChartOption(chartData)} height="27rem" onTimeRangeSelect={onSelectTimeWindow} timeRangeBounds={queryWindow} />
				) : isSeriesLoading || isFetching ? (
					<Spinner label={t("panel.loadingTcpSeries")} layout="panel" size="lg" />
				) : (
					<div className={styles.emptyState}>{t("panel.noTcpSeries")}</div>
				)}
			</Panel>
		</div>
	);
}

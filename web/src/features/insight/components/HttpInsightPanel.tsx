import type { CheckDefinition } from "@/features/checks/data/checks";
import { HttpLatestResultPanel } from "@/features/insight/components/HttpLatestResultPanel";
import type { Probe } from "@/features/probes/data/probes";
import { i18n } from "@/i18n";
import type { HttpInsightResponse, HttpSeriesKey, HttpSeriesResponse, LatestHttpResult } from "@/shared/api/types";
import { formatMs, formatPercent } from "@/shared/utils/insightFormatters";
import type { ChartOption } from "@/shared/visualizations/chartOptions";
import { ChartPanel } from "@/shared/visualizations/ChartPanel";
import { chartAxisLabel, chartTheme, chartTooltipTextStyle } from "@/shared/visualizations/chartTheme";
import { BodyCopy, Panel, Spinner } from "@netstamp/ui";
import { useTranslation } from "react-i18next";
import styles from "./PingInsightPanel.module.css";

const insightT = i18n.getFixedT(null, "insight") as (key: string) => string;

function points(data: HttpSeriesResponse | undefined, key: HttpSeriesKey) {
	return data?.series[key]?.points ?? [];
}
function chartOption(data: HttpSeriesResponse | undefined): ChartOption {
	const theme = chartTheme();
	const metrics: Array<[HttpSeriesKey, string]> = [
		["total_avg", insightT("http.total")],
		["ttfb_avg", "TTFB"],
		["dns_avg", "DNS"],
		["connect_avg", insightT("http.connect")],
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
	latestResult?: LatestHttpResult;
	nowMs: number;
	isLoading: boolean;
	isLatestLoading: boolean;
	isFetching: boolean;
	onSelectTimeWindow: (value: { from: number; to: number }) => void;
}
export function HttpInsightPanel({ selectedProbe, selectedTarget, insightData, seriesData, latestResult, nowMs, isLoading, isLatestLoading, isFetching, onSelectTimeWindow }: Props) {
	const { t } = useTranslation("insight");
	if (!selectedProbe || !selectedTarget)
		return (
			<Panel tone="deep" title={t("panel.noHttpTitle")}>
				<BodyCopy>{t("panel.noHttpDescription")}</BodyCopy>
			</Panel>
		);
	if (isLoading && !insightData && !seriesData)
		return (
			<Panel tone="deep" title={t("panel.httpTitle")}>
				<Spinner label={t("panel.loadingHttp")} layout="panel" size="lg" />
			</Panel>
		);
	const summary = insightData?.summary;
	const meta = seriesData?.meta ?? insightData?.meta;
	const metrics = [
		{ label: t("panel.averageTotal"), value: formatMs(summary?.averageTotalMs), detail: t("panel.requestTotal") },
		{ label: t("panel.maxTotal"), value: formatMs(summary?.maxTotalMs), detail: t("panel.requestTotal") },
		{ label: t("panel.averageTtfb"), value: formatMs(summary?.averageTtfbMs), detail: t("panel.serverWait") },
		{ label: t("panel.failure"), value: formatPercent(summary?.failurePercent), detail: t("panel.timeoutError") },
		{ label: t("panel.success"), value: formatPercent(summary?.successRate), detail: t("panel.successful") },
		{
			label: t("panel.certificateFloor"),
			value: summary?.certificateDaysRemaining == null ? "-" : t("http.daysShort", { count: Math.floor(summary.certificateDaysRemaining) }),
			detail: t("panel.minimumInRange")
		}
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
			<HttpLatestResultPanel latestResult={latestResult} target={selectedTarget.target} nowMs={nowMs} isLoading={isLatestLoading} isFetching={isFetching} />
			<Panel tone="deep" title={`${selectedProbe.name} -> ${selectedTarget.target}`}>
				<div className={styles.chartMeta}>
					<span>{isFetching ? t("panel.syncingSeries") : t("panel.points", { count: meta?.totalPoints ?? 0 })}</span>
					<span>{[meta?.source, meta?.resolution].filter(Boolean).join(" / ")}</span>
				</div>
				{hasData ? (
					<ChartPanel option={chartOption(seriesData)} height="27rem" onTimeRangeSelect={onSelectTimeWindow} timeRangeBounds={meta ? { from: meta.from, to: meta.to } : undefined} />
				) : (
					<div className={styles.emptyState}>{t("panel.noHttpSeries")}</div>
				)}
			</Panel>
		</div>
	);
}

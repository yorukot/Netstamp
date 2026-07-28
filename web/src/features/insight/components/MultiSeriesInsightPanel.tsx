import { HttpCertificateInventory } from "@/features/insight/components/HttpCertificateInventory";
import { hasTcpSeriesChartData, tcpSeriesChartData } from "@/features/insight/data/tcpInsightData";
import type { InsightPair, TimeWindow } from "@/features/insight/insightTypes";
import { projectQueries } from "@/shared/api/queries";
import type { HttpSeriesResponse, LatestHttpResult, PingSeriesResponse, TcpSeriesResponse } from "@/shared/api/types";
import { hasPingSeriesChartData, pingSeriesChartData } from "@/shared/utils/pingInsightData";
import { ChartPanel } from "@/shared/visualizations/ChartPanel";
import { multiHttpInsightChartOption, multiPingInsightChartOption, multiTcpInsightChartOption, type ChartOption, type InsightMultiSeriesLine } from "@/shared/visualizations/chartOptions";
import type { ChartTheme } from "@/shared/visualizations/chartTheme";
import { Panel, Spinner } from "@netstamp/ui";
import { useQueries } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import styles from "./MultiSeriesInsightPanel.module.css";

interface MultiSeriesInsightPanelProps {
	projectRef: string | null | undefined;
	pairs: InsightPair[];
	filters: TimeWindow;
	latestHTTPResults: LatestHttpResult[] | undefined;
	nowMs: number;
	isLatestHTTPLoading: boolean;
	isLatestHTTPFetching: boolean;
	onSelectTimeWindow: (timeWindow: TimeWindow) => void;
}

interface LegendLine extends InsightMultiSeriesLine {
	meta: string;
}

type MultiSeriesOptionBuilder = (lines: InsightMultiSeriesLine[], theme: ChartTheme) => ChartOption;

function pairLabel(pair: InsightPair) {
	return `${pair.probe.name} -> ${pair.check.target}`;
}

function pairMeta(pair: InsightPair) {
	return `${pair.check.name} · ${pair.probe.location}`;
}

function queryWindow(data: Array<PingSeriesResponse | TcpSeriesResponse | HttpSeriesResponse | undefined>, fallback: TimeWindow) {
	const meta = data.find(item => item?.meta)?.meta;
	return meta ? { from: meta.from, to: meta.to } : fallback;
}

function SeriesLegend({ lines }: { lines: LegendLine[] }) {
	const { t } = useTranslation("insight");
	return (
		<div className={styles.legend} aria-label={t("legend.series")}>
			{lines.map((line, index) => (
				<div className={styles.legendItem} key={line.id}>
					<span className={styles.legendSwatch} data-color-index={index % 8} aria-hidden="true" />
					<span className={styles.legendText}>
						<strong>{line.name}</strong>
						<small>{line.meta}</small>
					</span>
				</div>
			))}
		</div>
	);
}

function SeriesPanel({
	title,
	typeLabel,
	totalPoints,
	isLoading,
	isFetching,
	hasData,
	lines,
	getOption,
	filters,
	onSelectTimeWindow
}: {
	title: string;
	typeLabel: string;
	totalPoints: number;
	isLoading: boolean;
	isFetching: boolean;
	hasData: boolean;
	lines: LegendLine[];
	getOption: MultiSeriesOptionBuilder;
	filters: TimeWindow;
	onSelectTimeWindow: (timeWindow: TimeWindow) => void;
}) {
	const { t } = useTranslation("insight");

	return (
		<Panel tone="deep" title={title}>
			<div className={styles.chartMeta}>
				<span>{isFetching ? t("panel.syncingSeries") : t("panel.points", { count: totalPoints })}</span>
				<span>{t("multi.seriesCount", { count: lines.length, type: typeLabel })}</span>
			</div>
			{hasData ? (
				<>
					<ChartPanel getOption={theme => getOption(lines, theme)} height="27rem" onTimeRangeSelect={onSelectTimeWindow} timeRangeBounds={filters} />
					<SeriesLegend lines={lines} />
				</>
			) : (
				<>
					{isLoading || isFetching ? (
						<Spinner label={t("multi.loading", { type: typeLabel })} layout="panel" size="lg" />
					) : (
						<div className={styles.emptyState}>{t("multi.empty", { type: typeLabel })}</div>
					)}
				</>
			)}
		</Panel>
	);
}

export function MultiSeriesInsightPanel({ projectRef, pairs, filters, latestHTTPResults, nowMs, isLatestHTTPLoading, isLatestHTTPFetching, onSelectTimeWindow }: MultiSeriesInsightPanelProps) {
	const { t } = useTranslation("insight");
	const pingPairs = pairs.filter(pair => pair.check.type === "Ping");
	const tcpPairs = pairs.filter(pair => pair.check.type === "TCP");
	const httpPairs = pairs.filter(pair => pair.check.type === "HTTP");
	const pingSeriesQueries = useQueries({
		queries: pingPairs.map(pair => ({
			...projectQueries.pingSeries(projectRef || "", pair.probeId, pair.checkId, filters),
			enabled: Boolean(projectRef)
		}))
	});
	const tcpSeriesQueries = useQueries({
		queries: tcpPairs.map(pair => ({
			...projectQueries.tcpSeries(projectRef || "", pair.probeId, pair.checkId, filters),
			enabled: Boolean(projectRef)
		}))
	});
	const httpSeriesQueries = useQueries({ queries: httpPairs.map(pair => ({ ...projectQueries.httpSeries(projectRef || "", pair.probeId, pair.checkId, filters), enabled: Boolean(projectRef) })) });
	const pingLines: LegendLine[] = pingPairs.map((pair, index) => {
		const data = pingSeriesChartData(pingSeriesQueries[index]?.data);

		return {
			id: pair.key,
			name: pairLabel(pair),
			meta: pairMeta(pair),
			points: data.latencyAvg
		};
	});
	const tcpLines: LegendLine[] = tcpPairs.map((pair, index) => {
		const data = tcpSeriesChartData(tcpSeriesQueries[index]?.data);

		return {
			id: pair.key,
			name: pairLabel(pair),
			meta: pairMeta(pair),
			points: data.connectAvg
		};
	});
	const httpLines: LegendLine[] = httpPairs.map((pair, index) => ({
		id: pair.key,
		name: pairLabel(pair),
		meta: pairMeta(pair),
		points: httpSeriesQueries[index]?.data?.series.total_avg?.points ?? []
	}));
	const hasPingData = pingSeriesQueries.some(query => hasPingSeriesChartData(pingSeriesChartData(query.data)));
	const hasTcpData = tcpSeriesQueries.some(query => hasTcpSeriesChartData(tcpSeriesChartData(query.data)));
	const hasHTTPData = httpLines.some(line => line.points.length > 0);
	const pingTotalPoints = pingSeriesQueries.reduce((total, query) => total + (query.data?.meta.totalPoints ?? 0), 0);
	const tcpTotalPoints = tcpSeriesQueries.reduce((total, query) => total + (query.data?.meta.totalPoints ?? 0), 0);
	const httpTotalPoints = httpSeriesQueries.reduce((total, query) => total + (query.data?.meta.totalPoints ?? 0), 0);
	const pingWindow = queryWindow(
		pingSeriesQueries.map(query => query.data),
		filters
	);
	const tcpWindow = queryWindow(
		tcpSeriesQueries.map(query => query.data),
		filters
	);
	const httpWindow = queryWindow(
		httpSeriesQueries.map(query => query.data),
		filters
	);

	if (!pingPairs.length && !tcpPairs.length && !httpPairs.length) {
		return null;
	}

	return (
		<div className={styles.stack}>
			{httpPairs.length ? <HttpCertificateInventory pairs={httpPairs} latestResults={latestHTTPResults} nowMs={nowMs} isLoading={isLatestHTTPLoading} isFetching={isLatestHTTPFetching} /> : null}
			{pingPairs.length ? (
				<SeriesPanel
					title={t("multi.title", { type: "Ping", count: pingPairs.length })}
					typeLabel="Ping"
					totalPoints={pingTotalPoints}
					isLoading={pingSeriesQueries.some(query => query.isLoading)}
					isFetching={pingSeriesQueries.some(query => query.isFetching)}
					hasData={hasPingData}
					lines={pingLines}
					getOption={multiPingInsightChartOption}
					filters={pingWindow}
					onSelectTimeWindow={onSelectTimeWindow}
				/>
			) : null}
			{tcpPairs.length ? (
				<SeriesPanel
					title={t("multi.title", { type: "TCP", count: tcpPairs.length })}
					typeLabel="TCP"
					totalPoints={tcpTotalPoints}
					isLoading={tcpSeriesQueries.some(query => query.isLoading)}
					isFetching={tcpSeriesQueries.some(query => query.isFetching)}
					hasData={hasTcpData}
					lines={tcpLines}
					getOption={multiTcpInsightChartOption}
					filters={tcpWindow}
					onSelectTimeWindow={onSelectTimeWindow}
				/>
			) : null}
			{httpPairs.length ? (
				<SeriesPanel
					title={t("multi.title", { type: "HTTP", count: httpPairs.length })}
					typeLabel="HTTP"
					totalPoints={httpTotalPoints}
					isLoading={httpSeriesQueries.some(query => query.isLoading)}
					isFetching={httpSeriesQueries.some(query => query.isFetching)}
					hasData={hasHTTPData}
					lines={httpLines}
					getOption={multiHttpInsightChartOption}
					filters={httpWindow}
					onSelectTimeWindow={onSelectTimeWindow}
				/>
			) : null}
		</div>
	);
}

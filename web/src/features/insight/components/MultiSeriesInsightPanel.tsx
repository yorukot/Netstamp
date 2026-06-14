import { hasTcpSeriesChartData, tcpSeriesChartData } from "@/features/insight/data/tcpInsightData";
import type { InsightPair, TimeWindow } from "@/features/insight/insightTypes";
import { projectQueries } from "@/shared/api/queries";
import type { PingSeriesResponse, TcpSeriesResponse } from "@/shared/api/types";
import { LoadingState } from "@/shared/components/LoadingState";
import { formatCount } from "@/shared/utils/insightFormatters";
import { hasPingSeriesChartData, pingSeriesChartData } from "@/shared/utils/pingInsightData";
import { ChartPanel } from "@/shared/visualizations/ChartPanel";
import { insightSeriesColor, multiPingInsightChartOption, multiTcpInsightChartOption, type InsightMultiSeriesLine } from "@/shared/visualizations/chartOptions";
import { Panel } from "@netstamp/ui";
import { useQueries } from "@tanstack/react-query";
import styles from "./MultiSeriesInsightPanel.module.css";

interface MultiSeriesInsightPanelProps {
	projectRef: string | null | undefined;
	pairs: InsightPair[];
	filters: TimeWindow;
	onSelectTimeWindow: (timeWindow: TimeWindow) => void;
}

interface LegendLine extends InsightMultiSeriesLine {
	meta: string;
}

function pairLabel(pair: InsightPair) {
	return `${pair.probe.name} -> ${pair.check.target}`;
}

function pairMeta(pair: InsightPair) {
	return `${pair.check.name} · ${pair.probe.location}`;
}

function queryWindow(data: Array<PingSeriesResponse | TcpSeriesResponse | undefined>, fallback: TimeWindow) {
	const meta = data.find(item => item?.meta)?.meta;
	return meta ? { from: meta.from, to: meta.to } : fallback;
}

function SeriesLegend({ lines }: { lines: LegendLine[] }) {
	return (
		<div className={styles.legend} aria-label="Series color legend">
			{lines.map(line => (
				<div className={styles.legendItem} key={line.id}>
					<span className={styles.legendSwatch} style={{ backgroundColor: line.color }} aria-hidden="true" />
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
	unitLabel,
	totalPoints,
	isLoading,
	isFetching,
	hasData,
	lines,
	option,
	filters,
	onSelectTimeWindow
}: {
	title: string;
	unitLabel: string;
	totalPoints: number;
	isLoading: boolean;
	isFetching: boolean;
	hasData: boolean;
	lines: LegendLine[];
	option: ReturnType<typeof multiPingInsightChartOption>;
	filters: TimeWindow;
	onSelectTimeWindow: (timeWindow: TimeWindow) => void;
}) {
	return (
		<Panel tone="deep" title={title}>
			<div className={styles.chartMeta}>
				<span>{isFetching ? "syncing result series" : `${formatCount(totalPoints)} points`}</span>
				<span>
					{formatCount(lines.length)} {unitLabel} series
				</span>
			</div>
			{hasData ? (
				<>
					<ChartPanel option={option} height="27rem" onTimeRangeSelect={onSelectTimeWindow} timeRangeBounds={filters} />
					<SeriesLegend lines={lines} />
				</>
			) : (
				<>
					{isLoading || isFetching ? (
						<LoadingState label={`Loading ${unitLabel} series`} detail="Syncing result points for the selected assignments and time range." />
					) : (
						<div className={styles.emptyState}>No {unitLabel} series points were recorded for the selected assignments in this time range.</div>
					)}
				</>
			)}
		</Panel>
	);
}

export function MultiSeriesInsightPanel({ projectRef, pairs, filters, onSelectTimeWindow }: MultiSeriesInsightPanelProps) {
	const pingPairs = pairs.filter(pair => pair.check.type === "Ping");
	const tcpPairs = pairs.filter(pair => pair.check.type === "TCP");
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
	const pingLines: LegendLine[] = pingPairs.map((pair, index) => {
		const data = pingSeriesChartData(pingSeriesQueries[index]?.data);

		return {
			id: pair.key,
			name: pairLabel(pair),
			meta: pairMeta(pair),
			color: insightSeriesColor(index),
			points: data.latencyAvg
		};
	});
	const tcpLines: LegendLine[] = tcpPairs.map((pair, index) => {
		const data = tcpSeriesChartData(tcpSeriesQueries[index]?.data);

		return {
			id: pair.key,
			name: pairLabel(pair),
			meta: pairMeta(pair),
			color: insightSeriesColor(index),
			points: data.connectAvg
		};
	});
	const hasPingData = pingSeriesQueries.some(query => hasPingSeriesChartData(pingSeriesChartData(query.data)));
	const hasTcpData = tcpSeriesQueries.some(query => hasTcpSeriesChartData(tcpSeriesChartData(query.data)));
	const pingTotalPoints = pingSeriesQueries.reduce((total, query) => total + (query.data?.meta.totalPoints ?? 0), 0);
	const tcpTotalPoints = tcpSeriesQueries.reduce((total, query) => total + (query.data?.meta.totalPoints ?? 0), 0);
	const pingWindow = queryWindow(
		pingSeriesQueries.map(query => query.data),
		filters
	);
	const tcpWindow = queryWindow(
		tcpSeriesQueries.map(query => query.data),
		filters
	);

	if (!pingPairs.length && !tcpPairs.length) {
		return null;
	}

	return (
		<div className={styles.stack}>
			{pingPairs.length ? (
				<SeriesPanel
					title={`Ping series (${formatCount(pingPairs.length)} assignments)`}
					unitLabel="ping"
					totalPoints={pingTotalPoints}
					isLoading={pingSeriesQueries.some(query => query.isLoading)}
					isFetching={pingSeriesQueries.some(query => query.isFetching)}
					hasData={hasPingData}
					lines={pingLines}
					option={multiPingInsightChartOption(pingLines)}
					filters={pingWindow}
					onSelectTimeWindow={onSelectTimeWindow}
				/>
			) : null}
			{tcpPairs.length ? (
				<SeriesPanel
					title={`TCP series (${formatCount(tcpPairs.length)} assignments)`}
					unitLabel="TCP"
					totalPoints={tcpTotalPoints}
					isLoading={tcpSeriesQueries.some(query => query.isLoading)}
					isFetching={tcpSeriesQueries.some(query => query.isFetching)}
					hasData={hasTcpData}
					lines={tcpLines}
					option={multiTcpInsightChartOption(tcpLines)}
					filters={tcpWindow}
					onSelectTimeWindow={onSelectTimeWindow}
				/>
			) : null}
		</div>
	);
}

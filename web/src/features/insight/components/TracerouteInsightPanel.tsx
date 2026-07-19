import type { CheckDefinition } from "@/features/checks/data/checks";
import { buildHopDiagnostics, selectedTimelineValueLabel, tracerouteInsightTimelinePoints } from "@/features/insight/data/tracerouteInsightData";
import type { HopDiagnostic, TimeWindow } from "@/features/insight/insightTypes";
import type { Probe } from "@/features/probes/data/probes";
import type { TracerouteInsightResponse, TracerouteResult } from "@/shared/api/types";
import { formatMs, formatPercent } from "@/shared/utils/insightFormatters";
import { LatencyRail } from "@/shared/visualizations/LatencyRail";
import { RouteTopologyMap, type RouteTopologyEdge, type RouteTopologyNode } from "@/shared/visualizations/RouteTopologyMap";
import { RunTimeline } from "@/shared/visualizations/RunTimeline";
import { Badge, BodyCopy, DataTable, Panel, Spinner, type DataColumn } from "@netstamp/ui";
import type { TFunction } from "i18next";
import { useTranslation } from "react-i18next";
import styles from "./TracerouteInsightPanel.module.css";

function hopColumns(maxRtt: number, t: TFunction<"insight">): DataColumn<HopDiagnostic>[] {
	return [
		{ key: "hopIndex", label: t("panel.hop"), render: row => String(row.hopIndex).padStart(2, "0") },
		{
			key: "label",
			label: t("panel.node"),
			render: row => (
				<span className={styles.hopIdentity}>
					<strong>{row.label}</strong>
					{row.address !== row.label ? <span>{row.address}</span> : null}
				</span>
			)
		},
		{ key: "loss", label: t("panel.loss"), render: row => formatPercent(row.loss) },
		{
			key: "latency",
			label: t("panel.latency"),
			render: row => <LatencyRail minValue={row.minRtt} avgValue={row.avgRtt} maxValue={row.maxRtt} scaleMax={maxRtt} valueLabel={formatMs(row.avgRtt)} tone={row.tone} />
		},
		{ key: "medianRtt", label: t("panel.median"), render: row => formatMs(row.medianRtt) },
		{ key: "range", label: t("panel.range"), render: row => `${formatMs(row.minRtt)} / ${formatMs(row.maxRtt)}` },
		{ key: "sent", label: t("panel.sentReceived"), render: row => `${row.sent}/${row.received}` },
		{ key: "state", label: t("panel.state"), render: row => <Badge tone={row.tone}>{row.state}</Badge> }
	];
}

interface TracerouteInsightPanelProps {
	selectedProbe: Probe | null;
	selectedTarget: CheckDefinition | null;
	insight: TracerouteInsightResponse | undefined;
	runs: TracerouteResult[];
	topologyNodes: RouteTopologyNode[];
	topologyEdges: RouteTopologyEdge[];
	isInsightLoading: boolean;
	isRunsLoading: boolean;
	isTopologyLoading: boolean;
	selectedRunStartedAt: string;
	showTopology?: boolean;
	onSelectRun: (startedAt: string) => void;
	onSelectTimeWindow: (timeWindow: TimeWindow) => void;
}

export function TracerouteInsightPanel({
	selectedProbe,
	selectedTarget,
	insight,
	runs,
	topologyNodes,
	topologyEdges,
	isInsightLoading,
	isRunsLoading,
	isTopologyLoading,
	selectedRunStartedAt,
	showTopology = true,
	onSelectRun,
	onSelectTimeWindow
}: TracerouteInsightPanelProps) {
	const { t } = useTranslation("insight");
	const selectedRun = runs.find(run => run.startedAt === selectedRunStartedAt) || runs[0] || null;
	const diagnostics = buildHopDiagnostics(selectedRun);
	const timelinePoints = tracerouteInsightTimelinePoints(insight);
	const timeRangeBounds = insight?.query ? { from: insight.query.from, to: insight.query.to } : undefined;
	const hasTopology = topologyNodes.length > 0 && topologyEdges.length > 0;
	const totalRuns = insight?.query.totalRuns ?? runs.length;
	const isRouteLoading = isRunsLoading || isInsightLoading;

	function handleSelectTimelinePoint(point: (typeof timelinePoints)[number]) {
		if (point.runStartedAt) {
			onSelectRun(point.runStartedAt);
			return;
		}

		if (typeof point.rangeFromMs === "number" && typeof point.rangeToMs === "number" && point.rangeToMs > point.rangeFromMs) {
			onSelectTimeWindow({ from: point.rangeFromMs, to: point.rangeToMs });
		}
	}

	if (!selectedProbe || !selectedTarget) {
		return (
			<Panel tone="deep" title={t("panel.noRouteTitle")}>
				<BodyCopy>{t("panel.noRouteDescription")}</BodyCopy>
			</Panel>
		);
	}

	if (isRouteLoading && !runs.length && !timelinePoints.length) {
		return (
			<Panel tone="deep" title={t("panel.route")}>
				<Spinner label={t("panel.loadingRoute")} layout="panel" size="lg" />
			</Panel>
		);
	}

	if (!runs.length && !timelinePoints.length) {
		return (
			<Panel tone="deep" title={t("panel.noRunsTitle")}>
				<BodyCopy>{t("panel.noRunsDescription")}</BodyCopy>
			</Panel>
		);
	}

	return (
		<div className={styles.tracerouteStack}>
			<Panel className={styles.tracePanel} tone="deep" title={t("panel.traceTitle")}>
				<div className={styles.traceStack}>
					{diagnostics.length ? (
						<DataTable
							columns={hopColumns(Math.max(1, ...diagnostics.map(hop => hop.maxRtt ?? hop.avgRtt ?? 0)), t)}
							rows={diagnostics}
							density="compact"
							minWidth="68rem"
							maxHeight="28rem"
							getRowKey={row => row.id}
							emptyLabel={t("panel.noHopData")}
						/>
					) : (
						<BodyCopy>{t("panel.noHopRows")}</BodyCopy>
					)}
					<div className={styles.traceTimeline}>
						<div className={styles.traceTimelineHeader}>
							<span>{t("panel.runTimeline")}</span>
							<strong className="ns-title">{t("panel.runsInWindow", { count: totalRuns })}</strong>
						</div>
						<RunTimeline
							points={timelinePoints}
							selectedPointId={selectedRun?.startedAt}
							selectedValueLabel={selectedTimelineValueLabel(selectedRun, timelinePoints)}
							timeRangeBounds={timeRangeBounds}
							emptyState={isRouteLoading ? <Spinner label={t("panel.loadingTimeline")} layout="compact" size="lg" /> : <BodyCopy>{t("panel.noRunsInRange")}</BodyCopy>}
							onSelectPoint={handleSelectTimelinePoint}
							onSelectTimeRange={onSelectTimeWindow}
						/>
					</div>
				</div>
			</Panel>

			{showTopology ? (
				<Panel tone="deep" title={t("panel.routeGraph")}>
					{hasTopology ? (
						<RouteTopologyMap nodes={topologyNodes} edges={topologyEdges} />
					) : isTopologyLoading ? (
						<Spinner label={t("panel.loadingTopology")} layout="panel" size="lg" />
					) : (
						<BodyCopy>{t("panel.topologyUnavailable")}</BodyCopy>
					)}
				</Panel>
			) : null}
		</div>
	);
}

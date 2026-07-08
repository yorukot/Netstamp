import type { CheckDefinition } from "@/features/checks/data/checks";
import { buildHopDiagnostics, selectedTimelineValueLabel, tracerouteInsightTimelinePoints } from "@/features/insight/data/tracerouteInsightData";
import type { HopDiagnostic, TimeWindow } from "@/features/insight/insightTypes";
import type { Probe } from "@/features/probes/data/probes";
import type { TracerouteInsightResponse, TracerouteResult } from "@/shared/api/types";
import { formatMs, formatPercent } from "@/shared/utils/insightFormatters";
import { LatencyRail } from "@/shared/visualizations/LatencyRail";
import { RouteTopologyMap, type RouteTopologyEdge, type RouteTopologyNode } from "@/shared/visualizations/RouteTopologyMap";
import { RunTimeline } from "@/shared/visualizations/RunTimeline";
import { Badge, BodyCopy, DataTable, LoadingState, Panel, type DataColumn } from "@netstamp/ui";
import styles from "./TracerouteInsightPanel.module.css";

function hopColumns(maxRtt: number): DataColumn<HopDiagnostic>[] {
	return [
		{ key: "hopIndex", label: "Hop", render: row => String(row.hopIndex).padStart(2, "0") },
		{
			key: "label",
			label: "Node",
			render: row => (
				<span className={styles.hopIdentity}>
					<strong>{row.label}</strong>
					{row.address !== row.label ? <span>{row.address}</span> : null}
				</span>
			)
		},
		{ key: "loss", label: "Loss", render: row => formatPercent(row.loss) },
		{
			key: "latency",
			label: "Latency",
			render: row => <LatencyRail minValue={row.minRtt} avgValue={row.avgRtt} maxValue={row.maxRtt} scaleMax={maxRtt} valueLabel={formatMs(row.avgRtt)} tone={row.tone} />
		},
		{ key: "medianRtt", label: "Median", render: row => formatMs(row.medianRtt) },
		{ key: "range", label: "Range", render: row => `${formatMs(row.minRtt)} / ${formatMs(row.maxRtt)}` },
		{ key: "sent", label: "Sent/Recv", render: row => `${row.sent}/${row.received}` },
		{ key: "state", label: "State", render: row => <Badge tone={row.tone}>{row.state}</Badge> }
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
			<Panel tone="deep" title="No route selected">
				<BodyCopy>Select a probe and traceroute target to inspect route details.</BodyCopy>
			</Panel>
		);
	}

	if (isRouteLoading && !runs.length && !timelinePoints.length) {
		return (
			<Panel tone="deep" title="Loading route">
				<LoadingState label="Loading route" detail="Fetching traceroute runs, hops, and route timeline for this probe-target pair." />
			</Panel>
		);
	}

	if (!runs.length && !timelinePoints.length) {
		return (
			<Panel tone="deep" title="No traceroute runs">
				<BodyCopy>No traceroute results were recorded for this probe-target pair in the selected time range.</BodyCopy>
			</Panel>
		);
	}

	return (
		<div className={styles.tracerouteStack}>
			<Panel className={styles.tracePanel} tone="deep" title="Hop latency, loss, and run timeline">
				<div className={styles.traceStack}>
					{diagnostics.length ? (
						<DataTable
							columns={hopColumns(Math.max(1, ...diagnostics.map(hop => hop.maxRtt ?? hop.avgRtt ?? 0)))}
							rows={diagnostics}
							density="compact"
							minWidth="68rem"
							maxHeight="28rem"
							getRowKey={row => row.id}
							emptyLabel="No hop data"
						/>
					) : (
						<BodyCopy>This run did not include hop rows.</BodyCopy>
					)}
					<div className={styles.traceTimeline}>
						<div className={styles.traceTimelineHeader}>
							<span>Run timeline</span>
							<strong className="ns-title">{totalRuns} runs in window</strong>
						</div>
						<RunTimeline
							points={timelinePoints}
							selectedPointId={selectedRun?.startedAt}
							selectedValueLabel={selectedTimelineValueLabel(selectedRun, timelinePoints)}
							timeRangeBounds={timeRangeBounds}
							emptyState={
								isRouteLoading ? (
									<LoadingState label="Loading traceroute timeline" detail="Building run buckets for the selected time range." size="compact" />
								) : (
									<BodyCopy>No traceroute runs in this time range.</BodyCopy>
								)
							}
							onSelectPoint={handleSelectTimelinePoint}
							onSelectTimeRange={onSelectTimeWindow}
						/>
					</div>
				</div>
			</Panel>

			{showTopology ? (
				<Panel tone="deep" title="Aggregated route graph">
					{hasTopology ? (
						<RouteTopologyMap nodes={topologyNodes} edges={topologyEdges} />
					) : isTopologyLoading ? (
						<LoadingState label="Loading topology" detail="Aggregating traceroute hops into the route graph." />
					) : (
						<BodyCopy>Topology data is unavailable for the selected filters; hop rows still show the latest run.</BodyCopy>
					)}
				</Panel>
			) : null}
		</div>
	);
}

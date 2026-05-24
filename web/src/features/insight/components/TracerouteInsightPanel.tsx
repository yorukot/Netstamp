import type { CheckDefinition } from "@/features/checks/data/checks";
import { buildHopDiagnostics, selectedTimelineValueLabel, summarizeTraceroute, tracerouteTimelinePoints } from "@/features/insight/data/tracerouteInsightData";
import { formatCount, formatMs, formatPercent, formatTime } from "@/features/insight/insightFormatters";
import type { HopDiagnostic } from "@/features/insight/insightTypes";
import type { Probe } from "@/features/probes/data/probes";
import type { TracerouteResult } from "@/shared/api/types";
import { BodyCopy } from "@/shared/components/BodyCopy";
import { KeyValueGrid } from "@/shared/components/KeyValueGrid";
import { LatencyRail } from "@/shared/visualizations/LatencyRail";
import { RouteTopologyMap, type RouteTopologyEdge, type RouteTopologyNode } from "@/shared/visualizations/RouteTopologyMap";
import { RunTimeline } from "@/shared/visualizations/RunTimeline";
import { Badge, DataTable, Panel, type DataColumn } from "@netstamp/ui";
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
	runs: TracerouteResult[];
	topologyNodes: RouteTopologyNode[];
	topologyEdges: RouteTopologyEdge[];
	isRunsLoading: boolean;
	isTopologyLoading: boolean;
	selectedRunStartedAt: string;
	onSelectRun: (startedAt: string) => void;
}

export function TracerouteInsightPanel({
	selectedProbe,
	selectedTarget,
	runs,
	topologyNodes,
	topologyEdges,
	isRunsLoading,
	isTopologyLoading,
	selectedRunStartedAt,
	onSelectRun
}: TracerouteInsightPanelProps) {
	const selectedRun = runs.find(run => run.startedAt === selectedRunStartedAt) || runs[0] || null;
	const diagnostics = buildHopDiagnostics(selectedRun);
	const summary = summarizeTraceroute(runs, selectedRun);
	const timelinePoints = tracerouteTimelinePoints(runs);
	const hasTopology = topologyNodes.length > 0 && topologyEdges.length > 0;

	if (!selectedProbe || !selectedTarget) {
		return (
			<Panel tone="deep" eyebrow="Traceroute" title="No route selected">
				<BodyCopy>Select a probe and traceroute target to inspect route health.</BodyCopy>
			</Panel>
		);
	}

	if (isRunsLoading && !runs.length) {
		return (
			<Panel tone="deep" eyebrow="Traceroute" title="Loading route">
				<BodyCopy>Loading traceroute runs for this probe-target pair.</BodyCopy>
			</Panel>
		);
	}

	if (!runs.length) {
		return (
			<Panel tone="deep" eyebrow="Traceroute" title="No traceroute runs">
				<BodyCopy>No traceroute results were recorded for this probe-target pair in the selected time range.</BodyCopy>
			</Panel>
		);
	}

	return (
		<div className={styles.tracerouteStack}>
			<Panel tone="glass" eyebrow="Route health" title={`${selectedProbe.name} → ${selectedTarget.target}`}>
				<KeyValueGrid
					items={[
						{ label: "Destination", value: <Badge tone={summary.statusTone}>{summary.statusLabel}</Badge> },
						{ label: "Final RTT", value: formatMs(summary.finalRtt) },
						{ label: "Final loss", value: formatPercent(summary.finalLoss) },
						{ label: "Hops", value: String(selectedRun?.hopCount ?? "-") },
						{ label: "Path changes", value: formatCount(summary.pathChangeCount) },
						{ label: "First loss", value: summary.firstPropagatedLossHop ? `hop ${summary.firstPropagatedLossHop}` : "none" },
						{ label: "First RTT jump", value: summary.firstRttJumpHop ? `hop ${summary.firstRttJumpHop}` : "none" },
						{ label: "Run", value: selectedRun ? formatTime(selectedRun.startedAt) : "-" }
					]}
				/>
			</Panel>

			<Panel className={styles.tracePanel} tone="deep" eyebrow="Route trace" title="Hop latency, loss, and run timeline">
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
							<strong>{runs.length} runs in window</strong>
						</div>
						<RunTimeline
							points={timelinePoints}
							selectedPointId={selectedRun?.startedAt}
							selectedValueLabel={selectedTimelineValueLabel(selectedRun, timelinePoints)}
							emptyState={<BodyCopy>No traceroute runs in this time range.</BodyCopy>}
							onSelectPoint={onSelectRun}
						/>
					</div>
				</div>
			</Panel>

			<Panel tone="deep" eyebrow="Topology" title="Aggregated route graph">
				{hasTopology ? (
					<RouteTopologyMap nodes={topologyNodes} edges={topologyEdges} />
				) : (
					<BodyCopy>{isTopologyLoading ? "Loading topology for this route." : "Topology data is unavailable for the selected filters; hop rows still show the latest run."}</BodyCopy>
				)}
			</Panel>
		</div>
	);
}

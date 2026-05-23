import { mapApiChecks } from "@/features/checks/api/checkAdapters";
import { type CheckDefinition } from "@/features/checks/data/checks";
import { mapApiProbes } from "@/features/probes/api/probeAdapters";
import { type Probe } from "@/features/probes/data/probes";
import { projectQueries } from "@/shared/api/queries";
import { type PingInsightResponse, type TracerouteHop, type TracerouteResult, type TracerouteTopologyEdge, type TracerouteTopologyNode } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { BodyCopy } from "@/shared/components/BodyCopy";
import { ChartPanel } from "@/shared/components/ChartPanel";
import { KeyValueGrid } from "@/shared/components/KeyValueGrid";
import { PageStack } from "@/shared/components/PageStack";
import { ResponsiveGrid } from "@/shared/components/ResponsiveGrid";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { pingInsightChartOption, type ChartOption, type PingInsightChartBucket, type PingInsightSampleDensityCell } from "@/shared/utils/chartOptions";
import { classNames } from "@/shared/utils/classNames";
import { Badge, DataTable, Panel, SelectField, type BadgeTone, type DataColumn } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useMemo, type CSSProperties } from "react";
import { useSearchParams } from "react-router-dom";
import styles from "./InsightPage.module.css";

type InsightView = "probe" | "target";
type HopTone = Extract<BadgeTone, "success" | "warning" | "critical" | "muted">;

interface EntityDetail {
	label: string;
	value: string;
}

interface GraphCard {
	key: string;
	title: string;
	metric: string;
	selected: boolean;
}

interface SummaryMetric {
	label: string;
	value: string;
	detail: string;
}

interface HopDiagnostic {
	id: string;
	hopIndex: number;
	label: string;
	address: string;
	sent: number;
	received: number;
	loss: number;
	minRtt: number | null;
	avgRtt: number | null;
	medianRtt: number | null;
	maxRtt: number | null;
	sampleCount: number;
	state: string;
	tone: HopTone;
	routerOnlyLoss: boolean;
	propagatedLoss: boolean;
	rttJump: boolean;
	noReply: boolean;
	error: string;
}

interface TracerouteSummary {
	statusTone: BadgeTone;
	statusLabel: string;
	finalRtt: number | null;
	finalLoss: number | null;
	firstPropagatedLossHop: number | null;
	firstRttJumpHop: number | null;
	pathChangeCount: number;
}

interface TopologyGraphNode {
	id: string;
	name: string;
	label: string;
	seenCount: number;
	avgRttMs?: number;
	lossPercent?: number;
	x: number;
	y: number;
	symbolSize: number;
	itemStyle: { color: string; borderColor: string; borderWidth: number };
	labelLayout?: { hideOverlap: boolean };
}

interface TopologyGraphEdge {
	source: string;
	target: string;
	seenCount: number;
	avgRttMs?: number;
	lossPercent?: number;
	lineStyle: { color: string; width: number; opacity: number; curveness: number };
}

interface TopologyTooltipParam {
	dataType?: "node" | "edge";
	data?: Partial<TopologyGraphNode & TopologyGraphEdge>;
}

type TimelineStyle = CSSProperties & {
	"--ns-run-rtt"?: string;
	"--ns-run-loss"?: string;
	"--ns-run-hop"?: string;
};

type RailStyle = CSSProperties & {
	"--ns-hop-range-start"?: string;
	"--ns-hop-range-end"?: string;
	"--ns-hop-rtt"?: string;
};

interface TimeWindow {
	from: number;
	to: number;
}

interface ParsedInsightUrlState {
	view: InsightView;
	hasValidView: boolean;
	timeWindow: TimeWindow;
	hasValidTimeWindow: boolean;
	probeId: string;
	checkId: string;
	runStartedAt: string;
}

const timeOptions = [
	{ value: "15m", label: "Last 15 minutes" },
	{ value: "1h", label: "Last 1 hour" },
	{ value: "6h", label: "Last 6 hours" },
	{ value: "24h", label: "Last 24 hours" },
	{ value: "7d", label: "Last 7 days" },
	{ value: "30d", label: "Last 30 days" }
];

const customTimeOption = { value: "custom", label: "Custom range" };

const timeRangeDurations: Record<string, number> = {
	"15m": 15 * 60 * 1000,
	"1h": 60 * 60 * 1000,
	"6h": 6 * 60 * 60 * 1000,
	"24h": 24 * 60 * 60 * 1000,
	"7d": 7 * 24 * 60 * 60 * 1000,
	"30d": 30 * 24 * 60 * 60 * 1000
};

const viewOptions = [
	{ value: "probe", label: "Choose a probe" },
	{ value: "target", label: "Choose a target" }
];

const hopColumns: DataColumn<HopDiagnostic>[] = [
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
	{ key: "avgRtt", label: "Avg", render: row => formatMs(row.avgRtt) },
	{ key: "medianRtt", label: "Median", render: row => formatMs(row.medianRtt) },
	{ key: "range", label: "Range", render: row => `${formatMs(row.minRtt)} / ${formatMs(row.maxRtt)}` },
	{ key: "sent", label: "Sent/Recv", render: row => `${row.sent}/${row.received}` },
	{ key: "state", label: "State", render: row => <Badge tone={row.tone}>{row.state}</Badge> }
];

function timeLabel(value: string) {
	if (value === customTimeOption.value) {
		return customTimeOption.label;
	}

	return timeOptions.find(option => option.value === value)?.label || value;
}

function timeWindowForRange(value: string) {
	const to = Date.now();
	const from = to - (timeRangeDurations[value] ?? timeRangeDurations["24h"]);

	return { from, to };
}

function parseEpochMs(value: string | null) {
	if (!value) {
		return null;
	}

	const parsed = Number(value);

	return Number.isFinite(parsed) && parsed > 0 ? Math.trunc(parsed) : null;
}

function isInsightView(value: string | null): value is InsightView {
	return value === "probe" || value === "target";
}

function parseInsightUrlState(searchParams: URLSearchParams, fallbackTimeWindow: TimeWindow): ParsedInsightUrlState {
	const from = parseEpochMs(searchParams.get("from"));
	const to = parseEpochMs(searchParams.get("to"));
	const rawView = searchParams.get("view");
	const hasValidTimeWindow = from !== null && to !== null && from < to;
	const hasValidView = isInsightView(rawView);

	return {
		view: hasValidView ? rawView : "probe",
		hasValidView,
		timeWindow: hasValidTimeWindow ? { from, to } : fallbackTimeWindow,
		hasValidTimeWindow,
		probeId: searchParams.get("probeId") || "",
		checkId: searchParams.get("checkId") || "",
		runStartedAt: searchParams.get("runStartedAt") || ""
	};
}

function timeRangeForWindow(timeWindow: TimeWindow) {
	const duration = timeWindow.to - timeWindow.from;
	const option = timeOptions.find(candidate => timeRangeDurations[candidate.value] === duration);

	return option?.value || customTimeOption.value;
}

function detailsForProbe(probe: Probe): EntityDetail[] {
	return [
		{ label: "Status", value: probe.status },
		{ label: "Location", value: probe.location },
		{ label: "Network", value: probe.asn },
		{ label: "Last heartbeat", value: probe.lastHeartbeat }
	];
}

function detailsForTarget(check: CheckDefinition): EntityDetail[] {
	return [
		{ label: "Target", value: check.target },
		{ label: "Family", value: check.type },
		{ label: "Interval", value: check.interval },
		{ label: "Latest", value: check.latest }
	];
}

function formatTime(value: string) {
	return new Date(value).toLocaleString(undefined, { month: "short", day: "2-digit", hour: "2-digit", minute: "2-digit" });
}

function formatShortTime(value: string) {
	return new Date(value).toLocaleTimeString(undefined, { hour: "2-digit", minute: "2-digit" });
}

function formatMs(value: number | null | undefined) {
	if (typeof value !== "number" || !Number.isFinite(value)) {
		return "-";
	}

	return `${value >= 100 ? Math.round(value) : value.toFixed(1)}ms`;
}

function formatPercent(value: number | null | undefined) {
	if (typeof value !== "number" || !Number.isFinite(value)) {
		return "-";
	}

	return `${value >= 10 ? Math.round(value) : value.toFixed(1)}%`;
}

function formatCount(value: number | null | undefined) {
	if (typeof value !== "number" || !Number.isFinite(value)) {
		return "-";
	}

	return new Intl.NumberFormat().format(value);
}

function formatEpochMs(value: number | null | undefined) {
	if (typeof value !== "number" || !Number.isFinite(value)) {
		return "-";
	}

	return new Date(value).toLocaleString(undefined, { month: "short", day: "2-digit", hour: "2-digit", minute: "2-digit" });
}

function pingSuccessRate(summary: PingInsightResponse["summary"] | undefined) {
	if (!summary?.totalResults) {
		return undefined;
	}

	return (summary.successfulCount / summary.totalResults) * 100;
}

function pingSummaryMetrics(data: PingInsightResponse | undefined): SummaryMetric[] {
	const summary = data?.summary;
	const sampleCount = data?.sampleDensity.reduce((total, cell) => total + cell.sampleCount, 0) ?? 0;

	return [
		{ label: "Latest", value: formatMs(summary?.latestRttAvgMs), detail: summary?.latestStatus || "no result" },
		{ label: "Average", value: formatMs(summary?.avgRttMs), detail: "rtt avg" },
		{ label: "P95", value: formatMs(summary?.p95RttMs), detail: "sample percentile" },
		{ label: "P99", value: formatMs(summary?.p99RttMs), detail: "sample percentile" },
		{ label: "Max", value: formatMs(summary?.maxRttMs), detail: "rtt max" },
		{ label: "Loss", value: formatPercent(summary?.avgLossPercent), detail: "average" },
		{ label: "Success", value: formatPercent(pingSuccessRate(summary)), detail: `${formatCount(summary?.successfulCount)}/${formatCount(summary?.totalResults)}` },
		{ label: "Samples", value: formatCount(sampleCount), detail: `${formatCount(summary?.receivedCount)} replies` }
	];
}

function pingChartBuckets(data: PingInsightResponse | undefined): PingInsightChartBucket[] {
	return (data?.buckets ?? []).map(bucket => ({
		timestampMs: bucket.timestampMs,
		rttMinMs: bucket.rttMinMs,
		rttAvgMs: bucket.rttAvgMs,
		rttMedianMs: bucket.rttMedianMs,
		rttMaxMs: bucket.rttMaxMs,
		rttStddevMs: bucket.rttStddevMs,
		lossPercent: bucket.lossPercent,
		sentCount: bucket.sentCount,
		receivedCount: bucket.receivedCount,
		resultCount: bucket.resultCount
	}));
}

function pingSampleDensity(data: PingInsightResponse | undefined): PingInsightSampleDensityCell[] {
	return (data?.sampleDensity ?? []).map(cell => ({
		timestampMs: cell.timestampMs,
		rttBucketStartMs: cell.rttBucketStartMs,
		rttBucketEndMs: cell.rttBucketEndMs,
		sampleCount: cell.sampleCount
	}));
}

function orderedHops(run: TracerouteResult | null | undefined) {
	return [...(run?.hops ?? [])].sort((a, b) => a.hopIndex - b.hopIndex);
}

function hopNodeId(hop: TracerouteHop) {
	return hop.address || hop.hostname || `unknown:${hop.hopIndex}`;
}

function runPathSignature(run: TracerouteResult) {
	return orderedHops(run)
		.map(hop => hopNodeId(hop))
		.join(">");
}

function lastRespondingHop(run: TracerouteResult | null | undefined) {
	return orderedHops(run)
		.reverse()
		.find(hop => hop.receivedCount > 0 || typeof hop.rttAvgMs === "number");
}

function hopLabel(hop: TracerouteHop) {
	return hop.hostname || hop.address || `unknown hop ${hop.hopIndex}`;
}

function hasMeaningfulRttJump(previousAvg: number | null, currentAvg: number | null) {
	return typeof previousAvg === "number" && typeof currentAvg === "number" && currentAvg - previousAvg > 10 && currentAvg > previousAvg * 1.2;
}

function buildHopDiagnostics(run: TracerouteResult | null | undefined): HopDiagnostic[] {
	const hops = orderedHops(run);
	const lastResponding = [...hops].reverse().find(hop => hop.receivedCount > 0 || typeof hop.rttAvgMs === "number");
	const lastRespondingIndex = lastResponding?.hopIndex ?? null;
	let previousRespondingAvg: number | null = null;

	return hops.map(hop => {
		const avgRtt = typeof hop.rttAvgMs === "number" ? hop.rttAvgMs : null;
		const minRtt = typeof hop.rttMinMs === "number" ? hop.rttMinMs : avgRtt;
		const medianRtt = typeof hop.rttMedianMs === "number" ? hop.rttMedianMs : avgRtt;
		const maxRtt = typeof hop.rttMaxMs === "number" ? hop.rttMaxMs : avgRtt;
		const loss = typeof hop.lossPercent === "number" ? hop.lossPercent : 0;
		const noReply = hop.sentCount > 0 && hop.receivedCount === 0;
		const downstreamResponding = hops.filter(candidate => candidate.hopIndex > hop.hopIndex && (candidate.receivedCount > 0 || typeof candidate.rttAvgMs === "number"));
		const lossPropagatesDownstream = loss >= 1 && downstreamResponding.length > 0 && downstreamResponding.every(candidate => candidate.lossPercent >= 1);
		const finalHopLoss = loss >= 1 && hop.hopIndex === lastRespondingIndex;
		const propagatedLoss = lossPropagatesDownstream || finalHopLoss;
		const rttJump = hasMeaningfulRttJump(previousRespondingAvg, avgRtt);
		let tone: HopTone = "success";
		let state = "Clear";

		if (noReply) {
			tone = "muted";
			state = "No reply";
		} else if (propagatedLoss) {
			tone = "critical";
			state = finalHopLoss ? "Final loss" : "Propagated loss";
		} else if (rttJump) {
			tone = "warning";
			state = "RTT jump";
		} else if (loss >= 1) {
			tone = "warning";
			state = "Router-only loss";
		}

		if (avgRtt !== null) {
			previousRespondingAvg = avgRtt;
		}

		return {
			id: `${hop.hopIndex}-${hopNodeId(hop)}`,
			hopIndex: hop.hopIndex,
			label: hopLabel(hop),
			address: hop.address || hop.hostname || "unknown",
			sent: hop.sentCount,
			received: hop.receivedCount,
			loss,
			minRtt,
			avgRtt,
			medianRtt,
			maxRtt,
			sampleCount: hop.rttSamplesMs?.length ?? 0,
			state,
			tone,
			routerOnlyLoss: loss >= 1 && !propagatedLoss,
			propagatedLoss,
			rttJump,
			noReply,
			error: hop.errorMessage || hop.errorCode || ""
		};
	});
}

function summarizeTraceroute(runs: TracerouteResult[], selectedRun: TracerouteResult | null): TracerouteSummary {
	const diagnostics = buildHopDiagnostics(selectedRun);
	const finalHop = lastRespondingHop(selectedRun);
	const chronologicalRuns = [...runs].sort((a, b) => Date.parse(a.startedAt) - Date.parse(b.startedAt));
	let pathChangeCount = 0;
	let previousSignature = "";

	for (const run of chronologicalRuns) {
		const signature = runPathSignature(run);

		if (previousSignature && signature && signature !== previousSignature) {
			pathChangeCount += 1;
		}

		if (signature) {
			previousSignature = signature;
		}
	}

	if (!selectedRun) {
		return {
			statusTone: "muted",
			statusLabel: "No data",
			finalRtt: null,
			finalLoss: null,
			firstPropagatedLossHop: null,
			firstRttJumpHop: null,
			pathChangeCount
		};
	}

	const statusTone = selectedRun.status === "successful" && selectedRun.destinationReached ? "success" : selectedRun.status === "error" || selectedRun.status === "timeout" ? "critical" : "warning";

	return {
		statusTone,
		statusLabel: selectedRun.destinationReached ? "Reached" : selectedRun.status,
		finalRtt: finalHop?.rttAvgMs ?? null,
		finalLoss: finalHop?.lossPercent ?? null,
		firstPropagatedLossHop: diagnostics.find(hop => hop.propagatedLoss)?.hopIndex ?? null,
		firstRttJumpHop: diagnostics.find(hop => hop.rttJump)?.hopIndex ?? null,
		pathChangeCount
	};
}

function runFinalRtt(run: TracerouteResult) {
	return lastRespondingHop(run)?.rttAvgMs ?? null;
}

function runFinalLoss(run: TracerouteResult) {
	return lastRespondingHop(run)?.lossPercent ?? 0;
}

function escapeHtml(value: unknown) {
	return String(value ?? "")
		.replaceAll("&", "&amp;")
		.replaceAll("<", "&lt;")
		.replaceAll(">", "&gt;")
		.replaceAll('"', "&quot;")
		.replaceAll("'", "&#39;");
}

function topologyTone(lossPercent: number | undefined, avgRttMs: number | undefined) {
	if (typeof lossPercent === "number" && lossPercent >= 1) {
		return "#ff453a";
	}

	if (typeof avgRttMs === "number" && avgRttMs >= 100) {
		return "#ff9f0a";
	}

	return "#ff7a1a";
}

function topologyColumn(node: TracerouteTopologyNode, maxHop: number) {
	if (node.kind === "probe") {
		return 0;
	}

	if (typeof node.hopIndex === "number") {
		return node.hopIndex;
	}

	if (node.kind === "destination") {
		return maxHop + 1;
	}

	return maxHop + 2;
}

function topologyChartOption(nodes: TracerouteTopologyNode[], edges: TracerouteTopologyEdge[]): ChartOption {
	const maxHop = Math.max(0, ...nodes.map(node => node.hopIndex ?? 0));
	const sortedNodes = [...nodes].sort((a, b) => topologyColumn(a, maxHop) - topologyColumn(b, maxHop) || a.label.localeCompare(b.label));
	const nodeColumns = new Map<number, TracerouteTopologyNode[]>();
	const maxSeen = Math.max(1, ...nodes.map(node => node.seenCount), ...edges.map(edge => edge.seenCount));

	for (const node of sortedNodes) {
		const column = topologyColumn(node, maxHop);
		nodeColumns.set(column, [...(nodeColumns.get(column) ?? []), node]);
	}

	const graphNodes: TopologyGraphNode[] = sortedNodes.map(node => {
		const column = topologyColumn(node, maxHop);
		const siblings = nodeColumns.get(column) ?? [];
		const row = Math.max(
			0,
			siblings.findIndex(candidate => candidate.id === node.id)
		);
		const yOffset = row - (siblings.length - 1) / 2;
		const color = node.kind === "probe" ? "#c4ccd9" : topologyTone(node.lossPercent, node.avgRttMs);

		return {
			id: node.id,
			name: node.label.length > 22 ? `${node.label.slice(0, 19)}...` : node.label,
			label: node.label,
			seenCount: node.seenCount,
			avgRttMs: node.avgRttMs,
			lossPercent: node.lossPercent,
			x: column * 180,
			y: yOffset * 88,
			symbolSize: 18 + Math.min(34, (node.seenCount / maxSeen) * 34),
			itemStyle: {
				color,
				borderColor: "rgba(255,255,255,0.52)",
				borderWidth: 1
			},
			labelLayout: { hideOverlap: true }
		};
	});
	const knownNodeIds = new Set(graphNodes.map(node => node.id));
	const graphEdges: TopologyGraphEdge[] = edges
		.filter(edge => knownNodeIds.has(edge.source) && knownNodeIds.has(edge.target))
		.map(edge => ({
			source: edge.source,
			target: edge.target,
			seenCount: edge.seenCount,
			avgRttMs: edge.avgRttMs,
			lossPercent: edge.lossPercent,
			lineStyle: {
				color: topologyTone(edge.lossPercent, edge.avgRttMs),
				width: 1 + Math.min(5, (edge.seenCount / maxSeen) * 5),
				opacity: 0.22 + Math.min(0.56, edge.seenCount / maxSeen),
				curveness: 0.08
			}
		}));

	return {
		backgroundColor: "transparent",
		tooltip: {
			trigger: "item",
			backgroundColor: "rgba(10,13,18,0.96)",
			borderColor: "rgba(255,122,26,0.32)",
			textStyle: { color: "#fff7ec", fontFamily: "JetBrains Mono, monospace", fontSize: 11 },
			formatter: (params: TopologyTooltipParam) => {
				const data = params.data;

				if (params.dataType === "edge") {
					return [`edge`, `seen ${formatCount(data?.seenCount)}`, `avg ${formatMs(data?.avgRttMs)}`, `loss ${formatPercent(data?.lossPercent)}`].map(escapeHtml).join("<br/>");
				}

				return [data?.label || "node", `seen ${formatCount(data?.seenCount)}`, `avg ${formatMs(data?.avgRttMs)}`, `loss ${formatPercent(data?.lossPercent)}`].map(escapeHtml).join("<br/>");
			}
		},
		series: [
			{
				type: "graph",
				layout: "none",
				roam: true,
				data: graphNodes,
				links: graphEdges,
				edgeSymbol: ["none", "arrow"],
				edgeSymbolSize: [0, 8],
				label: {
					show: true,
					position: "bottom",
					color: "#ddd4c8",
					fontFamily: "JetBrains Mono, monospace",
					fontSize: 10
				},
				lineStyle: {
					color: "source"
				},
				emphasis: {
					focus: "adjacency"
				}
			}
		]
	};
}

function LatencyRail({ hops }: { hops: HopDiagnostic[] }) {
	const maxRtt = Math.max(1, ...hops.map(hop => hop.maxRtt ?? hop.avgRtt ?? 0));

	return (
		<div className={styles.latencyRail}>
			<div className={styles.railHeader}>
				<span>Latency rail</span>
				<strong>max {formatMs(maxRtt)}</strong>
			</div>
			<div className={styles.railRows}>
				{hops.map(hop => {
					const start = ((hop.minRtt ?? hop.avgRtt ?? 0) / maxRtt) * 100;
					const end = ((hop.maxRtt ?? hop.avgRtt ?? 0) / maxRtt) * 100;
					const avg = ((hop.avgRtt ?? 0) / maxRtt) * 100;
					const style: RailStyle = {
						"--ns-hop-range-start": `${Math.max(0, Math.min(100, start))}%`,
						"--ns-hop-range-end": `${Math.max(0, Math.min(100, end))}%`,
						"--ns-hop-rtt": `${Math.max(0, Math.min(100, avg))}%`
					};

					return (
						<div className={classNames(styles.railRow, styles[`railRow${hop.tone}`])} style={style} key={hop.id}>
							<span className={styles.railHopIndex}>{String(hop.hopIndex).padStart(2, "0")}</span>
							<span className={styles.railTrack}>
								{hop.avgRtt !== null ? <span className={styles.railRange} /> : null}
								{hop.avgRtt !== null ? <span className={styles.railPoint} /> : null}
							</span>
							<span className={styles.railValue}>{formatMs(hop.avgRtt)}</span>
						</div>
					);
				})}
			</div>
		</div>
	);
}

function RunTimeline({ runs, selectedRun, onSelect }: { runs: TracerouteResult[]; selectedRun: TracerouteResult | null; onSelect: (startedAt: string) => void }) {
	const chronologicalRuns = [...runs].sort((a, b) => Date.parse(a.startedAt) - Date.parse(b.startedAt));
	const visibleRuns = chronologicalRuns.slice(-36);
	const maxRtt = Math.max(1, ...visibleRuns.map(run => runFinalRtt(run) ?? 0));
	const maxHopCount = Math.max(1, ...visibleRuns.map(run => run.hopCount));
	const timelineRuns = visibleRuns.reduce<{ items: Array<{ run: TracerouteResult; changed: boolean }>; previousSignature: string }>(
		(accumulator, run) => {
			const signature = runPathSignature(run);
			const changed = Boolean(accumulator.previousSignature && signature && signature !== accumulator.previousSignature);

			return {
				items: [...accumulator.items, { run, changed }],
				previousSignature: signature || accumulator.previousSignature
			};
		},
		{ items: [], previousSignature: "" }
	).items;

	if (!visibleRuns.length) {
		return <BodyCopy>No traceroute runs in this time range.</BodyCopy>;
	}

	return (
		<div className={styles.timeline}>
			<div className={styles.timelineRuns}>
				{timelineRuns.map(({ run, changed }) => {
					const style: TimelineStyle = {
						"--ns-run-rtt": `${Math.max(4, ((runFinalRtt(run) ?? 0) / maxRtt) * 100)}%`,
						"--ns-run-loss": `${Math.max(0, Math.min(100, runFinalLoss(run)))}%`,
						"--ns-run-hop": `${Math.max(6, (run.hopCount / maxHopCount) * 100)}%`
					};

					return (
						<button
							type="button"
							className={classNames(styles.timelineRun, changed && styles.timelineChanged)}
							style={style}
							data-selected={selectedRun?.startedAt === run.startedAt || undefined}
							onClick={() => onSelect(run.startedAt)}
							aria-label={`Select traceroute run ${formatTime(run.startedAt)}`}
							key={run.startedAt}
						>
							<span className={styles.timelineRtt} />
							<span className={styles.timelineLoss} />
							<span className={styles.timelineHop} />
							<span className={styles.timelineTime}>{formatShortTime(run.startedAt)}</span>
						</button>
					);
				})}
			</div>
			<div className={styles.timelineLegend}>
				<span>RTT</span>
				<span>Loss</span>
				<span>Hops</span>
				<span>Route change</span>
			</div>
		</div>
	);
}

function TracerouteInsight({
	selectedProbe,
	selectedTarget,
	runs,
	topologyNodes,
	topologyEdges,
	isRunsLoading,
	isTopologyLoading,
	selectedRunStartedAt,
	onSelectRun
}: {
	selectedProbe: Probe | null;
	selectedTarget: CheckDefinition | null;
	runs: TracerouteResult[];
	topologyNodes: TracerouteTopologyNode[];
	topologyEdges: TracerouteTopologyEdge[];
	isRunsLoading: boolean;
	isTopologyLoading: boolean;
	selectedRunStartedAt: string;
	onSelectRun: (startedAt: string) => void;
}) {
	const selectedRun = runs.find(run => run.startedAt === selectedRunStartedAt) || runs[0] || null;
	const diagnostics = buildHopDiagnostics(selectedRun);
	const summary = summarizeTraceroute(runs, selectedRun);
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

			<ResponsiveGrid collapseAt="lg" className={styles.tracerouteMainGrid}>
				<Panel className={styles.tracePanel} tone="deep" eyebrow="Route trace" title="Hop latency and loss">
					{diagnostics.length ? (
						<div className={styles.traceGrid}>
							<DataTable columns={hopColumns} rows={diagnostics} density="compact" minWidth="56rem" maxHeight="28rem" getRowKey={row => row.id} emptyLabel="No hop data" />
							<LatencyRail hops={diagnostics} />
						</div>
					) : (
						<BodyCopy>This run did not include hop rows.</BodyCopy>
					)}
				</Panel>
				<Panel tone="deep" eyebrow="Run timeline" title={`${runs.length} runs in window`}>
					<RunTimeline runs={runs} selectedRun={selectedRun} onSelect={onSelectRun} />
				</Panel>
			</ResponsiveGrid>

			<Panel tone="deep" eyebrow="Topology" title="Aggregated route graph">
				{hasTopology ? (
					<ChartPanel option={topologyChartOption(topologyNodes, topologyEdges)} height="24rem" />
				) : (
					<BodyCopy>{isTopologyLoading ? "Loading topology for this route." : "Topology data is unavailable for the selected filters; hop rows still show the latest run."}</BodyCopy>
				)}
			</Panel>
		</div>
	);
}

function PingInsight({
	selectedProbe,
	selectedTarget,
	data,
	isLoading,
	isFetching,
	timeLabel
}: {
	selectedProbe: Probe | null;
	selectedTarget: CheckDefinition | null;
	data: PingInsightResponse | undefined;
	isLoading: boolean;
	isFetching: boolean;
	timeLabel: string;
}) {
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
					<ChartPanel option={pingInsightChartOption(buckets, density)} height="27rem" />
				) : (
					<div className={styles.emptyState}>No ping results were recorded for this probe-target pair in the selected time range.</div>
				)}
			</Panel>
		</div>
	);
}

export function InsightPage() {
	const { projectRef } = useCurrentProject();
	const [searchParams, setSearchParams] = useSearchParams();
	const searchParamString = searchParams.toString();
	const fallbackTimeWindow = useMemo(() => timeWindowForRange("24h"), []);
	const urlState = useMemo(() => parseInsightUrlState(new URLSearchParams(searchParamString), fallbackTimeWindow), [fallbackTimeWindow, searchParamString]);
	const probesQuery = useQuery({
		...projectQueries.probes(projectRef || ""),
		enabled: Boolean(projectRef),
		select: data => mapApiProbes(data.probes)
	});
	const checksQuery = useQuery({
		...projectQueries.checks(projectRef || ""),
		enabled: Boolean(projectRef),
		select: data => mapApiChecks(data.checks, probesQuery.data)
	});
	const probes = probesQuery.data || [];
	const checks = checksQuery.data || [];
	const timeWindow = urlState.timeWindow;
	const timeRange = timeRangeForWindow(timeWindow);
	const timePickerOptions = timeRange === customTimeOption.value ? [...timeOptions, customTimeOption] : timeOptions;
	const view = urlState.view;
	const selectedProbe = probes.find(probe => probe.id === urlState.probeId) || probes[0] || null;
	const selectedTarget = checks.find(check => check.id === urlState.checkId) || checks[0] || null;
	const selectedRunStartedAt = selectedTarget?.type === "Traceroute" ? urlState.runStartedAt : "";
	const pingInsightQuery = useQuery({
		...projectQueries.pingInsight(projectRef || "", selectedProbe?.id || "", selectedTarget?.id || "", timeWindow),
		enabled: Boolean(projectRef && selectedProbe && selectedTarget && selectedTarget.type === "Ping")
	});
	const tracerouteRunsQuery = useQuery({
		...projectQueries.tracerouteRuns(projectRef || "", selectedProbe?.id || "", selectedTarget?.id || "", { ...timeWindow, limit: 100 }),
		enabled: Boolean(projectRef && selectedProbe && selectedTarget && selectedTarget.type === "Traceroute")
	});
	const tracerouteTopologyQuery = useQuery({
		...projectQueries.tracerouteTopology(projectRef || "", {
			probeId: selectedProbe?.id,
			checkId: selectedTarget?.id,
			...timeWindow,
			limit: 100
		}),
		enabled: Boolean(projectRef && selectedProbe && selectedTarget && selectedTarget.type === "Traceroute")
	});
	const selectedTitle = view === "probe" ? selectedProbe?.name || "No probe selected" : selectedTarget?.target || "No target selected";
	const selectedDetails = view === "probe" ? (selectedProbe ? detailsForProbe(selectedProbe) : []) : selectedTarget ? detailsForTarget(selectedTarget) : [];
	const pickerOptions =
		view === "probe" ? probes.map(probe => ({ value: probe.id, label: `${probe.name} · ${probe.location}` })) : checks.map(check => ({ value: check.id, label: `${check.target} · ${check.type}` }));

	const graphCards: GraphCard[] =
		view === "probe"
			? checks.map(check => ({
					key: check.id,
					title: `${selectedProbe?.name || "probe"} → ${check.target}`,
					metric: check.type.toLowerCase(),
					selected: check.id === selectedTarget?.id
				}))
			: probes.map(probe => ({
					key: probe.id,
					title: `${probe.name} → ${selectedTarget?.target || "target"}`,
					metric: (selectedTarget?.type || "Ping").toLowerCase(),
					selected: probe.id === selectedProbe?.id
				}));

	useEffect(() => {
		if (!probes.length || !checks.length) {
			return;
		}

		const next = new URLSearchParams(searchParamString);
		let changed = false;
		const setParam = (key: string, value: string) => {
			if (next.get(key) !== value) {
				next.set(key, value);
				changed = true;
			}
		};
		const deleteParam = (key: string) => {
			if (next.has(key)) {
				next.delete(key);
				changed = true;
			}
		};
		const probeChanged = Boolean(selectedProbe && urlState.probeId !== selectedProbe.id);
		const checkChanged = Boolean(selectedTarget && urlState.checkId !== selectedTarget.id);

		if (!urlState.hasValidTimeWindow) {
			setParam("from", String(timeWindow.from));
			setParam("to", String(timeWindow.to));
		}

		if (!urlState.hasValidView) {
			setParam("view", view);
		}

		if (selectedProbe) {
			setParam("probeId", selectedProbe.id);
		}

		if (selectedTarget) {
			setParam("checkId", selectedTarget.id);
		}

		if (!urlState.hasValidTimeWindow || probeChanged || checkChanged || selectedTarget?.type !== "Traceroute") {
			deleteParam("runStartedAt");
		}

		if (changed) {
			setSearchParams(next, { replace: true });
		}
	}, [
		checks.length,
		probes.length,
		searchParamString,
		selectedProbe,
		selectedTarget,
		setSearchParams,
		timeWindow.from,
		timeWindow.to,
		urlState.checkId,
		urlState.hasValidTimeWindow,
		urlState.hasValidView,
		urlState.probeId,
		view
	]);

	function updateSearchParams(update: (next: URLSearchParams) => void, options: { replace?: boolean } = {}) {
		const next = new URLSearchParams(searchParamString);
		update(next);
		setSearchParams(next, { replace: options.replace ?? false });
	}

	function selectGraphCard(graph: GraphCard) {
		updateSearchParams(next => {
			next.delete("runStartedAt");

			if (view === "probe") {
				next.set("checkId", graph.key);
				return;
			}

			next.set("probeId", graph.key);
		});
	}

	return (
		<PageStack>
			<ScreenHeader eyebrow="Result insight" title="Insight" copy="Inspect ping latency spread, packet-loss density, and traceroute route diagnostics from result data." />

			<div className={styles.filters}>
				<SelectField
					label="Time"
					value={timeRange}
					onChange={event => {
						const nextRange = event.currentTarget.value;

						if (nextRange === customTimeOption.value) {
							return;
						}

						const nextTimeWindow = timeWindowForRange(nextRange);

						updateSearchParams(next => {
							next.set("from", String(nextTimeWindow.from));
							next.set("to", String(nextTimeWindow.to));
							next.delete("runStartedAt");
						});
					}}
					options={timePickerOptions}
				/>
				<SelectField
					label="View"
					value={view}
					onChange={event => {
						const nextView = event.currentTarget.value;

						if (!isInsightView(nextView)) {
							return;
						}

						updateSearchParams(next => {
							next.set("view", nextView);
						});
					}}
					options={viewOptions}
				/>
				<SelectField
					label={view === "probe" ? "Probe" : "Target"}
					value={view === "probe" ? selectedProbe?.id || "" : selectedTarget?.id || ""}
					onChange={event => {
						updateSearchParams(next => {
							next.delete("runStartedAt");

							if (view === "probe") {
								next.set("probeId", event.currentTarget.value);
								return;
							}

							next.set("checkId", event.currentTarget.value);
						});
					}}
					options={pickerOptions}
				/>
			</div>

			<ResponsiveGrid>
				<Panel tone="glass" eyebrow={view === "probe" ? "Probe" : "Target"} title={selectedTitle}>
					<KeyValueGrid items={selectedDetails} />
				</Panel>
				<Panel tone="glass" eyebrow={view === "probe" ? "Targets" : "Probes"} title={view === "probe" ? "Target list" : "Probe list"}>
					<div className={styles.entityList}>
						{graphCards.map(graph => (
							<button type="button" className={styles.entityButton} data-selected={graph.selected || undefined} onClick={() => selectGraphCard(graph)} key={graph.key}>
								<span>{graph.title}</span>
								<strong>{graph.metric}</strong>
							</button>
						))}
					</div>
				</Panel>
			</ResponsiveGrid>

			{selectedTarget?.type === "Traceroute" ? (
				<TracerouteInsight
					selectedProbe={selectedProbe}
					selectedTarget={selectedTarget}
					runs={tracerouteRunsQuery.data?.runs ?? []}
					topologyNodes={tracerouteTopologyQuery.data?.nodes ?? []}
					topologyEdges={tracerouteTopologyQuery.data?.edges ?? []}
					isRunsLoading={tracerouteRunsQuery.isLoading}
					isTopologyLoading={tracerouteTopologyQuery.isLoading}
					selectedRunStartedAt={selectedRunStartedAt}
					onSelectRun={startedAt =>
						updateSearchParams(next => {
							next.set("runStartedAt", startedAt);
						})
					}
				/>
			) : (
				<PingInsight
					selectedProbe={selectedProbe}
					selectedTarget={selectedTarget}
					data={pingInsightQuery.data}
					isLoading={pingInsightQuery.isLoading}
					isFetching={pingInsightQuery.isFetching}
					timeLabel={timeLabel(timeRange)}
				/>
			)}
		</PageStack>
	);
}

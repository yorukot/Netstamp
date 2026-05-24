import { formatInterval, mapApiChecks } from "@/features/checks/api/checkAdapters";
import { type CheckDefinition } from "@/features/checks/data/checks";
import { mapApiProbes } from "@/features/probes/api/probeAdapters";
import { type Probe, type ProbeStatus } from "@/features/probes/data/probes";
import { projectQueries } from "@/shared/api/queries";
import { type ApiProjectAssignment, type PingInsightResponse, type TracerouteHop, type TracerouteResult, type TracerouteTopologyEdge, type TracerouteTopologyNode } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { BodyCopy } from "@/shared/components/BodyCopy";
import { ChartPanel } from "@/shared/components/ChartPanel";
import { KeyValueGrid } from "@/shared/components/KeyValueGrid";
import { PageStack } from "@/shared/components/PageStack";
import { ResponsiveGrid } from "@/shared/components/ResponsiveGrid";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { pingInsightChartOption, type PingInsightChartBucket, type PingInsightSampleDensityCell } from "@/shared/utils/chartOptions";
import { classNames } from "@/shared/utils/classNames";
import { Badge, DataTable, Panel, SelectField, type BadgeTone, type DataColumn } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useMemo, useRef, useState, type CSSProperties } from "react";
import { useSearchParams } from "react-router-dom";
import styles from "./InsightPage.module.css";

type InsightMode = "overview" | "probe" | "target";
type HopTone = Extract<BadgeTone, "success" | "warning" | "critical" | "muted">;

interface EntityDetail {
	label: string;
	value: string;
}

interface InsightPair {
	key: string;
	probeId: string;
	checkId: string;
	probe: Probe;
	check: CheckDefinition;
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

interface TopologyRouteNode {
	id: string;
	name: string;
	label: string;
	address?: string;
	hostname?: string;
	kind: TracerouteTopologyNode["kind"];
	hopIndex?: number;
	hopLabel: string;
	seenCount: number;
	avgRttMs?: number;
	lossPercent?: number;
	x: number;
	y: number;
	severity: TopologySeverity;
}

interface TopologyRouteEdge {
	source: string;
	target: string;
	sourceLabel: string;
	targetLabel: string;
	seenCount: number;
	avgRttMs?: number;
	lossPercent?: number;
	x1: number;
	y1: number;
	x2: number;
	y2: number;
	color: string;
	width: number;
	opacity: number;
}

interface TopologyRouteLayout {
	nodes: TopologyRouteNode[];
	edges: TopologyRouteEdge[];
	viewWidth: number;
	viewHeight: number;
	routeStartX: number;
	routeEndX: number;
}

type TopologySeverity = "normal" | "warning" | "critical";
type TopologyDetailPlacement = "above" | "below";
type TopologyDetailTone = TopologySeverity | "agent" | "destination";

interface TopologyHoverDetail {
	id: string;
	title: string;
	subtitle?: string;
	rows: EntityDetail[];
	x: number;
	y: number;
	placement: TopologyDetailPlacement;
	tone: TopologyDetailTone;
}

type TimelinePointStyle = CSSProperties & {
	"--ns-timeline-x"?: string;
	"--ns-timeline-y"?: string;
};

type RailStyle = CSSProperties & {
	"--ns-hop-range-start"?: string;
	"--ns-hop-range-end"?: string;
	"--ns-hop-rtt"?: string;
};

type TopologyMapStyle = CSSProperties & {
	"--ns-topology-width"?: string;
	"--ns-topology-height"?: string;
};

type TopologyNodeStyle = CSSProperties & {
	"--ns-topology-node-x"?: string;
	"--ns-topology-node-y"?: string;
};

type TopologyDetailStyle = CSSProperties & {
	"--ns-topology-detail-x"?: string;
	"--ns-topology-detail-y"?: string;
};

interface TimeWindow {
	from: number;
	to: number;
}

interface ParsedInsightUrlState {
	mode: InsightMode;
	hasValidMode: boolean;
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

const modeOptions: Array<{ value: InsightMode; label: string }> = [
	{ value: "overview", label: "Overview" },
	{ value: "probe", label: "By probe" },
	{ value: "target", label: "By target" }
];

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
		{ key: "latency", label: "Latency", render: row => <LatencyRailCell hop={row} maxRtt={maxRtt} /> },
		{ key: "medianRtt", label: "Median", render: row => formatMs(row.medianRtt) },
		{ key: "range", label: "Range", render: row => `${formatMs(row.minRtt)} / ${formatMs(row.maxRtt)}` },
		{ key: "sent", label: "Sent/Recv", render: row => `${row.sent}/${row.received}` },
		{ key: "state", label: "State", render: row => <Badge tone={row.tone}>{row.state}</Badge> }
	];
}

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

function isInsightMode(value: string | null): value is InsightMode {
	return value === "overview" || value === "probe" || value === "target";
}

function parseInsightUrlState(searchParams: URLSearchParams, fallbackTimeWindow: TimeWindow): ParsedInsightUrlState {
	const from = parseEpochMs(searchParams.get("from"));
	const to = parseEpochMs(searchParams.get("to"));
	const rawMode = searchParams.get("mode") || searchParams.get("view");
	const hasValidTimeWindow = from !== null && to !== null && from < to;
	const hasValidMode = isInsightMode(rawMode);

	return {
		mode: hasValidMode ? rawMode : "overview",
		hasValidMode,
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

function pairKey(probeId: string, checkId: string) {
	return `${probeId}:${checkId}`;
}

function checkTypeFromApi(value: string | undefined): CheckDefinition["type"] {
	return value?.toLowerCase() === "traceroute" ? "Traceroute" : "Ping";
}

function fallbackProbe(assignment: ApiProjectAssignment): Probe {
	const location = assignment.probe?.locationName || "-";
	const status: ProbeStatus = assignment.probe?.enabled === false ? "Draining" : "Offline";

	return {
		id: assignment.probeId,
		name: assignment.probe?.name || assignment.probeId,
		status,
		location,
		publicIp: "-",
		asn: "-",
		provider: "Unlabeled",
		region: location,
		ipFamily: "-",
		lastHeartbeat: "never",
		labelTokens: assignment.probe?.labels?.map(label => `${label.key}:${label.value}`) ?? [],
		version: "-",
		uptime: "-",
		cpu: "-",
		memory: "-",
		queue: assignment.probe?.enabled === false ? "disabled" : "accepting jobs",
		loss: "-",
		capabilities: []
	};
}

function fallbackCheck(assignment: ApiProjectAssignment): CheckDefinition {
	const type = checkTypeFromApi(assignment.check?.type);
	const target = assignment.check?.target || assignment.checkId;

	return {
		id: assignment.checkId,
		name: assignment.check?.name || target,
		type,
		target,
		status: "Configured",
		interval: assignment.check ? formatInterval(assignment.check.intervalSeconds) : "-",
		latest: "-",
		assigned: 0,
		description: assignment.check?.description || "",
		fields: [
			["Target", target],
			["Type", type],
			["Interval", assignment.check ? formatInterval(assignment.check.intervalSeconds) : "-"]
		]
	};
}

function buildInsightPairs(assignments: ApiProjectAssignment[], probes: Probe[], checks: CheckDefinition[]): InsightPair[] {
	const probesByID = new Map(probes.map(probe => [probe.id, probe]));
	const checksByID = new Map(checks.map(check => [check.id, check]));
	const seen = new Set<string>();
	const pairs: InsightPair[] = [];

	for (const assignment of assignments) {
		const key = pairKey(assignment.probeId, assignment.checkId);

		if (seen.has(key)) {
			continue;
		}

		seen.add(key);
		const probe = probesByID.get(assignment.probeId) || fallbackProbe(assignment);
		const check = checksByID.get(assignment.checkId) || fallbackCheck(assignment);

		pairs.push({
			key,
			probeId: assignment.probeId,
			checkId: assignment.checkId,
			probe,
			check
		});
	}

	return pairs.sort((a, b) => a.probe.name.localeCompare(b.probe.name) || a.check.target.localeCompare(b.check.target));
}

function uniquePairsByProbe(pairs: InsightPair[]) {
	const values = new Map<string, { probe: Probe; count: number }>();

	for (const pair of pairs) {
		const current = values.get(pair.probeId);
		values.set(pair.probeId, { probe: pair.probe, count: (current?.count ?? 0) + 1 });
	}

	return [...values.values()].sort((a, b) => a.probe.name.localeCompare(b.probe.name));
}

function uniquePairsByCheck(pairs: InsightPair[]) {
	const values = new Map<string, { check: CheckDefinition; count: number }>();

	for (const pair of pairs) {
		const current = values.get(pair.checkId);
		values.set(pair.checkId, { check: pair.check, count: (current?.count ?? 0) + 1 });
	}

	return [...values.values()].sort((a, b) => a.check.target.localeCompare(b.check.target));
}

function fallbackPairForMode(pairs: InsightPair[], mode: InsightMode, probeId: string, checkId: string) {
	const exact = pairs.find(pair => pair.probeId === probeId && pair.checkId === checkId);
	if (exact) {
		return exact;
	}

	if (mode === "probe" && probeId) {
		return pairs.find(pair => pair.probeId === probeId) || pairs[0] || null;
	}

	if (mode === "target" && checkId) {
		return pairs.find(pair => pair.checkId === checkId) || pairs[0] || null;
	}

	return pairs[0] || null;
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

function topologyTone(lossPercent: number | undefined, avgRttMs: number | undefined) {
	if (typeof lossPercent === "number" && lossPercent >= 1) {
		return "#ff453a";
	}

	if (typeof avgRttMs === "number" && avgRttMs >= 100) {
		return "#ff9f0a";
	}

	return "#ff7a1a";
}

const topologyColumnGap = 168;
const topologyRowGap = 76;
const topologyPaddingX = 72;
const topologyMinWidth = 760;
const topologyMinHeight = 280;
const topologyDetailInsetX = 132;
const topologyDetailEstimatedHeight = 224;
const topologyDetailGap = 16;

function topologySeverity(lossPercent: number | undefined, avgRttMs: number | undefined): TopologySeverity {
	if (typeof lossPercent === "number" && lossPercent >= 1) {
		return "critical";
	}

	if (typeof avgRttMs === "number" && avgRttMs >= 100) {
		return "warning";
	}

	return "normal";
}

function topologyKindRank(kind: TracerouteTopologyNode["kind"]) {
	if (kind === "probe") {
		return 0;
	}

	if (kind === "destination") {
		return 2;
	}

	return 1;
}

function topologySiblingOffset(index: number) {
	if (index === 0) {
		return 0;
	}

	const distance = Math.ceil(index / 2);
	return index % 2 === 1 ? -distance : distance;
}

function compactTopologyLabel(value: string) {
	return value.length > 18 ? `${value.slice(0, 15)}...` : value;
}

function topologyHopLabel(node: Pick<TracerouteTopologyNode, "kind" | "hopIndex">) {
	if (node.kind === "probe") {
		return "agent";
	}

	if (typeof node.hopIndex === "number") {
		return `hop ${String(node.hopIndex).padStart(2, "0")}`;
	}

	return node.kind;
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

function topologyRouteLayout(nodes: TracerouteTopologyNode[], edges: TracerouteTopologyEdge[]): TopologyRouteLayout {
	const maxHop = Math.max(0, ...nodes.map(node => node.hopIndex ?? 0));
	const sortedNodes = [...nodes].sort((a, b) => topologyColumn(a, maxHop) - topologyColumn(b, maxHop) || b.seenCount - a.seenCount || a.label.localeCompare(b.label));
	const nodeColumns = new Map<number, TracerouteTopologyNode[]>();
	const maxSeen = Math.max(1, ...nodes.map(node => node.seenCount), ...edges.map(edge => edge.seenCount));

	for (const node of sortedNodes) {
		const column = topologyColumn(node, maxHop);
		nodeColumns.set(column, [...(nodeColumns.get(column) ?? []), node]);
	}

	for (const [column, siblings] of nodeColumns) {
		nodeColumns.set(
			column,
			[...siblings].sort((a, b) => topologyKindRank(a.kind) - topologyKindRank(b.kind) || b.seenCount - a.seenCount || a.label.localeCompare(b.label))
		);
	}

	const columns = [...nodeColumns.entries()].sort(([a], [b]) => a - b);
	const maxColumn = Math.max(1, ...columns.map(([column]) => column));
	const maxOffset = Math.max(0, ...columns.map(([, siblings]) => Math.ceil((siblings.length - 1) / 2)));
	const viewWidth = Math.max(topologyMinWidth, topologyPaddingX * 2 + maxColumn * topologyColumnGap);
	const viewHeight = Math.max(topologyMinHeight, 176 + maxOffset * topologyRowGap * 2);
	const centerY = viewHeight / 2 - 14;
	const routeNodes: TopologyRouteNode[] = columns
		.sort(([a], [b]) => a - b)
		.flatMap(([column, siblings]) =>
			siblings.map((node, index) => {
				const yOffset = topologySiblingOffset(index);

				return {
					id: node.id,
					name: compactTopologyLabel(node.label),
					label: node.label,
					address: node.address,
					hostname: node.hostname,
					kind: node.kind,
					hopIndex: node.hopIndex,
					hopLabel: topologyHopLabel(node),
					seenCount: node.seenCount,
					avgRttMs: node.avgRttMs,
					lossPercent: node.lossPercent,
					x: topologyPaddingX + column * topologyColumnGap,
					y: centerY + yOffset * topologyRowGap,
					severity: topologySeverity(node.lossPercent, node.avgRttMs)
				};
			})
		);
	const knownNodeIds = new Set(routeNodes.map(node => node.id));
	const routeNodeById = new Map(routeNodes.map(node => [node.id, node]));
	const routeEdges: TopologyRouteEdge[] = edges
		.filter(edge => knownNodeIds.has(edge.source) && knownNodeIds.has(edge.target))
		.map(edge => {
			const sourceNode = routeNodeById.get(edge.source);
			const targetNode = routeNodeById.get(edge.target);
			const seenRatio = edge.seenCount / maxSeen;

			return {
				source: edge.source,
				target: edge.target,
				sourceLabel: sourceNode?.label ?? edge.source,
				targetLabel: targetNode?.label ?? edge.target,
				seenCount: edge.seenCount,
				avgRttMs: edge.avgRttMs,
				lossPercent: edge.lossPercent,
				x1: sourceNode?.x ?? 0,
				y1: sourceNode?.y ?? 0,
				x2: targetNode?.x ?? 0,
				y2: targetNode?.y ?? 0,
				color: topologyTone(edge.lossPercent, edge.avgRttMs),
				width: 1.5 + Math.min(2.5, seenRatio * 2.5),
				opacity: 0.34 + Math.min(0.48, seenRatio)
			};
		});
	const routeStartX = Math.min(...routeNodes.map(node => node.x));
	const routeEndX = Math.max(...routeNodes.map(node => node.x));

	return { nodes: routeNodes, edges: routeEdges, viewWidth, viewHeight, routeStartX, routeEndX };
}

function topologyNodeTitle(node: TopologyRouteNode) {
	const primaryName = node.hostname || node.label;
	const secondaryName = node.hostname && node.label !== node.hostname ? node.label : null;

	return [primaryName, secondaryName, node.address, node.hopLabel, `seen ${formatCount(node.seenCount)}`, `avg ${formatMs(node.avgRttMs)}`, `loss ${formatPercent(node.lossPercent)}`]
		.filter(Boolean)
		.join("\n");
}

function topologyEdgeTitle(edge: TopologyRouteEdge) {
	return [`${edge.sourceLabel} -> ${edge.targetLabel}`, `seen ${formatCount(edge.seenCount)}`, `avg ${formatMs(edge.avgRttMs)}`, `loss ${formatPercent(edge.lossPercent)}`].join("\n");
}

function topologyAriaLabel(value: string) {
	return value.replace(/\s*\n\s*/g, ", ");
}

function clampNumber(value: number, min: number, max: number) {
	return Math.min(Math.max(value, min), max);
}

function topologyDetailPosition(x: number, y: number, layout: TopologyRouteLayout): Pick<TopologyHoverDetail, "x" | "y" | "placement"> {
	const maxX = Math.max(topologyDetailInsetX, layout.viewWidth - topologyDetailInsetX);
	const hasRoomBelow = y + topologyDetailGap + topologyDetailEstimatedHeight <= layout.viewHeight;
	const placement: TopologyDetailPlacement = hasRoomBelow ? "below" : "above";

	return {
		x: clampNumber(x, topologyDetailInsetX, maxX),
		y,
		placement
	};
}

function topologyNodeDetailTone(node: TopologyRouteNode): TopologyDetailTone {
	if (node.kind === "probe") {
		return "agent";
	}
	if (node.kind === "destination") {
		return "destination";
	}
	return node.severity;
}

function topologyNodeDetail(node: TopologyRouteNode, layout: TopologyRouteLayout): TopologyHoverDetail {
	const position = topologyDetailPosition(node.x, node.y, layout);
	const title = node.hostname || node.label;
	const subtitle = node.hostname && node.label !== node.hostname ? node.label : undefined;

	return {
		id: `node:${node.id}`,
		title,
		subtitle,
		x: position.x,
		y: position.y,
		placement: position.placement,
		tone: topologyNodeDetailTone(node),
		rows: [
			...(node.address ? [{ label: "address", value: node.address }] : []),
			{ label: "seen", value: formatCount(node.seenCount) },
			{ label: "avg rtt", value: formatMs(node.avgRttMs) },
			{ label: "loss", value: formatPercent(node.lossPercent) }
		]
	};
}

function topologyEdgeDetail(edge: TopologyRouteEdge, layout: TopologyRouteLayout): TopologyHoverDetail {
	const position = topologyDetailPosition((edge.x1 + edge.x2) / 2, (edge.y1 + edge.y2) / 2, layout);

	return {
		id: `edge:${edge.source}->${edge.target}`,
		title: `${edge.sourceLabel} -> ${edge.targetLabel}`,
		x: position.x,
		y: position.y,
		placement: position.placement,
		tone: topologySeverity(edge.lossPercent, edge.avgRttMs),
		rows: [
			{ label: "seen", value: formatCount(edge.seenCount) },
			{ label: "avg rtt", value: formatMs(edge.avgRttMs) },
			{ label: "loss", value: formatPercent(edge.lossPercent) }
		]
	};
}

function TopologyDetailCard({ detail }: { detail: TopologyHoverDetail }) {
	const style: TopologyDetailStyle = {
		"--ns-topology-detail-x": `${detail.x}px`,
		"--ns-topology-detail-y": `${detail.y}px`
	};

	return (
		<div className={classNames(styles.topologyDetail, styles[`topologyDetail${detail.tone}`])} style={style} data-placement={detail.placement} id="topology-detail-card">
			<strong>{detail.title}</strong>
			{detail.subtitle ? <span className={styles.topologyDetailSubtitle}>{detail.subtitle}</span> : null}
			<dl>
				{detail.rows.map(row => (
					<div key={`${detail.id}:${row.label}`}>
						<dt>{row.label}</dt>
						<dd>{row.value}</dd>
					</div>
				))}
			</dl>
		</div>
	);
}

function TopologyRouteMap({ nodes, edges }: { nodes: TracerouteTopologyNode[]; edges: TracerouteTopologyEdge[] }) {
	const shellRef = useRef<HTMLDivElement>(null);
	const viewportRef = useRef<HTMLDivElement>(null);
	const [activeDetail, setActiveDetail] = useState<TopologyHoverDetail | null>(null);
	const layout = topologyRouteLayout(nodes, edges);
	const centerY = layout.viewHeight / 2 - 14;
	const style: TopologyMapStyle = {
		"--ns-topology-width": `${layout.viewWidth}px`,
		"--ns-topology-height": `${layout.viewHeight}px`
	};
	const clearActiveDetail = () => setActiveDetail(null);
	const showActiveDetail = (detail: TopologyHoverDetail) => {
		const shell = shellRef.current;
		const viewport = viewportRef.current;

		if (!shell || !viewport) {
			setActiveDetail(detail);
			return;
		}

		const viewportRect = viewport.getBoundingClientRect();
		const viewportHeight = window.innerHeight || document.documentElement.clientHeight;
		const screenY = viewportRect.top + detail.y;
		const hasRoomAbove = screenY - topologyDetailGap - topologyDetailEstimatedHeight >= 0;
		const hasRoomBelow = screenY + topologyDetailGap + topologyDetailEstimatedHeight <= viewportHeight;
		const placement = hasRoomAbove && (!hasRoomBelow || screenY > viewportHeight / 2) ? "above" : "below";
		const maxVisibleX = Math.max(topologyDetailInsetX, viewport.clientWidth - topologyDetailInsetX);
		const visibleX = clampNumber(detail.x - viewport.scrollLeft, topologyDetailInsetX, maxVisibleX);

		setActiveDetail({
			...detail,
			x: viewport.offsetLeft + visibleX,
			y: viewport.offsetTop + detail.y,
			placement
		});
	};

	return (
		<div className={styles.topologyShell} ref={shellRef}>
			<div className={styles.topologyLegend} aria-hidden="true">
				<span data-tone="agent">agent</span>
				<span data-tone="normal">normal</span>
				<span data-tone="warning">high rtt</span>
				<span data-tone="critical">loss</span>
				<span data-tone="destination">destination</span>
			</div>
			<div className={styles.topologyViewport} ref={viewportRef} onScroll={clearActiveDetail}>
				<div className={styles.topologyMap} style={style}>
					<svg className={styles.topologySvg} viewBox={`0 0 ${layout.viewWidth} ${layout.viewHeight}`} role="img" aria-label="Aggregated traceroute topology">
						<line className={styles.topologyCenterLine} x1={layout.routeStartX} x2={layout.routeEndX} y1={centerY} y2={centerY} />
						{layout.edges.map(edge => {
							const edgeTitle = topologyEdgeTitle(edge);

							return (
								<g key={`${edge.source}->${edge.target}`}>
									<line className={styles.topologyEdge} x1={edge.x1} x2={edge.x2} y1={edge.y1} y2={edge.y2} stroke={edge.color} strokeWidth={edge.width} opacity={edge.opacity} />
									<line
										aria-label={topologyAriaLabel(edgeTitle)}
										aria-describedby={activeDetail?.id === `edge:${edge.source}->${edge.target}` ? "topology-detail-card" : undefined}
										className={styles.topologyEdgeHit}
										role="graphics-symbol"
										tabIndex={0}
										x1={edge.x1}
										x2={edge.x2}
										y1={edge.y1}
										y2={edge.y2}
										onBlur={clearActiveDetail}
										onFocus={() => showActiveDetail(topologyEdgeDetail(edge, layout))}
										onPointerEnter={() => showActiveDetail(topologyEdgeDetail(edge, layout))}
										onPointerLeave={clearActiveDetail}
									/>
								</g>
							);
						})}
					</svg>
					{layout.nodes.map(node => {
						const nodeStyle: TopologyNodeStyle = {
							"--ns-topology-node-x": `${node.x}px`,
							"--ns-topology-node-y": `${node.y}px`
						};
						const nodeTitle = topologyNodeTitle(node);

						return (
							<div
								className={classNames(styles.topologyNode, styles[`topologyNode${node.kind}`], styles[`topologyNode${node.severity}`])}
								style={nodeStyle}
								tabIndex={0}
								aria-describedby={activeDetail?.id === `node:${node.id}` ? "topology-detail-card" : undefined}
								aria-label={topologyAriaLabel(nodeTitle)}
								key={node.id}
								onBlur={clearActiveDetail}
								onFocus={() => showActiveDetail(topologyNodeDetail(node, layout))}
								onPointerEnter={() => showActiveDetail(topologyNodeDetail(node, layout))}
								onPointerLeave={clearActiveDetail}
							>
								<span className={styles.topologyNodeDot} />
								<span className={styles.topologyNodeLabel}>
									<strong>{node.name}</strong>
									<span>{node.hopLabel}</span>
								</span>
							</div>
						);
					})}
				</div>
			</div>
			{activeDetail ? <TopologyDetailCard detail={activeDetail} /> : null}
		</div>
	);
}

function LatencyRailCell({ hop, maxRtt }: { hop: HopDiagnostic; maxRtt: number }) {
	const start = ((hop.minRtt ?? hop.avgRtt ?? 0) / maxRtt) * 100;
	const end = ((hop.maxRtt ?? hop.avgRtt ?? 0) / maxRtt) * 100;
	const avg = ((hop.avgRtt ?? 0) / maxRtt) * 100;
	const style: RailStyle = {
		"--ns-hop-range-start": `${Math.max(0, Math.min(100, start))}%`,
		"--ns-hop-range-end": `${Math.max(0, Math.min(100, end))}%`,
		"--ns-hop-rtt": `${Math.max(0, Math.min(100, avg))}%`
	};

	return (
		<span className={classNames(styles.railCell, styles[`railCell${hop.tone}`])} style={style}>
			<span className={styles.railTrack}>
				{hop.avgRtt !== null ? <span className={styles.railRange} /> : null}
				{hop.avgRtt !== null ? <span className={styles.railPoint} /> : null}
			</span>
			<span className={styles.railValue}>{formatMs(hop.avgRtt)}</span>
		</span>
	);
}

function RunTimeline({ runs, selectedRun, onSelect }: { runs: TracerouteResult[]; selectedRun: TracerouteResult | null; onSelect: (startedAt: string) => void }) {
	const chronologicalRuns = [...runs].sort((a, b) => Date.parse(a.startedAt) - Date.parse(b.startedAt));
	const visibleRuns = chronologicalRuns.slice(-100);
	const firstRunTime = visibleRuns.length ? Date.parse(visibleRuns[0].startedAt) : 0;
	const lastRunTime = visibleRuns.length ? Date.parse(visibleRuns[visibleRuns.length - 1].startedAt) : firstRunTime;
	const runTimeSpan = Math.max(1, lastRunTime - firstRunTime);
	const rttValues = visibleRuns.map(run => runFinalRtt(run)).filter((value): value is number => typeof value === "number");
	const minRtt = rttValues.length ? Math.min(...rttValues) : 0;
	const maxRtt = rttValues.length ? Math.max(...rttValues) : 1;
	const rttSpan = Math.max(1, maxRtt - minRtt);
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
	const points = timelineRuns.map(({ run, changed }) => {
		const timestamp = Date.parse(run.startedAt);
		const rtt = runFinalRtt(run);
		const loss = runFinalLoss(run);
		const x = 6 + ((timestamp - firstRunTime) / runTimeSpan) * 88;
		const y = typeof rtt === "number" ? (maxRtt === minRtt ? 49 : 78 - ((rtt - minRtt) / rttSpan) * 58) : 82;

		return { run, changed, rtt, loss, x, y };
	});
	const polylinePoints = points.map(point => `${point.x},${point.y}`).join(" ");

	if (!visibleRuns.length) {
		return <BodyCopy>No traceroute runs in this time range.</BodyCopy>;
	}

	return (
		<div className={styles.timeline}>
			<div className={styles.timelineChart}>
				<svg className={styles.timelineSvg} viewBox="0 0 100 100" preserveAspectRatio="none" aria-hidden="true">
					<line className={styles.timelineGridLine} x1="6" x2="94" y1="20" y2="20" />
					<line className={styles.timelineGridLine} x1="6" x2="94" y1="49" y2="49" />
					<line className={styles.timelineAxisLine} x1="6" x2="94" y1="78" y2="78" />
					{polylinePoints ? <polyline className={styles.timelineLine} points={polylinePoints} /> : null}
				</svg>
				{points.map(point => {
					const style: TimelinePointStyle = {
						"--ns-timeline-x": `${point.x}%`,
						"--ns-timeline-y": `${point.y}%`
					};
					const selected = selectedRun?.startedAt === point.run.startedAt;

					return (
						<button
							type="button"
							className={styles.timelinePoint}
							style={style}
							data-selected={selected || undefined}
							data-loss={point.loss > 0 || undefined}
							data-changed={point.changed || undefined}
							onClick={() => onSelect(point.run.startedAt)}
							aria-label={`Select traceroute run ${formatTime(point.run.startedAt)} final RTT ${formatMs(point.rtt)} loss ${formatPercent(point.loss)}`}
							key={point.run.startedAt}
						>
							<span className={styles.timelinePointCore} />
							<span className={styles.timelinePointLabel}>
								{formatShortTime(point.run.startedAt)}
								<br />
								{formatMs(point.rtt)}
							</span>
						</button>
					);
				})}
				<div className={styles.timelineAxisLabels}>
					<span>{formatShortTime(visibleRuns[0].startedAt)}</span>
					<strong>{formatMs(selectedRun ? runFinalRtt(selectedRun) : points[points.length - 1]?.rtt)} selected</strong>
					<span>{formatShortTime(visibleRuns[visibleRuns.length - 1].startedAt)}</span>
				</div>
			</div>
			<div className={styles.timelineLegend}>
				<span>RTT line</span>
				<span>Loss</span>
				<span>Route change</span>
				<span>Selected run</span>
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
						<RunTimeline runs={runs} selectedRun={selectedRun} onSelect={onSelectRun} />
					</div>
				</div>
			</Panel>

			<Panel tone="deep" eyebrow="Topology" title="Aggregated route graph">
				{hasTopology ? (
					<TopologyRouteMap nodes={topologyNodes} edges={topologyEdges} />
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

function ModeControl({ mode, onChange }: { mode: InsightMode; onChange: (mode: InsightMode) => void }) {
	return (
		<div className={styles.modeField}>
			<span className={styles.modeLabel}>Mode</span>
			<div className={styles.modeControl} role="tablist" aria-label="Insight mode">
				{modeOptions.map(option => (
					<button
						type="button"
						role="tab"
						aria-selected={mode === option.value}
						className={styles.modeButton}
						data-selected={mode === option.value || undefined}
						onClick={() => onChange(option.value)}
						key={option.value}
					>
						{option.label}
					</button>
				))}
			</div>
		</div>
	);
}

function PairList({ pairs, selectedKey, onSelect, emptyLabel }: { pairs: InsightPair[]; selectedKey: string; onSelect: (pair: InsightPair) => void; emptyLabel: string }) {
	if (!pairs.length) {
		return <BodyCopy>{emptyLabel}</BodyCopy>;
	}

	return (
		<div className={styles.entityList}>
			{pairs.map(pair => (
				<button type="button" className={styles.entityButton} data-selected={pair.key === selectedKey || undefined} onClick={() => onSelect(pair)} key={pair.key}>
					<span>
						{pair.probe.name} → {pair.check.target}
					</span>
					<strong>
						{pair.check.type.toLowerCase()} · {pair.check.interval}
					</strong>
					<small className={styles.pairMeta}>
						{pair.probe.location} · {pair.check.name}
					</small>
				</button>
			))}
		</div>
	);
}

function GroupTopologyPanel({ title, nodes, edges, isLoading }: { title: string; nodes: TracerouteTopologyNode[]; edges: TracerouteTopologyEdge[]; isLoading: boolean }) {
	const hasTopology = nodes.length > 0 && edges.length > 0;

	return (
		<Panel tone="deep" eyebrow="Traceroute topology" title={title}>
			{hasTopology ? (
				<TopologyRouteMap nodes={nodes} edges={edges} />
			) : (
				<BodyCopy>{isLoading ? "Loading aggregated route graph." : "No traceroute topology is available for the selected scope and time range."}</BodyCopy>
			)}
		</Panel>
	);
}

function InsightPairDetail({
	pair,
	pingData,
	isPingLoading,
	isPingFetching,
	tracerouteRuns,
	topologyNodes,
	topologyEdges,
	isRunsLoading,
	isTopologyLoading,
	selectedRunStartedAt,
	onSelectRun,
	timeLabel
}: {
	pair: InsightPair | null;
	pingData: PingInsightResponse | undefined;
	isPingLoading: boolean;
	isPingFetching: boolean;
	tracerouteRuns: TracerouteResult[];
	topologyNodes: TracerouteTopologyNode[];
	topologyEdges: TracerouteTopologyEdge[];
	isRunsLoading: boolean;
	isTopologyLoading: boolean;
	selectedRunStartedAt: string;
	onSelectRun: (startedAt: string) => void;
	timeLabel: string;
}) {
	if (!pair) {
		return (
			<Panel tone="deep" eyebrow="Insight detail" title="No assignment selected">
				<BodyCopy>Select an active probe-check assignment to inspect result data.</BodyCopy>
			</Panel>
		);
	}

	return pair.check.type === "Traceroute" ? (
		<TracerouteInsight
			selectedProbe={pair.probe}
			selectedTarget={pair.check}
			runs={tracerouteRuns}
			topologyNodes={topologyNodes}
			topologyEdges={topologyEdges}
			isRunsLoading={isRunsLoading}
			isTopologyLoading={isTopologyLoading}
			selectedRunStartedAt={selectedRunStartedAt}
			onSelectRun={onSelectRun}
		/>
	) : (
		<PingInsight selectedProbe={pair.probe} selectedTarget={pair.check} data={pingData} isLoading={isPingLoading} isFetching={isPingFetching} timeLabel={timeLabel} />
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
	const assignmentsQuery = useQuery({
		...projectQueries.assignments(projectRef || ""),
		enabled: Boolean(projectRef),
		select: data => data.assignments
	});
	const probes = useMemo(() => probesQuery.data ?? [], [probesQuery.data]);
	const checks = useMemo(() => checksQuery.data ?? [], [checksQuery.data]);
	const assignments = useMemo(() => assignmentsQuery.data ?? [], [assignmentsQuery.data]);
	const pairs = useMemo(() => buildInsightPairs(assignments, probes, checks), [assignments, checks, probes]);
	const timeWindow = urlState.timeWindow;
	const timeRange = timeRangeForWindow(timeWindow);
	const timePickerOptions = timeRange === customTimeOption.value ? [...timeOptions, customTimeOption] : timeOptions;
	const mode = urlState.mode;
	const selectedPair = useMemo(() => fallbackPairForMode(pairs, mode, urlState.probeId, urlState.checkId), [mode, pairs, urlState.checkId, urlState.probeId]);
	const selectedRunStartedAt = selectedPair?.check.type === "Traceroute" ? urlState.runStartedAt : "";
	const probeOptions = useMemo(() => uniquePairsByProbe(pairs), [pairs]);
	const checkOptions = useMemo(() => uniquePairsByCheck(pairs), [pairs]);
	const selectedKey = selectedPair?.key || "";
	const probePairs = selectedPair ? pairs.filter(pair => pair.probeId === selectedPair.probeId) : [];
	const targetPairs = selectedPair ? pairs.filter(pair => pair.checkId === selectedPair.checkId) : [];
	const focusPairs = mode === "probe" ? probePairs : mode === "target" ? targetPairs : pairs;
	const hasGroupTraceroute =
		mode === "overview"
			? pairs.some(pair => pair.check.type === "Traceroute")
			: mode === "probe"
				? probePairs.some(pair => pair.check.type === "Traceroute")
				: selectedPair?.check.type === "Traceroute";
	const groupTopologyFilters = useMemo(() => {
		if (!hasGroupTraceroute) {
			return null;
		}

		if (mode === "probe" && selectedPair) {
			return { probeId: selectedPair.probeId, ...timeWindow, limit: 100 };
		}

		if (mode === "target" && selectedPair) {
			return { checkId: selectedPair.checkId, ...timeWindow, limit: 100 };
		}

		return { ...timeWindow, limit: 100 };
	}, [hasGroupTraceroute, mode, selectedPair, timeWindow]);
	const pingInsightQuery = useQuery({
		...projectQueries.pingInsight(projectRef || "", selectedPair?.probeId || "", selectedPair?.checkId || "", timeWindow),
		enabled: Boolean(projectRef && selectedPair && selectedPair.check.type === "Ping")
	});
	const tracerouteRunsQuery = useQuery({
		...projectQueries.tracerouteRuns(projectRef || "", selectedPair?.probeId || "", selectedPair?.checkId || "", { ...timeWindow, limit: 100 }),
		enabled: Boolean(projectRef && selectedPair && selectedPair.check.type === "Traceroute")
	});
	const pairTopologyQuery = useQuery({
		...projectQueries.tracerouteTopology(projectRef || "", {
			probeId: selectedPair?.probeId,
			checkId: selectedPair?.checkId,
			...timeWindow,
			limit: 100
		}),
		enabled: Boolean(projectRef && selectedPair && selectedPair.check.type === "Traceroute")
	});
	const groupTopologyQuery = useQuery({
		...projectQueries.tracerouteTopology(projectRef || "", groupTopologyFilters || {}),
		enabled: Boolean(projectRef && groupTopologyFilters)
	});
	const isSelectionLoading = Boolean(projectRef) && (assignmentsQuery.isLoading || probesQuery.isLoading || checksQuery.isLoading);
	const pingPairs = pairs.filter(pair => pair.check.type === "Ping").length;
	const traceroutePairs = pairs.filter(pair => pair.check.type === "Traceroute").length;

	useEffect(() => {
		if (!projectRef) {
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

		if (!urlState.hasValidTimeWindow) {
			setParam("from", String(timeWindow.from));
			setParam("to", String(timeWindow.to));
		}

		if (next.has("view")) {
			deleteParam("view");
		}

		if (!urlState.hasValidMode || next.get("mode") !== mode) {
			setParam("mode", mode);
		}

		if (!pairs.length) {
			deleteParam("probeId");
			deleteParam("checkId");
			deleteParam("runStartedAt");
			if (changed) {
				setSearchParams(next, { replace: true });
			}
			return;
		}

		if (selectedPair) {
			setParam("probeId", selectedPair.probeId);
			setParam("checkId", selectedPair.checkId);
		}

		if (!urlState.hasValidTimeWindow || selectedPair?.key !== pairKey(urlState.probeId, urlState.checkId) || selectedPair?.check.type !== "Traceroute") {
			deleteParam("runStartedAt");
		}

		if (changed) {
			setSearchParams(next, { replace: true });
		}
	}, [
		mode,
		pairs.length,
		projectRef,
		searchParamString,
		selectedPair,
		setSearchParams,
		timeWindow.from,
		timeWindow.to,
		urlState.checkId,
		urlState.hasValidTimeWindow,
		urlState.hasValidMode,
		urlState.probeId
	]);

	function updateSearchParams(update: (next: URLSearchParams) => void, options: { replace?: boolean } = {}) {
		const next = new URLSearchParams(searchParamString);
		update(next);
		setSearchParams(next, { replace: options.replace ?? false });
	}

	function selectPair(pair: InsightPair) {
		updateSearchParams(next => {
			next.set("probeId", pair.probeId);
			next.set("checkId", pair.checkId);
			next.delete("runStartedAt");
		});
	}

	const groupTopologyTitle =
		mode === "probe" && selectedPair ? `${selectedPair.probe.name} route graph` : mode === "target" && selectedPair ? `${selectedPair.check.target} route graph` : "Project route graph";
	const detail = (
		<InsightPairDetail
			pair={selectedPair}
			pingData={pingInsightQuery.data}
			isPingLoading={pingInsightQuery.isLoading}
			isPingFetching={pingInsightQuery.isFetching}
			tracerouteRuns={tracerouteRunsQuery.data?.runs ?? []}
			topologyNodes={pairTopologyQuery.data?.nodes ?? []}
			topologyEdges={pairTopologyQuery.data?.edges ?? []}
			isRunsLoading={tracerouteRunsQuery.isLoading}
			isTopologyLoading={pairTopologyQuery.isLoading}
			selectedRunStartedAt={selectedRunStartedAt}
			onSelectRun={startedAt =>
				updateSearchParams(next => {
					next.set("runStartedAt", startedAt);
				})
			}
			timeLabel={timeLabel(timeRange)}
		/>
	);
	const groupTopology = groupTopologyFilters ? (
		<GroupTopologyPanel title={groupTopologyTitle} nodes={groupTopologyQuery.data?.nodes ?? []} edges={groupTopologyQuery.data?.edges ?? []} isLoading={groupTopologyQuery.isLoading} />
	) : null;
	const modeBody =
		mode === "overview" ? (
			<>
				<ResponsiveGrid>
					<Panel tone="glass" eyebrow="Assignments" title={`${pairs.length} active paths`}>
						<PairList pairs={pairs} selectedKey={selectedKey} onSelect={selectPair} emptyLabel="No active assignments." />
					</Panel>
					<Panel tone="glass" eyebrow="Coverage" title="Assignment scope">
						<KeyValueGrid
							items={[
								{ label: "Probes", value: formatCount(probeOptions.length) },
								{ label: "Targets", value: formatCount(checkOptions.length) },
								{ label: "Ping", value: formatCount(pingPairs) },
								{ label: "Traceroute", value: formatCount(traceroutePairs) },
								{ label: "Selected", value: selectedPair ? `${selectedPair.probe.name} -> ${selectedPair.check.target}` : "-" }
							]}
						/>
					</Panel>
				</ResponsiveGrid>
				{groupTopology}
				{detail}
			</>
		) : (
			<>
				<ResponsiveGrid>
					<Panel
						tone="glass"
						eyebrow={mode === "probe" ? "Probe" : "Target"}
						title={mode === "probe" ? selectedPair?.probe.name || "No probe selected" : selectedPair?.check.target || "No target selected"}
					>
						<KeyValueGrid items={mode === "probe" && selectedPair ? detailsForProbe(selectedPair.probe) : mode === "target" && selectedPair ? detailsForTarget(selectedPair.check) : []} />
					</Panel>
					<Panel tone="glass" eyebrow={mode === "probe" ? "Assigned targets" : "Assigned probes"} title={mode === "probe" ? `${focusPairs.length} targets` : `${focusPairs.length} probes`}>
						<PairList pairs={focusPairs} selectedKey={selectedKey} onSelect={selectPair} emptyLabel="No assignments in this scope." />
					</Panel>
				</ResponsiveGrid>
				{groupTopology}
				{detail}
			</>
		);

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
				<ModeControl
					mode={mode}
					onChange={nextMode => {
						updateSearchParams(next => {
							next.set("mode", nextMode);
							next.delete("view");
							next.delete("runStartedAt");
						});
					}}
				/>
				{mode === "overview" ? (
					<div className={styles.overviewMeta}>
						<span>{isSelectionLoading ? "loading assignments" : `${formatCount(pairs.length)} active paths`}</span>
						<strong>{selectedPair ? `${selectedPair.probe.name} -> ${selectedPair.check.target}` : "No selection"}</strong>
					</div>
				) : mode === "probe" ? (
					<SelectField
						label="Probe"
						value={selectedPair?.probeId || ""}
						onChange={event => {
							const nextPair = pairs.find(pair => pair.probeId === event.currentTarget.value);
							if (nextPair) {
								selectPair(nextPair);
							}
						}}
						options={probeOptions.map(option => ({ value: option.probe.id, label: `${option.probe.name} · ${option.count} targets` }))}
					/>
				) : (
					<SelectField
						label="Target"
						value={selectedPair?.checkId || ""}
						onChange={event => {
							const nextPair = pairs.find(pair => pair.checkId === event.currentTarget.value);
							if (nextPair) {
								selectPair(nextPair);
							}
						}}
						options={checkOptions.map(option => ({ value: option.check.id, label: `${option.check.target} · ${option.count} probes` }))}
					/>
				)}
			</div>

			{isSelectionLoading && !pairs.length ? (
				<Panel tone="deep" eyebrow="Assignments" title="Loading active paths">
					<BodyCopy>Loading probe-check assignments for this project.</BodyCopy>
				</Panel>
			) : !pairs.length ? (
				<Panel tone="deep" eyebrow="Assignments" title="No active paths">
					<BodyCopy>Create or refresh check assignments before opening result insight.</BodyCopy>
				</Panel>
			) : (
				modeBody
			)}
		</PageStack>
	);
}

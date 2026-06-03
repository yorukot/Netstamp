import { formatInterval, mapApiChecks } from "@/features/checks/api/checkAdapters";
import { type CheckDefinition } from "@/features/checks/data/checks";
import { GroupTopologyPanel } from "@/features/insight/components/GroupTopologyPanel";
import { FocusChip, InsightTimeControl, SegmentedControl } from "@/features/insight/components/InsightControls";
import { PingInsightPanel } from "@/features/insight/components/PingInsightPanel";
import { TcpInsightPanel } from "@/features/insight/components/TcpInsightPanel";
import { TracerouteInsightPanel } from "@/features/insight/components/TracerouteInsightPanel";
import {
	type InsightCheckTypeFilter,
	type InsightGroupBy,
	type InsightPair,
	type InsightRefreshInterval,
	type InsightRelativeRange,
	type InsightTimeMode,
	type ParsedInsightUrlState,
	type TimeWindow
} from "@/features/insight/insightTypes";
import { mapApiProbes } from "@/features/probes/api/probeAdapters";
import { type Probe, type ProbeStatus } from "@/features/probes/data/probes";
import { projectQueries } from "@/shared/api/queries";
import { apiQueryKeys } from "@/shared/api/queryKeys";
import {
	type ApiLatestResult,
	type ApiProjectAssignment,
	type PingInsightResponse,
	type PingSeriesResponse,
	type TcpInsightResponse,
	type TcpSeriesResponse,
	type TracerouteInsightResponse,
	type TracerouteResult
} from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { BodyCopy } from "@/shared/components/BodyCopy";
import { FilterGrid } from "@/shared/components/FilterGrid";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { formatCount, formatEpochMs } from "@/shared/utils/insightFormatters";
import {
	isRelativeTimeRange as isInsightRelativeRange,
	parseEpochMs,
	relativeRangeForTimeWindow as relativeRangeForWindow,
	timeWindowForRelativeRange as timeWindowForRange
} from "@/shared/utils/timeRanges";
import { type RouteTopologyEdge, type RouteTopologyNode } from "@/shared/visualizations/RouteTopologyMap";
import { Badge, Button, DataTable, Panel, TextField, type BadgeTone, type DataColumn } from "@netstamp/ui";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
import styles from "./InsightPage.module.css";

const refreshDurations: Partial<Record<InsightRefreshInterval, number>> = {
	"10s": 10 * 1000,
	"30s": 30 * 1000,
	"1m": 60 * 1000,
	"5m": 5 * 60 * 1000
};

const checkTypeOptions: Array<{ value: InsightCheckTypeFilter; label: string }> = [
	{ value: "all", label: "All" },
	{ value: "ping", label: "Ping" },
	{ value: "tcp", label: "TCP" },
	{ value: "traceroute", label: "Traceroute" }
];

const groupByOptions: Array<{ value: InsightGroupBy; label: string }> = [
	{ value: "check", label: "By check" },
	{ value: "probe", label: "By probe" }
];

type GroupStatus = {
	label: string;
	tone: BadgeTone;
	rank: number;
};

interface LatestSummary {
	latestStartedAtMs: number | null;
	status: GroupStatus;
}

interface InsightGroupRow {
	key: string;
	id: string;
	groupBy: InsightGroupBy;
	label: string;
	secondary: string;
	pairs: InsightPair[];
	pairCount: number;
	probeCount: number;
	checkCount: number;
	pingCount: number;
	tcpCount: number;
	tracerouteCount: number;
	summary: LatestSummary;
	searchText: string;
}

function isInsightTimeMode(value: string | null): value is InsightTimeMode {
	return value === "relative" || value === "absolute";
}

function isInsightRefreshInterval(value: string | null): value is InsightRefreshInterval {
	return value === "off" || value === "10s" || value === "30s" || value === "1m" || value === "5m";
}

function isInsightCheckTypeFilter(value: string | null): value is InsightCheckTypeFilter {
	return value === "all" || value === "ping" || value === "tcp" || value === "traceroute";
}

function isInsightGroupBy(value: string | null): value is InsightGroupBy {
	return value === "check" || value === "probe";
}

function parseInsightUrlState(searchParams: URLSearchParams, now: number): ParsedInsightUrlState {
	const from = parseEpochMs(searchParams.get("from"));
	const to = parseEpochMs(searchParams.get("to"));
	const rawCheckType = searchParams.get("type");
	const rawGroupBy = searchParams.get("groupBy");
	const rawTimeMode = searchParams.get("timeMode");
	const rawTimeRange = searchParams.get("range");
	const rawRefresh = searchParams.get("refresh");
	const hasValidTimeWindow = from !== null && to !== null && from < to;
	const hasValidCheckType = isInsightCheckTypeFilter(rawCheckType);
	const hasValidGroupBy = isInsightGroupBy(rawGroupBy);
	const hasValidTimeMode = isInsightTimeMode(rawTimeMode);
	const hasValidTimeRange = isInsightRelativeRange(rawTimeRange);
	const hasValidRefresh = isInsightRefreshInterval(rawRefresh);
	const legacyRelativeRange = hasValidTimeWindow ? relativeRangeForWindow({ from, to }) : null;
	const timeMode: InsightTimeMode = hasValidTimeMode ? rawTimeMode : legacyRelativeRange ? "relative" : hasValidTimeWindow ? "absolute" : "relative";
	const timeRange: InsightRelativeRange = hasValidTimeRange ? rawTimeRange : (legacyRelativeRange ?? "24h");
	const timeWindow = timeMode === "relative" ? timeWindowForRange(timeRange, now) : hasValidTimeWindow ? { from, to } : timeWindowForRange("24h", now);

	return {
		checkType: hasValidCheckType ? rawCheckType : "all",
		hasValidCheckType,
		groupBy: hasValidGroupBy ? rawGroupBy : "check",
		hasValidGroupBy,
		timeMode,
		hasValidTimeMode,
		timeRange,
		hasValidTimeRange,
		timeWindow,
		hasValidTimeWindow: timeMode === "relative" || hasValidTimeWindow,
		refresh: hasValidRefresh ? rawRefresh : "off",
		hasValidRefresh,
		probeId: searchParams.get("probeId") || "",
		checkId: searchParams.get("checkId") || "",
		runStartedAt: searchParams.get("runStartedAt") || ""
	};
}

function pairKey(probeId: string, checkId: string) {
	return `${probeId}:${checkId}`;
}

function checkTypeFromApi(value: string | undefined): CheckDefinition["type"] {
	switch (value?.toLowerCase()) {
		case "tcp":
			return "TCP";
		case "traceroute":
			return "Traceroute";
		default:
			return "Ping";
	}
}

function pairCheckType(pair: InsightPair): Exclude<InsightCheckTypeFilter, "all"> {
	if (pair.check.type === "TCP") {
		return "tcp";
	}

	return pair.check.type === "Traceroute" ? "traceroute" : "ping";
}

function checkTypeFilterFromCheck(check: CheckDefinition): Exclude<InsightCheckTypeFilter, "all"> {
	if (check.type === "TCP") {
		return "tcp";
	}

	return check.type === "Traceroute" ? "traceroute" : "ping";
}

function matchesCheckType(pair: InsightPair, checkType: InsightCheckTypeFilter) {
	return checkType === "all" || pairCheckType(pair) === checkType;
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

	return pairs.sort((a, b) => a.check.target.localeCompare(b.check.target) || a.probe.name.localeCompare(b.probe.name));
}

function normalizeSearch(value: string) {
	return value.trim().toLowerCase();
}

function latestStartedAtMs(result: ApiLatestResult) {
	const startedAt = new Date(result.latestStartedAt).getTime();
	return Number.isFinite(startedAt) ? startedAt : null;
}

function summarizeLatestResults(results: ApiLatestResult[]): LatestSummary {
	const timeout = results.filter(result => result.latestStatus === "timeout").length;
	const error = results.filter(result => result.latestStatus === "error").length;
	const partial = results.filter(result => result.latestStatus === "partial").length;
	const latestStartedAt = results.reduce<number | null>((latest, result) => {
		const next = latestStartedAtMs(result);
		if (next === null) {
			return latest;
		}

		return latest === null || next > latest ? next : latest;
	}, null);
	let status: GroupStatus = { label: "No data", tone: "warning", rank: 2 };

	if (error || timeout) {
		status = { label: `${formatCount(error + timeout)} failing`, tone: "critical", rank: 0 };
	} else if (partial) {
		status = { label: `${formatCount(partial)} partial`, tone: "warning", rank: 1 };
	} else if (results.length) {
		status = { label: "Reporting", tone: "success", rank: 3 };
	}

	return {
		latestStartedAtMs: latestStartedAt,
		status
	};
}

function groupLatestResultsForPairs(results: ApiLatestResult[], pairs: InsightPair[]) {
	const pairKeys = new Set(pairs.map(pair => pair.key));

	return results.filter(result => pairKeys.has(pairKey(result.probeId, result.checkId)));
}

function buildSearchText(pairs: InsightPair[], label: string, secondary: string) {
	return [
		label,
		secondary,
		...pairs.flatMap(pair => [pair.probe.name, pair.probe.location, pair.probe.asn, pair.probe.provider, pair.check.name, pair.check.target, pair.check.description, ...pair.probe.labelTokens])
	]
		.join(" ")
		.toLowerCase();
}

function buildInsightGroups(pairs: InsightPair[], latestResults: ApiLatestResult[], groupBy: InsightGroupBy): InsightGroupRow[] {
	const grouped = new Map<string, InsightPair[]>();

	for (const pair of pairs) {
		const id = groupBy === "check" ? pair.checkId : pair.probeId;
		grouped.set(id, [...(grouped.get(id) ?? []), pair]);
	}

	return [...grouped.entries()]
		.map(([id, groupPairs]) => {
			const firstPair = groupPairs[0];
			const groupLatestResults = groupLatestResultsForPairs(latestResults, groupPairs);
			const probeCount = new Set(groupPairs.map(pair => pair.probeId)).size;
			const checkCount = new Set(groupPairs.map(pair => pair.checkId)).size;
			const pingCount = groupPairs.filter(pair => pair.check.type === "Ping").length;
			const tcpCount = groupPairs.filter(pair => pair.check.type === "TCP").length;
			const tracerouteCount = groupPairs.filter(pair => pair.check.type === "Traceroute").length;
			const label = groupBy === "check" ? firstPair.check.target : firstPair.probe.name;
			const secondary =
				groupBy === "check" ? `${firstPair.check.name} · ${firstPair.check.type} · ${formatCount(probeCount)} probes` : `${firstPair.probe.location} · ${formatCount(checkCount)} checks`;

			return {
				key: `${groupBy}:${id}`,
				id,
				groupBy,
				label,
				secondary,
				pairs: groupPairs,
				pairCount: groupPairs.length,
				probeCount,
				checkCount,
				pingCount,
				tcpCount,
				tracerouteCount,
				summary: summarizeLatestResults(groupLatestResults),
				searchText: buildSearchText(groupPairs, label, secondary)
			};
		})
		.sort((a, b) => {
			if (a.summary.status.rank !== b.summary.status.rank) {
				return a.summary.status.rank - b.summary.status.rank;
			}

			if (a.summary.latestStartedAtMs !== b.summary.latestStartedAtMs) {
				return (b.summary.latestStartedAtMs ?? 0) - (a.summary.latestStartedAtMs ?? 0);
			}

			return a.label.localeCompare(b.label);
		});
}

function scopePairs(pairs: InsightPair[], checkType: InsightCheckTypeFilter, probeId: string, checkId: string) {
	return pairs.filter(pair => matchesCheckType(pair, checkType) && (!probeId || pair.probeId === probeId) && (!checkId || pair.checkId === checkId));
}

function pairLatestResult(pair: InsightPair, latestResults: ApiLatestResult[]) {
	const latest = latestResults.find(result => result.probeId === pair.probeId && result.checkId === pair.checkId);

	if (!latest) {
		return { status: "No data", time: "-" };
	}

	return {
		status: latest.latestStatus,
		time: new Date(latest.latestStartedAt).toLocaleString(undefined, { month: "short", day: "2-digit", hour: "2-digit", minute: "2-digit" })
	};
}

function statusTone(status: string): BadgeTone {
	if (status === "successful" || status === "Reporting") {
		return "success";
	}

	if (status === "timeout" || status === "error" || status.includes("failing")) {
		return "critical";
	}

	if (status === "partial" || status === "No data") {
		return "warning";
	}

	return "neutral";
}

function GroupTitle({ row }: { row: InsightGroupRow }) {
	return (
		<div className={styles.rowTitle}>
			<strong>{row.label}</strong>
			<span>{row.secondary}</span>
		</div>
	);
}

function InsightPairDetail({
	pair,
	pingInsightData,
	pingSeriesData,
	tcpInsightData,
	tcpSeriesData,
	isPingInsightLoading,
	isPingSeriesLoading,
	isPingFetching,
	isTCPInsightLoading,
	isTCPSeriesLoading,
	isTCPFetching,
	tracerouteInsight,
	tracerouteRuns,
	topologyNodes,
	topologyEdges,
	isTracerouteInsightLoading,
	isRunsLoading,
	isTopologyLoading,
	selectedRunStartedAt,
	onSelectRun,
	onSelectTimeWindow
}: {
	pair: InsightPair | null;
	pingInsightData: PingInsightResponse | undefined;
	pingSeriesData: PingSeriesResponse | undefined;
	tcpInsightData: TcpInsightResponse | undefined;
	tcpSeriesData: TcpSeriesResponse | undefined;
	isPingInsightLoading: boolean;
	isPingSeriesLoading: boolean;
	isPingFetching: boolean;
	isTCPInsightLoading: boolean;
	isTCPSeriesLoading: boolean;
	isTCPFetching: boolean;
	tracerouteInsight: TracerouteInsightResponse | undefined;
	tracerouteRuns: TracerouteResult[];
	topologyNodes: RouteTopologyNode[];
	topologyEdges: RouteTopologyEdge[];
	isTracerouteInsightLoading: boolean;
	isRunsLoading: boolean;
	isTopologyLoading: boolean;
	selectedRunStartedAt: string;
	onSelectRun: (startedAt: string) => void;
	onSelectTimeWindow: (timeWindow: TimeWindow) => void;
}) {
	if (!pair) {
		return null;
	}

	if (pair.check.type === "TCP") {
		return (
			<TcpInsightPanel
				selectedProbe={pair.probe}
				selectedTarget={pair.check}
				insightData={tcpInsightData}
				seriesData={tcpSeriesData}
				isInsightLoading={isTCPInsightLoading}
				isSeriesLoading={isTCPSeriesLoading}
				isFetching={isTCPFetching}
				onSelectTimeWindow={onSelectTimeWindow}
			/>
		);
	}

	return pair.check.type === "Traceroute" ? (
		<TracerouteInsightPanel
			selectedProbe={pair.probe}
			selectedTarget={pair.check}
			insight={tracerouteInsight}
			runs={tracerouteRuns}
			topologyNodes={topologyNodes}
			topologyEdges={topologyEdges}
			isInsightLoading={isTracerouteInsightLoading}
			isRunsLoading={isRunsLoading}
			isTopologyLoading={isTopologyLoading}
			selectedRunStartedAt={selectedRunStartedAt}
			onSelectRun={onSelectRun}
			onSelectTimeWindow={onSelectTimeWindow}
		/>
	) : (
		<PingInsightPanel
			selectedProbe={pair.probe}
			selectedTarget={pair.check}
			insightData={pingInsightData}
			seriesData={pingSeriesData}
			isInsightLoading={isPingInsightLoading}
			isSeriesLoading={isPingSeriesLoading}
			isFetching={isPingFetching}
			onSelectTimeWindow={onSelectTimeWindow}
		/>
	);
}

export function InsightPage() {
	const { projectRef } = useCurrentProject();
	const queryClient = useQueryClient();
	const [searchParams, setSearchParams] = useSearchParams();
	const [search, setSearch] = useState("");
	const [nowMs, setNowMs] = useState(() => Date.now());
	const searchParamString = searchParams.toString();
	const urlState = useMemo(() => parseInsightUrlState(new URLSearchParams(searchParamString), nowMs), [nowMs, searchParamString]);
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
	const isSelectionLoading = Boolean(projectRef) && (assignmentsQuery.isLoading || probesQuery.isLoading || checksQuery.isLoading);
	const knownProbeIds = useMemo(() => new Set(pairs.map(pair => pair.probeId)), [pairs]);
	const knownCheckIds = useMemo(() => new Set(pairs.map(pair => pair.checkId)), [pairs]);
	const hasProbeFocus = Boolean(urlState.probeId);
	const hasCheckFocus = Boolean(urlState.checkId);
	const hasInvalidProbeFocus = hasProbeFocus && !isSelectionLoading && !knownProbeIds.has(urlState.probeId);
	const hasInvalidCheckFocus = hasCheckFocus && !isSelectionLoading && !knownCheckIds.has(urlState.checkId);
	const hasInvalidFocus = hasInvalidProbeFocus || hasInvalidCheckFocus;
	const activeProbeId = hasInvalidProbeFocus ? "" : urlState.probeId;
	const activeCheckId = hasInvalidCheckFocus ? "" : urlState.checkId;
	const timeWindow = urlState.timeWindow;
	const timeMode = urlState.timeMode;
	const timeRange = urlState.timeRange;
	const refresh = urlState.refresh;
	const checkType = urlState.checkType;
	const groupBy = urlState.groupBy;
	const resultWindowFilters = useMemo(() => ({ from: timeWindow.from, to: timeWindow.to }), [timeWindow.from, timeWindow.to]);
	const scopedPairs = useMemo(() => (hasInvalidFocus ? [] : scopePairs(pairs, checkType, activeProbeId, activeCheckId)), [activeCheckId, activeProbeId, checkType, hasInvalidFocus, pairs]);
	const exactPair = activeProbeId && activeCheckId && scopedPairs.length === 1 ? scopedPairs[0] : null;
	const selectedProbe = activeProbeId ? scopedPairs.find(pair => pair.probeId === activeProbeId)?.probe || probes.find(probe => probe.id === activeProbeId) || null : null;
	const selectedCheck = activeCheckId ? scopedPairs.find(pair => pair.checkId === activeCheckId)?.check || checks.find(check => check.id === activeCheckId) || null : null;
	const latestResultFilters = useMemo(
		() => ({
			...(checkType === "all" ? {} : { type: checkType }),
			...(activeProbeId ? { probeId: activeProbeId } : {}),
			...(activeCheckId ? { checkId: activeCheckId } : {})
		}),
		[activeCheckId, activeProbeId, checkType]
	);
	const latestResultsQuery = useQuery({
		...projectQueries.latestResults(projectRef || "", latestResultFilters),
		enabled: Boolean(projectRef && !isSelectionLoading && !hasInvalidFocus)
	});
	const latestResults = useMemo(() => latestResultsQuery.data?.results ?? [], [latestResultsQuery.data?.results]);
	const groups = useMemo(() => buildInsightGroups(scopedPairs, latestResults, groupBy), [groupBy, latestResults, scopedPairs]);
	const searchTerm = normalizeSearch(search);
	const visibleGroups = useMemo(() => (searchTerm ? groups.filter(group => group.searchText.includes(searchTerm)) : groups), [groups, searchTerm]);
	const selectedRunStartedAt = exactPair?.check.type === "Traceroute" ? urlState.runStartedAt : "";
	const canQueryPairDetail = Boolean(projectRef && exactPair);
	const canQueryTracerouteGroup = Boolean(projectRef && !exactPair && scopedPairs.some(pair => pair.check.type === "Traceroute") && !hasInvalidFocus);
	const tracerouteTopologyFilters = useMemo(
		() => ({
			...(activeProbeId ? { probeId: activeProbeId } : {}),
			...(activeCheckId ? { checkId: activeCheckId } : {}),
			...resultWindowFilters,
			limit: 100
		}),
		[activeCheckId, activeProbeId, resultWindowFilters]
	);
	const pingInsightQuery = useQuery({
		...projectQueries.pingInsight(projectRef || "", exactPair?.probeId || "", exactPair?.checkId || "", resultWindowFilters),
		enabled: Boolean(canQueryPairDetail && exactPair?.check.type === "Ping")
	});
	const pingSeriesQuery = useQuery({
		...projectQueries.pingSeries(projectRef || "", exactPair?.probeId || "", exactPair?.checkId || "", resultWindowFilters),
		enabled: Boolean(canQueryPairDetail && exactPair?.check.type === "Ping")
	});
	const tcpInsightQuery = useQuery({
		...projectQueries.tcpInsight(projectRef || "", exactPair?.probeId || "", exactPair?.checkId || "", resultWindowFilters),
		enabled: Boolean(canQueryPairDetail && exactPair?.check.type === "TCP")
	});
	const tcpSeriesQuery = useQuery({
		...projectQueries.tcpSeries(projectRef || "", exactPair?.probeId || "", exactPair?.checkId || "", resultWindowFilters),
		enabled: Boolean(canQueryPairDetail && exactPair?.check.type === "TCP")
	});
	const tracerouteInsightQuery = useQuery({
		...projectQueries.tracerouteInsight(projectRef || "", exactPair?.probeId || "", exactPair?.checkId || "", resultWindowFilters),
		enabled: Boolean(canQueryPairDetail && exactPair?.check.type === "Traceroute")
	});
	const tracerouteRunsQuery = useQuery({
		...projectQueries.tracerouteRuns(projectRef || "", exactPair?.probeId || "", exactPair?.checkId || "", { ...resultWindowFilters, limit: 200 }),
		enabled: Boolean(canQueryPairDetail && exactPair?.check.type === "Traceroute")
	});
	const pairTopologyQuery = useQuery({
		...projectQueries.tracerouteTopology(projectRef || "", {
			probeId: exactPair?.probeId,
			checkId: exactPair?.checkId,
			...resultWindowFilters,
			limit: 100
		}),
		enabled: Boolean(canQueryPairDetail && exactPair?.check.type === "Traceroute")
	});
	const groupTopologyQuery = useQuery({
		...projectQueries.tracerouteTopology(projectRef || "", tracerouteTopologyFilters),
		enabled: canQueryTracerouteGroup
	});
	const groupColumns: DataColumn<InsightGroupRow>[] = [
		{ key: "scope", label: groupBy === "check" ? "Check" : "Probe", render: row => <GroupTitle row={row} /> },
		{
			key: "status",
			label: "Status",
			render: row => <Badge tone={row.summary.status.tone}>{row.summary.status.label}</Badge>
		},
		{ key: "coverage", label: "Coverage", render: row => `${formatCount(row.probeCount)} probes · ${formatCount(row.checkCount)} checks` },
		{ key: "latest", label: "Last seen", render: row => formatEpochMs(row.summary.latestStartedAtMs) }
	];
	const pairColumns: DataColumn<InsightPair>[] = [
		{
			key: "probe",
			label: "Probe",
			render: pair => (
				<div className={styles.rowTitle}>
					<strong>{pair.probe.name}</strong>
					<span>{pair.probe.location}</span>
				</div>
			)
		},
		{
			key: "check",
			label: "Check",
			render: pair => (
				<div className={styles.rowTitle}>
					<strong>{pair.check.target}</strong>
					<span>{pair.check.name}</span>
				</div>
			)
		},
		{ key: "type", label: "Type", render: pair => <Badge tone="accent">{pair.check.type}</Badge> },
		{ key: "interval", label: "Interval", render: pair => pair.check.interval },
		{
			key: "latest",
			label: "Latest",
			render: pair => {
				const latest = pairLatestResult(pair, latestResults);

				return (
					<div className={styles.rowTitle}>
						<strong>
							<Badge tone={statusTone(latest.status)}>{latest.status}</Badge>
						</strong>
						<span>{latest.time}</span>
					</div>
				);
			}
		}
	];
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

		if (!urlState.hasValidTimeMode) {
			setParam("timeMode", timeMode);
		}

		if (timeMode === "relative") {
			setParam("range", timeRange);
			deleteParam("from");
			deleteParam("to");
		} else if (!urlState.hasValidTimeWindow) {
			setParam("from", String(timeWindow.from));
			setParam("to", String(timeWindow.to));
		}

		if (!urlState.hasValidRefresh) {
			setParam("refresh", refresh);
		}

		if (!urlState.hasValidCheckType) {
			setParam("type", checkType);
		}

		if (!urlState.hasValidGroupBy) {
			setParam("groupBy", groupBy);
		}

		deleteParam("mode");
		deleteParam("view");

		if (!exactPair || exactPair.check.type !== "Traceroute") {
			deleteParam("runStartedAt");
		}

		if (changed) {
			setSearchParams(next, { replace: true });
		}
	}, [
		checkType,
		exactPair,
		groupBy,
		projectRef,
		refresh,
		searchParamString,
		setSearchParams,
		timeMode,
		timeRange,
		timeWindow.from,
		timeWindow.to,
		urlState.hasValidCheckType,
		urlState.hasValidGroupBy,
		urlState.hasValidRefresh,
		urlState.hasValidTimeMode,
		urlState.hasValidTimeWindow
	]);

	function updateSearchParams(update: (next: URLSearchParams) => void, options: { replace?: boolean } = {}) {
		const next = new URLSearchParams(searchParamString);
		update(next);
		setSearchParams(next, { replace: options.replace ?? false });
	}

	const refreshProjectQueries = useCallback(() => {
		if (!projectRef) {
			return;
		}

		void queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.detail(projectRef) });
	}, [projectRef, queryClient]);

	const refreshInsight = useCallback(() => {
		if (timeMode === "relative") {
			setNowMs(Date.now());
		}

		refreshProjectQueries();
	}, [refreshProjectQueries, timeMode]);

	useEffect(() => {
		const refreshDuration = refreshDurations[refresh];

		if (!refreshDuration) {
			return;
		}

		const interval = window.setInterval(refreshInsight, refreshDuration);

		return () => window.clearInterval(interval);
	}, [refresh, refreshInsight]);

	function applyRelativeRange(range: InsightRelativeRange) {
		setNowMs(Date.now());
		updateSearchParams(next => {
			next.set("timeMode", "relative");
			next.set("range", range);
			next.delete("from");
			next.delete("to");
			next.delete("runStartedAt");
		});
		refreshProjectQueries();
	}

	function applyAbsoluteWindow(nextTimeWindow: TimeWindow) {
		const from = Math.trunc(nextTimeWindow.from);
		const to = Math.trunc(nextTimeWindow.to);

		if (!Number.isFinite(from) || !Number.isFinite(to) || to <= from) {
			return;
		}

		if (timeMode === "absolute" && timeWindow.from === from && timeWindow.to === to) {
			return;
		}

		updateSearchParams(next => {
			next.set("timeMode", "absolute");
			next.set("from", String(from));
			next.set("to", String(to));
			next.delete("range");
			next.delete("runStartedAt");
		});
	}

	function updateRefresh(nextRefresh: InsightRefreshInterval) {
		updateSearchParams(next => {
			next.set("refresh", nextRefresh);
		});
	}

	function selectGroup(row: InsightGroupRow) {
		updateSearchParams(next => {
			if (row.groupBy === "check") {
				if (activeCheckId === row.id) {
					next.delete("checkId");
				} else {
					next.set("checkId", row.id);
				}
			} else {
				if (activeProbeId === row.id) {
					next.delete("probeId");
				} else {
					next.set("probeId", row.id);
				}
			}
			next.delete("runStartedAt");
		});
	}

	function selectPair(pair: InsightPair) {
		updateSearchParams(next => {
			if (exactPair?.key === pair.key) {
				if (groupBy === "check") {
					next.delete("probeId");
				} else {
					next.delete("checkId");
				}
			} else {
				next.set("probeId", pair.probeId);
				next.set("checkId", pair.checkId);
			}
			next.delete("runStartedAt");
		});
	}

	function clearProbeFocus() {
		updateSearchParams(next => {
			next.delete("probeId");
			next.delete("runStartedAt");
		});
	}

	function clearCheckFocus() {
		updateSearchParams(next => {
			next.delete("checkId");
			next.delete("runStartedAt");
		});
	}

	function resetScope() {
		setSearch("");
		updateSearchParams(next => {
			next.set("type", "all");
			next.set("groupBy", "check");
			next.delete("probeId");
			next.delete("checkId");
			next.delete("runStartedAt");
		});
	}

	const scopeTitle = exactPair
		? `${exactPair.probe.name} -> ${exactPair.check.target}`
		: selectedProbe && selectedCheck
			? "No active assignment"
			: selectedProbe
				? selectedProbe.name
				: selectedCheck
					? selectedCheck.target
					: "Project scope";
	const groupTopologyTitle = selectedProbe
		? `${selectedProbe.name} route graph`
		: selectedCheck
			? `${selectedCheck.target} route graph`
			: checkType === "traceroute"
				? "Traceroute route graph"
				: "Project route graph";
	const pairDetail = (
		<InsightPairDetail
			pair={exactPair}
			pingInsightData={pingInsightQuery.data}
			pingSeriesData={pingSeriesQuery.data}
			tcpInsightData={tcpInsightQuery.data}
			tcpSeriesData={tcpSeriesQuery.data}
			isPingInsightLoading={pingInsightQuery.isLoading}
			isPingSeriesLoading={pingSeriesQuery.isLoading}
			isPingFetching={pingInsightQuery.isFetching || pingSeriesQuery.isFetching}
			isTCPInsightLoading={tcpInsightQuery.isLoading}
			isTCPSeriesLoading={tcpSeriesQuery.isLoading}
			isTCPFetching={tcpInsightQuery.isFetching || tcpSeriesQuery.isFetching}
			tracerouteInsight={tracerouteInsightQuery.data}
			tracerouteRuns={tracerouteRunsQuery.data?.runs ?? []}
			topologyNodes={pairTopologyQuery.data?.nodes ?? []}
			topologyEdges={pairTopologyQuery.data?.edges ?? []}
			isTracerouteInsightLoading={tracerouteInsightQuery.isLoading}
			isRunsLoading={tracerouteRunsQuery.isLoading}
			isTopologyLoading={pairTopologyQuery.isLoading}
			selectedRunStartedAt={selectedRunStartedAt}
			onSelectRun={startedAt =>
				updateSearchParams(next => {
					next.set("runStartedAt", startedAt);
				})
			}
			onSelectTimeWindow={applyAbsoluteWindow}
		/>
	);

	return (
		<PageStack>
			<ScreenHeader title="Insight" />

			<Panel
				tone="glass"
				title={scopeTitle}
				actions={
					<Button variant="outline" size="sm" onClick={resetScope}>
						Reset scope
					</Button>
				}
			>
				<FilterGrid className={styles.scopeBar}>
					<InsightTimeControl
						className={styles.scopeTimeControl}
						timeMode={timeMode}
						timeRange={timeRange}
						timeWindow={timeWindow}
						refresh={refresh}
						onApplyRelative={applyRelativeRange}
						onApplyAbsolute={applyAbsoluteWindow}
						onRefresh={refreshInsight}
						onRefreshChange={updateRefresh}
					/>
					<SegmentedControl
						label="Type"
						value={checkType}
						options={checkTypeOptions}
						onChange={nextType => {
							updateSearchParams(next => {
								next.set("type", nextType);
								next.delete("runStartedAt");
								if (nextType !== "all" && selectedCheck && checkTypeFilterFromCheck(selectedCheck) !== nextType) {
									next.delete("checkId");
								}
							});
						}}
					/>
					<SegmentedControl
						label="Group"
						value={groupBy}
						options={groupByOptions}
						onChange={nextGroupBy => {
							updateSearchParams(next => {
								next.set("groupBy", nextGroupBy);
							});
						}}
					/>
					<TextField label="Search" placeholder="check, probe, target, label, location" value={search} onChange={event => setSearch(event.currentTarget.value)} />
				</FilterGrid>
				<div className={styles.focusChips} aria-label="Active Insight scope">
					{hasProbeFocus ? (
						<FocusChip
							label="Probe"
							value={hasInvalidProbeFocus ? `Unknown probe ${urlState.probeId}` : selectedProbe?.name || urlState.probeId}
							invalid={hasInvalidProbeFocus}
							onClear={clearProbeFocus}
						/>
					) : null}
					{hasCheckFocus ? (
						<FocusChip
							label="Check"
							value={hasInvalidCheckFocus ? `Unknown check ${urlState.checkId}` : selectedCheck?.target || urlState.checkId}
							invalid={hasInvalidCheckFocus}
							onClear={clearCheckFocus}
						/>
					) : null}
					{!hasProbeFocus && !hasCheckFocus ? <span className={styles.scopeHint}>All active assignments in this project are included.</span> : null}
				</div>
			</Panel>

			{isSelectionLoading && !pairs.length ? (
				<Panel tone="deep" title="Loading active paths">
					<BodyCopy>Loading probe-check assignments for this project.</BodyCopy>
				</Panel>
			) : !pairs.length ? (
				<Panel tone="deep" title="No active paths">
					<BodyCopy>Create or refresh check assignments before opening result insight.</BodyCopy>
				</Panel>
			) : hasInvalidFocus ? (
				<Panel tone="deep" title="The shared scope is no longer valid">
					<BodyCopy>Clear the unknown probe or check chip to return to active assignments.</BodyCopy>
				</Panel>
			) : (
				<>
					<Panel tone="glass" title={groupBy === "check" ? `${formatCount(visibleGroups.length)} checks` : `${formatCount(visibleGroups.length)} probes`}>
						<DataTable
							columns={groupColumns}
							rows={visibleGroups}
							density="compact"
							minWidth="58rem"
							maxHeight="28rem"
							ariaLabel="Insight grouped scope"
							getRowKey={row => row.key}
							getRowAriaLabel={row => `Focus ${row.label}`}
							selectedKey={selectedCheck && groupBy === "check" ? `check:${selectedCheck.id}` : selectedProbe && groupBy === "probe" ? `probe:${selectedProbe.id}` : undefined}
							onRowClick={selectGroup}
							emptyLabel={searchTerm ? "No groups match the current search." : "No assignments match the current scope."}
						/>
					</Panel>

					<Panel tone="glass" title={exactPair ? "Selected assignment" : `${formatCount(scopedPairs.length)} assignments in scope`}>
						<DataTable
							columns={pairColumns}
							rows={scopedPairs}
							density="compact"
							minWidth="52rem"
							maxHeight="24rem"
							ariaLabel="Insight assignments"
							getRowKey={row => row.key}
							getRowAriaLabel={row => `Open ${row.probe.name} to ${row.check.target}`}
							selectedKey={exactPair?.key}
							onRowClick={selectPair}
							emptyLabel="No assignments match the current scope."
						/>
					</Panel>

					{canQueryTracerouteGroup ? (
						<GroupTopologyPanel title={groupTopologyTitle} nodes={groupTopologyQuery.data?.nodes ?? []} edges={groupTopologyQuery.data?.edges ?? []} isLoading={groupTopologyQuery.isLoading} />
					) : null}

					{pairDetail}
				</>
			)}
		</PageStack>
	);
}

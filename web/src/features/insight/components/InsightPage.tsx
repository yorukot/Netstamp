import { formatInterval, mapApiChecks } from "@/features/checks/api/checkAdapters";
import { type CheckDefinition } from "@/features/checks/data/checks";
import { GroupTopologyPanel } from "@/features/insight/components/GroupTopologyPanel";
import { AssignmentMultiSelect, FocusChip, InsightTimeControl, ScopeSelect, SegmentedControl, type AssignmentSelectOption } from "@/features/insight/components/InsightControls";
import { MultiSeriesInsightPanel } from "@/features/insight/components/MultiSeriesInsightPanel";
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
import { type ApiProjectAssignment, type PingInsightResponse, type PingSeriesResponse, type TcpInsightResponse, type TcpSeriesResponse } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { BodyCopy } from "@/shared/components/BodyCopy";
import { FilterGrid } from "@/shared/components/FilterGrid";
import { LoadingState } from "@/shared/components/LoadingState";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { formatCount } from "@/shared/utils/insightFormatters";
import {
	isRelativeTimeRange as isInsightRelativeRange,
	parseEpochMs,
	relativeRangeForTimeWindow as relativeRangeForWindow,
	timeWindowForRelativeRange as timeWindowForRange
} from "@/shared/utils/timeRanges";
import { Button, Panel, SelectField } from "@netstamp/ui";
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
	const assignmentKeys = Array.from(new Set(searchParams.getAll("assignment").filter(value => value.includes(":"))));

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
		assignmentKeys,
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
		provider: "Unlabeled",
		region: location,
		ipFamily: "-",
		lastHeartbeat: "never",
		lastHeartbeatAt: null,
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

function scopePairs(pairs: InsightPair[], checkType: InsightCheckTypeFilter, probeId: string, checkId: string) {
	return pairs.filter(pair => matchesCheckType(pair, checkType) && (!probeId || pair.probeId === probeId) && (!checkId || pair.checkId === checkId));
}

function assignmentLabel(pair: InsightPair) {
	return `${pair.check.name} / ${pair.probe.name} / ${pair.check.target}`;
}

function assignmentMeta(pair: InsightPair) {
	return `${pair.check.type} · ${pair.probe.location}`;
}

function assignmentSelectOption(pair: InsightPair): AssignmentSelectOption {
	const label = assignmentLabel(pair);
	const meta = assignmentMeta(pair);
	const searchText = normalizeSearch(
		[label, meta, pair.probe.name, pair.probe.location, pair.probe.provider, pair.check.name, pair.check.target, pair.check.description, ...pair.probe.labelTokens].join(" ")
	);

	return {
		value: pair.key,
		label,
		meta,
		searchText
	};
}

function uniqueProbeOptions(probes: Probe[], pairs: InsightPair[]): AssignmentSelectOption[] {
	const counts = new Map<string, number>();
	const options = new Map<string, AssignmentSelectOption>();

	for (const pair of pairs) {
		counts.set(pair.probeId, (counts.get(pair.probeId) ?? 0) + 1);
	}

	for (const probe of probes) {
		options.set(probe.id, {
			value: probe.id,
			label: probe.name,
			meta: probe.location,
			searchText: normalizeSearch([probe.name, probe.location, probe.provider, ...probe.labelTokens].join(" "))
		});
	}

	for (const pair of pairs) {
		if (options.has(pair.probeId)) {
			continue;
		}

		options.set(pair.probeId, {
			value: pair.probeId,
			label: pair.probe.name,
			meta: pair.probe.location,
			searchText: normalizeSearch([pair.probe.name, pair.probe.location, pair.probe.provider, ...pair.probe.labelTokens].join(" "))
		});
	}

	return [...options.values()]
		.map(option => ({
			...option,
			meta: `${option.meta} · ${formatCount(counts.get(option.value) ?? 0)} assignments`
		}))
		.sort((a, b) => a.label.localeCompare(b.label));
}

function uniqueCheckOptions(checks: CheckDefinition[], pairs: InsightPair[]): AssignmentSelectOption[] {
	const counts = new Map<string, number>();
	const options = new Map<string, AssignmentSelectOption>();

	for (const pair of pairs) {
		counts.set(pair.checkId, (counts.get(pair.checkId) ?? 0) + 1);
	}

	for (const check of checks) {
		options.set(check.id, {
			value: check.id,
			label: check.target,
			meta: `${check.name} · ${check.type}`,
			searchText: normalizeSearch([check.name, check.target, check.description, check.type].join(" "))
		});
	}

	for (const pair of pairs) {
		if (options.has(pair.checkId)) {
			continue;
		}

		options.set(pair.checkId, {
			value: pair.checkId,
			label: pair.check.target,
			meta: `${pair.check.name} · ${pair.check.type}`,
			searchText: normalizeSearch([pair.check.name, pair.check.target, pair.check.description, pair.check.type].join(" "))
		});
	}

	return [...options.values()]
		.map(option => ({
			...option,
			meta: `${option.meta} · ${formatCount(counts.get(option.value) ?? 0)} assignments`
		}))
		.sort((a, b) => a.label.localeCompare(b.label));
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
	onSelectTimeWindow: (timeWindow: TimeWindow) => void;
}) {
	if (!pair) {
		return null;
	}

	if (pair.check.type === "Traceroute") {
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

	return (
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
	const pairsByKey = useMemo(() => new Map(pairs.map(pair => [pair.key, pair])), [pairs]);
	const knownProbeIds = useMemo(() => new Set([...probes.map(probe => probe.id), ...pairs.map(pair => pair.probeId)]), [pairs, probes]);
	const knownCheckIds = useMemo(() => new Set([...checks.map(check => check.id), ...pairs.map(pair => pair.checkId)]), [checks, pairs]);
	const timeWindow = urlState.timeWindow;
	const timeMode = urlState.timeMode;
	const timeRange = urlState.timeRange;
	const refresh = urlState.refresh;
	const checkType = urlState.checkType;
	const groupBy = urlState.groupBy;
	const hasAssignmentSelection = urlState.assignmentKeys.length > 0;
	const hasProbeFocus = Boolean(urlState.probeId);
	const hasCheckFocus = Boolean(urlState.checkId);
	const hasInvalidProbeFocus = hasProbeFocus && !isSelectionLoading && !knownProbeIds.has(urlState.probeId);
	const hasInvalidCheckFocus = hasCheckFocus && !isSelectionLoading && !knownCheckIds.has(urlState.checkId);
	const requestedAssignmentKeys = useMemo(
		() => (hasAssignmentSelection ? urlState.assignmentKeys : urlState.probeId && urlState.checkId ? [pairKey(urlState.probeId, urlState.checkId)] : []),
		[hasAssignmentSelection, urlState.assignmentKeys, urlState.checkId, urlState.probeId]
	);
	const unknownAssignmentKeys = requestedAssignmentKeys.filter(key => !pairsByKey.has(key));
	const hasInvalidAssignmentFocus = requestedAssignmentKeys.length > 0 && !isSelectionLoading && unknownAssignmentKeys.length > 0;
	const hasInvalidFocus = hasInvalidProbeFocus || hasInvalidCheckFocus || hasInvalidAssignmentFocus;
	const activeProbeId = hasInvalidProbeFocus ? "" : urlState.probeId;
	const activeCheckId = hasInvalidCheckFocus ? "" : urlState.checkId;
	const resultWindowFilters = useMemo(() => ({ from: timeWindow.from, to: timeWindow.to }), [timeWindow.from, timeWindow.to]);
	const typeFilteredPairs = useMemo(() => pairs.filter(pair => matchesCheckType(pair, checkType)), [checkType, pairs]);
	const typeFilteredChecks = useMemo(() => checks.filter(check => checkType === "all" || checkTypeFilterFromCheck(check) === checkType), [checkType, checks]);
	const selectedPairKeys = useMemo(
		() =>
			requestedAssignmentKeys.filter(key => {
				const pair = pairsByKey.get(key);
				return pair ? matchesCheckType(pair, checkType) && (!activeProbeId || pair.probeId === activeProbeId) && (!activeCheckId || pair.checkId === activeCheckId) : false;
			}),
		[activeCheckId, activeProbeId, checkType, pairsByKey, requestedAssignmentKeys]
	);
	const selectedPairs = useMemo(() => selectedPairKeys.map(key => pairsByKey.get(key)).filter((pair): pair is InsightPair => Boolean(pair)), [pairsByKey, selectedPairKeys]);
	const legacyScopedPairs = useMemo(() => (hasInvalidFocus ? [] : scopePairs(pairs, checkType, activeProbeId, activeCheckId)), [activeCheckId, activeProbeId, checkType, hasInvalidFocus, pairs]);
	const hasResultScope = selectedPairKeys.length > 0 || Boolean(activeProbeId || activeCheckId);
	const scopedPairs = useMemo(() => (hasResultScope ? (selectedPairs.length ? selectedPairs : legacyScopedPairs) : []), [hasResultScope, legacyScopedPairs, selectedPairs]);
	const exactPair = selectedPairs.length === 1 ? selectedPairs[0] : activeProbeId && activeCheckId && scopedPairs.length === 1 ? scopedPairs[0] : null;
	const selectedProbe = exactPair?.probe ?? (activeProbeId ? scopedPairs.find(pair => pair.probeId === activeProbeId)?.probe || probes.find(probe => probe.id === activeProbeId) || null : null);
	const selectedCheck = exactPair?.check ?? (activeCheckId ? scopedPairs.find(pair => pair.checkId === activeCheckId)?.check || checks.find(check => check.id === activeCheckId) || null : null);
	const scopeOptions = useMemo(
		() => (groupBy === "check" ? uniqueCheckOptions(typeFilteredChecks, typeFilteredPairs) : uniqueProbeOptions(probes, typeFilteredPairs)),
		[groupBy, probes, typeFilteredChecks, typeFilteredPairs]
	);
	const assignmentOptions = useMemo(() => legacyScopedPairs.map(pair => assignmentSelectOption(pair)), [legacyScopedPairs]);
	const canQueryPairDetail = Boolean(projectRef && exactPair);
	const canQueryTracerouteDetail = Boolean(canQueryPairDetail && exactPair?.check.type === "Traceroute");
	const topologyProbeId = activeProbeId;
	const topologyCheckId = activeCheckId;
	const topologyProbe = topologyProbeId ? legacyScopedPairs.find(pair => pair.probeId === topologyProbeId)?.probe || probes.find(probe => probe.id === topologyProbeId) || null : null;
	const topologyCheck = topologyCheckId ? legacyScopedPairs.find(pair => pair.checkId === topologyCheckId)?.check || checks.find(check => check.id === topologyCheckId) || null : null;
	const hasTopologyScope = Boolean(topologyProbeId || topologyCheckId);
	const canQueryTracerouteGroup = Boolean(projectRef && hasResultScope && hasTopologyScope && scopedPairs.some(pair => pair.check.type === "Traceroute") && !hasInvalidFocus);
	const tracerouteTopologyFilters = useMemo(
		() => ({
			...(topologyProbeId ? { probeId: topologyProbeId } : {}),
			...(topologyCheckId ? { checkId: topologyCheckId } : {}),
			...resultWindowFilters,
			limit: 100
		}),
		[resultWindowFilters, topologyCheckId, topologyProbeId]
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
		enabled: canQueryTracerouteDetail
	});
	const tracerouteRunsQuery = useQuery({
		...projectQueries.tracerouteRuns(projectRef || "", exactPair?.probeId || "", exactPair?.checkId || "", { ...resultWindowFilters, limit: 200 }),
		enabled: canQueryTracerouteDetail
	});
	const groupTopologyQuery = useQuery({
		...projectQueries.tracerouteTopology(projectRef || "", tracerouteTopologyFilters),
		enabled: canQueryTracerouteGroup
	});
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

		if (urlState.assignmentKeys.length) {
			const normalizedAssignmentKeys = Array.from(new Set(urlState.assignmentKeys));
			if (next.getAll("assignment").join("\u0000") !== normalizedAssignmentKeys.join("\u0000")) {
				next.delete("assignment");
				normalizedAssignmentKeys.forEach(key => next.append("assignment", key));
				changed = true;
			}
		}

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
		urlState.assignmentKeys,
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

	function writeAssignmentParams(next: URLSearchParams, keys: string[]) {
		next.delete("assignment");
		keys.forEach(key => next.append("assignment", key));
	}

	function selectAssignments(keys: string[]) {
		updateSearchParams(next => {
			writeAssignmentParams(next, keys);
			next.delete("runStartedAt");
		});
	}

	function selectGroupScope(value: string) {
		updateSearchParams(next => {
			next.delete("assignment");
			next.delete("runStartedAt");

			if (groupBy === "check") {
				next.delete("probeId");
				if (value) {
					next.set("checkId", value);
				} else {
					next.delete("checkId");
				}
				return;
			}

			next.delete("checkId");
			if (value) {
				next.set("probeId", value);
			} else {
				next.delete("probeId");
			}
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
		updateSearchParams(next => {
			next.set("type", "all");
			next.set("groupBy", "check");
			next.delete("assignment");
			next.delete("probeId");
			next.delete("checkId");
			next.delete("runStartedAt");
		});
	}

	const scopeTitle =
		selectedPairs.length > 1
			? `${formatCount(selectedPairs.length)} selected assignments`
			: exactPair
				? `${exactPair.probe.name} -> ${exactPair.check.target}`
				: selectedProbe && selectedCheck
					? "No active assignment"
					: selectedProbe
						? selectedProbe.name
						: selectedCheck
							? selectedCheck.target
							: "Select scope";
	const groupTopologyTitle =
		topologyProbe && topologyCheck
			? `${topologyProbe.name} -> ${topologyCheck.target} route graph`
			: topologyProbe
				? `${topologyProbe.name} route graph`
				: topologyCheck
					? `${topologyCheck.target} route graph`
					: "Selected route graph";
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
			onSelectTimeWindow={applyAbsoluteWindow}
		/>
	);
	const tracerouteDetail =
		exactPair?.check.type === "Traceroute" ? (
			<TracerouteInsightPanel
				selectedProbe={exactPair.probe}
				selectedTarget={exactPair.check}
				insight={tracerouteInsightQuery.data}
				runs={tracerouteRunsQuery.data?.runs ?? []}
				topologyNodes={[]}
				topologyEdges={[]}
				isInsightLoading={tracerouteInsightQuery.isLoading || tracerouteInsightQuery.isFetching}
				isRunsLoading={tracerouteRunsQuery.isLoading || tracerouteRunsQuery.isFetching}
				isTopologyLoading={false}
				selectedRunStartedAt={urlState.runStartedAt}
				showTopology={false}
				onSelectRun={startedAt =>
					updateSearchParams(next => {
						next.set("runStartedAt", startedAt);
					})
				}
				onSelectTimeWindow={applyAbsoluteWindow}
			/>
		) : null;
	const detailPanels =
		selectedPairs.length > 1 ? (
			<MultiSeriesInsightPanel projectRef={projectRef} pairs={selectedPairs} filters={resultWindowFilters} onSelectTimeWindow={applyAbsoluteWindow} />
		) : exactPair?.check.type === "Traceroute" ? (
			tracerouteDetail
		) : (
			pairDetail
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
					<SelectField
						label="Type"
						value={checkType}
						options={checkTypeOptions}
						onChange={event => {
							const nextType = event.currentTarget.value as InsightCheckTypeFilter;

							updateSearchParams(next => {
								next.set("type", nextType);
								next.delete("runStartedAt");
								if (selectedPairKeys.length) {
									const nextKeys = selectedPairKeys.filter(key => {
										const pair = pairsByKey.get(key);
										return pair ? matchesCheckType(pair, nextType) : false;
									});
									writeAssignmentParams(next, nextKeys);
								}
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
								next.delete("assignment");
								next.delete("runStartedAt");
								if (nextGroupBy === "check") {
									next.delete("probeId");
								} else {
									next.delete("checkId");
								}
							});
						}}
					/>
				</FilterGrid>
				<div className={styles.scopeSelectorRow}>
					<ScopeSelect
						label={groupBy === "check" ? "Check" : "Probe"}
						placeholder={groupBy === "check" ? "Select check" : "Select probe"}
						options={scopeOptions}
						value={groupBy === "check" ? activeCheckId : activeProbeId}
						disabled={isSelectionLoading || !scopeOptions.length}
						onChange={selectGroupScope}
					/>
				</div>
				<div className={styles.assignmentSelectorRow}>
					<AssignmentMultiSelect
						label="Assignments"
						placeholder="Type probe, check, target, label, or location"
						options={assignmentOptions}
						selectedValues={selectedPairKeys}
						disabled={isSelectionLoading || !assignmentOptions.length}
						onChange={selectAssignments}
					/>
				</div>
				<div className={styles.focusChips} aria-label="Active Insight scope">
					{hasInvalidAssignmentFocus ? (
						<FocusChip
							label="Assignment"
							value={unknownAssignmentKeys.length === 1 ? `Unknown assignment ${unknownAssignmentKeys[0]}` : `${formatCount(unknownAssignmentKeys.length)} unknown assignments`}
							invalid
							onClear={() => selectAssignments([])}
						/>
					) : null}
					{hasInvalidProbeFocus ? (
						<FocusChip
							label="Probe"
							value={hasInvalidProbeFocus ? `Unknown probe ${urlState.probeId}` : selectedProbe?.name || urlState.probeId}
							invalid={hasInvalidProbeFocus}
							onClear={clearProbeFocus}
						/>
					) : null}
					{hasInvalidCheckFocus ? (
						<FocusChip
							label="Check"
							value={hasInvalidCheckFocus ? `Unknown check ${urlState.checkId}` : selectedCheck?.target || urlState.checkId}
							invalid={hasInvalidCheckFocus}
							onClear={clearCheckFocus}
						/>
					) : null}
					{!hasResultScope && !hasInvalidFocus ? <span className={styles.scopeHint}>Select a check, probe, or assignment to inspect results.</span> : null}
					{!selectedPairKeys.length && (activeProbeId || activeCheckId) && !hasInvalidFocus ? (
						<span className={styles.scopeHint}>{formatCount(legacyScopedPairs.length)} assignments available in this scope.</span>
					) : null}
					{selectedPairKeys.length ? <span className={styles.scopeHint}>{formatCount(selectedPairKeys.length)} assignments selected.</span> : null}
				</div>
			</Panel>

			{isSelectionLoading && !pairs.length ? (
				<Panel tone="deep" title="Loading active paths">
					<LoadingState label="Loading active paths" detail="Fetching probe-check assignments for this project." size="compact" />
				</Panel>
			) : !pairs.length ? (
				<Panel tone="deep" title="No active paths">
					<BodyCopy>Create or refresh check assignments before opening result insight.</BodyCopy>
				</Panel>
			) : hasInvalidFocus ? (
				<Panel tone="deep" title="The shared scope is no longer valid">
					<BodyCopy>Clear the unknown probe or check chip to return to active assignments.</BodyCopy>
				</Panel>
			) : !hasResultScope ? null : (
				<>
					{canQueryTracerouteGroup ? (
						<GroupTopologyPanel
							title={groupTopologyTitle}
							nodes={groupTopologyQuery.data?.nodes ?? []}
							edges={groupTopologyQuery.data?.edges ?? []}
							isLoading={groupTopologyQuery.isLoading || groupTopologyQuery.isFetching}
						/>
					) : null}

					{detailPanels}
				</>
			)}
		</PageStack>
	);
}

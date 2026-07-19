import { mapApiChecks } from "@/features/checks/api/checkAdapters";
import { GroupTopologyPanel } from "@/features/insight/components/GroupTopologyPanel";
import { HttpInsightPanel } from "@/features/insight/components/HttpInsightPanel";
import { CheckScopeSelect, InsightTimeControl, ProbeScopeSelect } from "@/features/insight/components/InsightControls";
import { MultiSeriesInsightPanel } from "@/features/insight/components/MultiSeriesInsightPanel";
import { PingInsightPanel } from "@/features/insight/components/PingInsightPanel";
import { TcpInsightPanel } from "@/features/insight/components/TcpInsightPanel";
import { TracerouteInsightPanel } from "@/features/insight/components/TracerouteInsightPanel";
import { latestHTTPResultForPair, latestHttpResultMap } from "@/features/insight/data/httpResultData";
import { buildInsightPairs, parseInsightUrlState, refreshDurations } from "@/features/insight/insightPageState";
import { type InsightPair, type InsightRefreshInterval, type InsightRelativeRange, type TimeWindow } from "@/features/insight/insightTypes";
import { mapApiProbes } from "@/features/probes/api/probeAdapters";
import { projectQueries } from "@/shared/api/queries";
import { apiQueryKeys } from "@/shared/api/queryKeys";
import {
	type HttpInsightResponse,
	type HttpSeriesResponse,
	type LatestHttpResult,
	type PingInsightResponse,
	type PingSeriesResponse,
	type TcpInsightResponse,
	type TcpSeriesResponse
} from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { formatCount } from "@/shared/utils/insightFormatters";
import { BodyCopy, Button, Panel, Spinner } from "@netstamp/ui";
import { ArrowRightIcon } from "@phosphor-icons/react/dist/csr/ArrowRight";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { useSearchParams } from "react-router-dom";
import styles from "./InsightPage.module.css";

function InsightPairDetail({
	pair,
	pingInsightData,
	pingSeriesData,
	tcpInsightData,
	tcpSeriesData,
	httpInsightData,
	httpSeriesData,
	httpLatestResult,
	nowMs,
	isPingInsightLoading,
	isPingSeriesLoading,
	isPingFetching,
	isTCPInsightLoading,
	isTCPSeriesLoading,
	isTCPFetching,
	isHTTPLoading,
	isHTTPLatestLoading,
	isHTTPFetching,
	onSelectTimeWindow
}: {
	pair: InsightPair | null;
	pingInsightData: PingInsightResponse | undefined;
	pingSeriesData: PingSeriesResponse | undefined;
	tcpInsightData: TcpInsightResponse | undefined;
	tcpSeriesData: TcpSeriesResponse | undefined;
	httpInsightData: HttpInsightResponse | undefined;
	httpSeriesData: HttpSeriesResponse | undefined;
	httpLatestResult: LatestHttpResult | undefined;
	nowMs: number;
	isPingInsightLoading: boolean;
	isPingSeriesLoading: boolean;
	isPingFetching: boolean;
	isTCPInsightLoading: boolean;
	isTCPSeriesLoading: boolean;
	isTCPFetching: boolean;
	isHTTPLoading: boolean;
	isHTTPLatestLoading: boolean;
	isHTTPFetching: boolean;
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
	if (pair.check.type === "HTTP") {
		return (
			<HttpInsightPanel
				selectedProbe={pair.probe}
				selectedTarget={pair.check}
				insightData={httpInsightData}
				seriesData={httpSeriesData}
				latestResult={httpLatestResult}
				nowMs={nowMs}
				isLoading={isHTTPLoading}
				isLatestLoading={isHTTPLatestLoading}
				isFetching={isHTTPFetching}
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
	const { t } = useTranslation("insight");
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
	const selectableProbes = useMemo(() => {
		const options = new Map(probes.map(probe => [probe.id, probe]));
		pairs.forEach(pair => options.set(pair.probeId, options.get(pair.probeId) ?? pair.probe));
		return Array.from(options.values()).sort((a, b) => a.name.localeCompare(b.name));
	}, [pairs, probes]);
	const selectableChecks = useMemo(() => {
		const options = new Map(checks.map(check => [check.id, check]));
		pairs.forEach(pair => options.set(pair.checkId, options.get(pair.checkId) ?? pair.check));
		return Array.from(options.values()).sort((a, b) => a.name.localeCompare(b.name) || a.target.localeCompare(b.target));
	}, [checks, pairs]);
	const knownProbeIds = useMemo(() => new Set(selectableProbes.map(probe => probe.id)), [selectableProbes]);
	const knownCheckIds = useMemo(() => new Set(selectableChecks.map(check => check.id)), [selectableChecks]);
	const timeWindow = urlState.timeWindow;
	const timeMode = urlState.timeMode;
	const timeRange = urlState.timeRange;
	const refresh = urlState.refresh;
	const legacySelectedPairs = useMemo(() => urlState.assignmentKeys.map(key => pairsByKey.get(key)).filter((pair): pair is InsightPair => Boolean(pair)), [pairsByKey, urlState.assignmentKeys]);
	const requestedProbeIds = useMemo(
		() => (urlState.probeIds.length ? urlState.probeIds : Array.from(new Set(legacySelectedPairs.map(pair => pair.probeId)))),
		[legacySelectedPairs, urlState.probeIds]
	);
	const requestedCheckIds = useMemo(
		() => (urlState.checkIds.length ? urlState.checkIds : Array.from(new Set(legacySelectedPairs.map(pair => pair.checkId)))),
		[legacySelectedPairs, urlState.checkIds]
	);
	const unknownProbeIds = requestedProbeIds.filter(id => !knownProbeIds.has(id));
	const unknownCheckIds = requestedCheckIds.filter(id => !knownCheckIds.has(id));
	const unknownAssignmentKeys = urlState.assignmentKeys.filter(key => !pairsByKey.has(key));
	const hasInvalidProbeFocus = !isSelectionLoading && unknownProbeIds.length > 0;
	const hasInvalidCheckFocus = !isSelectionLoading && unknownCheckIds.length > 0;
	const hasInvalidAssignmentFocus = !isSelectionLoading && unknownAssignmentKeys.length > 0;
	const hasInvalidFocus = hasInvalidProbeFocus || hasInvalidCheckFocus || hasInvalidAssignmentFocus;
	const activeProbeIds = useMemo(() => requestedProbeIds.filter(id => knownProbeIds.has(id)), [knownProbeIds, requestedProbeIds]);
	const activeCheckIds = useMemo(() => requestedCheckIds.filter(id => knownCheckIds.has(id)), [knownCheckIds, requestedCheckIds]);
	const activeProbeIdSet = useMemo(() => new Set(activeProbeIds), [activeProbeIds]);
	const activeCheckIdSet = useMemo(() => new Set(activeCheckIds), [activeCheckIds]);
	const usesLegacyAssignmentScope = !urlState.probeIds.length && !urlState.checkIds.length && urlState.assignmentKeys.length > 0;
	const resultWindowFilters = useMemo(() => ({ from: timeWindow.from, to: timeWindow.to }), [timeWindow.from, timeWindow.to]);
	const availableCheckIds = useMemo(
		() => new Set(pairs.filter(pair => !activeProbeIds.length || activeProbeIdSet.has(pair.probeId)).map(pair => pair.checkId)),
		[activeProbeIdSet, activeProbeIds.length, pairs]
	);
	const availableChecks = useMemo(() => selectableChecks.filter(check => availableCheckIds.has(check.id)), [availableCheckIds, selectableChecks]);
	const hasResultScope = activeProbeIds.length > 0 || activeCheckIds.length > 0;
	const selectedPairs = useMemo(
		() =>
			hasResultScope && !hasInvalidFocus
				? usesLegacyAssignmentScope
					? legacySelectedPairs
					: pairs.filter(pair => (!activeProbeIds.length || activeProbeIdSet.has(pair.probeId)) && (!activeCheckIds.length || activeCheckIdSet.has(pair.checkId)))
				: [],
		[activeCheckIdSet, activeCheckIds.length, activeProbeIdSet, activeProbeIds.length, hasInvalidFocus, hasResultScope, legacySelectedPairs, pairs, usesLegacyAssignmentScope]
	);
	const exactPair = selectedPairs.length === 1 ? selectedPairs[0] : null;
	const hasSelectedHTTPPairs = selectedPairs.some(pair => pair.check.type === "HTTP");
	const selectedProbe = activeProbeIds.length === 1 ? (selectableProbes.find(probe => probe.id === activeProbeIds[0]) ?? null) : null;
	const selectedCheck = activeCheckIds.length === 1 ? (selectableChecks.find(check => check.id === activeCheckIds[0]) ?? null) : null;
	const canQueryPairDetail = Boolean(projectRef && exactPair);
	const canQueryTracerouteDetail = Boolean(canQueryPairDetail && exactPair?.check.type === "Traceroute");
	const topologyProbeId = activeProbeIds.length === 1 ? activeProbeIds[0] : "";
	const topologyCheckId = activeCheckIds.length === 1 ? activeCheckIds[0] : "";
	const topologyProbe = topologyProbeId ? (selectableProbes.find(probe => probe.id === topologyProbeId) ?? null) : null;
	const topologyCheck = topologyCheckId ? (selectableChecks.find(check => check.id === topologyCheckId) ?? null) : null;
	const hasTopologyScope = Boolean(topologyProbeId || topologyCheckId);
	const canQueryTracerouteGroup = Boolean(projectRef && hasResultScope && hasTopologyScope && selectedPairs.some(pair => pair.check.type === "Traceroute") && !hasInvalidFocus);
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
	const httpInsightQuery = useQuery({
		...projectQueries.httpInsight(projectRef || "", exactPair?.probeId || "", exactPair?.checkId || "", resultWindowFilters),
		enabled: Boolean(canQueryPairDetail && exactPair?.check.type === "HTTP")
	});
	const httpSeriesQuery = useQuery({
		...projectQueries.httpSeries(projectRef || "", exactPair?.probeId || "", exactPair?.checkId || "", resultWindowFilters),
		enabled: Boolean(canQueryPairDetail && exactPair?.check.type === "HTTP")
	});
	const latestHTTPResultsQuery = useQuery({
		...projectQueries.latestHttpResults(projectRef || ""),
		enabled: Boolean(projectRef && hasSelectedHTTPPairs)
	});
	const latestHTTPResultsByPair = useMemo(() => latestHttpResultMap(latestHTTPResultsQuery.data?.results), [latestHTTPResultsQuery.data?.results]);
	const exactLatestHTTPResult = latestHTTPResultForPair(latestHTTPResultsByPair, exactPair);
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

		deleteParam("mode");
		deleteParam("view");
		deleteParam("type");
		deleteParam("groupBy");

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
		exactPair,
		projectRef,
		refresh,
		searchParamString,
		setSearchParams,
		timeMode,
		timeRange,
		timeWindow.from,
		timeWindow.to,
		urlState.assignmentKeys,
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

	function writeSelectionParams(next: URLSearchParams, key: "probeId" | "checkId", values: string[]) {
		next.delete(key);
		values.forEach(value => next.append(key, value));
	}

	function selectProbeScope(probeIds: string[]) {
		const nextProbeIds = Array.from(new Set(probeIds));
		const nextProbeIdSet = new Set(nextProbeIds);
		const allowedCheckIds = new Set(pairs.filter(pair => !nextProbeIds.length || nextProbeIdSet.has(pair.probeId)).map(pair => pair.checkId));
		const nextCheckIds = nextProbeIds.length ? activeCheckIds.filter(checkId => allowedCheckIds.has(checkId)) : activeCheckIds;

		updateSearchParams(next => {
			writeSelectionParams(next, "probeId", nextProbeIds);
			writeSelectionParams(next, "checkId", nextCheckIds);
			next.delete("assignment");
			next.delete("runStartedAt");
		});
	}

	function selectCheckScope(checkIds: string[]) {
		updateSearchParams(next => {
			writeSelectionParams(next, "checkId", Array.from(new Set(checkIds)));
			next.delete("assignment");
			next.delete("runStartedAt");
		});
	}

	function resetScope() {
		updateSearchParams(next => {
			next.delete("assignment");
			next.delete("probeId");
			next.delete("checkId");
			next.delete("type");
			next.delete("groupBy");
			next.delete("runStartedAt");
		});
	}

	const scopeTitle = hasInvalidFocus
		? t("invalidScope")
		: selectedPairs.length > 1
			? t("selectedAssignments", { count: formatCount(selectedPairs.length) })
			: exactPair
				? `${exactPair.probe.name} -> ${exactPair.check.target}`
				: selectedProbe && selectedCheck
					? t("noActiveAssignment")
					: selectedProbe
						? selectedProbe.name
						: selectedCheck
							? selectedCheck.target
							: hasResultScope
								? t("noActiveAssignment")
								: t("selectScope");
	const groupTopologyTitle =
		topologyProbe && topologyCheck
			? t("routeGraphPair", { probe: topologyProbe.name, target: topologyCheck.target })
			: topologyProbe
				? t("routeGraphProbe", { probe: topologyProbe.name })
				: topologyCheck
					? t("routeGraphCheck", { target: topologyCheck.target })
					: t("selectedRouteGraph");
	const pairDetail = (
		<InsightPairDetail
			pair={exactPair}
			pingInsightData={pingInsightQuery.data}
			pingSeriesData={pingSeriesQuery.data}
			tcpInsightData={tcpInsightQuery.data}
			tcpSeriesData={tcpSeriesQuery.data}
			httpInsightData={httpInsightQuery.data}
			httpSeriesData={httpSeriesQuery.data}
			httpLatestResult={exactLatestHTTPResult}
			nowMs={nowMs}
			isPingInsightLoading={pingInsightQuery.isLoading}
			isPingSeriesLoading={pingSeriesQuery.isLoading}
			isPingFetching={pingInsightQuery.isFetching || pingSeriesQuery.isFetching}
			isTCPInsightLoading={tcpInsightQuery.isLoading}
			isTCPSeriesLoading={tcpSeriesQuery.isLoading}
			isTCPFetching={tcpInsightQuery.isFetching || tcpSeriesQuery.isFetching}
			isHTTPLoading={httpInsightQuery.isLoading || httpSeriesQuery.isLoading}
			isHTTPLatestLoading={latestHTTPResultsQuery.isLoading}
			isHTTPFetching={httpInsightQuery.isFetching || httpSeriesQuery.isFetching || latestHTTPResultsQuery.isFetching}
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
			<MultiSeriesInsightPanel
				projectRef={projectRef}
				pairs={selectedPairs}
				filters={resultWindowFilters}
				latestHTTPResults={latestHTTPResultsQuery.data?.results}
				nowMs={nowMs}
				isLatestHTTPLoading={latestHTTPResultsQuery.isLoading}
				isLatestHTTPFetching={latestHTTPResultsQuery.isFetching}
				onSelectTimeWindow={applyAbsoluteWindow}
			/>
		) : exactPair?.check.type === "Traceroute" ? (
			tracerouteDetail
		) : (
			pairDetail
		);

	return (
		<PageStack>
			<ScreenHeader title={t("title")} />

			<Panel
				tone="glass"
				title={scopeTitle}
				actions={
					<Button variant="outline" size="sm" disabled={!hasResultScope && !hasInvalidFocus} onClick={resetScope}>
						{t("resetScope")}
					</Button>
				}
			>
				<div className={styles.scopeControls}>
					<div className={styles.scopeSelectionRow}>
						<ProbeScopeSelect probes={selectableProbes} selectedValues={activeProbeIds} disabled={isSelectionLoading || !selectableProbes.length} onChange={selectProbeScope} />
						<span className={styles.scopeArrow} aria-hidden="true">
							<ArrowRightIcon size="1.25rem" weight="bold" focusable="false" />
						</span>
						<CheckScopeSelect checks={availableChecks} selectedValues={activeCheckIds} disabled={isSelectionLoading || !activeProbeIds.length || !availableChecks.length} onChange={selectCheckScope} />
					</div>
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
				</div>
			</Panel>

			{isSelectionLoading && !pairs.length ? (
				<Panel tone="deep" title={t("activePaths")}>
					<Spinner label={t("loadingActivePaths")} layout="compact" size="lg" />
				</Panel>
			) : !pairs.length ? (
				<Panel tone="deep" title={t("noActivePaths")}>
					<BodyCopy>{t("noActivePathsDescription")}</BodyCopy>
				</Panel>
			) : hasInvalidFocus ? (
				<Panel tone="deep" title={t("invalidSharedScope")}>
					<BodyCopy>{t("invalidSharedScopeDescription")}</BodyCopy>
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

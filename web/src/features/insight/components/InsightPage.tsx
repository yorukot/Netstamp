import { formatInterval, mapApiChecks } from "@/features/checks/api/checkAdapters";
import { type CheckDefinition } from "@/features/checks/data/checks";
import { GroupTopologyPanel } from "@/features/insight/components/GroupTopologyPanel";
import { PingInsightPanel } from "@/features/insight/components/PingInsightPanel";
import { TracerouteInsightPanel } from "@/features/insight/components/TracerouteInsightPanel";
import { formatCount } from "@/features/insight/insightFormatters";
import { type EntityDetail, type InsightMode, type InsightPair, type ParsedInsightUrlState, type TimeWindow } from "@/features/insight/insightTypes";
import { mapApiProbes } from "@/features/probes/api/probeAdapters";
import { type Probe, type ProbeStatus } from "@/features/probes/data/probes";
import { projectQueries } from "@/shared/api/queries";
import { type ApiProjectAssignment, type PingInsightResponse, type TracerouteResult } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { BodyCopy } from "@/shared/components/BodyCopy";
import { KeyValueGrid } from "@/shared/components/KeyValueGrid";
import { PageStack } from "@/shared/components/PageStack";
import { ResponsiveGrid } from "@/shared/components/ResponsiveGrid";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { type RouteTopologyEdge, type RouteTopologyNode } from "@/shared/visualizations/RouteTopologyMap";
import { Panel, SelectField } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useMemo } from "react";
import { useSearchParams } from "react-router-dom";
import styles from "./InsightPage.module.css";

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
	topologyNodes: RouteTopologyNode[];
	topologyEdges: RouteTopologyEdge[];
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
		<TracerouteInsightPanel
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
		<PingInsightPanel selectedProbe={pair.probe} selectedTarget={pair.check} data={pingData} isLoading={isPingLoading} isFetching={isPingFetching} timeLabel={timeLabel} />
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
		urlState.hasValidMode,
		urlState.hasValidTimeWindow,
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
								{ label: "Selected", value: selectedPair ? `${selectedPair.probe.name} → ${selectedPair.check.target}` : "-" }
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
						<strong>{selectedPair ? `${selectedPair.probe.name} → ${selectedPair.check.target}` : "No selection"}</strong>
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

import { formatInterval } from "@/features/checks/api/checkAdapters";
import type { CheckDefinition } from "@/features/checks/data/checks";
import type { Probe, ProbeStatus } from "@/features/probes/data/probes";
import type { ApiProjectAssignment } from "@/shared/api/types";
import { formatCount } from "@/shared/utils/insightFormatters";
import {
	isRelativeTimeRange as isInsightRelativeRange,
	parseEpochMs,
	relativeRangeForTimeWindow as relativeRangeForWindow,
	timeWindowForRelativeRange as timeWindowForRange
} from "@/shared/utils/timeRanges";
import type { AssignmentSelectOption, InsightCheckTypeFilter, InsightGroupBy, InsightPair, InsightRefreshInterval, InsightRelativeRange, InsightTimeMode, ParsedInsightUrlState } from "./insightTypes";

export const refreshDurations: Partial<Record<InsightRefreshInterval, number>> = {
	"10s": 10 * 1000,
	"30s": 30 * 1000,
	"1m": 60 * 1000,
	"5m": 5 * 60 * 1000
};

export const checkTypeOptions: Array<{ value: InsightCheckTypeFilter; label: string }> = [
	{ value: "all", label: "All" },
	{ value: "ping", label: "Ping" },
	{ value: "tcp", label: "TCP" },
	{ value: "traceroute", label: "Traceroute" },
	{ value: "http", label: "HTTP / HTTPS" }
];

export const groupByOptions: Array<{ value: InsightGroupBy; label: string }> = [
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
	return value === "all" || value === "ping" || value === "tcp" || value === "traceroute" || value === "http";
}

function isInsightGroupBy(value: string | null): value is InsightGroupBy {
	return value === "check" || value === "probe";
}

export function parseInsightUrlState(searchParams: URLSearchParams, now: number): ParsedInsightUrlState {
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
	const probeIds = Array.from(new Set(searchParams.getAll("probeId").filter(Boolean)));
	const checkIds = Array.from(new Set(searchParams.getAll("checkId").filter(Boolean)));

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
		probeIds,
		checkIds,
		probeId: probeIds[0] || "",
		checkId: checkIds[0] || "",
		runStartedAt: searchParams.get("runStartedAt") || ""
	};
}

export function pairKey(probeId: string, checkId: string) {
	return `${probeId}:${checkId}`;
}

function checkTypeFromApi(value: string | undefined): CheckDefinition["type"] {
	switch (value?.toLowerCase()) {
		case "tcp":
			return "TCP";
		case "traceroute":
			return "Traceroute";
		case "http":
			return "HTTP";
		default:
			return "Ping";
	}
}

function pairCheckType(pair: InsightPair): Exclude<InsightCheckTypeFilter, "all"> {
	if (pair.check.type === "TCP") {
		return "tcp";
	}
	if (pair.check.type === "HTTP") return "http";

	return pair.check.type === "Traceroute" ? "traceroute" : "ping";
}

export function checkTypeFilterFromCheck(check: CheckDefinition): Exclude<InsightCheckTypeFilter, "all"> {
	if (check.type === "TCP") {
		return "tcp";
	}
	if (check.type === "HTTP") return "http";

	return check.type === "Traceroute" ? "traceroute" : "ping";
}

export function matchesCheckType(pair: InsightPair, checkType: InsightCheckTypeFilter) {
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

export function buildInsightPairs(assignments: ApiProjectAssignment[], probes: Probe[], checks: CheckDefinition[]): InsightPair[] {
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

export function scopePairs(pairs: InsightPair[], checkType: InsightCheckTypeFilter, probeId: string, checkId: string) {
	return pairs.filter(pair => matchesCheckType(pair, checkType) && (!probeId || pair.probeId === probeId) && (!checkId || pair.checkId === checkId));
}

function assignmentLabel(pair: InsightPair) {
	return `${pair.check.name} / ${pair.probe.name} / ${pair.check.target}`;
}

function assignmentMeta(pair: InsightPair) {
	return `${pair.check.type} · ${pair.probe.location}`;
}

export function assignmentSelectOption(pair: InsightPair): AssignmentSelectOption {
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

export function uniqueProbeOptions(probes: Probe[], pairs: InsightPair[]): AssignmentSelectOption[] {
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

export function uniqueCheckOptions(checks: CheckDefinition[], pairs: InsightPair[]): AssignmentSelectOption[] {
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

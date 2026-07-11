import { checkConfigSummaryFields } from "@/features/checks/data/checkConfig";
import type { CheckDefinition, CheckType } from "@/features/checks/data/checks";
import type { ApiCheck, ApiProjectAssignment } from "@/shared/api/types";

function mapCheckType(type: string): CheckType {
	switch (type.toLowerCase()) {
		case "http":
			return "HTTP";
		case "tcp":
			return "TCP";
		case "traceroute":
			return "Traceroute";
		default:
			return "Ping";
	}
}

export function formatInterval(seconds: number) {
	return `${seconds}s`;
}

export interface IntervalValidation {
	value: number;
	error: string;
}

export function validateIntervalSeconds(value: string): IntervalValidation {
	const trimmed = value.trim();
	const normalized = trimmed.endsWith("s") ? trimmed.slice(0, -1) : trimmed;

	if (!normalized) {
		return { value: Number.NaN, error: "Interval is required." };
	}

	if (!/^\d+$/.test(normalized)) {
		return { value: Number.NaN, error: "Interval must be whole seconds, for example 30s." };
	}

	const parsed = Number.parseInt(normalized, 10);

	if (!Number.isFinite(parsed) || parsed < 1) {
		return { value: parsed, error: "Interval must be at least 1 second." };
	}

	if (parsed > 86400) {
		return { value: parsed, error: "Interval must be at most 86400 seconds." };
	}

	return { value: parsed, error: "" };
}

export function parseIntervalSeconds(value: string) {
	const validation = validateIntervalSeconds(value);

	if (validation.error) {
		throw new Error(validation.error);
	}

	return validation.value;
}

export function mapApiCheck(check: ApiCheck, assigned = 0): CheckDefinition {
	const type = mapCheckType(check.type);

	return {
		id: check.id,
		name: check.name,
		type,
		target: check.target,
		interval: formatInterval(check.intervalSeconds),
		latest: "-",
		assigned,
		description: check.description || "",
		fields: [["Target", check.target], ["Type", type], ["Interval", formatInterval(check.intervalSeconds)], ["Labels", String(check.labels?.length ?? 0)], ...checkConfigSummaryFields(check)]
	};
}

export function mapApiChecks(checks: ApiCheck[] | null | undefined, probes: Array<unknown> | null | undefined): CheckDefinition[] {
	return (checks ?? []).map(check => mapApiCheck(check, probes?.length ?? 0));
}

export function mapApiChecksWithAssignments(checks: ApiCheck[] | null | undefined, assignments: ApiProjectAssignment[] | null | undefined): CheckDefinition[] {
	const assignmentCounts = new Map<string, number>();

	for (const assignment of assignments ?? []) {
		assignmentCounts.set(assignment.checkId, (assignmentCounts.get(assignment.checkId) ?? 0) + 1);
	}

	return (checks ?? []).map(check => mapApiCheck(check, assignmentCounts.get(check.id) ?? 0));
}

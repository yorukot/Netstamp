import { checkConfigSummaryFields } from "@/features/checks/data/checkConfig";
import type { CheckDefinition, CheckType } from "@/features/checks/data/checks";
import type { ApiCheck, ApiProjectAssignment } from "@/shared/api/types";

function mapCheckType(type: string): CheckType {
	switch (type.toLowerCase()) {
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

export function parseIntervalSeconds(value: string) {
	const parsed = Number.parseInt(value, 10);
	return Number.isFinite(parsed) && parsed > 0 ? parsed : 30;
}

export function mapApiCheck(check: ApiCheck, assigned = 0): CheckDefinition {
	const type = mapCheckType(check.type);

	return {
		id: check.id,
		name: check.name,
		type,
		target: check.target,
		status: "Configured",
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

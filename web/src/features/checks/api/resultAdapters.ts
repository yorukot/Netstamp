import type { CheckDefinition } from "@/features/checks/data/checks";
import type { AssignedRow } from "@/features/probes/components/types";
import type { Probe } from "@/features/probes/data/probes";
import type { ApiProjectAssignment } from "@/shared/api/types";
import { formatInterval } from "./checkAdapters";

function namesById<TItem extends { id: string; name: string }>(items: TItem[]) {
	return new Map(items.map(item => [item.id, item.name]));
}

export function mapApiAssignments(assignments: ApiProjectAssignment[] | null | undefined, probes: Probe[], checks: CheckDefinition[]): AssignedRow[] {
	const probeNames = namesById(probes);
	const checksById = new Map(checks.map(check => [check.id, check]));

	return (assignments ?? []).map(assignment => {
		const check = checksById.get(assignment.checkId);

		return {
			probe: assignment.probe?.name || probeNames.get(assignment.probeId) || assignment.probeId,
			check: assignment.check?.name || check?.name || assignment.checkId,
			type: check?.type || "Ping",
			interval: check?.interval || (assignment.check ? formatInterval(assignment.check.intervalSeconds) : "-"),
			latest: check?.latest || "-"
		};
	});
}

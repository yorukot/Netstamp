import type { Probe } from "@/features/probes/data/probes";
import type { AssignedRow } from "@/shared/domain/assignments";

export function filterProbes(source: Probe[], search: string) {
	const term = search.trim().toLowerCase();
	return source.filter(probe => {
		const searchable = [probe.name, probe.location, probe.publicIp, probe.provider, probe.region, ...probe.labelTokens].join(" ").toLowerCase();

		return !term || searchable.includes(term);
	});
}

export function expandAssignedRows(rows: AssignedRow[]) {
	return rows;
}

import type { Probe } from "@/features/probes/data/probes";
import type { AssignedRow } from "@/shared/domain/assignments";
import type { ProbeSort } from "./types";

export function filterProbes(source: Probe[], search: string, sortKey: ProbeSort) {
	const term = search.trim().toLowerCase();
	const filtered = source.filter(probe => {
		const searchable = [probe.name, probe.location, probe.publicIp, probe.provider, probe.region, ...probe.labelTokens].join(" ").toLowerCase();

		return !term || searchable.includes(term);
	});

	if (sortKey === "name") {
		return [...filtered].sort((left, right) => left.name.localeCompare(right.name));
	}

	return [...filtered].sort((left, right) => {
		if (left.lastHeartbeatAt === null && right.lastHeartbeatAt === null) {
			return left.name.localeCompare(right.name);
		}
		if (left.lastHeartbeatAt === null) {
			return 1;
		}
		if (right.lastHeartbeatAt === null) {
			return -1;
		}

		return right.lastHeartbeatAt - left.lastHeartbeatAt || left.name.localeCompare(right.name);
	});
}

export function expandAssignedRows(rows: AssignedRow[]) {
	return rows;
}

import type { Probe, ProbeStatus } from "@/features/probes/data/probes";
import type { AssignedRow, ProbeSort } from "./types";

function asnNumber(asn: string) {
	return Number(asn.replace(/\D/g, "")) || 0;
}

export function filterProbes(source: Probe[], search: string, statusFilter: "all" | ProbeStatus, providerFilter: string, sortKey: ProbeSort) {
	const term = search.trim().toLowerCase();
	const filtered = source.filter(probe => {
		const searchable = [probe.name, probe.location, probe.publicIp, probe.asn, probe.provider, probe.region, ...probe.tags].join(" ").toLowerCase();

		return (!term || searchable.includes(term)) && (statusFilter === "all" || probe.status === statusFilter) && (providerFilter === "all" || probe.provider === providerFilter);
	});

	if (sortKey === "name") {
		return filtered.sort((left, right) => left.name.localeCompare(right.name));
	}

	if (sortKey === "asn") {
		return filtered.sort((left, right) => asnNumber(left.asn) - asnNumber(right.asn));
	}

	return filtered;
}

export function expandAssignedRows(rows: AssignedRow[]) {
	return rows;
}

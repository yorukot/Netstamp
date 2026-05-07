import type { Probe, ProbeStatus } from "@/features/probes/data/probes";
import type { AssignedRow, ProbeSort } from "./types";

const assignedOverflowRowCount = 18;

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
	if (!rows.length) {
		return [];
	}

	return Array.from({ length: assignedOverflowRowCount }, (_, index) => {
		const row = rows[index % rows.length];
		const suffix = String(index + 1).padStart(2, "0");

		return {
			...row,
			check: index < rows.length ? row.check : `${row.check}-${suffix}`
		};
	});
}

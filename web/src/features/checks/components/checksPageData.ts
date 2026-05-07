import { assignments, checks, results, type CheckDefinition } from "@/features/checks/data/checks";

export interface LogRow {
	time: string;
	check: string;
	probe: string;
	status: string;
	latency: string;
	event: string;
}

export const checkRows: CheckDefinition[] = Array.from({ length: 21 }, (_, index) => {
	const check = checks[index % checks.length];

	if (index < checks.length) {
		return check;
	}

	return {
		...check,
		id: `${check.id}-${index + 1}`,
		name: `${check.name}-${String(index + 1).padStart(2, "0")}`,
		status: index % 4 === 0 ? "Warning" : check.status,
		interval: index % 2 === 0 ? check.interval : "45s",
		jitter: index % 2 === 0 ? check.jitter : "6s",
		assigned: Math.max(1, check.assigned - index)
	};
});

const latestLogs: LogRow[] = [
	...results.map(([time, probe, check, status, latency, , metadata]) => ({ time, probe, check, status, latency, event: metadata })),
	{ time: "2026-05-06 14:23:30", check: "api-latency", probe: "sfo-lab-05", status: "success", latency: "55ms", event: "icmp_seq=531" },
	{ time: "2026-05-06 14:23:18", check: "root-dns-a", probe: "fra-bm-02", status: "success", latency: "41ms", event: "NOERROR" },
	{ time: "2026-05-06 14:22:55", check: "validator-route", probe: "ams-edge-01", status: "partial", latency: "88ms", event: "hop 9 ttl exceeded" },
	{ time: "2026-05-06 14:22:42", check: "api-latency", probe: "sin-probe-04", status: "warning", latency: "84ms", event: "latency above baseline" },
	{ time: "2026-05-06 14:22:20", check: "root-dns-a", probe: "sfo-lab-05", status: "success", latency: "37ms", event: "NOERROR" },
	{ time: "2026-05-06 14:22:04", check: "api-latency", probe: "fra-bm-02", status: "success", latency: "39ms", event: "icmp_seq=530" }
].slice(0, 10);

export function logsForCheck(check: CheckDefinition, selectedProbes: string[]) {
	const existingLogs = latestLogs.filter(log => log.check === check.id || log.check === check.name);
	const probePool = selectedProbes.length ? selectedProbes : assignedProbeNames(check.id);
	const fallbackProbes = probePool.length ? probePool : ["controller"];
	const generatedLogs: LogRow[] = Array.from({ length: 10 }, (_, index) => ({
		time: `2026-05-06 14:${String(21 - Math.floor(index / 2)).padStart(2, "0")}:${String(58 - index * 4).padStart(2, "0")}`,
		check: check.name,
		probe: fallbackProbes[index % fallbackProbes.length],
		status: index % 5 === 0 ? "partial" : "success",
		latency: check.type === "DNS" ? `${35 + index * 3}ms` : `${42 + index * 5}ms`,
		event: check.type === "Traceroute" ? `path sample ${index + 1}` : `fetch sample ${index + 1}`
	}));

	return [...existingLogs, ...generatedLogs].slice(0, 10);
}

export function assignedProbeNames(checkId: string) {
	return assignments.filter(([, check]) => check === checkId).map(([probe]) => probe);
}

export function displayProbeSelection(selectedProbes: string[]) {
	if (!selectedProbes.length) {
		return "No probes assigned";
	}

	if (selectedProbes.length === 1) {
		return selectedProbes[0];
	}

	return `${selectedProbes.length} probes selected`;
}

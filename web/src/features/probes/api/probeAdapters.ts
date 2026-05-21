import type { Probe, ProbeStatus } from "@/features/probes/data/probes";
import type { ApiProbe } from "@/shared/api/types";

const fallbackCoordinates: Array<[number, number]> = [
	[120.6736, 24.1477],
	[8.6821, 50.1109],
	[-74.006, 40.7128],
	[103.8198, 1.3521],
	[-122.4194, 37.7749]
];

function formatRelativeTime(value: string | null | undefined) {
	if (!value) {
		return "never";
	}

	const elapsedSeconds = Math.max(0, Math.floor((Date.now() - new Date(value).getTime()) / 1000));

	if (elapsedSeconds < 60) {
		return `${elapsedSeconds}s ago`;
	}

	const elapsedMinutes = Math.floor(elapsedSeconds / 60);

	if (elapsedMinutes < 60) {
		return `${elapsedMinutes}m ago`;
	}

	const elapsedHours = Math.floor(elapsedMinutes / 60);
	return `${elapsedHours}h ago`;
}

function mapProbeStatus(probe: ApiProbe): ProbeStatus {
	const state = probe.status?.state?.toLowerCase() ?? "offline";

	if (!probe.enabled || state.includes("drain")) {
		return "Draining";
	}

	if (state.includes("online") || state.includes("active") || state.includes("healthy") || state.includes("ready")) {
		return "Online";
	}

	return "Offline";
}

function labelValue(probe: ApiProbe, key: string) {
	return probe.labels?.find(label => label.key.toLowerCase() === key)?.value || "";
}

export function mapApiProbe(probe: ApiProbe, index: number): Probe {
	const coordinates =
		typeof probe.longitude === "number" && typeof probe.latitude === "number" ? ([probe.longitude, probe.latitude] as [number, number]) : fallbackCoordinates[index % fallbackCoordinates.length];
	const status = probe.status;
	const publicIp = status?.publicV4 || status?.publicV6 || status?.addrs?.[0] || "-";
	const tags = probe.labels?.map(label => `${label.key}:${label.value}`) ?? [];

	return {
		id: probe.id,
		name: probe.name,
		status: mapProbeStatus(probe),
		location: labelValue(probe, "location") || probe.subdivisionCode || `${coordinates[1].toFixed(2)}, ${coordinates[0].toFixed(2)}`,
		publicIp,
		asn: status?.as || labelValue(probe, "as") || "-",
		provider: labelValue(probe, "provider") || "Unlabeled",
		region: probe.subdivisionCode || "unassigned",
		ipFamily: status?.publicV4 && status.publicV6 ? "IPv4 / IPv6" : status?.publicV6 ? "IPv6" : "IPv4",
		lastHeartbeat: formatRelativeTime(status?.lastSeenAt),
		tags,
		version: status?.agentVersion || "-",
		uptime: "-",
		cpu: "-",
		memory: "-",
		queue: probe.enabled ? "accepting jobs" : "disabled",
		loss: "-",
		coordinates,
		capabilities: ["ping"]
	};
}

export function mapApiProbes(probes: ApiProbe[] | null | undefined): Probe[] {
	return (probes ?? []).map(mapApiProbe);
}

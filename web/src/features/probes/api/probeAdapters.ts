import { coordinateSummary } from "@/features/probes/data/probeLocation";
import type { Probe, ProbeStatus } from "@/features/probes/data/probes";
import type { ApiProbe } from "@/shared/api/types";

function parseTimestamp(value: string | null | undefined) {
	if (!value) {
		return null;
	}

	const timestamp = new Date(value).getTime();
	return Number.isFinite(timestamp) ? timestamp : null;
}

function formatRelativeTime(timestamp: number | null) {
	if (timestamp === null) {
		return "never";
	}

	const elapsedSeconds = Math.max(0, Math.floor((Date.now() - timestamp) / 1000));

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
	void index;

	const coordinates = typeof probe.longitude === "number" && typeof probe.latitude === "number" ? ([probe.longitude, probe.latitude] as [number, number]) : undefined;
	const status = probe.status;
	const probeStatus = mapProbeStatus(probe);
	const supportsV4 = Boolean(status?.publicV4);
	const supportsV6 = Boolean(status?.publicV6);
	const publicIp = status?.publicV4 || status?.publicV6 || "-";
	const visibleLabels = probe.labels?.filter(label => label.key.toLowerCase() !== "as") ?? [];
	const labelTokens = visibleLabels.map(label => `${label.key}:${label.value}`);
	const ipFamily = supportsV4 && supportsV6 ? "IPv4 / IPv6" : supportsV4 ? "IPv4" : supportsV6 ? "IPv6" : "-";
	const location = probe.locationName || coordinateSummary(probe.latitude, probe.longitude) || "-";
	const lastHeartbeatAt = parseTimestamp(status?.lastSeenAt);

	return {
		id: probe.id,
		name: probe.name,
		status: probeStatus,
		location,
		publicIp,
		provider: labelValue(probe, "provider") || "Unlabeled",
		region: location,
		ipFamily,
		lastHeartbeat: formatRelativeTime(lastHeartbeatAt),
		lastHeartbeatAt,
		labelTokens,
		version: status?.agentVersion || "-",
		uptime: "-",
		cpu: "-",
		memory: "-",
		queue: probe.enabled ? "accepting jobs" : "disabled",
		loss: "-",
		coordinates,
		capabilities: []
	};
}

export function mapApiProbes(probes: ApiProbe[] | null | undefined): Probe[] {
	return (probes ?? []).map(mapApiProbe);
}

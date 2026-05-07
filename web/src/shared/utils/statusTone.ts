import type { BadgeTone } from "@netstamp/ui";

export function toneForStatus(status: unknown): BadgeTone {
	const value = String(status).toLowerCase();

	if (value.includes("online") || value.includes("healthy") || value.includes("success")) {
		return "success";
	}

	if (value.includes("warn") || value.includes("changed") || value.includes("draining") || value.includes("partial")) {
		return "warning";
	}

	if (value.includes("offline") || value.includes("critical") || value.includes("error") || value.includes("expired")) {
		return "critical";
	}

	return "neutral";
}

import type { CheckType } from "@/features/checks/data/checks";
import type { ApiCheck, CreateCheckInput } from "@/shared/api/types";
import type { CheckTypeFilter } from "./ChecksTable";

export function isCheckTypeFilter(value: string | null): value is CheckTypeFilter {
	return value === "all" || value === "ping" || value === "tcp" || value === "traceroute";
}

export function pathWithSearch(path: string, search: string) {
	return search ? `${path}?${search}` : path;
}

export function checkTypeFromApi(type: string): CheckType {
	switch (type) {
		case "tcp":
			return "TCP";
		case "traceroute":
			return "Traceroute";
		default:
			return "Ping";
	}
}

function copiedCheckName(name: string) {
	const base = name.trim() || "Check";
	const suffix = " copy";
	const maxBaseLength = Math.max(1, 128 - suffix.length);

	return `${base.slice(0, maxBaseLength)}${suffix}`;
}

export function duplicateCheckInput(check: ApiCheck): CreateCheckInput {
	const body: CreateCheckInput = {
		intervalSeconds: check.intervalSeconds,
		name: copiedCheckName(check.name),
		target: check.target,
		type: check.type
	};

	if (check.selector) {
		body.selector = check.selector;
	}
	if (check.description) {
		body.description = check.description;
	}
	if (check.labels.length) {
		body.labelIds = check.labels.map(label => label.id);
	}
	if (check.pingConfig) {
		body.pingConfig = { ...check.pingConfig };
	}
	if (check.tcpConfig) {
		body.tcpConfig = { ...check.tcpConfig };
	}
	if (check.tracerouteConfig) {
		body.tracerouteConfig = { ...check.tracerouteConfig };
	}

	return body;
}

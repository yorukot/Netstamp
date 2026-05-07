import type { CheckType } from "@/features/checks/data/checks";

export type ProbeView = "grid" | "map";
export type ProbeSort = "heartbeat" | "name" | "asn";
export type DetectionMode = "manual" | "auto";

export interface AssignedRow {
	probe: string;
	check: string;
	type: CheckType;
	interval: string;
	jitter: string;
	latest: string;
}

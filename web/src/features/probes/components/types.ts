import type { CheckType } from "@/features/checks/data/checks";

export type ProbeView = "grid" | "map";
export type ProbeSort = "heartbeat" | "name" | "asn";

export interface AssignedRow {
	probe: string;
	check: string;
	type: CheckType;
	interval: string;
	latest: string;
}

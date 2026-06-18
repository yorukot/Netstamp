import type { CheckType } from "./checks";

export interface AssignedRow {
	probe: string;
	check: string;
	type: CheckType;
	interval: string;
	latest: string;
}

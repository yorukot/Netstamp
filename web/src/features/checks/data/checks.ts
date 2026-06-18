import type { CheckType } from "@/shared/domain/checks";

export type { CheckType };

export interface CheckDefinition {
	id: string;
	name: string;
	type: CheckType;
	target: string;
	status: string;
	interval: string;
	latest: string;
	assigned: number;
	description: string;
	fields: Array<[label: string, value: string]>;
}

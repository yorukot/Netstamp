export type CheckType = "Ping" | "Traceroute" | "DNS";

export interface CheckDefinition {
	id: string;
	name: string;
	type: CheckType;
	target: string;
	status: string;
	interval: string;
	jitter: string;
	latest: string;
	assigned: number;
	description: string;
	fields: Array<[label: string, value: string]>;
}

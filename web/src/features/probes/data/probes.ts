export type ProbeStatus = "Online" | "Offline" | "Draining";

export interface Probe {
	id: string;
	name: string;
	status: ProbeStatus;
	location: string;
	publicIp: string;
	provider: string;
	region: string;
	ipFamily: string;
	lastHeartbeat: string;
	lastHeartbeatAt: number | null;
	labelTokens: string[];
	version: string;
	uptime: string;
	cpu: string;
	memory: string;
	queue: string;
	loss: string;
	coordinates?: [number, number];
	capabilities: string[];
}

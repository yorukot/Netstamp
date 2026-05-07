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

export type AssignmentTuple = [probe: string, check: string, type: CheckType, interval: string, jitter: string, latest: string];
export type ResultTuple = [time: string, probe: string, check: string, status: string, latency: string, loss: string, metadata: string];

export const checks: CheckDefinition[] = [
	{
		id: "api-latency",
		name: "api-latency",
		type: "Ping",
		target: "api.netstamp.io",
		status: "Healthy",
		interval: "30s",
		jitter: "4s",
		latest: "42ms",
		assigned: 42,
		description: "Latency and loss to public controller API.",
		fields: [
			["Target", "api.netstamp.io"],
			["IP version", "IPv4 / IPv6"],
			["Packet count", "5"],
			["Interval", "30s"],
			["Timeout", "2s"],
			["Packet size", "56 bytes"],
			["Source interface", "auto"],
			["Fallback status", "Using system ping fallback"]
		]
	},
	{
		id: "validator-route",
		name: "validator-route",
		type: "Traceroute",
		target: "validator-03.mainnet.example",
		status: "Path changed",
		interval: "120s",
		jitter: "16s",
		latest: "hash changed",
		assigned: 18,
		description: "Route fingerprint for validator RPC egress.",
		fields: [
			["Target", "validator-03.mainnet.example"],
			["IP version", "IPv6 preferred"],
			["Max hops", "32"],
			["Queries per hop", "3"],
			["Timeout", "3s"],
			["Protocol", "UDP"],
			["Path hash", "0x8fa3 → 0xc12e"],
			["Recent diff", "Transit changed at hop 9"]
		]
	},
	{
		id: "root-dns-a",
		name: "root-dns-a",
		type: "DNS",
		target: "netstamp.io A",
		status: "Warning",
		interval: "60s",
		jitter: "8s",
		latest: "SERVFAIL burst",
		assigned: 33,
		description: "Resolver correctness and latency for public hostname.",
		fields: [
			["Query name", "netstamp.io"],
			["Record type", "A"],
			["Resolver", "1.1.1.1"],
			["Transport", "UDP + TCP fallback"],
			["Timeout", "2s"],
			["Attempts", "2"],
			["IP version", "IPv4"],
			["RCODE distribution", "NOERROR 98.4% / SERVFAIL 1.6%"]
		]
	}
];

export const assignments: AssignmentTuple[] = [
	["ams-edge-01", "api-latency", "Ping", "30s", "4s", "42ms"],
	["fra-bm-02", "validator-route", "Traceroute", "120s", "16s", "Path hash changed"],
	["sin-probe-04", "root-dns-a", "DNS", "60s", "8s", "38ms"],
	["sfo-lab-05", "api-latency", "Ping", "30s", "4s", "55ms"]
];

export const results: ResultTuple[] = [
	["2026-05-06 14:24:18", "ams-edge-01", "api-latency", "success", "42ms", "0.00%", "icmp_seq=532"],
	["2026-05-06 14:24:06", "fra-bm-02", "validator-route", "partial", "91ms", "0.00%", "path hash changed"],
	["2026-05-06 14:23:52", "sin-probe-04", "root-dns-a", "warning", "118ms", "0.00%", "SERVFAIL"],
	["2026-05-06 14:23:45", "nyc-vps-03", "api-latency", "error", "-", "100%", "Probe heartbeat expired"]
];

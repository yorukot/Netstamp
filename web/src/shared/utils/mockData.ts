import type { AppRoute } from "@/routes/routeTypes";
import type { BadgeTone } from "@netstamp/ui";

export interface SidebarItem {
	label: string;
	route: AppRoute;
}

export interface CurrentUser {
	name: string;
	username: string;
	email: string;
	role: string;
	gravatarUrl: string;
}

export type ProbeStatus = "Online" | "Offline" | "Draining";

export interface Probe {
	id: string;
	name: string;
	status: ProbeStatus;
	location: string;
	publicIp: string;
	asn: string;
	provider: string;
	region: string;
	ipFamily: string;
	lastHeartbeat: string;
	tags: string[];
	version: string;
	uptime: string;
	cpu: string;
	memory: string;
	queue: string;
	loss: string;
	coordinates: [number, number];
	capabilities: string[];
}

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

export interface AlertRecord {
	type: string;
	check: string;
	probe: string;
	severity: "critical" | "warning";
	triggered: string;
	status: string;
}

export type MemberTuple = [name: string, email: string, role: string, lastActive: string];
export type SystemStateTuple = [title: string, copy: string];

export const sidebarItems: SidebarItem[] = [
	{ label: "Dashboard", route: "dashboard" },
	{ label: "Probes", route: "probes" },
	{ label: "Insight", route: "insight" },
	{ label: "Checks", route: "checks" },
	{ label: "Alerts", route: "alerts" },
	{ label: "team", route: "team" }
];

export const currentUser: CurrentUser = {
	name: "Elvis Mao",
	username: "elvis",
	email: "elvis@netstamp.dev",
	role: "Admin",
	gravatarUrl: "https://gravatar.com/avatar/f5a410169cdb93933383e6e54ac33b82e417fe84ffc4ed742adafd800cc07ab2?s=160&d=identicon"
};

export const probes: Probe[] = [
	{
		id: "ams-edge-01",
		name: "ams-edge-01",
		status: "Online",
		location: "Taichung, Taiwan",
		publicIp: "142.250.196.206",
		asn: "AS13335",
		provider: "Cloudflare",
		region: "ap-east blueprint zone",
		ipFamily: "IPv4 / IPv6",
		lastHeartbeat: "18s ago",
		tags: ["Apple", "Home"],
		version: "v1.0.0",
		uptime: "18d 04h",
		cpu: "18%",
		memory: "42%",
		queue: "12 jobs",
		loss: "0.08%",
		coordinates: [120.6736, 24.1477],
		capabilities: ["raw ICMP", "DNS TCP fallback", "IPv6", "system ping fallback"]
	},
	{
		id: "fra-bm-02",
		name: "fra-bm-02",
		status: "Online",
		location: "Frankfurt, Germany",
		publicIp: "45.76.88.19",
		asn: "AS20473",
		provider: "Vultr",
		region: "eu-central",
		ipFamily: "IPv4 / IPv6",
		lastHeartbeat: "22s ago",
		tags: ["Bare metal", "IX"],
		version: "v1.0.0",
		uptime: "41d 13h",
		cpu: "12%",
		memory: "35%",
		queue: "4 jobs",
		loss: "0.00%",
		coordinates: [8.6821, 50.1109],
		capabilities: ["raw ICMP", "privileged traceroute", "IPv6"]
	},
	{
		id: "nyc-vps-03",
		name: "nyc-vps-03",
		status: "Offline",
		location: "New York, United States",
		publicIp: "159.203.88.44",
		asn: "AS14061",
		provider: "DigitalOcean",
		region: "us-east",
		ipFamily: "IPv4",
		lastHeartbeat: "17m ago",
		tags: ["VPS", "Edge"],
		version: "v0.9.8",
		uptime: "0d 00h",
		cpu: "0%",
		memory: "0%",
		queue: "expired",
		loss: "100%",
		coordinates: [-74.006, 40.7128],
		capabilities: ["system ping fallback", "DNS UDP"]
	},
	{
		id: "sin-probe-04",
		name: "sin-probe-04",
		status: "Online",
		location: "Singapore",
		publicIp: "103.253.144.21",
		asn: "AS45102",
		provider: "Alibaba Cloud",
		region: "ap-southeast",
		ipFamily: "IPv4 / IPv6",
		lastHeartbeat: "9s ago",
		tags: ["Web3", "Validator"],
		version: "v1.0.0",
		uptime: "7d 18h",
		cpu: "21%",
		memory: "51%",
		queue: "19 jobs",
		loss: "0.24%",
		coordinates: [103.8198, 1.3521],
		capabilities: ["raw ICMP", "DNS DoT", "IPv6"]
	},
	{
		id: "sfo-lab-05",
		name: "sfo-lab-05",
		status: "Draining",
		location: "San Francisco, United States",
		publicIp: "198.51.100.88",
		asn: "AS6939",
		provider: "Hurricane Electric",
		region: "us-west lab",
		ipFamily: "IPv6",
		lastHeartbeat: "45s ago",
		tags: ["Lab", "IPv6"],
		version: "v1.0.0",
		uptime: "12d 09h",
		cpu: "27%",
		memory: "48%",
		queue: "draining",
		loss: "0.18%",
		coordinates: [-122.4194, 37.7749],
		capabilities: ["IPv6", "DNS TCP fallback"]
	}
];

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

export const alerts: AlertRecord[] = [
	{
		type: "packet loss threshold exceeded",
		check: "api-latency",
		probe: "nyc-vps-03",
		severity: "critical",
		triggered: "17m ago",
		status: "active"
	},
	{
		type: "latency threshold exceeded",
		check: "api-latency",
		probe: "sin-probe-04",
		severity: "warning",
		triggered: "22m ago",
		status: "acknowledged"
	},
	{
		type: "traceroute path change",
		check: "validator-route",
		probe: "fra-bm-02",
		severity: "warning",
		triggered: "28m ago",
		status: "active"
	},
	{
		type: "abnormal DNS response code",
		check: "root-dns-a",
		probe: "sin-probe-04",
		severity: "warning",
		triggered: "31m ago",
		status: "active"
	},
	{
		type: "probe offline",
		check: "fleet",
		probe: "nyc-vps-03",
		severity: "critical",
		triggered: "17m ago",
		status: "active"
	}
];

export const members: MemberTuple[] = [
	["Elvis Mao", "elvis@netstamp.dev", "Admin", "2m ago"],
	["Mika Sato", "mika@vector.example", "Owner", "8m ago"],
	["Noah Chen", "noah@vector.example", "Member", "34m ago"],
	["NOC Robot", "nocbot@vector.example", "Viewer", "streaming"]
];

export const systemStates: SystemStateTuple[] = [
	["Loading", "Fetching controller topology and probe health."],
	["Submitting", "Saving assignment schedule and jitter envelope."],
	["Empty", "No probes match the selected region filter."],
	["No results", "No measurements were recorded in this time window."],
	["API error", "Controller returned 503 while syncing snapshot."],
	["Validation error", "Assignment already active for this probe and check."],
	["Permission denied", "Admin role is required for dangerous operations."],
	["Token expired", "Re-authenticate before rotating registration tokens."],
	["Probe registration failure", "Waiting for first heartbeat exceeded 15 minutes."]
];

export const latencyData = [42, 39, 44, 49, 46, 62, 58, 53, 57, 64, 60, 71, 68, 52, 47, 44];
export const lossData = [0.08, 0.12, 0.06, 0.0, 0.18, 0.24, 0.11, 0.04, 0.08, 0.14, 0.09, 0.05];
export const dnsData = [31, 29, 33, 48, 42, 37, 35, 41, 68, 58, 44, 39];
export const routeDiffData = [1, 0, 0, 2, 1, 4, 1, 0, 2, 1, 3, 1];

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

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

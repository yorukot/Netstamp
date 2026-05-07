export interface AlertRecord {
	type: string;
	check: string;
	probe: string;
	severity: "critical" | "warning";
	triggered: string;
	status: string;
}

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

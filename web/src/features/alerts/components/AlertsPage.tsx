import { ActionRow } from "@/shared/components/ActionRow";
import { KeyValueGrid } from "@/shared/components/KeyValueGrid";
import { PageStack } from "@/shared/components/PageStack";
import { ResponsiveGrid } from "@/shared/components/ResponsiveGrid";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { toneForStatus } from "@/shared/utils/statusTone";
import { Badge, Button, DataTable, Panel, type DataColumn } from "@netstamp/ui";

interface AlertRecord {
	type: string;
	check: string;
	probe: string;
	severity: "critical" | "warning";
	triggered: string;
	status: string;
}

const alertColumns: DataColumn<AlertRecord>[] = [
	{ key: "type", label: "Alert type" },
	{ key: "check", label: "Check" },
	{ key: "probe", label: "Probe" },
	{ key: "severity", label: "Severity", render: row => <Badge tone={toneForStatus(row.severity)}>{row.severity}</Badge> },
	{ key: "triggered", label: "Triggered" },
	{ key: "status", label: "Status", render: row => <Badge tone={toneForStatus(row.status)}>{row.status}</Badge> }
];

export function AlertsPage() {
	const alerts: AlertRecord[] = [];

	return (
		<PageStack>
			<ScreenHeader eyebrow="Alerting" title="Alerts (TBD)" copy="Packet loss, latency, traceroute path change, DNS query errors, abnormal response codes, probe offline, and heartbeat expiry." />

			<ResponsiveGrid>
				<Panel tone="glass" eyebrow="Alert list" title="Active and historical events">
					<DataTable columns={alertColumns} rows={alerts} getRowKey={row => `${row.type}-${row.probe}`} />
				</Panel>
				<Panel tone="deep" eyebrow="Alert detail" title="No alert selected">
					<KeyValueGrid
						items={[
							{ label: "Affected probe", value: "-" },
							{ label: "Affected check", value: "-" },
							{ label: "Threshold", value: "-" },
							{ label: "State", value: "-" }
						]}
					/>
					<ActionRow>
						<Button variant="secondary" disabled>
							Open result history
						</Button>
						<Button variant="danger" disabled>
							Silence 30m
						</Button>
					</ActionRow>
				</Panel>
			</ResponsiveGrid>
		</PageStack>
	);
}

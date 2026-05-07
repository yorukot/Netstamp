import { alerts, type AlertRecord } from "@/features/alerts/data/alerts";
import { ActionRow } from "@/shared/components/ActionRow";
import { KeyValueGrid } from "@/shared/components/KeyValueGrid";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { toneForStatus } from "@/shared/utils/statusTone";
import { Badge, Button, DataTable, Panel, type DataColumn } from "@netstamp/ui";
import styles from "./AlertsPage.module.css";

const alertColumns: DataColumn<AlertRecord>[] = [
	{ key: "type", label: "Alert type" },
	{ key: "check", label: "Check" },
	{ key: "probe", label: "Probe" },
	{ key: "severity", label: "Severity", render: row => <Badge tone={toneForStatus(row.severity)}>{row.severity}</Badge> },
	{ key: "triggered", label: "Triggered" },
	{ key: "status", label: "Status", render: row => <Badge tone={toneForStatus(row.status)}>{row.status}</Badge> }
];

export function AlertsPage() {
	return (
		<PageStack>
			<ScreenHeader eyebrow="Alerting" title="Alerts (TBD)" copy="Packet loss, latency, traceroute path change, DNS query errors, abnormal response codes, probe offline, and heartbeat expiry." />

			<div className={styles.alertGrid}>
				<Panel tone="glass" eyebrow="Alert list" title="Active and historical events">
					<DataTable columns={alertColumns} rows={alerts} getRowKey={row => `${row.type}-${row.probe}`} />
				</Panel>
				<Panel tone="deep" eyebrow="Alert detail" title="packet loss threshold exceeded">
					<KeyValueGrid
						items={[
							{ label: "Affected probe", value: "nyc-vps-03" },
							{ label: "Affected check", value: "api-latency" },
							{ label: "Threshold", value: "loss > 5% for 5m" },
							{ label: "State", value: "active" }
						]}
					/>
					<ActionRow>
						<Button variant="secondary">Open result history</Button>
						<Button variant="danger">Silence 30m</Button>
					</ActionRow>
				</Panel>
			</div>
		</PageStack>
	);
}

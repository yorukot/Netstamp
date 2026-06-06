import type { Probe, ProbeStatus } from "@/features/probes/data/probes";
import { FilterGrid } from "@/shared/components/FilterGrid";
import { Badge, DataTable, Panel, SelectField, TextField, type BadgeTone, type DataColumn } from "@netstamp/ui";
import styles from "./ProbeList.module.css";
import type { ProbeSort } from "./types";

const statusTones: Record<ProbeStatus, BadgeTone> = {
	Online: "success",
	Draining: "warning",
	Offline: "critical"
};
const sortOptions: Array<{ value: ProbeSort; label: string }> = [
	{ value: "heartbeat", label: "Last heartbeat" },
	{ value: "name", label: "Probe name" }
];

const probeColumns: DataColumn<Probe>[] = [
	{ key: "name", label: "Probe name" },
	{ key: "status", label: "Status", render: probe => <Badge tone={statusTones[probe.status]}>{probe.status}</Badge> },
	{ key: "location", label: "Location" },
	{ key: "publicIp", label: "Public IP" },
	{ key: "ipFamily", label: "Support IP Family" },
	{ key: "lastHeartbeat", label: "Last heartbeat" },
	{
		key: "labelTokens",
		label: "Labels",
		render: probe => (
			<span className={styles.labelList}>
				{probe.labelTokens.map(labelToken => (
					<Badge key={labelToken} tone="muted" dot={false}>
						{labelToken}
					</Badge>
				))}
			</span>
		)
	},
	{ key: "version", label: "Version" }
];

interface ProbeListProps {
	probes: Probe[];
	selectedId: string;
	search: string;
	sortKey: ProbeSort;
	onSearchChange: (value: string) => void;
	onSortChange: (value: ProbeSort) => void;
	onSelect: (probeId: string) => void;
}

export function ProbeList({ probes, selectedId, search, sortKey, onSearchChange, onSortChange, onSelect }: ProbeListProps) {
	return (
		<Panel className={styles.panel} tone="glass" title="Probe list" aria-label="Probe list">
			<div className={styles.listStack}>
				<FilterGrid className={styles.filters}>
					<TextField label="Search" placeholder="probe name, location, provider, label" value={search} onChange={event => onSearchChange(event.currentTarget.value)} />
					<SelectField label="Sort" value={sortKey} options={sortOptions} onChange={event => onSortChange(event.currentTarget.value as ProbeSort)} />
				</FilterGrid>

				<DataTable
					ariaLabel="Probes"
					columns={probeColumns}
					rows={probes}
					density="compact"
					minWidth="62rem"
					maxHeight="min(28rem, 46svh)"
					getRowKey={probe => probe.id}
					selectedKey={selectedId}
					onRowClick={probe => onSelect(probe.id)}
					emptyLabel="No probes found"
				/>
			</div>
		</Panel>
	);
}

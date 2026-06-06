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
const statusFilterOptions: Array<{ value: "all" | ProbeStatus; label: string }> = [
	{ value: "all", label: "All statuses" },
	{ value: "Online", label: "Online" },
	{ value: "Draining", label: "Draining" },
	{ value: "Offline", label: "Offline" }
];
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
	providerOptions: string[];
	selectedId: string;
	search: string;
	statusFilter: "all" | ProbeStatus;
	providerFilter: string;
	sortKey: ProbeSort;
	onSearchChange: (value: string) => void;
	onStatusChange: (value: "all" | ProbeStatus) => void;
	onProviderChange: (value: string) => void;
	onSortChange: (value: ProbeSort) => void;
	onSelect: (probeId: string) => void;
}

export function ProbeList({
	probes,
	providerOptions,
	selectedId,
	search,
	statusFilter,
	providerFilter,
	sortKey,
	onSearchChange,
	onStatusChange,
	onProviderChange,
	onSortChange,
	onSelect
}: ProbeListProps) {
	const providerFilterOptions = [
		{ value: "all", label: "All providers" },
		...providerOptions.map(provider => ({
			value: provider,
			label: provider
		}))
	];

	return (
		<Panel className={styles.panel} tone="glass" title="Probe list" aria-label="Probe list">
			<div className={styles.listStack}>
				<FilterGrid className={styles.filters}>
					<TextField label="Search" placeholder="probe name, location, provider, label" value={search} onChange={event => onSearchChange(event.currentTarget.value)} />
					<SelectField label="Status" value={statusFilter} options={statusFilterOptions} onChange={event => onStatusChange(event.currentTarget.value as "all" | ProbeStatus)} />
					<SelectField label="Provider" value={providerFilter} options={providerFilterOptions} onChange={event => onProviderChange(event.currentTarget.value)} />
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

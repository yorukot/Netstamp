import type { Probe, ProbeStatus } from "@/features/probes/data/probes";
import { classNames } from "@/shared/utils/classNames";
import { Badge, DataTable, Input, Panel, Select, type BadgeTone, type DataColumn } from "@netstamp/ui";
import styles from "./ProbeList.module.css";
import type { ProbeSort } from "./types";

const statusTones: Record<ProbeStatus, BadgeTone> = {
	Online: "success",
	Draining: "warning",
	Offline: "critical"
};

const probeColumns: DataColumn<Probe>[] = [
	{ key: "name", label: "Probe name" },
	{ key: "status", label: "Status", render: probe => <Badge tone={statusTones[probe.status]}>{probe.status}</Badge> },
	{ key: "location", label: "Location" },
	{ key: "publicIp", label: "Public IP" },
	{ key: "asn", label: "AS" },
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
	return (
		<Panel
			className={styles.panel}
			tone="glass"
			title="Probe list"
			actions={
				<div className={styles.filters}>
					<Input
						variant="compact"
						frameClassName={styles.controlFrame}
						className={styles.control}
						aria-label="Search probes"
						placeholder="Search"
						value={search}
						onChange={event => onSearchChange(event.currentTarget.value)}
					/>
					<Select
						variant="compact"
						frameClassName={styles.controlFrame}
						className={styles.control}
						aria-label="Filter status"
						value={statusFilter}
						onChange={event => onStatusChange(event.currentTarget.value as "all" | ProbeStatus)}
					>
						<option value="all">Status</option>
						<option value="Online">Online</option>
						<option value="Draining">Draining</option>
						<option value="Offline">Offline</option>
					</Select>
					<Select
						variant="compact"
						frameClassName={styles.controlFrame}
						className={styles.control}
						aria-label="Filter provider"
						value={providerFilter}
						onChange={event => onProviderChange(event.currentTarget.value)}
					>
						<option value="all">Provider</option>
						{providerOptions.map(provider => (
							<option key={provider} value={provider}>
								{provider}
							</option>
						))}
					</Select>
					<Select
						variant="compact"
						frameClassName={classNames(styles.controlFrame, styles.sortControl)}
						className={styles.control}
						aria-label="Sort probes"
						value={sortKey}
						onChange={event => onSortChange(event.currentTarget.value as ProbeSort)}
					>
						<option value="heartbeat">Sort: Last Heartbeat</option>
						<option value="name">Sort: Probe Name</option>
						<option value="asn">Sort: AS</option>
					</Select>
				</div>
			}
			aria-label="Probe list"
		>
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
		</Panel>
	);
}

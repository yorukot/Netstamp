import type { Probe, ProbeStatus } from "@/features/probes/data/probes";
import { FilterGrid } from "@/shared/components/FilterGrid";
import { Badge, DataTable, Panel, PopoverContent, PopoverPortal, PopoverRoot, PopoverTrigger, SelectField, TextField, type BadgeTone, type DataColumn } from "@netstamp/ui";
import type { MouseEvent } from "react";
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
const visibleLabelCount = 2;

function stopRowSelection(event: MouseEvent) {
	event.stopPropagation();
}

function ProbeLabels({ labels }: { labels: string[] }) {
	const visibleLabels = labels.slice(0, visibleLabelCount);
	const hiddenCount = labels.length - visibleLabels.length;

	if (!labels.length) {
		return <span className={styles.labelEmpty}>-</span>;
	}

	return (
		<span className={styles.labelList}>
			{visibleLabels.map(labelToken => (
				<Badge className={styles.labelBadge} title={labelToken} key={labelToken} tone="muted" dot={false}>
					{labelToken}
				</Badge>
			))}
			{hiddenCount > 0 ? (
				<PopoverRoot>
					<span className={styles.labelOverflow}>
						<PopoverTrigger asChild>
							<button type="button" className={styles.labelOverflowButton} aria-label={`Show all ${labels.length} labels`} onClick={stopRowSelection}>
								+{hiddenCount}
							</button>
						</PopoverTrigger>
						<span className={styles.labelHoverCard} aria-hidden="true">
							<ProbeLabelGrid labels={labels} />
						</span>
					</span>
					<PopoverPortal>
						<PopoverContent className={styles.labelPopover} align="start" side="left" sideOffset={8} collisionPadding={8} onClick={stopRowSelection}>
							<ProbeLabelGrid labels={labels} />
						</PopoverContent>
					</PopoverPortal>
				</PopoverRoot>
			) : null}
		</span>
	);
}

function ProbeLabelGrid({ labels }: { labels: string[] }) {
	return (
		<span className={styles.labelGrid}>
			{labels.map(labelToken => (
				<Badge className={styles.labelBadge} title={labelToken} key={labelToken} tone="muted" dot={false}>
					{labelToken}
				</Badge>
			))}
		</span>
	);
}

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
		render: probe => <ProbeLabels labels={probe.labelTokens} />
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

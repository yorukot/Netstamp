import { formatProbeHeartbeat } from "@/features/probes/api/probeAdapters";
import type { Probe, ProbeStatus } from "@/features/probes/data/probes";
import { Badge, DataTable, FilterGrid, PopoverContent, PopoverPortal, PopoverRoot, PopoverTrigger, TextField, type BadgeTone, type DataColumn } from "@netstamp/ui";
import { useEffect, useState, type MouseEvent } from "react";
import styles from "./ProbeList.module.css";

const statusTones: Record<ProbeStatus, BadgeTone> = {
	Online: "success",
	Draining: "warning",
	Offline: "critical"
};
const statusSortRank: Record<ProbeStatus, number> = {
	Online: 0,
	Draining: 1,
	Offline: 2
};
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

function HeartbeatValue({ timestamp }: { timestamp: number | null }) {
	const [now, setNow] = useState(() => Date.now());

	useEffect(() => {
		const interval = window.setInterval(() => setNow(Date.now()), 30 * 1000);
		return () => window.clearInterval(interval);
	}, []);

	return <span title={timestamp ? new Date(timestamp).toLocaleString() : "No heartbeat recorded"}>{formatProbeHeartbeat(timestamp, now)}</span>;
}

const probeColumns: DataColumn<Probe>[] = [
	{ key: "name", label: "Probe name", sortable: true },
	{ key: "status", label: "Status", sortable: true, sortValue: probe => statusSortRank[probe.status], render: probe => <Badge tone={statusTones[probe.status]}>{probe.status}</Badge> },
	{ key: "location", label: "Location", sortable: true },
	{ key: "publicIp", label: "Public IP", sortable: true },
	{ key: "ipFamily", label: "Support IP Family", sortable: true },
	{
		key: "lastHeartbeat",
		label: "Last heartbeat",
		sortable: true,
		sortValue: probe => probe.lastHeartbeatAt ?? Number.NEGATIVE_INFINITY,
		render: probe => <HeartbeatValue timestamp={probe.lastHeartbeatAt} />
	},
	{
		key: "labelTokens",
		label: "Labels",
		sortable: true,
		sortValue: probe => probe.labelTokens.join(" "),
		render: probe => <ProbeLabels labels={probe.labelTokens} />
	},
	{ key: "version", label: "Version", sortable: true }
];

interface ProbeListProps {
	probes: Probe[];
	selectedId: string;
	search: string;
	onSearchChange: (value: string) => void;
	onSelect: (probeId: string) => void;
}

export function ProbeList({ probes, selectedId, search, onSearchChange, onSelect }: ProbeListProps) {
	return (
		<div className={styles.listStack}>
			<FilterGrid className={styles.filters}>
				<TextField label="Search" placeholder="probe name, location, provider, label" value={search} onChange={event => onSearchChange(event.currentTarget.value)} />
			</FilterGrid>

			<DataTable
				ariaLabel="Probes"
				columns={probeColumns}
				rows={probes}
				minWidth="62rem"
				maxHeight="min(28rem, 46svh)"
				defaultSort={{ key: "lastHeartbeat", direction: "desc" }}
				getRowKey={probe => probe.id}
				selectedKey={selectedId}
				onRowClick={probe => onSelect(probe.id)}
				emptyLabel="No probes found"
			/>
		</div>
	);
}

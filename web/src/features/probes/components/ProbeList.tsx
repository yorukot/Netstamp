import type { Probe, ProbeStatus } from "@/features/probes/data/probes";
import { useLocaleFormat } from "@/i18n/format";
import { Badge, DataTable, FilterGrid, PopoverContent, PopoverPortal, PopoverRoot, PopoverTrigger, TextField, type BadgeTone, type DataColumn } from "@netstamp/ui";
import { useEffect, useState, type MouseEvent } from "react";
import { useTranslation } from "react-i18next";
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
	const { t } = useTranslation("probes");
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
							<button type="button" className={styles.labelOverflowButton} aria-label={t("table.showLabels", { count: labels.length })} onClick={stopRowSelection}>
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
	const { t } = useTranslation("probes");
	const { dateTime, relativeTime } = useLocaleFormat();
	const [now, setNow] = useState(() => Date.now());

	useEffect(() => {
		const interval = window.setInterval(() => setNow(Date.now()), 30 * 1000);
		return () => window.clearInterval(interval);
	}, []);

	if (timestamp === null) {
		return <span title={t("table.noHeartbeat")}>{t("relative.never")}</span>;
	}

	const elapsedSeconds = Math.max(0, Math.floor((now - timestamp) / 1000));
	const value =
		elapsedSeconds < 60
			? relativeTime(-elapsedSeconds, "second")
			: elapsedSeconds < 3600
				? relativeTime(-Math.floor(elapsedSeconds / 60), "minute")
				: relativeTime(-Math.floor(elapsedSeconds / 3600), "hour");

	return <span title={dateTime(timestamp)}>{value}</span>;
}

interface ProbeListProps {
	probes: Probe[];
	selectedId: string;
	search: string;
	onSearchChange: (value: string) => void;
	onSelect: (probeId: string) => void;
}

export function ProbeList({ probes, selectedId, search, onSearchChange, onSelect }: ProbeListProps) {
	const { t } = useTranslation("probes");
	const statusLabel = (status: ProbeStatus) => t(`status.${status.toLowerCase() as "online" | "draining" | "offline"}`);
	const probeColumns: DataColumn<Probe>[] = [
		{ key: "name", label: t("table.name"), sortable: true },
		{
			key: "status",
			label: t("table.status"),
			sortable: true,
			sortValue: probe => statusSortRank[probe.status],
			render: probe => <Badge tone={statusTones[probe.status]}>{statusLabel(probe.status)}</Badge>
		},
		{ key: "location", label: t("table.location"), sortable: true },
		{ key: "publicIp", label: t("table.publicIp"), sortable: true },
		{ key: "ipFamily", label: t("table.ipFamily"), sortable: true },
		{
			key: "lastHeartbeat",
			label: t("table.lastHeartbeat"),
			sortable: true,
			sortValue: probe => probe.lastHeartbeatAt ?? Number.NEGATIVE_INFINITY,
			render: probe => <HeartbeatValue timestamp={probe.lastHeartbeatAt} />
		},
		{
			key: "labelTokens",
			label: t("table.labels"),
			sortable: true,
			sortValue: probe => probe.labelTokens.join(" "),
			render: probe => <ProbeLabels labels={probe.labelTokens} />
		},
		{ key: "version", label: t("table.version"), sortable: true }
	];

	return (
		<div className={styles.listStack}>
			<FilterGrid className={styles.filters}>
				<TextField label={t("search.label")} placeholder={t("search.placeholder")} value={search} onChange={event => onSearchChange(event.currentTarget.value)} />
			</FilterGrid>

			<DataTable
				ariaLabel={t("table.aria")}
				columns={probeColumns}
				rows={probes}
				minWidth="62rem"
				maxHeight="min(28rem, 46svh)"
				defaultSort={{ key: "lastHeartbeat", direction: "desc" }}
				getRowKey={probe => probe.id}
				selectedKey={selectedId}
				onRowClick={probe => onSelect(probe.id)}
				emptyLabel={t("table.empty")}
			/>
		</div>
	);
}

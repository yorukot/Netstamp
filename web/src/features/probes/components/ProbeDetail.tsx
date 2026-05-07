import type { Probe } from "@/features/probes/data/probes";
import { classNames } from "@/shared/utils/classNames";
import { Badge, Checkbox, DataTable, Surface, TextField, type DataColumn } from "@netstamp/ui";
import { useState } from "react";
import styles from "./ProbeDetail.module.css";
import { expandAssignedRows } from "./probeUtils";
import type { AssignedRow, DetectionMode } from "./types";

const assignedColumns: DataColumn<AssignedRow>[] = [
	{ key: "check", label: "Assigned check" },
	{ key: "type", label: "Type", render: row => <Badge tone="neutral">{row.type}</Badge> },
	{ key: "interval", label: "Interval" },
	{ key: "jitter", label: "Jitter" },
	{ key: "latest", label: "Latest" }
];

interface ProbeDetailProps {
	probe: Probe;
	assignedRows: AssignedRow[];
	floating?: boolean;
}

export function ProbeDetail({ probe, assignedRows, floating = false }: ProbeDetailProps) {
	const [probeName, setProbeName] = useState(probe.name);
	const [probeLocation, setProbeLocation] = useState(probe.location);
	const [probeAsn, setProbeAsn] = useState(probe.asn);
	const [locationMode, setLocationMode] = useState<DetectionMode>("manual");
	const [asMode, setAsMode] = useState<DetectionMode>("auto");
	const probeAssignments = assignedRows.filter(row => row.probe === probe.name);
	const baseRows = probeAssignments.length ? probeAssignments : assignedRows.filter(row => row.check === "api-latency");
	const detailRows = expandAssignedRows(baseRows);

	function toggleLocationMode() {
		const nextMode = locationMode === "manual" ? "auto" : "manual";

		setLocationMode(nextMode);

		if (nextMode === "auto") {
			setProbeLocation(probe.location);
		}
	}

	function toggleAsMode() {
		const nextMode = asMode === "manual" ? "auto" : "manual";

		setAsMode(nextMode);

		if (nextMode === "auto") {
			setProbeAsn(probe.asn);
		}
	}

	return (
		<Surface as="section" tone="matte" cut="lg" padding="lg" className={classNames(styles.card, floating && styles.floating)} aria-label="Probe detail">
			<div className={styles.header}>
				<span>Probe detail</span>
				<strong>
					{probe.name}
					<small> · uptime {probe.uptime}</small>
				</strong>
			</div>

			<div className={styles.fieldGrid}>
				<TextField className={styles.input} label="Probe name" value={probeName} onChange={event => setProbeName(event.currentTarget.value)} />
				<div className={styles.inputWithMode}>
					<TextField
						className={styles.input}
						label="Location (keywords search)"
						value={probeLocation}
						disabled={locationMode === "auto"}
						onChange={event => setProbeLocation(event.currentTarget.value)}
					/>
					<ModeToggle mode={locationMode} label="location detect mode" onToggle={toggleLocationMode} />
				</div>
				<div className={styles.inputWithMode}>
					<TextField className={styles.input} label="AS" value={probeAsn} disabled={asMode === "auto"} onChange={event => setProbeAsn(event.currentTarget.value)} />
					<ModeToggle mode={asMode} label="AS detect mode" onToggle={toggleAsMode} />
				</div>
			</div>

			<DataTable
				ariaLabel="Assigned checks"
				columns={assignedColumns}
				rows={detailRows}
				density="compact"
				minWidth="31rem"
				maxHeight="11.75rem"
				getRowKey={(row, index) => `${row.probe}-${row.check}-${index}`}
			/>
		</Surface>
	);
}

interface ModeToggleProps {
	mode: DetectionMode;
	label: string;
	onToggle: () => void;
}

function ModeToggle({ mode, label, onToggle }: ModeToggleProps) {
	const modeClass = mode === "manual" ? styles.modeToggleManual : styles.modeToggleAuto;

	return (
		<label className={classNames(styles.modeToggle, modeClass)}>
			<Checkbox className={styles.modeInput} checked={mode === "auto"} aria-label={label} onChange={onToggle} />
			<span className={styles.modePill}>
				<span className={styles.modeDot} aria-hidden="true" />
				<span>{mode}</span>
			</span>
		</label>
	);
}

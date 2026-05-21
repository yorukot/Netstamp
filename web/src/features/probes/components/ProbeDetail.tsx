import { mapApiProbe } from "@/features/probes/api/probeAdapters";
import type { Probe } from "@/features/probes/data/probes";
import { probeSecretUpdateCommand } from "@/shared/api/installAssets";
import { useDeleteProjectProbeMutation, useRotateProjectProbeSecretMutation, useUpdateProjectProbeMutation } from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import { useConfirm } from "@/shared/components/confirmContext";
import { classNames } from "@/shared/utils/classNames";
import { Badge, Button, Checkbox, DataTable, Surface, Terminal, TextField, type DataColumn } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
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
	projectRef?: string | null;
	onDeleted?: () => void;
}

export function ProbeDetail({ probe, assignedRows, floating = false, projectRef, onDeleted }: ProbeDetailProps) {
	const confirm = useConfirm();
	const detailQuery = useQuery({
		...projectQueries.probeDetail(projectRef || "", probe.id),
		enabled: Boolean(projectRef && probe.id),
		select: data => mapApiProbe(data.probe, 0)
	});
	const activeProbe = detailQuery.data || probe;
	const [probeName, setProbeName] = useState(activeProbe.name);
	const [probeLocation, setProbeLocation] = useState(activeProbe.location);
	const [probeAsn, setProbeAsn] = useState(activeProbe.asn);
	const [locationMode, setLocationMode] = useState<DetectionMode>("manual");
	const [asMode, setAsMode] = useState<DetectionMode>("auto");
	const [rotatedSecret, setRotatedSecret] = useState("");
	const updateProbeMutation = useUpdateProjectProbeMutation(projectRef);
	const deleteProbeMutation = useDeleteProjectProbeMutation(projectRef);
	const rotateSecretMutation = useRotateProjectProbeSecretMutation(projectRef);
	const probeAssignments = assignedRows.filter(row => row.probe === activeProbe.name);
	const detailRows = expandAssignedRows(probeAssignments);
	const rotatedSecretCommand = rotatedSecret ? probeSecretUpdateCommand({ probeId: activeProbe.id, probeSecret: rotatedSecret }) : "";

	function toggleLocationMode() {
		const nextMode = locationMode === "manual" ? "auto" : "manual";

		setLocationMode(nextMode);

		if (nextMode === "auto") {
			setProbeLocation(activeProbe.location);
		}
	}

	function toggleAsMode() {
		const nextMode = asMode === "manual" ? "auto" : "manual";

		setAsMode(nextMode);

		if (nextMode === "auto") {
			setProbeAsn(activeProbe.asn);
		}
	}

	async function deleteProbe() {
		const confirmed = await confirm({
			title: `Delete ${activeProbe.name}?`,
			message: "This removes the probe from the fleet and stops future check assignments for it.",
			confirmLabel: "Delete probe",
			tone: "danger"
		});

		if (!confirmed) {
			return;
		}

		deleteProbeMutation.mutate(activeProbe.id, { onSuccess: () => onDeleted?.() });
	}

	return (
		<Surface as="section" tone="matte" cut="lg" padding="lg" className={classNames(styles.card, floating && styles.floating)} aria-label="Probe detail">
			<div className={styles.header}>
				<span>Probe detail</span>
				<strong>
					{activeProbe.name}
					<small> · uptime {activeProbe.uptime}</small>
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

			<div className={styles.actions}>
				<Button disabled={!projectRef || updateProbeMutation.isPending || !probeName} onClick={() => updateProbeMutation.mutate({ probeId: activeProbe.id, body: { name: probeName } })}>
					{updateProbeMutation.isPending ? "Saving" : "Save probe"}
				</Button>
				<Button
					variant="outline"
					disabled={!projectRef || rotateSecretMutation.isPending}
					onClick={() => rotateSecretMutation.mutate(activeProbe.id, { onSuccess: data => setRotatedSecret(data.secret) })}
				>
					{rotateSecretMutation.isPending ? "Rotating" : "Rotate secret"}
				</Button>
				<Button variant="danger" disabled={!projectRef || deleteProbeMutation.isPending} onClick={() => void deleteProbe()}>
					{deleteProbeMutation.isPending ? "Deleting" : "Delete probe"}
				</Button>
			</div>

			{rotatedSecret ? (
				<div className={classNames("ns-cut-frame", styles.secretPanel)}>
					<span>New probe secret</span>
					<code>{rotatedSecret}</code>
					<p>Run this on the probe host to rewrite the systemd service environment with the rotated credential.</p>
					<Terminal title="update command" meta="run on probe host">
						{rotatedSecretCommand}
					</Terminal>
				</div>
			) : null}

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

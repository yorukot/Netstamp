import type { Probe } from "@/features/probes/data/probes";
import { mapApiProbe } from "@/features/probes/api/probeAdapters";
import { apiQueryKeys } from "@/shared/api/queryKeys";
import { deleteProjectProbe, projectQueries, rotateProjectProbeSecret, updateProjectProbe } from "@/shared/api/queries";
import { classNames } from "@/shared/utils/classNames";
import { Badge, Button, Checkbox, DataTable, Surface, TextField, type DataColumn } from "@netstamp/ui";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
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
	const queryClient = useQueryClient();
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
	const updateProbeMutation = useMutation({
		mutationFn: () => updateProjectProbe(projectRef || "", activeProbe.id, { name: probeName }),
		onSuccess: () => {
			if (projectRef) {
				queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.probes(projectRef) });
				queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.probeDetail(projectRef, activeProbe.id) });
			}
		}
	});
	const deleteProbeMutation = useMutation({
		mutationFn: () => deleteProjectProbe(projectRef || "", activeProbe.id),
		onSuccess: () => {
			if (projectRef) {
				queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.probes(projectRef) });
			}

			onDeleted?.();
		}
	});
	const rotateSecretMutation = useMutation({
		mutationFn: () => rotateProjectProbeSecret(projectRef || "", activeProbe.id),
		onSuccess: data => setRotatedSecret(data.secret)
	});
	const probeAssignments = assignedRows.filter(row => row.probe === activeProbe.name);
	const baseRows = probeAssignments.length ? probeAssignments : assignedRows.filter(row => row.check === "api-latency");
	const detailRows = expandAssignedRows(baseRows);

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

	function deleteProbe() {
		if (!window.confirm(`Delete probe ${activeProbe.name}?`)) {
			return;
		}

		deleteProbeMutation.mutate();
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
				<Button disabled={!projectRef || updateProbeMutation.isPending || !probeName} onClick={() => updateProbeMutation.mutate()}>
					{updateProbeMutation.isPending ? "Saving" : "Save probe"}
				</Button>
				<Button variant="outline" disabled={!projectRef || rotateSecretMutation.isPending} onClick={() => rotateSecretMutation.mutate()}>
					{rotateSecretMutation.isPending ? "Rotating" : "Rotate secret"}
				</Button>
				<Button variant="danger" disabled={!projectRef || deleteProbeMutation.isPending} onClick={deleteProbe}>
					{deleteProbeMutation.isPending ? "Deleting" : "Delete probe"}
				</Button>
			</div>

			{rotatedSecret ? (
				<div className={classNames("ns-cut-frame", styles.secretPanel)}>
					<span>New probe secret</span>
					<code>{rotatedSecret}</code>
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

import { mapApiProbe } from "@/features/probes/api/probeAdapters";
import {
	coordinateInputError,
	formatCoordinate,
	parseCoordinateInput,
	searchNominatimLocation,
	type CoordinateInputMode,
	type GeocodeStatus,
	type ProbeCoordinates
} from "@/features/probes/data/probeLocation";
import type { Probe } from "@/features/probes/data/probes";
import { ApiError } from "@/shared/api/client";
import { probeSecretUpdateCommand } from "@/shared/api/installAssets";
import { createProjectLabel, useDeleteProjectProbeMutation, useRotateProjectProbeSecretMutation, useUpdateProjectProbeMutation } from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import { apiQueryKeys } from "@/shared/api/queryKeys";
import type { ApiLabel, ApiProbe } from "@/shared/api/types";
import { CloseButton } from "@/shared/components/CloseButton";
import { useConfirm } from "@/shared/components/confirmContext";
import { pushErrorToast } from "@/shared/toast/toastStore";
import { classNames } from "@/shared/utils/classNames";
import { Badge, Button, DataTable, Surface, Terminal, TextField, type DataColumn } from "@netstamp/ui";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useRef, useState } from "react";
import { LocationPreviewMap } from "./LocationPreviewMap";
import styles from "./ProbeDetail.module.css";
import { expandAssignedRows } from "./probeUtils";
import type { AssignedRow } from "./types";

const assignedColumns: DataColumn<AssignedRow>[] = [
	{ key: "check", label: "Assigned check" },
	{ key: "type", label: "Type", render: row => <Badge tone="neutral">{row.type}</Badge> },
	{ key: "interval", label: "Interval" },
	{ key: "latest", label: "Latest" }
];

function initialLatitude(probe: Probe) {
	return probe.coordinates ? formatCoordinate(probe.coordinates[1]) : "";
}

function initialLongitude(probe: Probe) {
	return probe.coordinates ? formatCoordinate(probe.coordinates[0]) : "";
}

async function writeClipboardText(value: string) {
	if (navigator.clipboard?.writeText) {
		await navigator.clipboard.writeText(value);
		return;
	}

	const textarea = document.createElement("textarea");
	textarea.value = value;
	textarea.setAttribute("readonly", "");
	textarea.style.position = "fixed";
	textarea.style.left = "-9999px";
	document.body.appendChild(textarea);
	textarea.select();
	document.execCommand("copy");
	textarea.remove();
}

interface ProbeDetailProps {
	probe: Probe;
	assignedRows: AssignedRow[];
	floating?: boolean;
	projectRef?: string | null;
	onClose?: () => void;
	onDeleted?: () => void;
}

export function ProbeDetail({ probe, assignedRows, floating = false, projectRef, onClose, onDeleted }: ProbeDetailProps) {
	const detailQuery = useQuery({
		...projectQueries.probeDetail(projectRef || "", probe.id),
		enabled: Boolean(projectRef && probe.id)
	});
	const activeApiProbe = detailQuery.data?.probe ?? null;
	const activeProbe = activeApiProbe ? mapApiProbe(activeApiProbe, 0) : probe;
	const formKey = `${activeProbe.id}:${activeApiProbe?.updatedAt ?? "pending"}`;

	return (
		<ProbeDetailContent
			key={formKey}
			activeProbe={activeProbe}
			activeApiProbe={activeApiProbe}
			assignedRows={assignedRows}
			floating={floating}
			projectRef={projectRef}
			onClose={onClose}
			onDeleted={onDeleted}
		/>
	);
}

interface ProbeDetailContentProps {
	activeProbe: Probe;
	activeApiProbe: ApiProbe | null;
	assignedRows: AssignedRow[];
	floating?: boolean;
	projectRef?: string | null;
	onClose?: () => void;
	onDeleted?: () => void;
}

function ProbeDetailContent({ activeProbe, activeApiProbe, assignedRows, floating = false, projectRef, onClose, onDeleted }: ProbeDetailContentProps) {
	const confirm = useConfirm();
	const queryClient = useQueryClient();
	const copyTimeoutRef = useRef<number | null>(null);
	const labelsQuery = useQuery({
		...projectQueries.labels(projectRef || ""),
		enabled: Boolean(projectRef)
	});
	const [probeName, setProbeName] = useState(activeProbe.name);
	const [coordinateInputMode, setCoordinateInputMode] = useState<CoordinateInputMode>("search");
	const [locationSearch, setLocationSearch] = useState(activeProbe.location === "-" ? "" : activeProbe.location);
	const [locationName, setLocationName] = useState(activeProbe.location === "-" ? "" : activeProbe.location);
	const [latitudeInput, setLatitudeInput] = useState(initialLatitude(activeProbe));
	const [longitudeInput, setLongitudeInput] = useState(initialLongitude(activeProbe));
	const [geocodeStatus, setGeocodeStatus] = useState<GeocodeStatus>("idle");
	const [geocodeError, setGeocodeError] = useState("");
	const [probeAsn, setProbeAsn] = useState(activeProbe.asn);
	const [rotatedSecret, setRotatedSecret] = useState("");
	const [secretCommandCopied, setSecretCommandCopied] = useState(false);
	const [savingProbe, setSavingProbe] = useState(false);
	const geocodeAbortRef = useRef<AbortController | null>(null);
	const updateProbeMutation = useUpdateProjectProbeMutation(projectRef);
	const deleteProbeMutation = useDeleteProjectProbeMutation(projectRef);
	const rotateSecretMutation = useRotateProjectProbeSecretMutation(projectRef);
	const probeAssignments = assignedRows.filter(row => row.probe === activeProbe.name);
	const detailRows = expandAssignedRows(probeAssignments);
	const rotatedSecretCommand = rotatedSecret ? probeSecretUpdateCommand({ probeId: activeProbe.id, probeSecret: rotatedSecret }) : "";
	const latitude = parseCoordinateInput(latitudeInput);
	const longitude = parseCoordinateInput(longitudeInput);
	const latitudeError = coordinateInputMode === "manual" ? coordinateInputError("Latitude", latitudeInput, -90, 90) : "";
	const longitudeError = coordinateInputMode === "manual" ? coordinateInputError("Longitude", longitudeInput, -180, 180) : "";
	const visibleLatitudeError = latitudeInput.trim() ? latitudeError : "";
	const visibleLongitudeError = longitudeInput.trim() ? longitudeError : "";
	const searchCoordinatesReady = coordinateInputMode === "search" && latitude !== null && longitude !== null;
	const manualCoordinatesReady = coordinateInputMode === "manual" && latitude !== null && longitude !== null && !latitudeError && !longitudeError;
	const coordinatesReady = searchCoordinatesReady || manualCoordinatesReady;
	const selectedCoordinates: ProbeCoordinates | null = coordinatesReady && latitude !== null && longitude !== null ? { latitude, longitude } : null;
	const canSearchLocation = coordinateInputMode === "search" && locationSearch.trim().length > 0 && geocodeStatus !== "searching";
	const locationInputInvalid = coordinateInputMode === "manual" && Boolean(latitudeInput.trim() || longitudeInput.trim()) && !manualCoordinatesReady;
	const previewLocationName = locationName.trim() || "Manual coordinates";

	useEffect(() => {
		return () => {
			if (copyTimeoutRef.current) {
				window.clearTimeout(copyTimeoutRef.current);
			}
			geocodeAbortRef.current?.abort();
		};
	}, []);

	function clearResolvedCoordinates() {
		geocodeAbortRef.current?.abort();
		geocodeAbortRef.current = null;
		setLocationName("");
		setLatitudeInput("");
		setLongitudeInput("");
		setGeocodeStatus("idle");
		setGeocodeError("");
	}

	function updateLocationSearch(value: string) {
		setLocationSearch(value);
		clearResolvedCoordinates();
	}

	function updateCoordinateInputMode(nextMode: CoordinateInputMode) {
		if (nextMode === coordinateInputMode) {
			return;
		}

		setCoordinateInputMode(nextMode);
		setGeocodeStatus("idle");
		setGeocodeError("");

		if (nextMode === "search") {
			setLocationName("");
			setLatitudeInput("");
			setLongitudeInput("");
		}
	}

	async function searchLocation() {
		const query = locationSearch.trim();

		if (!query || geocodeStatus === "searching") {
			return;
		}

		geocodeAbortRef.current?.abort();
		const abortController = new AbortController();
		geocodeAbortRef.current = abortController;

		setGeocodeStatus("searching");
		setGeocodeError("");
		setLocationName("");
		setLatitudeInput("");
		setLongitudeInput("");

		try {
			const result = await searchNominatimLocation(query, abortController.signal);

			setLocationName(result.locationName);
			setLatitudeInput(formatCoordinate(result.coordinates.latitude));
			setLongitudeInput(formatCoordinate(result.coordinates.longitude));
			setGeocodeStatus("resolved");
		} catch (error) {
			if (error instanceof DOMException && error.name === "AbortError") {
				return;
			}

			setGeocodeStatus("error");
			setGeocodeError(error instanceof Error ? error.message : "Location search failed.");
		} finally {
			if (geocodeAbortRef.current === abortController) {
				geocodeAbortRef.current = null;
			}
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

	async function ensureProjectLabel(key: string, value: string, labels: ApiLabel[]) {
		const normalizedValue = value.trim();
		if (!projectRef || !normalizedValue || normalizedValue === "-") {
			return null;
		}

		const existing = labels.find(label => label.key.toLowerCase() === key && label.value === normalizedValue);
		if (existing) {
			return existing.id;
		}

		try {
			const created = await createProjectLabel(projectRef, { key, value: normalizedValue });
			return created.label.id;
		} catch (error) {
			if (!(error instanceof ApiError && error.status === 409)) {
				throw error;
			}

			const refreshed = await labelsQuery.refetch();
			return refreshed.data?.labels.find(label => label.key.toLowerCase() === key && label.value === normalizedValue)?.id ?? null;
		}
	}

	async function saveProbe() {
		if (!projectRef || !activeApiProbe || !probeName.trim()) {
			return;
		}

		setSavingProbe(true);
		try {
			const projectLabels = labelsQuery.data?.labels ?? [];
			const currentLabels = activeApiProbe?.labels ?? [];
			const labelIds = currentLabels.filter(label => label.key.toLowerCase() !== "as").map(label => label.id);
			const asLabelId = await ensureProjectLabel("as", probeAsn, projectLabels);

			for (const labelId of [asLabelId]) {
				if (labelId && !labelIds.includes(labelId)) {
					labelIds.push(labelId);
				}
			}

			const body = {
				name: probeName.trim(),
				...(locationName.trim() ? { locationName: locationName.trim() } : {}),
				...(selectedCoordinates ? { latitude: selectedCoordinates.latitude, longitude: selectedCoordinates.longitude } : {}),
				labelIds
			};

			await updateProbeMutation.mutateAsync({
				probeId: activeProbe.id,
				body
			});
			await queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.labels(projectRef) });
		} catch (error) {
			pushErrorToast(error instanceof Error ? error.message : "Probe could not be saved.");
		} finally {
			setSavingProbe(false);
		}
	}

	async function copyRotatedSecretCommand() {
		if (!rotatedSecretCommand) {
			return;
		}

		try {
			await writeClipboardText(rotatedSecretCommand);
		} catch {
			setSecretCommandCopied(false);
			return;
		}

		setSecretCommandCopied(true);

		if (copyTimeoutRef.current) {
			window.clearTimeout(copyTimeoutRef.current);
		}

		copyTimeoutRef.current = window.setTimeout(() => {
			setSecretCommandCopied(false);
			copyTimeoutRef.current = null;
		}, 1800);
	}

	return (
		<Surface as="section" tone="matte" cut="lg" padding="lg" className={classNames(styles.card, floating && styles.floating)} aria-label="Probe detail">
			<div className={styles.header}>
				<strong>
					{activeProbe.name}
					<small> · uptime {activeProbe.uptime}</small>
				</strong>
				{onClose ? <CloseButton ariaLabel="Close probe options" onClick={onClose} /> : null}
			</div>

			<div className={styles.fieldGrid}>
				<TextField className={styles.input} label="Probe name" value={probeName} onChange={event => setProbeName(event.currentTarget.value)} />
				<TextField className={styles.input} label="AS" value={probeAsn} onChange={event => setProbeAsn(event.currentTarget.value)} />
			</div>

			<div className={styles.locationEditor}>
				<div className={styles.locationMode} role="group" aria-label="Coordinate input mode">
					<Button
						type="button"
						variant={coordinateInputMode === "search" ? "secondary" : "ghost"}
						size="sm"
						aria-pressed={coordinateInputMode === "search"}
						onClick={() => updateCoordinateInputMode("search")}
					>
						Search name
					</Button>
					<Button
						type="button"
						variant={coordinateInputMode === "manual" ? "secondary" : "ghost"}
						size="sm"
						aria-pressed={coordinateInputMode === "manual"}
						onClick={() => updateCoordinateInputMode("manual")}
					>
						Manual coordinates
					</Button>
				</div>

				<div className={styles.locationBody}>
					<div className={styles.locationControls}>
						{coordinateInputMode === "search" ? (
							<div className={styles.locationSearch}>
								<TextField
									className={styles.input}
									label="Location search"
									value={locationSearch}
									placeholder="Taipei, Taiwan"
									disabled={geocodeStatus === "searching"}
									error={geocodeStatus === "error" ? geocodeError : undefined}
									onChange={event => updateLocationSearch(event.currentTarget.value)}
								/>
								<Button type="button" variant="outline" size="sm" disabled={!canSearchLocation} onClick={() => void searchLocation()}>
									{geocodeStatus === "searching" ? "Searching" : "Search"}
								</Button>
							</div>
						) : (
							<div className={styles.manualLocationFields}>
								<TextField className={styles.input} label="Location name" value={locationName} placeholder="Taipei, Taiwan" onChange={event => setLocationName(event.currentTarget.value)} />
								<div className={styles.coordinateGrid}>
									<TextField
										className={styles.input}
										label="Latitude"
										type="number"
										inputMode="decimal"
										step="any"
										value={latitudeInput}
										placeholder="25.037520"
										error={visibleLatitudeError || undefined}
										onChange={event => setLatitudeInput(event.currentTarget.value)}
									/>
									<TextField
										className={styles.input}
										label="Longitude"
										type="number"
										inputMode="decimal"
										step="any"
										value={longitudeInput}
										placeholder="121.563680"
										error={visibleLongitudeError || undefined}
										onChange={event => setLongitudeInput(event.currentTarget.value)}
									/>
								</div>
							</div>
						)}

						{selectedCoordinates ? null : (
							<p className={styles.locationStatus} aria-live="polite">
								{coordinateInputMode === "search" ? "Search for a place to update this probe location." : "Enter valid decimal coordinates to preview this probe location."}
							</p>
						)}
					</div>

					{selectedCoordinates ? (
						<LocationPreviewMap coordinates={selectedCoordinates} locationName={previewLocationName} probeName={probeName.trim() || activeProbe.name} className={styles.locationPreview} />
					) : null}
				</div>
			</div>

			<div className={styles.actions}>
				<Button
					disabled={!projectRef || !activeApiProbe || updateProbeMutation.isPending || savingProbe || !probeName || locationInputInvalid || geocodeStatus === "searching"}
					onClick={() => void saveProbe()}
				>
					{updateProbeMutation.isPending || savingProbe ? "Saving" : "Save probe"}
				</Button>
				<Button
					variant="outline"
					disabled={!projectRef || rotateSecretMutation.isPending}
					onClick={() =>
						rotateSecretMutation.mutate(activeProbe.id, {
							onSuccess: data => {
								setRotatedSecret(data.secret);
								setSecretCommandCopied(false);
							}
						})
					}
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
					<p>Rewrite the systemd service environment with the rotated credential.</p>
					<Terminal
						title="update command"
						meta={
							<Button type="button" variant="ghost" size="sm" disabled={!rotatedSecretCommand} onClick={() => void copyRotatedSecretCommand()}>
								{secretCommandCopied ? "Copied" : "Copy to clipboard"}
							</Button>
						}
					>
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

import { mapApiProbe } from "@/features/probes/api/probeAdapters";
import {
	coordinateInputError,
	formatCoordinate,
	LocationSearchError,
	parseCoordinateInput,
	searchNominatimLocation,
	type CoordinateInputMode,
	type GeocodeStatus,
	type ProbeCoordinates
} from "@/features/probes/data/probeLocation";
import type { Probe } from "@/features/probes/data/probes";
import { probeReinstallCommand, probeSecretUpdateCommand, probeUpgradeCommand } from "@/shared/api/installAssets";
import { useDeleteProjectProbeMutation, useRotateProjectProbeSecretMutation, useUpdateProjectProbeMutation } from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import type { ApiLabel, ApiProbe } from "@/shared/api/types";
import { CloseButton } from "@/shared/components/CloseButton";
import { useConfirm } from "@/shared/components/confirmContext";
import { UnsavedChangesBar } from "@/shared/components/UnsavedChangesBar";
import type { AssignedRow } from "@/shared/domain/assignments";
import { pushErrorToast } from "@/shared/toast/toastStore";
import { classNames } from "@/shared/utils/classNames";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { Badge, Button, Checkbox, CodeBlock, DataTable, FieldLabel, SegmentedControl, Spinner, TextField, type DataColumn } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { LocationPreviewMap } from "./LocationPreviewMap";
import styles from "./ProbeDetail.module.css";
import { expandAssignedRows } from "./probeUtils";

type ProbeServiceCommandMode = "reinstall" | "upgrade";

function initialLatitude(probe: Probe) {
	return probe.coordinates ? formatCoordinate(probe.coordinates[1]) : "";
}

function initialLongitude(probe: Probe) {
	return probe.coordinates ? formatCoordinate(probe.coordinates[0]) : "";
}

function probeServiceCommand(mode: ProbeServiceCommandMode) {
	if (mode === "reinstall") {
		return probeReinstallCommand();
	}
	return probeUpgradeCommand();
}

function labelToken(label: Pick<ApiLabel, "key" | "value">) {
	return `${label.key}:${label.value}`;
}

function sortLabels(left: ApiLabel, right: ApiLabel) {
	return left.key.localeCompare(right.key) || left.value.localeCompare(right.value);
}

function sameStringSet(left: string[], right: string[]) {
	if (left.length !== right.length) {
		return false;
	}

	const rightSet = new Set(right);
	return left.every(value => rightSet.has(value));
}

interface ProbeLabelGroup {
	key: string;
	labels: ApiLabel[];
}

function groupProbeLabels(labels: ApiLabel[]): ProbeLabelGroup[] {
	const groups = new Map<string, ApiLabel[]>();

	for (const label of labels) {
		const values = groups.get(label.key);
		if (values) {
			values.push(label);
		} else {
			groups.set(label.key, [label]);
		}
	}

	return Array.from(groups, ([key, groupedLabels]) => ({ key, labels: groupedLabels }));
}

function ProbeLabelPicker({ labels, selectedLabelSet, onToggle }: { labels: ApiLabel[]; selectedLabelSet: Set<string>; onToggle: (labelId: string, checked: boolean) => void }) {
	const { t } = useTranslation("probes");
	const groups = useMemo(() => groupProbeLabels(labels), [labels]);
	const [requestedKey, setRequestedKey] = useState(() => groups.find(group => group.labels.some(label => selectedLabelSet.has(label.id)))?.key ?? groups[0]?.key ?? "");
	const activeGroup = groups.find(group => group.key === requestedKey) ?? groups[0];

	if (!activeGroup) {
		return null;
	}

	return (
		<div className={styles.labelPicker}>
			<div className={["ns-scrollbar", styles.labelKeyList].join(" ")} role="group" aria-label={t("detail.labelKeys")}>
				{groups.map(group => {
					const selectedValues = group.labels.filter(label => selectedLabelSet.has(label.id)).map(label => label.value);
					const selectedSummary = selectedValues.length ? selectedValues.join(", ") : t("detail.noValues");
					const active = group.key === activeGroup.key;

					return (
						<button key={group.key} type="button" className={styles.labelKeyOption} aria-pressed={active} data-active={active || undefined} onClick={() => setRequestedKey(group.key)}>
							<span className={styles.labelKeyName}>{group.key}</span>
							<span className={styles.labelKeySummary}>{selectedSummary}</span>
						</button>
					);
				})}
			</div>

			<section className={styles.labelValuePanel} aria-label={t("detail.labelValues", { key: activeGroup.key })}>
				<div className={["ns-scrollbar", styles.labelValueList].join(" ")} role="group" aria-label={t("detail.selectValues", { key: activeGroup.key })}>
					{activeGroup.labels.map(label => {
						const selected = selectedLabelSet.has(label.id);

						return (
							<label className={styles.labelValueOption} data-selected={selected || undefined} title={labelToken(label)} key={label.id}>
								<Checkbox checked={selected} onChange={event => onToggle(label.id, event.currentTarget.checked)} />
								<span>{label.value}</span>
							</label>
						);
					})}
				</div>
			</section>
		</div>
	);
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
	const { t } = useTranslation(["probes", "common"]);
	const confirm = useConfirm();
	const [probeName, setProbeName] = useState(activeProbe.name);
	const [coordinateInputMode, setCoordinateInputMode] = useState<CoordinateInputMode>("search");
	const [locationSearch, setLocationSearch] = useState(activeProbe.location === "-" ? "" : activeProbe.location);
	const [locationName, setLocationName] = useState(activeProbe.location === "-" ? "" : activeProbe.location);
	const [latitudeInput, setLatitudeInput] = useState(initialLatitude(activeProbe));
	const [longitudeInput, setLongitudeInput] = useState(initialLongitude(activeProbe));
	const [geocodeStatus, setGeocodeStatus] = useState<GeocodeStatus>("idle");
	const [geocodeError, setGeocodeError] = useState("");
	const [rotatedSecret, setRotatedSecret] = useState("");
	const [serviceCommandMode, setServiceCommandMode] = useState<ProbeServiceCommandMode | null>(null);
	const [savingProbe, setSavingProbe] = useState(false);
	const [selectedLabelIds, setSelectedLabelIds] = useState(() => activeApiProbe?.labels.map(label => label.id) ?? []);
	const geocodeAbortRef = useRef<AbortController | null>(null);
	const updateProbeMutation = useUpdateProjectProbeMutation(projectRef, { suppressGlobalErrorToast: true });
	const deleteProbeMutation = useDeleteProjectProbeMutation(projectRef);
	const rotateSecretMutation = useRotateProjectProbeSecretMutation(projectRef);
	const labelsQuery = useQuery({
		...projectQueries.labels(projectRef || ""),
		enabled: Boolean(projectRef)
	});
	const probeAssignments = assignedRows.filter(row => row.probe === activeProbe.name);
	const detailRows = expandAssignedRows(probeAssignments);
	const initialLocationName = activeProbe.location === "-" ? "" : activeProbe.location;
	const initialLatitudeInput = initialLatitude(activeProbe);
	const initialLongitudeInput = initialLongitude(activeProbe);
	const initialLabelIds = useMemo(() => activeApiProbe?.labels.map(label => label.id) ?? [], [activeApiProbe?.labels]);
	const rotatedSecretCommand = rotatedSecret ? probeSecretUpdateCommand({ probeId: activeProbe.id, probeSecret: rotatedSecret }) : "";
	const serviceCommandOptions: Array<{ value: ProbeServiceCommandMode; label: string }> = [
		{ value: "reinstall", label: t("detail.reinstall") },
		{ value: "upgrade", label: t("detail.upgrade") }
	];
	const coordinateModeOptions: Array<{ value: CoordinateInputMode; label: string }> = [
		{ value: "search", label: t("location.searchName") },
		{ value: "manual", label: t("location.manualCoordinates") }
	];
	const assignedColumns: DataColumn<AssignedRow>[] = [
		{ key: "check", label: t("detail.assignedCheck") },
		{ key: "type", label: t("detail.type"), render: row => <Badge tone="neutral">{row.type}</Badge> },
		{ key: "interval", label: t("detail.interval") },
		{ key: "latest", label: t("detail.latest") }
	];
	const selectedServiceCommand = serviceCommandMode ? serviceCommandOptions.find(option => option.value === serviceCommandMode) : null;
	const serviceCommand = serviceCommandMode ? probeServiceCommand(serviceCommandMode) : "";
	const latitude = parseCoordinateInput(latitudeInput);
	const longitude = parseCoordinateInput(longitudeInput);
	const latitudeError = coordinateInputMode === "manual" ? coordinateInputError(latitudeInput, -90, 90) : null;
	const longitudeError = coordinateInputMode === "manual" ? coordinateInputError(longitudeInput, -180, 180) : null;
	const visibleLatitudeError = latitudeInput.trim() && latitudeError ? t(`location.${latitudeError}`, { label: t("location.latitude"), min: -90, max: 90 }) : "";
	const visibleLongitudeError = longitudeInput.trim() && longitudeError ? t(`location.${longitudeError}`, { label: t("location.longitude"), min: -180, max: 180 }) : "";
	const searchCoordinatesReady = coordinateInputMode === "search" && latitude !== null && longitude !== null;
	const manualCoordinatesReady = coordinateInputMode === "manual" && latitude !== null && longitude !== null && !latitudeError && !longitudeError;
	const coordinatesReady = searchCoordinatesReady || manualCoordinatesReady;
	const selectedCoordinates: ProbeCoordinates | null = coordinatesReady && latitude !== null && longitude !== null ? { latitude, longitude } : null;
	const canSearchLocation = coordinateInputMode === "search" && locationSearch.trim().length > 0 && geocodeStatus !== "searching";
	const locationInputInvalid = coordinateInputMode === "manual" && Boolean(latitudeInput.trim() || longitudeInput.trim()) && !manualCoordinatesReady;
	const previewLocationName = locationName.trim() || t("location.manual");
	const hasProbeChanges = Boolean(
		activeApiProbe &&
		(probeName !== activeProbe.name ||
			locationName !== initialLocationName ||
			latitudeInput !== initialLatitudeInput ||
			longitudeInput !== initialLongitudeInput ||
			!sameStringSet(selectedLabelIds, initialLabelIds))
	);
	const labelOptions = useMemo(() => {
		const labelsById = new Map<string, ApiLabel>();

		for (const label of activeApiProbe?.labels ?? []) {
			labelsById.set(label.id, label);
		}
		for (const label of labelsQuery.data?.labels ?? []) {
			labelsById.set(label.id, label);
		}

		return Array.from(labelsById.values()).sort(sortLabels);
	}, [activeApiProbe?.labels, labelsQuery.data?.labels]);
	const selectedLabelSet = useMemo(() => new Set(selectedLabelIds), [selectedLabelIds]);

	useEffect(() => {
		return () => {
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

	function updateServiceCommandMode(nextMode: ProbeServiceCommandMode) {
		setServiceCommandMode(nextMode);
	}

	function toggleLabel(labelId: string, checked: boolean) {
		setSelectedLabelIds(current => {
			if (checked) {
				return current.includes(labelId) ? current : [...current, labelId];
			}

			return current.filter(currentLabelId => currentLabelId !== labelId);
		});
	}

	function resetProbeDraft() {
		geocodeAbortRef.current?.abort();
		geocodeAbortRef.current = null;
		setProbeName(activeProbe.name);
		setCoordinateInputMode("search");
		setLocationSearch(initialLocationName);
		setLocationName(initialLocationName);
		setLatitudeInput(initialLatitudeInput);
		setLongitudeInput(initialLongitudeInput);
		setGeocodeStatus("idle");
		setGeocodeError("");
		setSelectedLabelIds(initialLabelIds);
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
			setGeocodeError(t(`location.${error instanceof LocationSearchError ? error.code : "searchFailed"}`));
		} finally {
			if (geocodeAbortRef.current === abortController) {
				geocodeAbortRef.current = null;
			}
		}
	}

	async function deleteProbe() {
		const confirmed = await confirm({
			title: t("detail.deleteQuestion"),
			message: t("detail.deleteMessage"),
			confirmLabel: t("detail.delete"),
			confirmationText: activeProbe.name,
			confirmationLabel: t("wizard.name"),
			tone: "danger"
		});

		if (!confirmed) {
			return;
		}

		deleteProbeMutation.mutate(activeProbe.id, { onSuccess: () => onDeleted?.() });
	}

	async function rotateSecret() {
		const confirmed = await confirm({
			title: t("detail.rotateQuestion", { name: activeProbe.name }),
			message: t("detail.rotateMessage"),
			confirmLabel: t("detail.rotate"),
			tone: "danger"
		});

		if (!confirmed) {
			return;
		}

		rotateSecretMutation.mutate(activeProbe.id, {
			onSuccess: data => {
				setRotatedSecret(data.secret);
			}
		});
	}

	async function saveProbe() {
		if (!projectRef || !activeApiProbe || !probeName.trim()) {
			return;
		}

		setSavingProbe(true);
		try {
			const body = {
				name: probeName.trim(),
				...(locationName.trim() ? { locationName: locationName.trim() } : {}),
				...(selectedCoordinates ? { latitude: selectedCoordinates.latitude, longitude: selectedCoordinates.longitude } : {}),
				labelIds: selectedLabelIds
			};

			await updateProbeMutation.mutateAsync({
				probeId: activeProbe.id,
				body
			});
		} catch (error) {
			pushErrorToast(requestErrorMessage(error, t("detail.saveError")));
		} finally {
			setSavingProbe(false);
		}
	}

	return (
		<section className={classNames(styles.card, floating && styles.floating)} aria-label={t("detail.aria")}>
			<div className={styles.header}>
				<strong className="ns-title">
					{activeProbe.name}
					<small> · {t("detail.uptime", { uptime: activeProbe.uptime })}</small>
				</strong>
				{onClose ? <CloseButton ariaLabel={t("detail.close")} onClick={onClose} /> : null}
			</div>

			<div className={styles.fieldGrid}>
				<TextField className={styles.input} label={t("wizard.name")} value={probeName} onChange={event => setProbeName(event.currentTarget.value)} />
			</div>

			<div className={styles.labelEditor}>
				<FieldLabel>{t("detail.labels")}</FieldLabel>
				{labelOptions.length ? (
					<ProbeLabelPicker labels={labelOptions} selectedLabelSet={selectedLabelSet} onToggle={toggleLabel} />
				) : labelsQuery.isLoading ? (
					<Spinner label={t("detail.loadingLabels")} layout="compact" size="md" />
				) : (
					<p className={styles.labelNotice}>{t("detail.noLabels")}</p>
				)}
			</div>

			<div className={styles.locationEditor}>
				<SegmentedControl
					className={styles.locationMode}
					size="sm"
					ariaLabel={t("location.coordinateMode")}
					value={coordinateInputMode}
					options={coordinateModeOptions}
					onValueChange={nextMode => updateCoordinateInputMode(nextMode as CoordinateInputMode)}
				/>

				<div className={styles.locationBody}>
					<div className={styles.locationControls}>
						{coordinateInputMode === "search" ? (
							<div className={styles.locationSearch}>
								<TextField
									className={styles.input}
									label={t("location.search")}
									value={locationSearch}
									placeholder={t("probes:location.namePlaceholder")}
									disabled={geocodeStatus === "searching"}
									error={geocodeStatus === "error" ? geocodeError : undefined}
									onChange={event => updateLocationSearch(event.currentTarget.value)}
								/>
								<Button type="button" variant="outline" size="sm" disabled={!canSearchLocation} onClick={() => void searchLocation()}>
									{geocodeStatus === "searching" ? t("location.searching") : t("location.searchAction")}
								</Button>
							</div>
						) : (
							<div className={styles.manualLocationFields}>
								<TextField
									className={styles.input}
									label={t("location.name")}
									value={locationName}
									placeholder={t("location.namePlaceholder")}
									onChange={event => setLocationName(event.currentTarget.value)}
								/>
								<div className={styles.coordinateGrid}>
									<TextField
										className={styles.input}
										label={t("location.latitude")}
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
										label={t("location.longitude")}
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
								{coordinateInputMode === "search" ? t("location.searchToUpdate") : t("location.coordinatesToPreview")}
							</p>
						)}
					</div>

					{selectedCoordinates ? (
						<LocationPreviewMap coordinates={selectedCoordinates} locationName={previewLocationName} probeName={probeName.trim() || activeProbe.name} className={styles.locationPreview} />
					) : null}
				</div>
			</div>

			<UnsavedChangesBar
				show={hasProbeChanges}
				saving={updateProbeMutation.isPending || savingProbe}
				disabled={!projectRef || !activeApiProbe || !probeName || locationInputInvalid || geocodeStatus === "searching"}
				onReset={resetProbeDraft}
				onSave={() => void saveProbe()}
			/>

			<div className={styles.actions}>
				<Button variant="outline" disabled={!projectRef || rotateSecretMutation.isPending} onClick={() => void rotateSecret()}>
					{rotateSecretMutation.isPending ? t("detail.rotating") : t("detail.rotate")}
				</Button>
				<SegmentedControl
					className={styles.serviceCommandMode}
					ariaLabel={t("detail.serviceCommand")}
					value={serviceCommandMode ?? ""}
					options={serviceCommandOptions}
					onValueChange={nextMode => updateServiceCommandMode(nextMode as ProbeServiceCommandMode)}
				/>
				<Button variant="danger" disabled={!projectRef || deleteProbeMutation.isPending} onClick={() => void deleteProbe()}>
					{deleteProbeMutation.isPending ? t("detail.deleting") : t("detail.delete")}
				</Button>
			</div>

			{serviceCommandMode && selectedServiceCommand ? (
				<div className={styles.serviceCommandPanel}>
					<span>{t("detail.serviceCommand")}</span>
					<CodeBlock title={t("detail.commandTitle", { action: selectedServiceCommand.label })} copyDisabled={!serviceCommand}>
						{serviceCommand}
					</CodeBlock>
				</div>
			) : null}

			{rotatedSecret ? (
				<div className={styles.secretPanel}>
					<span>{t("detail.newSecret")}</span>
					<code>{rotatedSecret}</code>
					<p>{t("detail.secretDescription")}</p>
					<CodeBlock title={t("detail.updateCommand")} copyDisabled={!rotatedSecretCommand}>
						{rotatedSecretCommand}
					</CodeBlock>
				</div>
			) : null}

			<DataTable
				ariaLabel={t("detail.assignedChecks")}
				columns={assignedColumns}
				rows={detailRows}
				density="compact"
				minWidth="31rem"
				maxHeight="11.75rem"
				getRowKey={(row, index) => `${row.probe}-${row.check}-${index}`}
			/>
		</section>
	);
}

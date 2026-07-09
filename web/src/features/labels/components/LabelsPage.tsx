import { pathForLabelDetail, pathForRoute } from "@/routes/routePaths";
import { useDeleteProjectLabelMutation, useSaveProjectLabelMutation } from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import type { ApiCheck, ApiLabel, ApiProbe } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { useConfirm } from "@/shared/components/confirmContext";
import { EditorDrawer } from "@/shared/components/EditorDrawer";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { UnsavedChangesBar } from "@/shared/components/UnsavedChangesBar";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { Badge, Button, FilterGrid, Panel, SelectField, TextField } from "@netstamp/ui";
import { PlusIcon } from "@phosphor-icons/react/dist/csr/Plus";
import { useQuery } from "@tanstack/react-query";
import { type KeyboardEvent, useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import styles from "./LabelsPage.module.css";

interface LabelUsage {
	probeNames: string[];
	checkNames: string[];
}

interface LabelRow {
	id: string;
	key: string;
	value: string;
	token: string;
	updatedAt: string;
	probeCount: number;
	checkCount: number;
	probeNames: string[];
	checkNames: string[];
}

interface LabelGroup {
	key: string;
	rows: LabelRow[];
	probeCount: number;
	checkCount: number;
}

type LabelEditorMode = "idle" | "create";

const emptyLabels: ApiLabel[] = [];
const emptyProbes: ApiProbe[] = [];
const emptyChecks: ApiCheck[] = [];

function formatUpdatedAt(value: string) {
	const date = new Date(value);

	if (Number.isNaN(date.getTime())) {
		return "-";
	}

	return date.toLocaleString();
}

function labelToken(label: Pick<ApiLabel, "key" | "value">) {
	return `${label.key}:${label.value}`;
}

function sortedUnique(values: string[]) {
	return Array.from(new Set(values.filter(Boolean))).sort((a, b) => a.localeCompare(b));
}

function buildUsage(labels: ApiLabel[], probes: ApiProbe[], checks: ApiCheck[]) {
	const usage = new Map<string, LabelUsage>();

	for (const label of labels) {
		usage.set(label.id, { probeNames: [], checkNames: [] });
	}

	for (const probe of probes) {
		for (const label of probe.labels ?? []) {
			usage.get(label.id)?.probeNames.push(probe.name);
		}
	}

	for (const check of checks) {
		for (const label of check.labels ?? []) {
			usage.get(label.id)?.checkNames.push(check.name);
		}
	}

	return usage;
}

function usageNames(names: string[]) {
	if (!names.length) {
		return "None";
	}

	return names.slice(0, 3).join(", ") + (names.length > 3 ? ` +${names.length - 3}` : "");
}

function countLabel(count: number, singular: string) {
	return `${count} ${singular}${count === 1 ? "" : "s"}`;
}

function renderUsage(count: number, names: string[], tone: "accent" | "success") {
	const summary = usageNames(names);

	return (
		<div className={styles.usageSummary} title={summary}>
			<Badge tone={count ? tone : "muted"} dot={false}>
				{count}
			</Badge>
			<span className={count ? styles.usageNames : styles.emptyUsage}>{summary}</span>
		</div>
	);
}

function buildLabelGroups(rows: LabelRow[]) {
	const groupDrafts = new Map<
		string,
		{
			key: string;
			rows: LabelRow[];
			probeNames: Set<string>;
			checkNames: Set<string>;
		}
	>();

	for (const row of rows) {
		const draft = groupDrafts.get(row.key) ?? {
			key: row.key,
			rows: [],
			probeNames: new Set<string>(),
			checkNames: new Set<string>()
		};

		draft.rows.push(row);
		for (const name of row.probeNames) {
			draft.probeNames.add(name);
		}
		for (const name of row.checkNames) {
			draft.checkNames.add(name);
		}

		groupDrafts.set(row.key, draft);
	}

	return Array.from(groupDrafts.values())
		.map<LabelGroup>(draft => ({
			key: draft.key,
			rows: draft.rows.sort((left, right) => left.value.localeCompare(right.value)),
			probeCount: draft.probeNames.size,
			checkCount: draft.checkNames.size
		}))
		.sort((left, right) => left.key.localeCompare(right.key));
}

export function LabelsPage() {
	const { projectRef } = useCurrentProject();
	const { labelId = "" } = useParams();
	const navigate = useNavigate();
	const confirm = useConfirm();
	const saveLabelMutation = useSaveProjectLabelMutation(projectRef);
	const deleteLabelMutation = useDeleteProjectLabelMutation(projectRef);
	const labelsQuery = useQuery({
		...projectQueries.labels(projectRef || ""),
		enabled: Boolean(projectRef)
	});
	const probesQuery = useQuery({
		...projectQueries.probes(projectRef || ""),
		enabled: Boolean(projectRef)
	});
	const checksQuery = useQuery({
		...projectQueries.checks(projectRef || ""),
		enabled: Boolean(projectRef)
	});
	const labels = labelsQuery.data?.labels ?? emptyLabels;
	const probes = probesQuery.data?.probes ?? emptyProbes;
	const checks = checksQuery.data?.checks ?? emptyChecks;
	const [editorMode, setEditorMode] = useState<LabelEditorMode>("idle");
	const [draftLabelId, setDraftLabelId] = useState("");
	const [draftKey, setDraftKey] = useState("");
	const [draftValue, setDraftValue] = useState("");
	const [search, setSearch] = useState("");
	const [keyFilter, setKeyFilter] = useState("all");
	const usageByLabelID = useMemo(() => buildUsage(labels, probes, checks), [labels, probes, checks]);
	const rows = useMemo<LabelRow[]>(
		() =>
			labels.map(label => {
				const usage = usageByLabelID.get(label.id) ?? { probeNames: [], checkNames: [] };
				const probeNames = sortedUnique(usage.probeNames);
				const checkNames = sortedUnique(usage.checkNames);

				return {
					id: label.id,
					key: label.key,
					value: label.value,
					token: labelToken(label),
					updatedAt: formatUpdatedAt(label.updatedAt),
					probeCount: probeNames.length,
					checkCount: checkNames.length,
					probeNames,
					checkNames
				};
			}),
		[labels, usageByLabelID]
	);
	const keyOptions = useMemo(() => [{ value: "all", label: "All keys" }, ...sortedUnique(labels.map(label => label.key)).map(key => ({ value: key, label: key }))], [labels]);
	const filteredRows = useMemo(() => {
		const query = search.trim().toLowerCase();

		return rows.filter(row => {
			const matchesKey = keyFilter === "all" || row.key === keyFilter;
			const matchesSearch = !query || [row.key, row.value, row.token, ...row.probeNames, ...row.checkNames].some(value => value.toLowerCase().includes(query));

			return matchesKey && matchesSearch;
		});
	}, [keyFilter, rows, search]);
	const labelGroups = useMemo(() => buildLabelGroups(filteredRows), [filteredRows]);
	const allLabelGroups = useMemo(() => buildLabelGroups(rows), [rows]);
	const selectedLabel = labels.find(label => label.id === labelId) ?? null;
	const selectedRow = rows.find(row => row.id === labelId) ?? null;
	const isCreating = editorMode === "create";
	const isEditing = !isCreating && Boolean(selectedLabel);
	const isEditorOpen = isCreating || isEditing;
	const hasSelectedDraft = Boolean(selectedLabel && draftLabelId === selectedLabel.id);
	const activeDraftKey = isCreating || hasSelectedDraft ? draftKey : (selectedLabel?.key ?? "");
	const activeDraftValue = isCreating || hasSelectedDraft ? draftValue : (selectedLabel?.value ?? "");
	const mutationError = saveLabelMutation.error ?? deleteLabelMutation.error;
	const hasLabelChanges = isCreating ? Boolean(activeDraftKey || activeDraftValue) : Boolean(selectedLabel && (selectedLabel.key !== activeDraftKey || selectedLabel.value !== activeDraftValue));
	const canSave = Boolean(
		projectRef && activeDraftKey.trim() && activeDraftValue.trim() && (!selectedLabel || selectedLabel.key !== activeDraftKey.trim() || selectedLabel.value !== activeDraftValue.trim())
	);
	const emptyLabel = projectRef ? (labelsQuery.isLoading ? "Loading labels" : "No labels match this view") : "Select a project to manage labels";
	const labelPanelTitle = search.trim() || keyFilter !== "all" ? `${labelGroups.length} keys / ${filteredRows.length} labels` : `${allLabelGroups.length} keys / ${rows.length} labels`;

	useEffect(() => {
		if (!projectRef || !labelId || labelsQuery.isPending || labelsQuery.isError || selectedLabel) {
			return;
		}

		navigate(pathForRoute("labels", { projectRef }), { replace: true });
	}, [labelId, labelsQuery.isError, labelsQuery.isPending, navigate, projectRef, selectedLabel]);

	function prepareSelectedLabelEdit() {
		if (isCreating || !selectedLabel || hasSelectedDraft) {
			return;
		}

		setDraftLabelId(selectedLabel.id);
		setDraftKey(selectedLabel.key);
		setDraftValue(selectedLabel.value);
	}

	function updateDraftKey(value: string) {
		prepareSelectedLabelEdit();
		setDraftKey(value);
	}

	function updateDraftValue(value: string) {
		prepareSelectedLabelEdit();
		setDraftValue(value);
	}

	function selectLabel(row: LabelRow) {
		setEditorMode("idle");
		setDraftLabelId("");
		saveLabelMutation.reset();
		deleteLabelMutation.reset();
		navigate(pathForLabelDetail(projectRef, row.id));
	}

	function startNewLabel(prefillKey = "") {
		setEditorMode("create");
		setDraftLabelId("__new__");
		setDraftKey(prefillKey);
		setDraftValue("");
		saveLabelMutation.reset();
		deleteLabelMutation.reset();
		navigate(pathForRoute("labels", { projectRef }));
	}

	function openLabelFromKeyboard(event: KeyboardEvent<HTMLTableRowElement>, row: LabelRow) {
		if (event.key !== "Enter" && event.key !== " ") {
			return;
		}

		event.preventDefault();
		selectLabel(row);
	}

	function closeEditor() {
		setEditorMode("idle");
		setDraftLabelId("");
		setDraftKey("");
		setDraftValue("");
		saveLabelMutation.reset();
		deleteLabelMutation.reset();
		navigate(pathForRoute("labels", { projectRef }));
	}

	function saveLabel() {
		if (!canSave) {
			return;
		}

		saveLabelMutation.mutate(
			{
				labelId: selectedLabel?.id,
				body: {
					key: activeDraftKey.trim(),
					value: activeDraftValue.trim()
				}
			},
			{
				onSuccess: data => {
					setEditorMode("idle");
					setDraftLabelId(data.label.id);
					setDraftKey(data.label.key);
					setDraftValue(data.label.value);
					navigate(pathForLabelDetail(projectRef, data.label.id));
				}
			}
		);
	}

	function resetLabelDraft() {
		saveLabelMutation.reset();

		if (isCreating || !selectedLabel) {
			setDraftLabelId("__new__");
			setDraftKey("");
			setDraftValue("");
			return;
		}

		setDraftLabelId(selectedLabel.id);
		setDraftKey(selectedLabel.key);
		setDraftValue(selectedLabel.value);
	}

	async function deleteLabel(row: LabelRow) {
		const confirmed = await confirm({
			title: `Delete ${row.token}?`,
			message: "This removes the label from future probe and check matching, then refreshes project assignments.",
			confirmLabel: "Delete label",
			tone: "danger"
		});

		if (!confirmed) {
			return;
		}

		deleteLabelMutation.mutate(row.id, {
			onSuccess: () => {
				if (labelId === row.id) {
					closeEditor();
				}
			}
		});
	}

	return (
		<PageStack>
			<ScreenHeader title="Labels" actions={<Button onClick={() => startNewLabel()}>New label</Button>} />

			<div className={styles.labelsGrid}>
				<Panel tone="glass" title={labelPanelTitle}>
					<div className={styles.listStack}>
						<FilterGrid className={styles.filters}>
							<TextField label="Search" placeholder="region:tokyo, provider, edge" value={search} disabled={!projectRef} onChange={event => setSearch(event.currentTarget.value)} />
							<SelectField label="Key" value={keyFilter} disabled={!projectRef} options={keyOptions} onChange={event => setKeyFilter(event.currentTarget.value)} />
						</FilterGrid>
						<div className={["ns-frame", styles.groupedTableFrame].join(" ")}>
							<div className={["ns-scrollbar", styles.groupedTableScroller].join(" ")}>
								<table className={styles.groupedTable} aria-label="Project labels grouped by key">
									<thead>
										<tr>
											<th>Value</th>
											<th>Probes</th>
											<th>Checks</th>
											<th>Updated</th>
										</tr>
									</thead>
									{labelGroups.length ? (
										labelGroups.map(group => (
											<tbody key={group.key}>
												<tr className={styles.groupRow}>
													<th className={styles.groupHeaderCell} scope="rowgroup" colSpan={4}>
														<div className={styles.groupHeader}>
															<div className={styles.groupHeading}>
																<strong translate="no">{group.key}</strong>
																<span>
																	{countLabel(group.rows.length, "value")} · {countLabel(group.probeCount, "probe")} · {countLabel(group.checkCount, "check")}
																</span>
															</div>
															<Button type="button" variant="secondary" size="sm" onClick={() => startNewLabel(group.key)}>
																<PlusIcon className={styles.addValueIcon} size={14} weight="bold" aria-hidden="true" focusable="false" />
																Add value
															</Button>
														</div>
													</th>
												</tr>
												{group.rows.map(row => {
													const selected = !isCreating && labelId === row.id;

													return (
														<tr
															key={row.id}
															className={[styles.labelValueRow, selected && styles.selectedLabelValueRow].filter(Boolean).join(" ")}
															aria-label={`Open label ${row.token}`}
															aria-selected={selected || undefined}
															tabIndex={0}
															onClick={() => selectLabel(row)}
															onKeyDown={event => openLabelFromKeyboard(event, row)}
														>
															<td>
																<div className={styles.valueCell}>
																	<strong translate="no">{row.value}</strong>
																	<span translate="no">{row.token}</span>
																</div>
															</td>
															<td>{renderUsage(row.probeCount, row.probeNames, "success")}</td>
															<td>{renderUsage(row.checkCount, row.checkNames, "accent")}</td>
															<td className={styles.updatedCell}>{row.updatedAt}</td>
														</tr>
													);
												})}
											</tbody>
										))
									) : (
										<tbody>
											<tr>
												<td className={styles.emptyState} colSpan={4}>
													{emptyLabel}
												</td>
											</tr>
										</tbody>
									)}
								</table>
							</div>
						</div>
					</div>
				</Panel>

				{isEditorOpen ? (
					<EditorDrawer open title={isEditing ? selectedRow?.token || "Label" : "New label"} ariaLabel="Label editor" onClose={closeEditor}>
						<div className={styles.editorStack}>
							<div className={styles.editorForm}>
								<TextField label="Key" placeholder="region" value={activeDraftKey} disabled={!projectRef} onChange={event => updateDraftKey(event.currentTarget.value)} />
								<TextField label="Value" placeholder="tokyo" value={activeDraftValue} disabled={!projectRef} onChange={event => updateDraftValue(event.currentTarget.value)} />
							</div>

							{mutationError ? <p className={styles.errorNotice}>{requestErrorMessage(mutationError, "Label operation failed.")}</p> : null}

							<UnsavedChangesBar show={hasLabelChanges} saving={saveLabelMutation.isPending} disabled={!canSave} onReset={resetLabelDraft} onSave={saveLabel} />

							<div className={styles.editorActions}>
								<Button variant="danger" disabled={!selectedRow || deleteLabelMutation.isPending} onClick={() => selectedRow && void deleteLabel(selectedRow)}>
									{deleteLabelMutation.isPending ? "Deleting" : "Delete selected"}
								</Button>
							</div>

							<div className={styles.usagePanel}>
								<div>
									<span>Probe usage</span>
									<strong>{selectedRow?.probeCount ?? 0}</strong>
									<p>{usageNames(selectedRow?.probeNames ?? [])}</p>
								</div>
								<div>
									<span>Check usage</span>
									<strong>{selectedRow?.checkCount ?? 0}</strong>
									<p>{usageNames(selectedRow?.checkNames ?? [])}</p>
								</div>
							</div>
						</div>
					</EditorDrawer>
				) : null}
			</div>
		</PageStack>
	);
}

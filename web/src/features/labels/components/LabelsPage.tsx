import { pathForLabelDetail, pathForRoute } from "@/routes/routePaths";
import { useDeleteProjectLabelMutation, useSaveProjectLabelMutation } from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import type { ApiCheck, ApiLabel, ApiProbe } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { useConfirm } from "@/shared/components/confirmContext";
import { EditorDrawer } from "@/shared/components/EditorDrawer";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { ActionRow, Badge, Button, DataTable, FilterGrid, Panel, SelectField, TextField, type DataColumn } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useMemo, useState } from "react";
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
			const matchesSearch = !query || row.key.toLowerCase().includes(query) || row.value.toLowerCase().includes(query) || row.token.toLowerCase().includes(query);

			return matchesKey && matchesSearch;
		});
	}, [keyFilter, rows, search]);
	const selectedLabel = labels.find(label => label.id === labelId) ?? null;
	const selectedRow = rows.find(row => row.id === labelId) ?? null;
	const isCreating = editorMode === "create";
	const isEditing = !isCreating && Boolean(selectedLabel);
	const isEditorOpen = isCreating || isEditing;
	const hasSelectedDraft = Boolean(selectedLabel && draftLabelId === selectedLabel.id);
	const activeDraftKey = isCreating || hasSelectedDraft ? draftKey : (selectedLabel?.key ?? "");
	const activeDraftValue = isCreating || hasSelectedDraft ? draftValue : (selectedLabel?.value ?? "");
	const mutationError = saveLabelMutation.error ?? deleteLabelMutation.error;
	const canSave = Boolean(
		projectRef && activeDraftKey.trim() && activeDraftValue.trim() && (!selectedLabel || selectedLabel.key !== activeDraftKey.trim() || selectedLabel.value !== activeDraftValue.trim())
	);
	const emptyLabel = projectRef ? (labelsQuery.isLoading ? "Loading labels" : "No labels match this view") : "Select a project to manage labels";
	const columns: DataColumn<LabelRow>[] = [
		{
			key: "label",
			label: "Label",
			render: row => (
				<div className={styles.labelPair}>
					<Badge tone="accent" dot={false}>
						{row.key}
					</Badge>
					<strong>{row.value}</strong>
				</div>
			)
		},
		{
			key: "probeCount",
			label: "Probes",
			render: row => (
				<span className={styles.usageCell} title={usageNames(row.probeNames)}>
					<Badge tone={row.probeCount ? "success" : "muted"}>{row.probeCount}</Badge>
				</span>
			)
		},
		{
			key: "checkCount",
			label: "Checks",
			render: row => (
				<span className={styles.usageCell} title={usageNames(row.checkNames)}>
					<Badge tone={row.checkCount ? "accent" : "muted"}>{row.checkCount}</Badge>
				</span>
			)
		},
		{ key: "updatedAt", label: "Updated" },
		{
			key: "delete",
			label: "Delete",
			render: row => (
				<Button
					variant="danger"
					size="sm"
					disabled={deleteLabelMutation.isPending}
					onClick={event => {
						event.stopPropagation();
						void deleteLabel(row);
					}}
				>
					Delete
				</Button>
			)
		}
	];

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

	function startNewLabel() {
		setEditorMode("create");
		setDraftLabelId("__new__");
		setDraftKey("");
		setDraftValue("");
		saveLabelMutation.reset();
		deleteLabelMutation.reset();
		navigate(pathForRoute("labels", { projectRef }));
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
			<ScreenHeader title="Labels" actions={<Button onClick={startNewLabel}>New label</Button>} />

			<div className={styles.labelsGrid}>
				<Panel tone="glass" title={`${rows.length} labels`}>
					<div className={styles.listStack}>
						<FilterGrid className={styles.filters}>
							<TextField label="Search" placeholder="region:tokyo, provider, edge" value={search} disabled={!projectRef} onChange={event => setSearch(event.currentTarget.value)} />
							<SelectField label="Key" value={keyFilter} disabled={!projectRef} options={keyOptions} onChange={event => setKeyFilter(event.currentTarget.value)} />
						</FilterGrid>
						<DataTable
							ariaLabel="Project labels"
							columns={columns}
							rows={filteredRows}
							getRowKey={row => row.id}
							getRowAriaLabel={row => `Open label ${row.token}`}
							selectedKey={isCreating ? undefined : labelId}
							onRowClick={selectLabel}
							emptyLabel={emptyLabel}
						/>
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

							<ActionRow className={styles.editorActions}>
								<Button disabled={!canSave || saveLabelMutation.isPending} onClick={saveLabel}>
									{saveLabelMutation.isPending ? "Saving" : isEditing ? "Save label" : "Create label"}
								</Button>
								<Button variant="secondary" disabled={!projectRef} onClick={startNewLabel}>
									Reset
								</Button>
								<Button variant="danger" disabled={!selectedRow || deleteLabelMutation.isPending} onClick={() => selectedRow && void deleteLabel(selectedRow)}>
									{deleteLabelMutation.isPending ? "Deleting" : "Delete selected"}
								</Button>
							</ActionRow>

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

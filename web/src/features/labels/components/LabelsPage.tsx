import { useDeleteProjectLabelMutation, useSaveProjectLabelMutation } from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import type { ApiCheck, ApiLabel, ApiProbe } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { ActionRow } from "@/shared/components/ActionRow";
import { useConfirm } from "@/shared/components/confirmContext";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { classNames } from "@/shared/utils/classNames";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { Badge, Button, DataTable, Panel, SelectField, TextField, type DataColumn } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import { useMemo, useState } from "react";
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
	const [selectedLabelId, setSelectedLabelId] = useState("");
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
	const selectedLabel = labels.find(label => label.id === selectedLabelId) ?? null;
	const selectedRow = rows.find(row => row.id === selectedLabelId) ?? null;
	const isEditing = Boolean(selectedLabel);
	const mutationError = saveLabelMutation.error ?? deleteLabelMutation.error;
	const canSave = Boolean(projectRef && draftKey.trim() && draftValue.trim() && (!selectedLabel || selectedLabel.key !== draftKey.trim() || selectedLabel.value !== draftValue.trim()));
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

	function selectLabel(row: LabelRow) {
		setSelectedLabelId(row.id);
		setDraftKey(row.key);
		setDraftValue(row.value);
		saveLabelMutation.reset();
		deleteLabelMutation.reset();
	}

	function startNewLabel() {
		setSelectedLabelId("");
		setDraftKey("");
		setDraftValue("");
		saveLabelMutation.reset();
		deleteLabelMutation.reset();
	}

	function saveLabel() {
		if (!canSave) {
			return;
		}

		saveLabelMutation.mutate(
			{
				labelId: selectedLabel?.id,
				body: {
					key: draftKey.trim(),
					value: draftValue.trim()
				}
			},
			{
				onSuccess: data => {
					setSelectedLabelId(data.label.id);
					setDraftKey(data.label.key);
					setDraftValue(data.label.value);
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
				if (selectedLabelId === row.id) {
					startNewLabel();
				}
			}
		});
	}

	return (
		<PageStack>
			<ScreenHeader
				eyebrow="Label management"
				title="Labels"
				copy="Manage project label keys and values that drive probe grouping, check selectors, and assignment refreshes."
				actions={<Button onClick={startNewLabel}>New label</Button>}
			/>

			<div className={styles.labelsGrid}>
				<Panel tone="glass" eyebrow="Registry" title={`${rows.length} labels`}>
					<div className={styles.listStack}>
						<div className={styles.filters}>
							<TextField label="Search" placeholder="region:tokyo, provider, edge" value={search} disabled={!projectRef} onChange={event => setSearch(event.currentTarget.value)} />
							<SelectField label="Key" value={keyFilter} disabled={!projectRef} options={keyOptions} onChange={event => setKeyFilter(event.currentTarget.value)} />
						</div>
						<DataTable
							ariaLabel="Project labels"
							columns={columns}
							rows={filteredRows}
							getRowKey={row => row.id}
							getRowAriaLabel={row => `Edit label ${row.token}`}
							selectedKey={selectedLabelId}
							onRowClick={selectLabel}
							emptyLabel={emptyLabel}
						/>
					</div>
				</Panel>

				<Panel className={styles.editorPanel} tone="glass" eyebrow={isEditing ? "Edit label" : "Create label"} title={isEditing ? selectedRow?.token : "New label"}>
					<div className={styles.editorStack}>
						<div className={styles.editorForm}>
							<TextField label="Key" placeholder="region" value={draftKey} disabled={!projectRef} onChange={event => setDraftKey(event.currentTarget.value)} />
							<TextField label="Value" placeholder="tokyo" value={draftValue} disabled={!projectRef} onChange={event => setDraftValue(event.currentTarget.value)} />
						</div>

						{mutationError ? <p className={classNames("ns-cut-frame", styles.errorNotice)}>{requestErrorMessage(mutationError, "Label operation failed.")}</p> : null}

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
				</Panel>
			</div>
		</PageStack>
	);
}

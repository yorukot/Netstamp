import { useLocaleFormat } from "@/i18n/format";
import { pathForLabelDetail, pathForRoute } from "@/routes/routePaths";
import { useDeleteProjectLabelMutation, useSaveProjectLabelMutation } from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import type { ApiCheck, ApiLabel, ApiProbe } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { useConfirm } from "@/shared/components/confirmContext";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { Badge, Button, DialogContent, DialogOverlay, DialogPortal, DialogRoot, DialogTitle, FilterGrid, SelectField, Spinner, TextField } from "@netstamp/ui";
import { PlusIcon } from "@phosphor-icons/react/dist/csr/Plus";
import { useQuery } from "@tanstack/react-query";
import { type AnimationEvent, type FormEvent, type KeyboardEvent, useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
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

type LabelEditorMode = "idle" | "create" | "addValue";

const emptyLabels: ApiLabel[] = [];
const emptyProbes: ApiProbe[] = [];
const emptyChecks: ApiCheck[] = [];

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

function usageNames(names: string[], emptyLabel: string) {
	if (!names.length) {
		return emptyLabel;
	}

	return names.slice(0, 3).join(", ") + (names.length > 3 ? ` +${names.length - 3}` : "");
}

function renderUsage(count: number, names: string[], tone: "accent" | "success", emptyLabel: string) {
	const summary = usageNames(names, emptyLabel);

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
	const { t } = useTranslation(["labels", "common"]);
	const format = useLocaleFormat();
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
	const [isEditorDismissed, setIsEditorDismissed] = useState(false);
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
					updatedAt: Number.isNaN(new Date(label.updatedAt).getTime()) ? "-" : format.dateTime(label.updatedAt),
					probeCount: probeNames.length,
					checkCount: checkNames.length,
					probeNames,
					checkNames
				};
			}),
		[format, labels, usageByLabelID]
	);
	const keyOptions = useMemo(() => [{ value: "all", label: t("allKeys") }, ...sortedUnique(labels.map(label => label.key)).map(key => ({ value: key, label: key }))], [labels, t]);
	const filteredRows = useMemo(() => {
		const query = search.trim().toLowerCase();

		return rows.filter(row => {
			const matchesKey = keyFilter === "all" || row.key === keyFilter;
			const matchesSearch = !query || [row.key, row.value, row.token, ...row.probeNames, ...row.checkNames].some(value => value.toLowerCase().includes(query));

			return matchesKey && matchesSearch;
		});
	}, [keyFilter, rows, search]);
	const labelGroups = useMemo(() => buildLabelGroups(filteredRows), [filteredRows]);
	const selectedLabel = labels.find(label => label.id === labelId) ?? null;
	const selectedRow = rows.find(row => row.id === labelId) ?? null;
	const isNewLabel = editorMode === "create";
	const isAddingValue = editorMode === "addValue";
	const isCreating = isNewLabel || isAddingValue;
	const isEditing = !isCreating && Boolean(selectedLabel);
	const isEditorOpen = (isCreating || isEditing) && !isEditorDismissed;
	const hasSelectedDraft = Boolean(selectedLabel && draftLabelId === selectedLabel.id);
	const activeDraftKey = isCreating || hasSelectedDraft ? draftKey : (selectedLabel?.key ?? "");
	const activeDraftValue = isCreating || hasSelectedDraft ? draftValue : (selectedLabel?.value ?? "");
	const mutationError = saveLabelMutation.error ?? deleteLabelMutation.error;
	const canSave = Boolean(projectRef && activeDraftKey.trim() && activeDraftValue.trim() && (isCreating || (selectedLabel && selectedLabel.value !== activeDraftValue.trim())));
	const emptyLabel = projectRef ? labelsQuery.isLoading ? <Spinner label={t("loading")} layout="compact" size="lg" /> : t("noMatch") : t("selectProject");
	const editorTitle = isNewLabel ? t("editor.new") : isAddingValue ? t("editor.add", { key: activeDraftKey }) : selectedLabel ? t("editor.edit", { key: selectedLabel.key }) : t("editor.title");
	const editorSubmitLabel = saveLabelMutation.isPending ? t("editor.saving") : isEditing ? t("editor.saveValue") : isNewLabel ? t("editor.create") : t("editor.addValue");

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

	function updateDraftValue(value: string) {
		prepareSelectedLabelEdit();
		setDraftValue(value);
	}

	function selectLabel(row: LabelRow) {
		setEditorMode("idle");
		setIsEditorDismissed(false);
		setDraftLabelId("");
		saveLabelMutation.reset();
		deleteLabelMutation.reset();
		navigate(pathForLabelDetail(projectRef, row.id));
	}

	function startNewLabel(prefillKey = "") {
		setEditorMode(prefillKey ? "addValue" : "create");
		setIsEditorDismissed(false);
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
		setIsEditorDismissed(true);
	}

	function finishClosingEditor(event: AnimationEvent<HTMLFormElement>) {
		if (event.target !== event.currentTarget || event.currentTarget.dataset.state !== "closed") {
			return;
		}

		setEditorMode("idle");
		setDraftLabelId("");
		setDraftKey("");
		setDraftValue("");
		saveLabelMutation.reset();
		deleteLabelMutation.reset();
		navigate(pathForRoute("labels", { projectRef }), { replace: true });
	}

	function submitLabel(event: FormEvent) {
		event.preventDefault();
		saveLabel();
	}

	function saveLabel() {
		if (!canSave) {
			return;
		}

		const body = {
			key: activeDraftKey.trim(),
			value: activeDraftValue.trim()
		};

		saveLabelMutation.mutate(
			{
				labelId: isCreating ? undefined : selectedLabel?.id,
				body
			},
			{
				onSuccess: () => {
					closeEditor();
				}
			}
		);
	}

	async function deleteLabel(row: LabelRow) {
		const confirmed = await confirm({
			title: t("editor.deleteQuestion", { value: row.value }),
			message: t("editor.deleteMessage", { key: row.key }),
			confirmLabel: t("editor.delete"),
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
			<ScreenHeader title={t("title")} actions={<Button onClick={() => startNewLabel()}>{t("new")}</Button>} />

			<div className={styles.listStack}>
				<FilterGrid className={styles.filters}>
					<TextField label={t("search")} placeholder={t("searchPlaceholder")} value={search} disabled={!projectRef} onChange={event => setSearch(event.currentTarget.value)} />
					<SelectField label={t("key")} value={keyFilter} disabled={!projectRef} options={keyOptions} onChange={event => setKeyFilter(event.currentTarget.value)} />
				</FilterGrid>
				<div className={["ns-frame", styles.groupedTableFrame].join(" ")}>
					<div className={["ns-scrollbar", styles.groupedTableScroller].join(" ")}>
						<table className={styles.groupedTable} aria-label={t("groupedAria")}>
							<thead>
								<tr>
									<th>{t("value")}</th>
									<th>{t("probes")}</th>
									<th>{t("checks")}</th>
									<th>{t("updated")}</th>
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
															{t("valueCount", { count: group.rows.length })} · {t("probeCount", { count: group.probeCount })} · {t("checkCount", { count: group.checkCount })}
														</span>
													</div>
													<Button type="button" variant="secondary" size="sm" onClick={() => startNewLabel(group.key)}>
														<PlusIcon className={styles.addValueIcon} size={14} weight="bold" aria-hidden="true" focusable="false" />
														{t("addValue")}
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
													aria-label={t("openValue", { key: row.key, value: row.value })}
													aria-selected={selected || undefined}
													tabIndex={0}
													onClick={() => selectLabel(row)}
													onKeyDown={event => openLabelFromKeyboard(event, row)}
												>
													<td>
														<div className={styles.valueCell}>
															<strong translate="no">{row.value}</strong>
														</div>
													</td>
													<td>{renderUsage(row.probeCount, row.probeNames, "success", t("none"))}</td>
													<td>{renderUsage(row.checkCount, row.checkNames, "accent", t("none"))}</td>
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

			<DialogRoot
				open={isEditorOpen}
				onOpenChange={open => {
					if (!open) {
						closeEditor();
					}
				}}
			>
				<DialogPortal>
					<DialogOverlay onMouseDown={closeEditor}>
						<DialogContent asChild>
							<form className={styles.popup} onSubmit={submitLabel} onMouseDown={event => event.stopPropagation()} onAnimationEnd={finishClosingEditor}>
								<div className={styles.popupHeader}>
									<DialogTitle asChild>
										<strong>{editorTitle}</strong>
									</DialogTitle>
								</div>

								<TextField
									label={t("key")}
									placeholder="region"
									value={activeDraftKey}
									disabled={!projectRef || saveLabelMutation.isPending}
									readOnly={!isNewLabel}
									required
									onChange={event => setDraftKey(event.currentTarget.value)}
									autoFocus={isNewLabel}
								/>

								<TextField
									label={t("value")}
									placeholder="tokyo"
									value={activeDraftValue}
									disabled={!projectRef || saveLabelMutation.isPending}
									required
									onChange={event => updateDraftValue(event.currentTarget.value)}
									autoFocus={!isNewLabel}
								/>

								{mutationError ? <p className={styles.errorNotice}>{requestErrorMessage(mutationError, t("editor.operationFailed"))}</p> : null}

								<div className={styles.popupActions}>
									{selectedRow ? (
										<Button type="button" variant="danger" disabled={deleteLabelMutation.isPending || saveLabelMutation.isPending} onClick={() => void deleteLabel(selectedRow)}>
											{deleteLabelMutation.isPending ? t("editor.deleting") : t("editor.delete")}
										</Button>
									) : null}
									<div className={styles.primaryActions}>
										<Button type="button" variant="ghost" disabled={saveLabelMutation.isPending || deleteLabelMutation.isPending} onClick={closeEditor}>
											{t("common:actions.cancel")}
										</Button>
										<Button type="submit" disabled={!canSave || saveLabelMutation.isPending || deleteLabelMutation.isPending}>
											{editorSubmitLabel}
										</Button>
									</div>
								</div>
							</form>
						</DialogContent>
					</DialogOverlay>
				</DialogPortal>
			</DialogRoot>
		</PageStack>
	);
}

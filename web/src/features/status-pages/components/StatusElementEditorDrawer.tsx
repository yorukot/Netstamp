import { checkTypeLabel, rootFolderOptions } from "@/features/status-pages/api/statusPageAdapters";
import type { ApiProjectAssignment, ApiPublicStatusElement, CreatePublicStatusElementInput, PublicStatusChartRange, PublicStatusElementChartMode, PublicStatusElementKind } from "@/shared/api/types";
import { EditorDrawer } from "@/shared/components/EditorDrawer";
import { UnsavedChangesBar } from "@/shared/components/UnsavedChangesBar";
import { Badge, Checkbox, SelectField, TextAreaField, TextField } from "@netstamp/ui";
import { useMemo, useState, type FormEvent } from "react";
import styles from "./StatusElementEditorDrawer.module.css";

type StatusElementSubmitInput = CreatePublicStatusElementInput;
type AssignmentSelectionMode = NonNullable<CreatePublicStatusElementInput["assignmentSelectionMode"]>;

interface StatusElementFormInput {
	parentElementId?: string;
	kind: PublicStatusElementKind;
	checkId?: string;
	assignmentSelectionMode?: AssignmentSelectionMode;
	assignmentIds: string[];
	title: string;
	description: string;
	sortOrder: number;
	chartMode: PublicStatusElementChartMode;
	chartRange?: PublicStatusChartRange;
}

interface PublicAssignmentOption {
	id: string;
	checkId: string;
	checkName: string;
	checkType: "ping" | "tcp";
	target: string;
	probeId: string;
	probeName: string;
	probeLocationName?: string;
}

interface StatusElementEditorDrawerProps {
	open: boolean;
	element?: ApiPublicStatusElement | null;
	elements: ApiPublicStatusElement[];
	assignments: ApiProjectAssignment[];
	saving: boolean;
	onClose: () => void;
	onSubmit: (body: StatusElementSubmitInput) => void;
}

const kindOptions: Array<{ value: PublicStatusElementKind; label: string }> = [
	{ value: "assignment_group", label: "Assignment group" },
	{ value: "folder", label: "Folder" }
];

const assignmentScopeOptions: Array<{ value: AssignmentSelectionMode; label: string }> = [
	{ value: "all_check", label: "All assignments for one check" },
	{ value: "selected_assignments", label: "Selected assignments" }
];

const chartModeOptions: Array<{ value: PublicStatusElementChartMode; label: string }> = [
	{ value: "inherit", label: "Inherit" },
	{ value: "off", label: "Off" },
	{ value: "compact", label: "Compact" }
];

const chartRangeOptions: Array<{ value: string; label: string }> = [
	{ value: "", label: "Inherit range" },
	{ value: "24h", label: "24 hours" },
	{ value: "7d", label: "7 days" },
	{ value: "30d", label: "30 days" }
];

function initialState(element?: ApiPublicStatusElement | null): StatusElementFormInput {
	return {
		parentElementId: element?.parentElementId,
		kind: element?.kind ?? "assignment_group",
		checkId: element?.checkId,
		assignmentSelectionMode: element?.assignmentSelectionMode ?? "all_check",
		assignmentIds: element?.assignmentIds ?? [],
		title: element?.title ?? "",
		description: element?.description ?? "",
		sortOrder: element?.sortOrder ?? 0,
		chartMode: element?.chartMode ?? "inherit",
		chartRange: element?.chartRange
	};
}

function sameValue(left: unknown, right: unknown) {
	return JSON.stringify(left) === JSON.stringify(right);
}

function publicAssignmentOption(assignment: ApiProjectAssignment): PublicAssignmentOption | null {
	if (!assignment.check || !assignment.probe) {
		return null;
	}
	if (assignment.check.type !== "ping" && assignment.check.type !== "tcp") {
		return null;
	}
	return {
		id: assignment.id,
		checkId: assignment.checkId,
		checkName: assignment.check.name,
		checkType: assignment.check.type,
		target: assignment.check.target,
		probeId: assignment.probeId,
		probeName: assignment.probe.name,
		probeLocationName: assignment.probe.locationName
	};
}

function compareAssignmentOptions(a: PublicAssignmentOption, b: PublicAssignmentOption) {
	return a.checkName.localeCompare(b.checkName) || a.probeName.localeCompare(b.probeName) || a.id.localeCompare(b.id);
}

function checkOptions(assignments: PublicAssignmentOption[]) {
	const byID = new Map<string, PublicAssignmentOption>();
	for (const assignment of assignments) {
		if (!byID.has(assignment.checkId)) {
			byID.set(assignment.checkId, assignment);
		}
	}
	return [...byID.values()]
		.sort((a, b) => a.checkName.localeCompare(b.checkName))
		.map(check => ({ value: check.checkId, label: `${check.checkName} / ${checkTypeLabel(check.checkType)} / ${check.target}` }));
}

function probeOptions(assignments: PublicAssignmentOption[]) {
	const byID = new Map<string, PublicAssignmentOption>();
	for (const assignment of assignments) {
		if (!byID.has(assignment.probeId)) {
			byID.set(assignment.probeId, assignment);
		}
	}
	return [...byID.values()]
		.sort((a, b) => a.probeName.localeCompare(b.probeName))
		.map(probe => ({ value: probe.probeId, label: probe.probeLocationName ? `${probe.probeName} / ${probe.probeLocationName}` : probe.probeName }));
}

export function StatusElementEditorDrawer({ open, element, elements, assignments, saving, onClose, onSubmit }: StatusElementEditorDrawerProps) {
	const initialForm = useMemo(() => initialState(element), [element]);
	const [form, setForm] = useState<StatusElementFormInput>(() => initialForm);
	const [error, setError] = useState("");
	const [checkFilter, setCheckFilter] = useState(element?.checkId ?? "");
	const [probeFilter, setProbeFilter] = useState("");
	const [search, setSearch] = useState("");
	const isEditing = Boolean(element);
	const availableParents = useMemo(() => rootFolderOptions(elements.filter(candidate => candidate.id !== element?.id)), [element?.id, elements]);
	const assignmentOptions = useMemo(
		() =>
			assignments
				.map(publicAssignmentOption)
				.filter((assignment): assignment is PublicAssignmentOption => Boolean(assignment))
				.sort(compareAssignmentOptions),
		[assignments]
	);
	const allCheckOptions = useMemo(() => checkOptions(assignmentOptions), [assignmentOptions]);
	const allProbeOptions = useMemo(() => probeOptions(assignmentOptions), [assignmentOptions]);
	const filteredAssignments = useMemo(() => {
		const needle = search.trim().toLowerCase();
		return assignmentOptions.filter(assignment => {
			if (checkFilter && assignment.checkId !== checkFilter) {
				return false;
			}
			if (probeFilter && assignment.probeId !== probeFilter) {
				return false;
			}
			if (!needle) {
				return true;
			}
			return [assignment.checkName, assignment.checkType, assignment.target, assignment.probeName, assignment.probeLocationName].some(value => value?.toLowerCase().includes(needle));
		});
	}, [assignmentOptions, checkFilter, probeFilter, search]);
	const selectedAssignmentIDs = useMemo(() => new Set(form.assignmentIds), [form.assignmentIds]);
	const parentOptions = [{ value: "", label: "Root" }, ...availableParents];
	const isFolder = form.kind === "folder";
	const selectionMode = form.assignmentSelectionMode ?? "all_check";
	const hasChanges = !sameValue(form, initialForm);

	function update<K extends keyof StatusElementFormInput>(key: K, value: StatusElementFormInput[K]) {
		setForm(current => ({ ...current, [key]: value }));
	}

	function handleKindChange(kind: PublicStatusElementKind) {
		setForm(current => ({
			...current,
			kind,
			parentElementId: kind === "folder" ? undefined : current.parentElementId,
			checkId: kind === "folder" ? undefined : current.checkId,
			assignmentSelectionMode: kind === "folder" ? undefined : (current.assignmentSelectionMode ?? "all_check"),
			assignmentIds: kind === "folder" ? [] : current.assignmentIds
		}));
	}

	function handleSelectionModeChange(mode: AssignmentSelectionMode) {
		setForm(current => ({
			...current,
			assignmentSelectionMode: mode,
			checkId: mode === "all_check" ? current.checkId || checkFilter || allCheckOptions[0]?.value : undefined,
			assignmentIds: mode === "selected_assignments" ? current.assignmentIds : []
		}));
	}

	function toggleAssignment(assignmentID: string, checked: boolean) {
		setForm(current => {
			const next = new Set(current.assignmentIds);
			if (checked) {
				next.add(assignmentID);
			} else {
				next.delete(assignmentID);
			}
			return { ...current, assignmentIds: [...next] };
		});
	}

	function resetForm() {
		setForm(initialForm);
		setError("");
	}

	function handleSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		const title = form.title.trim();
		const description = form.description.trim();
		const sortOrder = Number(form.sortOrder);

		if (!Number.isFinite(sortOrder) || sortOrder < 0) {
			setError("Order must be a non-negative number.");
			return;
		}
		if (form.kind === "assignment_group" && selectionMode === "all_check" && !form.checkId) {
			setError("Select a check for this assignment group.");
			return;
		}
		if (form.kind === "assignment_group" && selectionMode === "selected_assignments" && form.assignmentIds.length === 0) {
			setError("Select at least one assignment for this group.");
			return;
		}

		onSubmit({
			kind: form.kind,
			parentElementId: form.kind === "assignment_group" ? form.parentElementId || undefined : undefined,
			checkId: form.kind === "assignment_group" && selectionMode === "all_check" ? form.checkId : undefined,
			assignmentSelectionMode: form.kind === "assignment_group" ? selectionMode : undefined,
			assignmentIds: form.kind === "assignment_group" && selectionMode === "selected_assignments" ? form.assignmentIds : undefined,
			title: title || undefined,
			description: description || undefined,
			sortOrder,
			chartMode: form.chartMode || "inherit",
			chartRange: form.chartRange || undefined
		});
	}

	return (
		<EditorDrawer open={open} title={isEditing ? "Edit Element" : "New Element"} ariaLabel={isEditing ? "Edit status page element" : "Create status page element"} onClose={onClose}>
			<form id="status-element-editor-form" className={styles.form} onSubmit={handleSubmit}>
				{error ? <div className={styles.formError}>{error}</div> : null}
				<SelectField label="Element type" value={form.kind} disabled={isEditing} options={kindOptions} onChange={event => handleKindChange(event.currentTarget.value as PublicStatusElementKind)} />
				{isFolder ? null : (
					<>
						<SelectField label="Folder" value={form.parentElementId || ""} options={parentOptions} onChange={event => update("parentElementId", event.currentTarget.value || undefined)} />
						<SelectField label="Scope" value={selectionMode} options={assignmentScopeOptions} onChange={event => handleSelectionModeChange(event.currentTarget.value as AssignmentSelectionMode)} />
						{selectionMode === "all_check" ? (
							<SelectField
								label="Check"
								value={form.checkId || ""}
								options={[{ value: "", label: "Select check", disabled: true }, ...allCheckOptions]}
								onChange={event => update("checkId", event.currentTarget.value || undefined)}
							/>
						) : (
							<div className={styles.assignmentPicker}>
								<div className={styles.assignmentFilters}>
									<SelectField label="Check" value={checkFilter} options={[{ value: "", label: "All checks" }, ...allCheckOptions]} onChange={event => setCheckFilter(event.currentTarget.value)} />
									<SelectField label="Probe" value={probeFilter} options={[{ value: "", label: "All probes" }, ...allProbeOptions]} onChange={event => setProbeFilter(event.currentTarget.value)} />
									<TextField label="Search" value={search} onChange={event => setSearch(event.currentTarget.value)} />
								</div>
								<div className={styles.assignmentPickerHeader}>
									<span>{form.assignmentIds.length} selected</span>
									<Badge tone="neutral">{filteredAssignments.length} visible</Badge>
								</div>
								<div className={styles.assignmentList}>
									{filteredAssignments.length ? (
										filteredAssignments.map(assignment => (
											<label key={assignment.id} className={styles.assignmentRow}>
												<Checkbox checked={selectedAssignmentIDs.has(assignment.id)} onChange={event => toggleAssignment(assignment.id, event.currentTarget.checked)} />
												<span className={styles.assignmentCopy}>
													<strong>{assignment.checkName}</strong>
													<span>
														{checkTypeLabel(assignment.checkType)} / {assignment.target} / {assignment.probeName}
													</span>
												</span>
											</label>
										))
									) : (
										<div className={styles.assignmentEmpty}>No assignments match the current filters.</div>
									)}
								</div>
							</div>
						)}
					</>
				)}
				<TextField label="Title override" value={form.title} maxLength={1024} onChange={event => update("title", event.currentTarget.value)} />
				<TextAreaField label="Description override" value={form.description} maxLength={1024} rows={4} onChange={event => update("description", event.currentTarget.value)} />
				<TextField label="Order" type="number" min={0} value={String(form.sortOrder)} onChange={event => update("sortOrder", Number(event.currentTarget.value))} />
				<SelectField
					label="Chart mode"
					value={form.chartMode || "inherit"}
					options={chartModeOptions}
					onChange={event => update("chartMode", event.currentTarget.value as PublicStatusElementChartMode)}
				/>
				<SelectField
					label="Chart range"
					value={form.chartRange || ""}
					options={chartRangeOptions}
					onChange={event => update("chartRange", (event.currentTarget.value || undefined) as PublicStatusChartRange | undefined)}
				/>
				<UnsavedChangesBar show={hasChanges} saveType="submit" saving={saving} onReset={resetForm} />
			</form>
		</EditorDrawer>
	);
}

import { rootFolderOptions } from "@/features/status-pages/api/statusPageAdapters";
import type { ApiCheck, ApiPublicStatusElement, CreatePublicStatusElementInput, PublicStatusChartRange, PublicStatusElementChartMode, PublicStatusElementKind } from "@/shared/api/types";
import { EditorDrawer } from "@/shared/components/EditorDrawer";
import { Button, SelectField, TextAreaField, TextField } from "@netstamp/ui";
import { useMemo, useState, type FormEvent } from "react";
import styles from "./StatusElementEditorDrawer.module.css";

type StatusElementFormInput = CreatePublicStatusElementInput;

interface StatusElementEditorDrawerProps {
	open: boolean;
	element?: ApiPublicStatusElement | null;
	elements: ApiPublicStatusElement[];
	checks: ApiCheck[];
	saving: boolean;
	onClose: () => void;
	onSubmit: (body: StatusElementFormInput) => void;
}

const kindOptions: Array<{ value: PublicStatusElementKind; label: string }> = [
	{ value: "folder", label: "Folder" },
	{ value: "check", label: "Check" }
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

function checkLabel(check: ApiCheck) {
	return `${check.name} / ${check.type.toUpperCase()} / ${check.target}`;
}

function initialState(element?: ApiPublicStatusElement | null): StatusElementFormInput {
	return {
		parentElementId: element?.parentElementId,
		kind: element?.kind ?? "check",
		checkId: element?.checkId,
		title: element?.title ?? "",
		description: element?.description ?? "",
		sortOrder: element?.sortOrder ?? 0,
		chartMode: element?.chartMode ?? "inherit",
		chartRange: element?.chartRange
	};
}

export function StatusElementEditorDrawer({ open, element, elements, checks, saving, onClose, onSubmit }: StatusElementEditorDrawerProps) {
	const [form, setForm] = useState<StatusElementFormInput>(() => initialState(element));
	const [error, setError] = useState("");
	const isEditing = Boolean(element);
	const availableParents = useMemo(() => rootFolderOptions(elements.filter(candidate => candidate.id !== element?.id)), [element?.id, elements]);
	const checkOptions = useMemo(() => checks.map(check => ({ value: check.id, label: checkLabel(check) })), [checks]);
	const parentOptions = [{ value: "", label: "Root" }, ...availableParents];
	const isFolder = form.kind === "folder";

	function update<K extends keyof StatusElementFormInput>(key: K, value: StatusElementFormInput[K]) {
		setForm(current => ({ ...current, [key]: value }));
	}

	function handleKindChange(kind: PublicStatusElementKind) {
		setForm(current => ({
			...current,
			kind,
			parentElementId: kind === "folder" ? undefined : current.parentElementId,
			checkId: kind === "folder" ? undefined : current.checkId
		}));
	}

	function handleSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		const title = form.title?.trim();
		const description = form.description?.trim();
		const sortOrder = Number(form.sortOrder);

		if (!Number.isFinite(sortOrder) || sortOrder < 0) {
			setError("Order must be a non-negative number.");
			return;
		}
		if (form.kind === "check" && !form.checkId) {
			setError("Select a check for this element.");
			return;
		}

		onSubmit({
			kind: form.kind,
			parentElementId: form.kind === "check" ? form.parentElementId || undefined : undefined,
			checkId: form.kind === "check" ? form.checkId : undefined,
			title: title || undefined,
			description: description || undefined,
			sortOrder,
			chartMode: form.chartMode || "inherit",
			chartRange: form.chartRange || undefined
		});
	}

	return (
		<EditorDrawer
			open={open}
			title={isEditing ? "Edit Element" : "New Element"}
			ariaLabel={isEditing ? "Edit status page element" : "Create status page element"}
			onClose={onClose}
			actions={
				<Button type="submit" form="status-element-editor-form" disabled={saving}>
					{saving ? "Saving" : "Save"}
				</Button>
			}
		>
			<form id="status-element-editor-form" className={styles.form} onSubmit={handleSubmit}>
				{error ? <div className={styles.formError}>{error}</div> : null}
				<SelectField label="Element type" value={form.kind} disabled={isEditing} options={kindOptions} onChange={event => handleKindChange(event.currentTarget.value as PublicStatusElementKind)} />
				{isFolder ? null : (
					<>
						<SelectField label="Folder" value={form.parentElementId || ""} options={parentOptions} onChange={event => update("parentElementId", event.currentTarget.value || undefined)} />
						<SelectField
							label="Check"
							value={form.checkId || ""}
							options={[{ value: "", label: "Select check", disabled: true }, ...checkOptions]}
							onChange={event => update("checkId", event.currentTarget.value || undefined)}
						/>
					</>
				)}
				<TextField label="Title override" value={form.title || ""} maxLength={1024} onChange={event => update("title", event.currentTarget.value)} />
				<TextAreaField label="Description override" value={form.description || ""} maxLength={1024} rows={4} onChange={event => update("description", event.currentTarget.value)} />
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
			</form>
		</EditorDrawer>
	);
}

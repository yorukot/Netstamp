import type { ApiPublicStatusPage, CreatePublicStatusPageInput, PublicStatusChartMode, PublicStatusChartRange } from "@/shared/api/types";
import { EditorDrawer } from "@/shared/components/EditorDrawer";
import { UnsavedChangesBar } from "@/shared/components/UnsavedChangesBar";
import { Checkbox, SelectField, TextAreaField, TextField } from "@netstamp/ui";
import { useMemo, useState, type FormEvent } from "react";
import styles from "./StatusPageEditorDrawer.module.css";

type StatusPageFormInput = CreatePublicStatusPageInput;

interface StatusPageEditorDrawerProps {
	open: boolean;
	page?: ApiPublicStatusPage | null;
	saving: boolean;
	onClose: () => void;
	onSubmit: (body: StatusPageFormInput) => void;
}

const chartModeOptions: Array<{ value: PublicStatusChartMode; label: string }> = [
	{ value: "off", label: "Charts off" },
	{ value: "compact", label: "Compact charts" }
];

const chartRangeOptions: Array<{ value: PublicStatusChartRange; label: string }> = [
	{ value: "24h", label: "24 hours" },
	{ value: "7d", label: "7 days" },
	{ value: "30d", label: "30 days" }
];

function defaultSlug(title: string) {
	return title
		.trim()
		.toLowerCase()
		.replace(/[^a-z0-9]+/g, "-")
		.replace(/^-+|-+$/g, "")
		.slice(0, 64);
}

function initialState(page?: ApiPublicStatusPage | null): StatusPageFormInput {
	return {
		slug: page?.slug ?? "",
		title: page?.title ?? "",
		description: page?.description ?? "",
		enabled: page?.enabled ?? true,
		footerText: page?.footerText ?? "",
		bannerImageUrl: page?.bannerImageUrl ?? "",
		theme: page?.theme ?? "auto",
		showTargets: page?.showTargets ?? false,
		showProbeNames: page?.showProbeNames ?? false,
		showProbeLocations: page?.showProbeLocations ?? false,
		showIncidentHistory: page?.showIncidentHistory ?? true,
		showGeneratedAt: page?.showGeneratedAt ?? true,
		customCss: page?.customCss ?? "",
		defaultChartMode: page?.defaultChartMode ?? "off",
		defaultChartRange: page?.defaultChartRange ?? "24h"
	};
}

function sameValue(left: unknown, right: unknown) {
	return JSON.stringify(left) === JSON.stringify(right);
}

export function StatusPageEditorDrawer({ open, page, saving, onClose, onSubmit }: StatusPageEditorDrawerProps) {
	const initialForm = useMemo(() => initialState(page), [page]);
	const [form, setForm] = useState<StatusPageFormInput>(() => initialForm);
	const [error, setError] = useState("");
	const isEditing = Boolean(page);
	const hasChanges = !sameValue(form, initialForm);

	function update<K extends keyof StatusPageFormInput>(key: K, value: StatusPageFormInput[K]) {
		setForm(current => ({ ...current, [key]: value }));
	}

	function handleTitleChange(value: string) {
		setForm(current => ({
			...current,
			title: value,
			slug: isEditing || current.slug ? current.slug : defaultSlug(value)
		}));
	}

	function resetForm() {
		setForm(initialForm);
		setError("");
	}

	function handleSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		const slug = form.slug.trim();
		const title = form.title.trim();
		const description = form.description?.trim();

		if (!title || !slug) {
			setError("Title and slug are required.");
			return;
		}
		if (!/^[a-z0-9-]+$/.test(slug)) {
			setError("Slug can only contain lowercase letters, numbers, and hyphens.");
			return;
		}

		onSubmit({
			...form,
			slug,
			title,
			description: description || undefined,
			enabled: Boolean(form.enabled)
		});
	}

	return (
		<EditorDrawer open={open} title={isEditing ? "Edit Status Page" : "New Status Page"} ariaLabel={isEditing ? "Edit status page" : "Create status page"} onClose={onClose}>
			<form id="status-page-editor-form" className={styles.form} onSubmit={handleSubmit}>
				{error ? <div className={styles.formError}>{error}</div> : null}
				<TextField label="Title" value={form.title} maxLength={128} onChange={event => handleTitleChange(event.currentTarget.value)} />
				<TextField label="Slug" helper="Public URL uses /status/:slug." value={form.slug} maxLength={64} onChange={event => update("slug", event.currentTarget.value)} />
				<TextAreaField label="Description" value={form.description || ""} maxLength={1024} rows={4} onChange={event => update("description", event.currentTarget.value)} />
				<label className={styles.checkboxRow}>
					<Checkbox checked={Boolean(form.enabled)} onChange={event => update("enabled", event.currentTarget.checked)} />
					<span>Enabled</span>
				</label>
				<SelectField
					label="Default chart mode"
					value={form.defaultChartMode}
					options={chartModeOptions}
					onChange={event => update("defaultChartMode", event.currentTarget.value as PublicStatusChartMode)}
				/>
				<SelectField
					label="Default chart range"
					value={form.defaultChartRange}
					options={chartRangeOptions}
					onChange={event => update("defaultChartRange", event.currentTarget.value as PublicStatusChartRange)}
				/>
				<UnsavedChangesBar show={hasChanges} saveType="submit" saving={saving} onReset={resetForm} />
			</form>
		</EditorDrawer>
	);
}

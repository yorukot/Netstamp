import { Badge, SelectField, Spinner, TextField } from "@netstamp/ui";
import type { TFunction } from "i18next";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { StatusAssignmentPicker } from "./StatusAssignmentPicker";
import styles from "./StatusAssignmentScopeField.module.css";
import {
	statusAssignmentScopeForMode,
	unavailableAssignmentIds,
	type StatusAssignmentOption,
	type StatusAssignmentScope,
	type StatusAssignmentSelectionMode,
	type StatusCheckOption
} from "./statusAssignmentModel";

export interface StatusAssignmentScopeFieldProps {
	scope: StatusAssignmentScope;
	checks: StatusCheckOption[];
	assignments: StatusAssignmentOption[];
	loading?: boolean;
	unavailable: boolean;
	checkPickerVariant?: "browser" | "select";
	onChange: (scope: StatusAssignmentScope) => void;
	onCheckChange?: (check: StatusCheckOption | undefined) => void;
}

const checkCategoryLabel = (value: string, translate: TFunction<"status">) => {
	switch (value.toLowerCase()) {
		case "all":
			return translate("builder.all");
		case "http":
			return "HTTP";
		case "tcp":
			return "TCP";
		case "traceroute":
			return translate("builder.trace");
		default:
			return value;
	}
};

const StatusCheckBrowser = ({ checks, selectedId, onChange }: { checks: StatusCheckOption[]; selectedId?: string; onChange: (check: StatusCheckOption) => void }) => {
	const { t } = useTranslation("status");
	const [category, setCategory] = useState("all");
	const [search, setSearch] = useState("");
	const categories = ["all", ...Array.from(new Set(checks.map(check => check.type))).sort((left, right) => left.localeCompare(right))];
	const normalizedSearch = search.trim().toLocaleLowerCase();
	const visibleChecks = checks.filter(
		check => (category === "all" || check.type === category) && (!normalizedSearch || `${check.name} ${check.target} ${check.type}`.toLocaleLowerCase().includes(normalizedSearch))
	);

	return (
		<div className={styles.checkPicker}>
			<div className={styles.checkCategories} role="list" aria-label={t("builder.checkCategories")}>
				{categories.map(value => {
					const count = value === "all" ? checks.length : checks.filter(check => check.type === value).length;
					return (
						<button key={value} type="button" className={styles.checkCategory} data-selected={category === value} onClick={() => setCategory(value)}>
							<span>{checkCategoryLabel(value, t)}</span>
							<Badge tone={category === value ? "accent" : "neutral"}>{count}</Badge>
						</button>
					);
				})}
			</div>
			<div className={styles.checkChoices}>
				<TextField label={t("builder.searchChecks")} placeholder={t("builder.searchPlaceholder")} value={search} onChange={event => setSearch(event.currentTarget.value)} />
				<div className={styles.checkChoiceList} role="listbox" aria-label={t("builder.checks")}>
					{visibleChecks.map(check => (
						<button
							key={check.id}
							type="button"
							className={styles.checkChoice}
							role="option"
							aria-selected={selectedId === check.id}
							data-selected={selectedId === check.id}
							onClick={() => onChange(check)}
						>
							<strong>{check.name}</strong>
							<span>{check.target}</span>
						</button>
					))}
					{!visibleChecks.length ? <p className={styles.notice}>{t("builder.noChecksMatch")}</p> : null}
				</div>
			</div>
		</div>
	);
};

export const StatusAssignmentScopeField = ({ scope, checks, assignments, loading = false, unavailable, checkPickerVariant = "select", onChange, onCheckChange }: StatusAssignmentScopeFieldProps) => {
	const { t } = useTranslation("status");
	const currentScope: StatusAssignmentScope = {
		checkId: scope.checkId,
		assignmentSelectionMode: scope.assignmentSelectionMode,
		assignmentIds: scope.assignmentIds
	};
	const selectionMode = currentScope.assignmentSelectionMode ?? "all_check";
	const selectedIds = currentScope.assignmentIds ?? [];
	const missingAssignmentIds = unavailable ? [] : unavailableAssignmentIds(selectedIds, assignments);
	const missingAssignmentSet = new Set(missingAssignmentIds);
	const validAssignmentIds = selectedIds.filter(assignmentId => !missingAssignmentSet.has(assignmentId));

	const changeSelectionMode = (nextMode: StatusAssignmentSelectionMode) => {
		onChange(statusAssignmentScopeForMode(currentScope, nextMode, checks, assignments));
	};

	const changeCheck = (check: StatusCheckOption | undefined) => {
		onChange({ ...currentScope, checkId: check?.id });
		onCheckChange?.(check);
	};

	return (
		<>
			<SelectField
				label={t("builder.assignmentScope")}
				value={selectionMode}
				options={[
					{ value: "all_check", label: t("builder.assignmentScopes.allCheck") },
					{ value: "selected_assignments", label: t("builder.assignmentScopes.selected") }
				]}
				onChange={event => changeSelectionMode(event.currentTarget.value as StatusAssignmentSelectionMode)}
			/>
			{loading ? <Spinner label={t("builder.loadingChecks")} layout="compact" size="sm" /> : null}
			{!loading && !unavailable && selectionMode === "all_check" && checks.length ? (
				checkPickerVariant === "browser" ? (
					<StatusCheckBrowser checks={checks} selectedId={scope.checkId} onChange={changeCheck} />
				) : (
					<SelectField
						label={t("builder.check")}
						value={scope.checkId ?? ""}
						options={[{ value: "", label: t("builder.selectCheckOption"), disabled: true }, ...checks.map(check => ({ value: check.id, label: `${check.name} / ${check.type}` }))]}
						onChange={event => changeCheck(checks.find(check => check.id === event.currentTarget.value))}
					/>
				)
			) : null}
			{!loading && !unavailable && selectionMode === "selected_assignments" ? (
				<StatusAssignmentPicker
					assignments={assignments}
					selectedIds={selectedIds}
					unavailableIds={missingAssignmentIds}
					onChange={assignmentIds => onChange({ ...currentScope, assignmentIds })}
					onRemoveUnavailable={() => onChange({ ...currentScope, assignmentIds: validAssignmentIds })}
				/>
			) : null}
			{!loading && unavailable ? <p className={styles.notice}>{t("builder.assignmentPicker.loadError")}</p> : null}
			{!loading && !unavailable && selectionMode === "all_check" && !checks.length ? <p className={styles.notice}>{t("builder.noChecks")}</p> : null}
		</>
	);
};

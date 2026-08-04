import { Button, CategorizedMultiSelect, type CategorizedMultiSelectCategory } from "@netstamp/ui";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import type { StatusAssignmentOption } from "./statusAssignmentModel";
import styles from "./StatusAssignmentPicker.module.css";

const searchText = (assignment: StatusAssignmentOption) =>
	[assignment.checkName, assignment.checkType, assignment.checkTarget, assignment.probeName, assignment.probeLocationName].filter(Boolean).join(" ").toLocaleLowerCase();

const checkDescription = (assignment: StatusAssignmentOption) =>
	[assignment.checkType.toUpperCase(), assignment.checkTarget, assignment.probeName, assignment.probeLocationName].filter(Boolean).join(" / ");

const probeDescription = (assignment: StatusAssignmentOption) =>
	[assignment.probeLocationName, assignment.checkName, assignment.checkType.toUpperCase(), assignment.checkTarget].filter(Boolean).join(" / ");

export const StatusAssignmentPicker = ({
	assignments,
	selectedIds,
	unavailableIds = [],
	disabled,
	onChange,
	onRemoveUnavailable
}: {
	assignments: StatusAssignmentOption[];
	selectedIds: string[];
	unavailableIds?: string[];
	disabled?: boolean;
	onChange: (assignmentIds: string[]) => void;
	onRemoveUnavailable: () => void;
}) => {
	const { t } = useTranslation("status");
	const categories = useMemo<CategorizedMultiSelectCategory[]>(() => {
		const byCheck = [...assignments]
			.sort((left, right) => left.checkName.localeCompare(right.checkName) || left.probeName.localeCompare(right.probeName) || left.id.localeCompare(right.id))
			.map(assignment => ({
				value: assignment.id,
				label: assignment.checkName,
				description: checkDescription(assignment),
				searchText: searchText(assignment),
				selectionValues: [assignment.id]
			}));
		const byProbe = [...assignments]
			.sort((left, right) => left.probeName.localeCompare(right.probeName) || left.checkName.localeCompare(right.checkName) || left.id.localeCompare(right.id))
			.map(assignment => ({
				value: assignment.id,
				label: assignment.probeName,
				description: probeDescription(assignment),
				searchText: searchText(assignment),
				selectionValues: [assignment.id]
			}));

		return [
			{ value: "check", label: t("builder.assignmentPicker.byCheck"), items: byCheck },
			{ value: "probe", label: t("builder.assignmentPicker.byProbe"), items: byProbe }
		];
	}, [assignments, t]);
	const selected = new Set(selectedIds);
	const selectedAssignments = assignments.filter(assignment => selected.has(assignment.id));
	const single = selectedIds.length === 1 ? selectedAssignments[0] : undefined;
	const valueLabel = single ? `${single.checkName} / ${single.probeName}` : t("builder.assignmentPicker.selected", { count: selectedIds.length });

	return (
		<div className={styles.root}>
			<CategorizedMultiSelect
				label={t("builder.assignments")}
				placeholder={t("builder.assignmentPicker.placeholder")}
				valueLabel={valueLabel}
				categories={categories}
				selectedValues={selectedIds}
				disabled={disabled || !assignments.length}
				searchPlaceholder={t("builder.assignmentPicker.search")}
				selectAllLabel={t("builder.assignmentPicker.selectAll")}
				clearVisibleLabel={t("builder.assignmentPicker.clearVisible")}
				emptyLabel={t("builder.assignmentPicker.empty")}
				optionsAriaLabel={t("builder.assignmentPicker.options")}
				categoriesAriaLabel={t("builder.assignmentPicker.grouping")}
				selectItemAriaLabel={item => t("builder.assignmentPicker.selectItem", { item: item.label })}
				onValueChange={onChange}
			/>
			{!assignments.length ? <p className={styles.notice}>{t("builder.assignmentPicker.noneAvailable")}</p> : null}
			{unavailableIds.length ? (
				<div className={styles.warning} role="alert">
					<p>{t("builder.assignmentPicker.unavailable", { count: unavailableIds.length })}</p>
					<Button type="button" variant="ghost" size="sm" onClick={onRemoveUnavailable}>
						{t("builder.assignmentPicker.removeUnavailable")}
					</Button>
				</div>
			) : null}
		</div>
	);
};

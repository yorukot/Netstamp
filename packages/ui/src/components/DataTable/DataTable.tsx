import { ArrowsDownUpIcon } from "@phosphor-icons/react/dist/csr/ArrowsDownUp";
import { SortAscendingIcon } from "@phosphor-icons/react/dist/csr/SortAscending";
import { SortDescendingIcon } from "@phosphor-icons/react/dist/csr/SortDescending";
import type { CSSProperties, KeyboardEvent, MouseEvent, ReactNode } from "react";
import { useMemo, useState } from "react";
import { Checkbox } from "../Field/Field";
import styles from "./DataTable.module.css";

export interface DataColumn<Row extends object = Record<string, unknown>> {
	key: string;
	label: ReactNode;
	render?: (row: Row, index: number) => ReactNode;
	sortable?: boolean;
	sortValue?: (row: Row) => string | number | null | undefined;
	className?: string;
	headerClassName?: string;
}

export type DataTableSortDirection = "asc" | "desc";

export interface DataTableSortState {
	key: string;
	direction: DataTableSortDirection;
}

export interface DataTableProps<Row extends object = Record<string, unknown>> {
	columns: DataColumn<Row>[];
	rows: Row[];
	className?: string;
	density?: "normal" | "compact";
	minWidth?: string;
	maxHeight?: string;
	style?: CSSProperties;
	ariaLabel?: string;
	getRowKey?: (row: Row, index: number) => string;
	getRowAriaLabel?: (row: Row, index: number) => string;
	onRowClick?: (row: Row) => void;
	selectedKey?: string;
	selectable?: boolean;
	selectedRowKeys?: readonly string[] | ReadonlySet<string> | Record<string, boolean>;
	onSelectedRowKeysChange?: (keys: string[]) => void;
	sort?: DataTableSortState | null;
	defaultSort?: DataTableSortState | null;
	onSortChange?: (sort: DataTableSortState | null) => void;
	rowActions?: (row: Row, index: number) => ReactNode;
	rowActionsLabel?: ReactNode;
	rowActionsClassName?: string;
	rowActionsHeaderClassName?: string;
	batchLabel?: ReactNode;
	batchActions?: ReactNode;
	emptyLabel?: ReactNode;
}

type DataTableStyle = CSSProperties & {
	"--ns-data-table-min-width"?: string;
	"--ns-data-table-max-height"?: string;
};

export function DataTable<Row extends object>({
	columns,
	rows,
	className,
	density = "normal",
	minWidth,
	maxHeight,
	style,
	ariaLabel,
	getRowKey,
	getRowAriaLabel,
	onRowClick,
	selectedKey,
	selectable = false,
	selectedRowKeys,
	onSelectedRowKeysChange,
	sort,
	defaultSort = null,
	onSortChange,
	rowActions,
	rowActionsLabel = "Actions",
	rowActionsClassName,
	rowActionsHeaderClassName,
	batchLabel,
	batchActions,
	emptyLabel = "No results"
}: DataTableProps<Row>) {
	const wrapStyle: DataTableStyle = { ...style };
	const [internalSort, setInternalSort] = useState<DataTableSortState | null>(defaultSort);
	const activeSort = sort === undefined ? internalSort : sort;
	const selectedSet = useMemo(() => selectedKeysToSet(selectedRowKeys), [selectedRowKeys]);
	const visibleRowKeys = useMemo(() => rows.map((row, index) => rowKeyFor(row, index, getRowKey)), [getRowKey, rows]);
	const selectedArray = useMemo(() => Array.from(selectedSet), [selectedSet]);
	const sortedRows = useMemo(() => sortRows(rows, columns, activeSort), [activeSort, columns, rows]);
	const allVisibleSelected = visibleRowKeys.length > 0 && visibleRowKeys.every(rowKey => selectedSet.has(rowKey));
	const someVisibleSelected = visibleRowKeys.some(rowKey => selectedSet.has(rowKey));
	const shouldRenderBatch = selectable && selectedSet.size > 0 && (batchLabel || batchActions);
	const totalColumnCount = columns.length + (selectable ? 1 : 0) + (rowActions ? 1 : 0);

	if (minWidth) {
		wrapStyle["--ns-data-table-min-width"] = minWidth;
	}

	if (maxHeight) {
		wrapStyle["--ns-data-table-max-height"] = maxHeight;
	}

	function commitSort(nextSort: DataTableSortState | null) {
		if (sort === undefined) {
			setInternalSort(nextSort);
		}

		onSortChange?.(nextSort);
	}

	function toggleSort(column: DataColumn<Row>) {
		if (!column.sortable) {
			return;
		}

		if (activeSort?.key !== column.key) {
			commitSort({ key: column.key, direction: "asc" });
			return;
		}

		if (activeSort.direction === "asc") {
			commitSort({ key: column.key, direction: "desc" });
			return;
		}

		commitSort(null);
	}

	function updateSelectedKeys(nextSelectedSet: Set<string>) {
		onSelectedRowKeysChange?.(Array.from(nextSelectedSet));
	}

	function toggleVisibleRows(checked: boolean) {
		const nextSelectedSet = new Set(selectedSet);

		for (const rowKey of visibleRowKeys) {
			if (checked) {
				nextSelectedSet.add(rowKey);
			} else {
				nextSelectedSet.delete(rowKey);
			}
		}

		updateSelectedKeys(nextSelectedSet);
	}

	function toggleRowSelection(rowKey: string, checked: boolean) {
		const nextSelectedSet = new Set(selectedSet);

		if (checked) {
			nextSelectedSet.add(rowKey);
		} else {
			nextSelectedSet.delete(rowKey);
		}

		updateSelectedKeys(nextSelectedSet);
	}

	function handleRowKeyDown(event: KeyboardEvent<HTMLTableRowElement>, row: Row) {
		if (!onRowClick || (event.key !== "Enter" && event.key !== " ")) {
			return;
		}

		event.preventDefault();
		onRowClick(row);
	}

	function stopRowAction(event: MouseEvent<HTMLElement>) {
		event.stopPropagation();
	}

	return (
		<div className={["ns-frame", styles.frame, styles[density], className].filter(Boolean).join(" ")} style={wrapStyle}>
			{shouldRenderBatch ? (
				<div className={styles.batchBar}>
					<div className={styles.batchLabel}>{batchLabel ?? `${selectedSet.size} selected`}</div>
					{batchActions ? <div className={styles.batchActions}>{batchActions}</div> : null}
				</div>
			) : null}
			<div className={["ns-scrollbar", styles.scroller].join(" ")}>
				<table className={styles.table} aria-label={ariaLabel} aria-multiselectable={selectable || undefined}>
					<thead>
						<tr>
							{selectable ? (
								<th className={styles.selectionHeader}>
									<Checkbox
										ref={input => {
											if (input) {
												input.indeterminate = someVisibleSelected && !allVisibleSelected;
											}
										}}
										aria-label="Select visible rows"
										checked={allVisibleSelected}
										onChange={event => toggleVisibleRows(event.currentTarget.checked)}
									/>
								</th>
							) : null}
							{columns.map(column => (
								<th className={column.headerClassName} key={column.key} aria-sort={ariaSortValue(activeSort, column)}>
									{column.sortable ? (
										<button type="button" className={styles.sortButton} onClick={() => toggleSort(column)}>
											<span>{column.label}</span>
											<span className={[styles.sortIndicator, activeSort?.key === column.key && styles.sortIndicatorActive].filter(Boolean).join(" ")} aria-hidden="true">
												<SortIndicator sort={activeSort} column={column} />
											</span>
										</button>
									) : (
										column.label
									)}
								</th>
							))}
							{rowActions ? <th className={[styles.actionsHeader, rowActionsHeaderClassName].filter(Boolean).join(" ")}>{rowActionsLabel}</th> : null}
						</tr>
					</thead>
					<tbody>
						{sortedRows.length ? (
							sortedRows.map(({ row, index }) => {
								const rowKey = rowKeyFor(row, index, getRowKey);
								const selected = selectedKey === rowKey;
								const rowChecked = selectedSet.has(rowKey);
								const interactive = Boolean(onRowClick);

								return (
									<tr
										key={rowKey}
										className={[selected && styles.selected, interactive && styles.interactive].filter(Boolean).join(" ") || undefined}
										aria-label={interactive ? (getRowAriaLabel?.(row, index) ?? `Select row ${rowKey}`) : undefined}
										aria-selected={selected || rowChecked || undefined}
										tabIndex={interactive ? 0 : undefined}
										onClick={onRowClick ? () => onRowClick(row) : undefined}
										onKeyDown={event => handleRowKeyDown(event, row)}
									>
										{selectable ? (
											<td className={styles.selectionCell} onClick={stopRowAction}>
												<Checkbox
													aria-label={`Select ${getRowAriaLabel?.(row, index) ?? `row ${rowKey}`}`}
													checked={rowChecked}
													onChange={event => toggleRowSelection(rowKey, event.currentTarget.checked)}
												/>
											</td>
										) : null}
										{columns.map(column => (
											<td className={column.className} key={column.key}>
												{column.render ? column.render(row, index) : String((row as Record<string, unknown>)[column.key] ?? "")}
											</td>
										))}
										{rowActions ? (
											<td className={[styles.actionsCell, rowActionsClassName].filter(Boolean).join(" ")} onClick={stopRowAction}>
												{rowActions(row, index)}
											</td>
										) : null}
									</tr>
								);
							})
						) : (
							<tr>
								<td className={styles.empty} colSpan={totalColumnCount}>
									{emptyLabel}
								</td>
							</tr>
						)}
					</tbody>
				</table>
			</div>
		</div>
	);
}

function selectedKeysToSet(value: DataTableProps["selectedRowKeys"]) {
	if (!value) {
		return new Set<string>();
	}

	if (value instanceof Set) {
		return new Set(value);
	}

	if (Array.isArray(value)) {
		return new Set(value);
	}

	return new Set(
		Object.entries(value)
			.filter(([, selected]) => selected)
			.map(([key]) => key)
	);
}

function rowKeyFor<Row extends object>(row: Row, index: number, getRowKey: DataTableProps<Row>["getRowKey"]) {
	return getRowKey ? getRowKey(row, index) : String(index);
}

function dataTableValue<Row extends object>(row: Row, column: DataColumn<Row>) {
	return column.sortValue ? column.sortValue(row) : ((row as Record<string, unknown>)[column.key] as string | number | null | undefined);
}

function compareValues(left: string | number | null | undefined, right: string | number | null | undefined) {
	if (left === right) {
		return 0;
	}

	if (left === null || left === undefined || (typeof left === "number" && Number.isNaN(left))) {
		return 1;
	}

	if (right === null || right === undefined || (typeof right === "number" && Number.isNaN(right))) {
		return -1;
	}

	if (typeof left === "number" && typeof right === "number") {
		return left - right;
	}

	return String(left).localeCompare(String(right), undefined, { numeric: true, sensitivity: "base" });
}

function sortRows<Row extends object>(rows: Row[], columns: DataColumn<Row>[], sort: DataTableSortState | null | undefined) {
	const indexedRows = rows.map((row, index) => ({ row, index }));

	if (!sort) {
		return indexedRows;
	}

	const column = columns.find(item => item.key === sort.key);

	if (!column?.sortable) {
		return indexedRows;
	}

	return [...indexedRows].sort((left, right) => {
		const compared = compareValues(dataTableValue(left.row, column), dataTableValue(right.row, column));
		return sort.direction === "asc" ? compared : -compared;
	});
}

function ariaSortValue<Row extends object>(sort: DataTableSortState | null | undefined, column: DataColumn<Row>) {
	if (!column.sortable || sort?.key !== column.key) {
		return undefined;
	}

	return sort.direction === "asc" ? "ascending" : "descending";
}

function SortIndicator<Row extends object>({ sort, column }: { sort: DataTableSortState | null | undefined; column: DataColumn<Row> }) {
	if (sort?.key !== column.key) {
		return <ArrowsDownUpIcon size={14} weight="bold" focusable="false" />;
	}

	if (sort.direction === "asc") {
		return <SortAscendingIcon size={14} weight="bold" focusable="false" />;
	}

	return <SortDescendingIcon size={14} weight="bold" focusable="false" />;
}

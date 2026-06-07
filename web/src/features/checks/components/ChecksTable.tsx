import type { CheckDefinition } from "@/features/checks/data/checks";
import { classNames } from "@/shared/utils/classNames";
import { Badge, Button, Checkbox } from "@netstamp/ui";
import { ArrowsDownUp, CaretDown, CaretUp, Copy, PencilSimple, Trash } from "@phosphor-icons/react";
import {
	flexRender,
	getCoreRowModel,
	getFilteredRowModel,
	getSortedRowModel,
	useReactTable,
	type Column,
	type ColumnDef,
	type OnChangeFn,
	type RowSelectionState,
	type SortingState
} from "@tanstack/react-table";
import { useMemo, useState, type MouseEvent, type ReactNode } from "react";
import styles from "./ChecksPage.module.css";

export type CheckTypeFilter = "all" | "ping" | "tcp" | "traceroute";

interface ChecksTableProps {
	actionDisabled?: boolean;
	checks: CheckDefinition[];
	onDeleteCheck: (check: CheckDefinition) => void;
	onDuplicateCheck: (check: CheckDefinition) => void;
	onOpenCheck: (check: CheckDefinition) => void;
	onRowSelectionChange: OnChangeFn<RowSelectionState>;
	rowSelection: RowSelectionState;
	search: string;
	selectedKey: string;
	typeFilter: CheckTypeFilter;
}

const typeFilterLabels: Record<Exclude<CheckTypeFilter, "all">, CheckDefinition["type"]> = {
	ping: "Ping",
	tcp: "TCP",
	traceroute: "Traceroute"
};

function intervalValue(value: string) {
	const parsed = Number.parseInt(value, 10);
	return Number.isFinite(parsed) ? parsed : 0;
}

function SortHeader({ column, label }: { column: Column<CheckDefinition>; label: string }) {
	const sorted = column.getIsSorted();
	const Icon = sorted === "asc" ? CaretUp : sorted === "desc" ? CaretDown : ArrowsDownUp;

	return (
		<button type="button" className={styles.tableSortButton} onClick={column.getToggleSortingHandler()}>
			<span>{label}</span>
			<Icon size={13} weight="bold" aria-hidden="true" focusable="false" />
		</button>
	);
}

export function ChecksTable({ actionDisabled, checks, onDeleteCheck, onDuplicateCheck, onOpenCheck, onRowSelectionChange, rowSelection, search, selectedKey, typeFilter }: ChecksTableProps) {
	const [sorting, setSorting] = useState<SortingState>([]);
	const columns = useMemo<ColumnDef<CheckDefinition>[]>(
		() => [
			{
				id: "select",
				enableSorting: false,
				header: ({ table }) => {
					const visibleRows = table.getRowModel().rows;
					const allVisibleSelected = visibleRows.length > 0 && visibleRows.every(row => row.getIsSelected());
					const someVisibleSelected = visibleRows.some(row => row.getIsSelected());

					return (
						<Checkbox
							ref={input => {
								if (input) {
									input.indeterminate = someVisibleSelected && !allVisibleSelected;
								}
							}}
							aria-label="Select visible checks"
							checked={allVisibleSelected}
							onChange={event => {
								const { checked } = event.currentTarget;
								for (const row of visibleRows) {
									row.toggleSelected(checked);
								}
							}}
						/>
					);
				},
				cell: ({ row }) => <Checkbox aria-label={`Select ${row.original.name}`} checked={row.getIsSelected()} onClick={event => event.stopPropagation()} onChange={row.getToggleSelectedHandler()} />
			},
			{
				accessorKey: "name",
				header: ({ column }) => <SortHeader column={column} label="Check name" />,
				cell: ({ row }) => (
					<div className={styles.checkNameCell}>
						<strong>{row.original.name}</strong>
						<span>{row.original.description || row.original.status}</span>
					</div>
				)
			},
			{
				accessorKey: "type",
				filterFn: (row, _columnId, value) => value === "all" || row.original.type === typeFilterLabels[value as Exclude<CheckTypeFilter, "all">],
				header: ({ column }) => <SortHeader column={column} label="Type" />,
				cell: ({ row }) => <Badge tone="accent">{row.original.type}</Badge>
			},
			{
				accessorKey: "target",
				header: ({ column }) => <SortHeader column={column} label="Target" />
			},
			{
				accessorKey: "interval",
				sortingFn: (rowA, rowB) => intervalValue(rowA.original.interval) - intervalValue(rowB.original.interval),
				header: ({ column }) => <SortHeader column={column} label="Interval" />
			},
			{
				accessorKey: "assigned",
				header: ({ column }) => <SortHeader column={column} label="Assigned probes" />,
				cell: ({ row }) => <Badge tone={row.original.assigned ? "accent" : "muted"}>{row.original.assigned}</Badge>
			},
			{
				id: "actions",
				enableSorting: false,
				header: "Actions",
				cell: ({ row }) => (
					<div className={styles.rowActions}>
						<IconAction label={`Open ${row.original.name}`} onClick={() => onOpenCheck(row.original)} disabled={actionDisabled}>
							<PencilSimple size={15} weight="bold" aria-hidden="true" focusable="false" />
						</IconAction>
						<IconAction label={`Duplicate ${row.original.name}`} onClick={() => onDuplicateCheck(row.original)} disabled={actionDisabled}>
							<Copy size={15} weight="bold" aria-hidden="true" focusable="false" />
						</IconAction>
						<IconAction label={`Delete ${row.original.name}`} onClick={() => onDeleteCheck(row.original)} disabled={actionDisabled} danger>
							<Trash size={15} weight="bold" aria-hidden="true" focusable="false" />
						</IconAction>
					</div>
				)
			}
		],
		[actionDisabled, onDeleteCheck, onDuplicateCheck, onOpenCheck]
	);
	const columnFilters = useMemo(() => (typeFilter === "all" ? [] : [{ id: "type", value: typeFilter }]), [typeFilter]);
	// TanStack Table exposes instance helpers that React Compiler intentionally skips.
	// eslint-disable-next-line react-hooks/incompatible-library
	const table = useReactTable({
		columns,
		data: checks,
		enableRowSelection: true,
		getCoreRowModel: getCoreRowModel(),
		getFilteredRowModel: getFilteredRowModel(),
		getRowId: row => row.id,
		getSortedRowModel: getSortedRowModel(),
		globalFilterFn: (row, _columnId, value) => {
			const needle = String(value ?? "")
				.trim()
				.toLowerCase();
			if (!needle) {
				return true;
			}

			return [row.original.name, row.original.target, row.original.description, row.original.type].some(field => (field ?? "").toLowerCase().includes(needle));
		},
		onRowSelectionChange,
		onSortingChange: setSorting,
		state: {
			columnFilters,
			globalFilter: search,
			rowSelection,
			sorting
		}
	});

	return (
		<div className={classNames("ns-cut-frame", styles.checkTableFrame)}>
			<div className={classNames("ns-scrollbar", styles.checkTableScroller)}>
				<table className={styles.checkTable} aria-label="Project checks">
					<thead>
						{table.getHeaderGroups().map(headerGroup => (
							<tr key={headerGroup.id}>
								{headerGroup.headers.map(header => (
									<th key={header.id}>{header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}</th>
								))}
							</tr>
						))}
					</thead>
					<tbody>
						{table.getRowModel().rows.length ? (
							table.getRowModel().rows.map(row => {
								const selected = row.id === selectedKey;

								return (
									<tr
										key={row.id}
										className={classNames(styles.checkTableRow, selected && styles.checkTableRowSelected)}
										aria-selected={selected || undefined}
										tabIndex={0}
										onClick={() => onOpenCheck(row.original)}
										onKeyDown={event => {
											if (event.key === "Enter" || event.key === " ") {
												event.preventDefault();
												onOpenCheck(row.original);
											}
										}}
									>
										{row.getVisibleCells().map(cell => (
											<td key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</td>
										))}
									</tr>
								);
							})
						) : (
							<tr>
								<td className={styles.checkTableEmpty} colSpan={columns.length}>
									No checks match the current filters.
								</td>
							</tr>
						)}
					</tbody>
				</table>
			</div>
		</div>
	);
}

function IconAction({ children, danger, disabled, label, onClick }: { children: ReactNode; danger?: boolean; disabled?: boolean; label: string; onClick: () => void }) {
	function handleClick(event: MouseEvent<HTMLButtonElement>) {
		event.stopPropagation();
		onClick();
	}

	return (
		<Button
			className={classNames(styles.iconAction, danger && styles.iconActionDanger)}
			type="button"
			variant="ghost"
			size="sm"
			aria-label={label}
			title={label}
			disabled={disabled}
			onClick={handleClick}
		>
			{children}
		</Button>
	);
}

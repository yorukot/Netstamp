import type { CSSProperties, KeyboardEvent, ReactNode } from "react";
import styles from "./DataTable.module.css";

export interface DataColumn<Row extends object = Record<string, unknown>> {
	key: string;
	label: ReactNode;
	render?: (row: Row, index: number) => ReactNode;
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
	emptyLabel = "No results"
}: DataTableProps<Row>) {
	const wrapStyle: DataTableStyle = { ...style };

	if (minWidth) {
		wrapStyle["--ns-data-table-min-width"] = minWidth;
	}

	if (maxHeight) {
		wrapStyle["--ns-data-table-max-height"] = maxHeight;
	}

	function handleRowKeyDown(event: KeyboardEvent<HTMLTableRowElement>, row: Row) {
		if (!onRowClick || (event.key !== "Enter" && event.key !== " ")) {
			return;
		}

		event.preventDefault();
		onRowClick(row);
	}

	return (
		<div className={["ns-cut-frame", styles.frame, styles[density], className].filter(Boolean).join(" ")} style={wrapStyle}>
			<div className={["ns-scrollbar", styles.scroller].join(" ")}>
				<table className={styles.table} aria-label={ariaLabel}>
					<thead>
						<tr>
							{columns.map(column => (
								<th key={column.key}>{column.label}</th>
							))}
						</tr>
					</thead>
					<tbody>
						{rows.length ? (
							rows.map((row, index) => {
								const rowKey = getRowKey ? getRowKey(row, index) : String(index);
								const selected = selectedKey === rowKey;
								const interactive = Boolean(onRowClick);

								return (
									<tr
										key={rowKey}
										className={[selected && styles.selected, interactive && styles.interactive].filter(Boolean).join(" ") || undefined}
										role={interactive ? "button" : undefined}
										aria-label={interactive ? (getRowAriaLabel?.(row, index) ?? `Select row ${rowKey}`) : undefined}
										aria-selected={selected || undefined}
										tabIndex={interactive ? 0 : undefined}
										onClick={onRowClick ? () => onRowClick(row) : undefined}
										onKeyDown={event => handleRowKeyDown(event, row)}
									>
										{columns.map(column => (
											<td key={column.key}>{column.render ? column.render(row, index) : String((row as Record<string, unknown>)[column.key] ?? "")}</td>
										))}
									</tr>
								);
							})
						) : (
							<tr>
								<td className={styles.empty} colSpan={columns.length}>
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

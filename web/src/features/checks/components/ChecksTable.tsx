import type { CheckDefinition } from "@/features/checks/data/checks";
import { classNames } from "@/shared/utils/classNames";
import { Badge, Button, DataTable, IconButton, type DataColumn } from "@netstamp/ui";
import { CopyIcon } from "@phosphor-icons/react/dist/csr/Copy";
import { PencilSimpleIcon } from "@phosphor-icons/react/dist/csr/PencilSimple";
import { TrashIcon } from "@phosphor-icons/react/dist/csr/Trash";
import { useMemo, type MouseEvent, type ReactNode } from "react";
import styles from "./ChecksPage.module.css";

export type CheckTypeFilter = "all" | "ping" | "tcp" | "traceroute";
export type CheckRowSelectionState = Record<string, boolean>;

interface ChecksTableProps {
	actionDisabled?: boolean;
	batchDeleteDisabled?: boolean;
	batchDeletePending?: boolean;
	checks: CheckDefinition[];
	onDeleteCheck: (check: CheckDefinition) => void;
	onDeleteSelectedChecks: () => void;
	onDuplicateCheck: (check: CheckDefinition) => void;
	onOpenCheck: (check: CheckDefinition) => void;
	onRowSelectionChange: (rowSelection: CheckRowSelectionState) => void;
	rowSelection: CheckRowSelectionState;
	search: string;
	selectedKey: string;
	selectedSummary: ReactNode;
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

function checkMatchesSearch(check: CheckDefinition, search: string) {
	const needle = search.trim().toLowerCase();

	if (!needle) {
		return true;
	}

	return [check.name, check.target, check.description, check.type].some(field => (field ?? "").toLowerCase().includes(needle));
}

function checkMatchesType(check: CheckDefinition, typeFilter: CheckTypeFilter) {
	return typeFilter === "all" || check.type === typeFilterLabels[typeFilter];
}

function rowSelectionKeys(rowSelection: CheckRowSelectionState) {
	return Object.entries(rowSelection)
		.filter(([, selected]) => selected)
		.map(([key]) => key);
}

function keysToRowSelection(keys: string[]): CheckRowSelectionState {
	return Object.fromEntries(keys.map(key => [key, true]));
}

export function ChecksTable({
	actionDisabled,
	batchDeleteDisabled,
	batchDeletePending,
	checks,
	onDeleteCheck,
	onDeleteSelectedChecks,
	onDuplicateCheck,
	onOpenCheck,
	onRowSelectionChange,
	rowSelection,
	search,
	selectedKey,
	selectedSummary,
	typeFilter
}: ChecksTableProps) {
	const filteredChecks = useMemo(() => checks.filter(check => checkMatchesType(check, typeFilter) && checkMatchesSearch(check, search)), [checks, search, typeFilter]);
	const selectedRowKeys = useMemo(() => rowSelectionKeys(rowSelection), [rowSelection]);
	const columns = useMemo<DataColumn<CheckDefinition>[]>(
		() => [
			{
				key: "name",
				label: "Check name",
				sortable: true,
				render: check => (
					<div className={styles.checkNameCell}>
						<strong>{check.name}</strong>
						<span>{check.description || check.status}</span>
					</div>
				)
			},
			{
				key: "type",
				label: "Type",
				sortable: true,
				render: check => <Badge tone="accent">{check.type}</Badge>
			},
			{
				key: "target",
				label: "Target",
				sortable: true
			},
			{
				key: "interval",
				label: "Interval",
				sortable: true,
				sortValue: check => intervalValue(check.interval)
			},
			{
				key: "assigned",
				label: "Assigned probes",
				sortable: true,
				render: check => <Badge tone={check.assigned ? "accent" : "muted"}>{check.assigned}</Badge>
			}
		],
		[]
	);

	return (
		<DataTable
			ariaLabel="Project checks"
			className={styles.checkTableFrame}
			columns={columns}
			rows={filteredChecks}
			density="compact"
			minWidth="72rem"
			getRowKey={check => check.id}
			getRowAriaLabel={check => `Open check ${check.name}`}
			onRowClick={onOpenCheck}
			selectedKey={selectedKey}
			selectable
			selectedRowKeys={selectedRowKeys}
			onSelectedRowKeysChange={keys => onRowSelectionChange(keysToRowSelection(keys))}
			defaultSort={{ key: "name", direction: "asc" }}
			batchLabel={selectedSummary}
			batchActions={
				<Button type="button" variant="danger" size="sm" disabled={batchDeleteDisabled || batchDeletePending} onClick={onDeleteSelectedChecks}>
					{batchDeletePending ? "Deleting" : "Delete selected"}
				</Button>
			}
			rowActions={check => (
				<div className={styles.rowActions}>
					<IconAction label={`Open ${check.name}`} onClick={() => onOpenCheck(check)} disabled={actionDisabled}>
						<PencilSimpleIcon size={15} weight="bold" aria-hidden="true" focusable="false" />
					</IconAction>
					<IconAction label={`Duplicate ${check.name}`} onClick={() => onDuplicateCheck(check)} disabled={actionDisabled}>
						<CopyIcon size={15} weight="bold" aria-hidden="true" focusable="false" />
					</IconAction>
					<IconAction label={`Delete ${check.name}`} onClick={() => onDeleteCheck(check)} disabled={actionDisabled} danger>
						<TrashIcon size={15} weight="bold" aria-hidden="true" focusable="false" />
					</IconAction>
				</div>
			)}
			emptyLabel="No checks match the current filters."
		/>
	);
}

function IconAction({ children, danger, disabled, label, onClick }: { children: ReactNode; danger?: boolean; disabled?: boolean; label: string; onClick: () => void }) {
	function handleClick(event: MouseEvent<HTMLButtonElement>) {
		event.stopPropagation();
		onClick();
	}

	return (
		<IconButton className={classNames(styles.iconAction, danger && styles.iconActionDanger)} aria-label={label} disabled={disabled} danger={danger} onClick={handleClick}>
			{children}
		</IconButton>
	);
}

import type { CheckDefinition } from "@/features/checks/data/checks";
import { classNames } from "@/shared/utils/classNames";
import { Badge, Button, DataTable, IconButton, PopoverAnchor, PopoverContent, PopoverPortal, PopoverRoot, type DataColumn } from "@netstamp/ui";
import { CopyIcon } from "@phosphor-icons/react/dist/csr/Copy";
import { InfoIcon } from "@phosphor-icons/react/dist/csr/Info";
import { PencilSimpleIcon } from "@phosphor-icons/react/dist/csr/PencilSimple";
import { TrashIcon } from "@phosphor-icons/react/dist/csr/Trash";
import { useMemo, useState, type MouseEvent, type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import styles from "./ChecksPage.module.css";

export type CheckTypeFilter = "all" | "ping" | "tcp" | "traceroute" | "http";
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
	traceroute: "Traceroute",
	http: "HTTP"
};

const checkTypeBadgeClasses: Record<CheckDefinition["type"], string> = {
	Ping: styles.checkTypePing,
	TCP: styles.checkTypeTcp,
	Traceroute: styles.checkTypeTraceroute,
	HTTP: styles.checkTypeHttp
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

function stopRowSelection(event: MouseEvent) {
	event.stopPropagation();
}

function tooltipDescription(description: string) {
	return description.replace(/\s*\r?\n\s*/g, " ").trim();
}

function CheckDescriptionHint({ check }: { check: CheckDefinition }) {
	const { t } = useTranslation("checks");
	const [hovered, setHovered] = useState(false);
	const [pinned, setPinned] = useState(false);
	const open = hovered || pinned;

	if (!check.description) {
		return null;
	}

	const description = tooltipDescription(check.description);

	function handleOpenChange(nextOpen: boolean) {
		if (!nextOpen) {
			setHovered(false);
			setPinned(false);
		}
	}

	function handleTriggerClick(event: MouseEvent<HTMLButtonElement>) {
		stopRowSelection(event);

		if (pinned) {
			setHovered(false);
			setPinned(false);
			return;
		}

		setPinned(true);
	}

	return (
		<PopoverRoot open={open} onOpenChange={handleOpenChange}>
			<PopoverAnchor asChild>
				<span className={styles.descriptionHint} onMouseEnter={() => setHovered(true)} onMouseLeave={() => setHovered(false)} onFocus={() => setHovered(true)} onBlur={() => setHovered(false)}>
					<button
						type="button"
						className={classNames(styles.descriptionTrigger, open && styles.descriptionTriggerOpen)}
						aria-label={t("table.showDescription", { name: check.name })}
						aria-expanded={open}
						onClick={handleTriggerClick}
					>
						<InfoIcon size={13} weight="bold" aria-hidden="true" focusable="false" />
					</button>
				</span>
			</PopoverAnchor>
			<PopoverPortal>
				<PopoverContent className={styles.descriptionPopover} align="start" side="top" sideOffset={8} collisionPadding={8} onClick={stopRowSelection} onOpenAutoFocus={event => event.preventDefault()}>
					{description}
				</PopoverContent>
			</PopoverPortal>
		</PopoverRoot>
	);
}

function CheckTypeBadge({ type }: { type: CheckDefinition["type"] }) {
	return <Badge className={classNames(styles.checkTypeBadge, checkTypeBadgeClasses[type])}>{type}</Badge>;
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
	const { t } = useTranslation("checks");
	const filteredChecks = useMemo(() => checks.filter(check => checkMatchesType(check, typeFilter) && checkMatchesSearch(check, search)), [checks, search, typeFilter]);
	const selectedRowKeys = useMemo(() => rowSelectionKeys(rowSelection), [rowSelection]);
	const columns = useMemo<DataColumn<CheckDefinition>[]>(
		() => [
			{
				key: "name",
				label: t("table.name"),
				sortable: true,
				render: check => (
					<div className={styles.checkNameCell}>
						<strong>{check.name}</strong>
						<CheckDescriptionHint check={check} />
					</div>
				)
			},
			{
				key: "type",
				label: t("table.type"),
				sortable: true,
				render: check => <CheckTypeBadge type={check.type} />
			},
			{
				key: "target",
				label: t("table.target"),
				sortable: true
			},
			{
				key: "interval",
				label: t("table.interval"),
				sortable: true,
				sortValue: check => intervalValue(check.interval)
			},
			{
				key: "assigned",
				label: t("table.assigned"),
				sortable: true,
				render: check => <Badge tone={check.assigned ? "accent" : "muted"}>{check.assigned}</Badge>
			}
		],
		[t]
	);

	return (
		<DataTable
			ariaLabel={t("table.aria")}
			className={styles.checkTableFrame}
			columns={columns}
			rows={filteredChecks}
			density="compact"
			minWidth="72rem"
			getRowKey={check => check.id}
			getRowAriaLabel={check => t("table.open", { name: check.name })}
			onRowClick={onOpenCheck}
			selectedKey={selectedKey}
			selectable
			selectedRowKeys={selectedRowKeys}
			onSelectedRowKeysChange={keys => onRowSelectionChange(keysToRowSelection(keys))}
			defaultSort={{ key: "name", direction: "asc" }}
			batchLabel={selectedSummary}
			batchActions={
				<Button type="button" variant="danger" size="sm" disabled={batchDeleteDisabled || batchDeletePending} onClick={onDeleteSelectedChecks}>
					{batchDeletePending ? t("deleting") : t("table.deleteSelected")}
				</Button>
			}
			rowActions={check => (
				<div className={styles.rowActions}>
					<IconAction label={t("table.open", { name: check.name })} onClick={() => onOpenCheck(check)} disabled={actionDisabled}>
						<PencilSimpleIcon size={15} weight="bold" aria-hidden="true" focusable="false" />
					</IconAction>
					<IconAction label={t("table.duplicate", { name: check.name })} onClick={() => onDuplicateCheck(check)} disabled={actionDisabled}>
						<CopyIcon size={15} weight="bold" aria-hidden="true" focusable="false" />
					</IconAction>
					<IconAction label={t("table.delete", { name: check.name })} onClick={() => onDeleteCheck(check)} disabled={actionDisabled} danger>
						<TrashIcon size={15} weight="bold" aria-hidden="true" focusable="false" />
					</IconAction>
				</div>
			)}
			emptyLabel={t("table.empty")}
		/>
	);
}

function IconAction({ children, danger, disabled, label, onClick }: { children: ReactNode; danger?: boolean; disabled?: boolean; label: string; onClick: () => void }) {
	function handleClick(event: MouseEvent<HTMLButtonElement>) {
		stopRowSelection(event);
		onClick();
	}

	return (
		<IconButton className={classNames(styles.iconAction, danger && styles.iconActionDanger)} aria-label={label} disabled={disabled} danger={danger} onClick={handleClick}>
			{children}
		</IconButton>
	);
}

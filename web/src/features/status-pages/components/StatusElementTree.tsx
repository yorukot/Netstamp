import { chartModeLabel, chartRangeLabel, checkTypeLabel, type ElementTreeNode } from "@/features/status-pages/api/statusPageAdapters";
import type { ApiPublicStatusElement } from "@/shared/api/types";
import { Badge, Button } from "@netstamp/ui";
import { PencilSimple, Trash } from "@phosphor-icons/react";
import styles from "./StatusElementTree.module.css";

interface StatusElementTreeProps {
	nodes: ElementTreeNode[];
	onEdit: (element: ApiPublicStatusElement) => void;
	onDelete: (element: ApiPublicStatusElement) => void;
}

export function StatusElementTree({ nodes, onEdit, onDelete }: StatusElementTreeProps) {
	if (!nodes.length) {
		return <div className={styles.empty}>No elements have been added to this status page.</div>;
	}

	return (
		<div className={styles.tree}>
			{nodes.map(node => (
				<ElementNode key={node.id} node={node} onDelete={onDelete} onEdit={onEdit} />
			))}
		</div>
	);
}

interface ElementNodeProps {
	node: ElementTreeNode;
	onEdit: (element: ApiPublicStatusElement) => void;
	onDelete: (element: ApiPublicStatusElement) => void;
}

function ElementNode({ node, onEdit, onDelete }: ElementNodeProps) {
	const isFolder = node.kind === "folder";
	const title = node.title || node.checkName || (isFolder ? "Untitled folder" : "Assignment group");
	const scopeLabel =
		node.kind === "assignment_group" && node.assignmentSelectionMode === "selected_assignments"
			? `${node.assignmentIds.length} selected`
			: node.kind === "assignment_group" && node.checkName
				? `${checkTypeLabel(node.checkType)} / ${node.checkName}`
				: null;

	return (
		<div className={styles.node} data-kind={node.kind}>
			<div className={styles.nodeMain}>
				<div className={styles.nodeCopy}>
					<div className={styles.nodeTitle}>
						<strong>{title}</strong>
						<Badge tone={isFolder ? "accent" : "neutral"}>{isFolder ? "Folder" : "Assignment group"}</Badge>
					</div>
					<div className={styles.nodeMeta}>
						<span>Order {node.sortOrder}</span>
						{scopeLabel ? <span>{scopeLabel}</span> : null}
						{node.checkTarget ? <span>{node.checkTarget}</span> : null}
						<span>{chartModeLabel(node.chartMode)}</span>
						{node.chartRange ? <span>{chartRangeLabel(node.chartRange)}</span> : null}
					</div>
				</div>
				<div className={styles.nodeActions}>
					<Button type="button" variant="ghost" size="sm" onClick={() => onEdit(node)}>
						<PencilSimple aria-hidden="true" />
						Edit
					</Button>
					<Button type="button" variant="ghost" size="sm" onClick={() => onDelete(node)}>
						<Trash aria-hidden="true" />
						Delete
					</Button>
				</div>
			</div>
			{node.children.length ? (
				<div className={styles.children}>
					{node.children.map(child => (
						<ElementNode key={child.id} node={child} onDelete={onDelete} onEdit={onEdit} />
					))}
				</div>
			) : null}
		</div>
	);
}

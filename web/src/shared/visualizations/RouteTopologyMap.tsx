import { i18n } from "@/i18n";
import { useRef, useState, type CSSProperties } from "react";
import { useTranslation } from "react-i18next";
import styles from "./RouteTopologyMap.module.css";
import { TopologyDetailCard, TopologyEdgeLayer, TopologyLegend, TopologyNodeLayer } from "./RouteTopologyMapParts";

const insightT = i18n.getFixedT(null, "insight") as (key: string) => string;

export type RouteTopologyNodeKind = "probe" | "hop" | "destination" | "unknown";
type TopologySeverity = "normal" | "warning" | "critical";
type TopologyDetailPlacement = "above" | "below";
type TopologyDetailTone = TopologySeverity | "agent" | "destination";

export interface RouteTopologyNode {
	id: string;
	kind: RouteTopologyNodeKind;
	label: string;
	address?: string;
	hostname?: string;
	hopIndex?: number;
	seenCount: number;
	avgRttMs?: number;
	lossPercent?: number;
}

export interface RouteTopologyEdge {
	source: string;
	target: string;
	seenCount: number;
	avgRttMs?: number;
	lossPercent?: number;
}

export interface TopologyRouteNode {
	id: string;
	name: string;
	label: string;
	address?: string;
	hostname?: string;
	kind: RouteTopologyNodeKind;
	hopIndex?: number;
	hopLabel: string;
	seenCount: number;
	avgRttMs?: number;
	lossPercent?: number;
	x: number;
	y: number;
	severity: TopologySeverity;
}

export interface TopologyRouteEdge {
	source: string;
	target: string;
	sourceLabel: string;
	targetLabel: string;
	seenCount: number;
	avgRttMs?: number;
	lossPercent?: number;
	x1: number;
	y1: number;
	x2: number;
	y2: number;
	color: string;
	width: number;
	opacity: number;
}

interface TopologyRouteLayout {
	nodes: TopologyRouteNode[];
	edges: TopologyRouteEdge[];
	viewWidth: number;
	viewHeight: number;
	routeStartX: number;
	routeEndX: number;
}

type TopologyColumn = [number, RouteTopologyNode[]];

interface TopologyWeightedNeighbor {
	id: string;
	weight: number;
}

interface TopologyNeighborScore {
	value: number;
	dominantValue: number;
	weight: number;
	hasScore: boolean;
}

export interface TopologyHoverDetail {
	id: string;
	title: string;
	subtitle?: string;
	rows: Array<{ label: string; value: string }>;
	x: number;
	y: number;
	placement: TopologyDetailPlacement;
	tone: TopologyDetailTone;
}

type TopologyMapStyle = CSSProperties & {
	"--ns-topology-width"?: string;
	"--ns-topology-height"?: string;
};

const topologyColumnGap = 168;
const topologyRowGap = 76;
const topologyPaddingX = 72;
const topologyMinWidth = 760;
const topologyMinHeight = 280;
const topologyDetailInsetX = 132;
const topologyDetailEstimatedHeight = 224;
const topologyDetailGap = 16;

function formatMs(value: number | null | undefined) {
	if (typeof value !== "number" || !Number.isFinite(value)) {
		return "-";
	}

	return `${value >= 100 ? Math.round(value) : value.toFixed(1)}ms`;
}

function formatPercent(value: number | null | undefined) {
	if (typeof value !== "number" || !Number.isFinite(value)) {
		return "-";
	}

	return `${value >= 10 ? Math.round(value) : value.toFixed(1)}%`;
}

function formatCount(value: number | null | undefined) {
	if (typeof value !== "number" || !Number.isFinite(value)) {
		return "-";
	}

	return new Intl.NumberFormat().format(value);
}

function topologyTone(lossPercent: number | undefined, avgRttMs: number | undefined) {
	if (typeof lossPercent === "number" && lossPercent >= 1) {
		return "var(--ns-critical)";
	}

	if (typeof avgRttMs === "number" && avgRttMs >= 100) {
		return "var(--ns-warning)";
	}

	return "var(--ns-success)";
}

function topologySeverity(lossPercent: number | undefined, avgRttMs: number | undefined): TopologySeverity {
	if (typeof lossPercent === "number" && lossPercent >= 1) {
		return "critical";
	}

	if (typeof avgRttMs === "number" && avgRttMs >= 100) {
		return "warning";
	}

	return "normal";
}

function topologyKindRank(kind: RouteTopologyNodeKind) {
	if (kind === "probe") {
		return 0;
	}

	if (kind === "destination") {
		return 2;
	}

	return 1;
}

function topologySiblingOffset(index: number, total: number) {
	return index - (total - 1) / 2;
}

function compactTopologyLabel(value: string) {
	return value.length > 18 ? `${value.slice(0, 15)}...` : value;
}

function topologyHopLabel(node: Pick<RouteTopologyNode, "kind" | "hopIndex">) {
	if (node.kind === "probe") {
		return "agent";
	}

	if (typeof node.hopIndex === "number") {
		return `hop ${String(node.hopIndex).padStart(2, "0")}`;
	}

	return node.kind;
}

function topologyColumn(node: RouteTopologyNode, maxHop: number) {
	if (node.kind === "probe") {
		return 0;
	}

	if (typeof node.hopIndex === "number") {
		return node.hopIndex;
	}

	if (node.kind === "destination") {
		return maxHop + 1;
	}

	return maxHop + 2;
}

function topologyBaseNodeCompare(a: RouteTopologyNode, b: RouteTopologyNode) {
	return topologyKindRank(a.kind) - topologyKindRank(b.kind) || b.seenCount - a.seenCount || a.label.localeCompare(b.label);
}

function topologyOrderIndex(columns: TopologyColumn[]) {
	const order = new Map<string, number>();

	for (const [, siblings] of columns) {
		siblings.forEach((node, index) => order.set(node.id, index));
	}

	return order;
}

function topologyNeighborScore(id: string, neighbors: Map<string, TopologyWeightedNeighbor[]>, order: Map<string, number>): TopologyNeighborScore {
	const values = neighbors.get(id) ?? [];
	let weightedSum = 0;
	let totalWeight = 0;
	let dominantWeight = 0;
	let dominantValue = 0;

	for (const neighbor of values) {
		const neighborOrder = order.get(neighbor.id);

		if (typeof neighborOrder !== "number") {
			continue;
		}

		const weight = Math.max(1, neighbor.weight);
		weightedSum += neighborOrder * weight;
		totalWeight += weight;

		if (weight > dominantWeight) {
			dominantWeight = weight;
			dominantValue = neighborOrder;
		}
	}

	if (totalWeight === 0) {
		return { value: 0, dominantValue: 0, weight: 0, hasScore: false };
	}

	return {
		value: weightedSum / totalWeight,
		dominantValue,
		weight: totalWeight,
		hasScore: true
	};
}

function topologyReorderSiblings(siblings: RouteTopologyNode[], neighbors: Map<string, TopologyWeightedNeighbor[]>, order: Map<string, number>) {
	const currentIndex = new Map(siblings.map((node, index) => [node.id, index]));

	return [...siblings].sort((a, b) => {
		const aScore = topologyNeighborScore(a.id, neighbors, order);
		const bScore = topologyNeighborScore(b.id, neighbors, order);
		const aValue = aScore.hasScore ? aScore.value : (currentIndex.get(a.id) ?? 0);
		const bValue = bScore.hasScore ? bScore.value : (currentIndex.get(b.id) ?? 0);
		const scoreDelta = aValue - bValue;

		if (Math.abs(scoreDelta) > 0.0001) {
			return scoreDelta;
		}

		const dominantDelta = aScore.dominantValue - bScore.dominantValue;
		if (aScore.hasScore && bScore.hasScore && Math.abs(dominantDelta) > 0.0001) {
			return dominantDelta;
		}

		const weightDelta = bScore.weight - aScore.weight;
		if (Math.abs(weightDelta) > 0.0001) {
			return weightDelta;
		}

		return (currentIndex.get(a.id) ?? 0) - (currentIndex.get(b.id) ?? 0) || topologyBaseNodeCompare(a, b);
	});
}

function topologyOrderedColumns(nodeColumns: Map<number, RouteTopologyNode[]>, edges: RouteTopologyEdge[]): TopologyColumn[] {
	const columns: TopologyColumn[] = [...nodeColumns.entries()].sort(([a], [b]) => a - b).map(([column, siblings]) => [column, [...siblings].sort(topologyBaseNodeCompare)]);
	const knownNodeIds = new Set(columns.flatMap(([, siblings]) => siblings.map(node => node.id)));
	const incoming = new Map<string, TopologyWeightedNeighbor[]>();
	const outgoing = new Map<string, TopologyWeightedNeighbor[]>();

	for (const edge of edges) {
		if (!knownNodeIds.has(edge.source) || !knownNodeIds.has(edge.target)) {
			continue;
		}

		incoming.set(edge.target, [...(incoming.get(edge.target) ?? []), { id: edge.source, weight: edge.seenCount }]);
		outgoing.set(edge.source, [...(outgoing.get(edge.source) ?? []), { id: edge.target, weight: edge.seenCount }]);
	}

	let order = topologyOrderIndex(columns);
	for (let pass = 0; pass < 8; pass++) {
		for (let index = 1; index < columns.length; index++) {
			columns[index] = [columns[index][0], topologyReorderSiblings(columns[index][1], incoming, order)];
			order = topologyOrderIndex(columns);
		}

		for (let index = columns.length - 2; index >= 0; index--) {
			columns[index] = [columns[index][0], topologyReorderSiblings(columns[index][1], outgoing, order)];
			order = topologyOrderIndex(columns);
		}
	}

	return columns;
}

function topologyRouteLayout(nodes: RouteTopologyNode[], edges: RouteTopologyEdge[]): TopologyRouteLayout {
	const maxHop = Math.max(0, ...nodes.map(node => node.hopIndex ?? 0));
	const sortedNodes = [...nodes].sort((a, b) => topologyColumn(a, maxHop) - topologyColumn(b, maxHop) || b.seenCount - a.seenCount || a.label.localeCompare(b.label));
	const nodeColumns = new Map<number, RouteTopologyNode[]>();
	const maxSeen = Math.max(1, ...nodes.map(node => node.seenCount), ...edges.map(edge => edge.seenCount));

	for (const node of sortedNodes) {
		const column = topologyColumn(node, maxHop);
		nodeColumns.set(column, [...(nodeColumns.get(column) ?? []), node]);
	}

	const columns = topologyOrderedColumns(nodeColumns, edges);
	const maxColumn = Math.max(1, ...columns.map(([column]) => column));
	const maxOffset = Math.max(0, ...columns.map(([, siblings]) => (siblings.length - 1) / 2));
	const viewWidth = Math.max(topologyMinWidth, topologyPaddingX * 2 + maxColumn * topologyColumnGap);
	const viewHeight = Math.max(topologyMinHeight, 176 + maxOffset * topologyRowGap * 2);
	const centerY = viewHeight / 2 - 14;
	const routeNodes: TopologyRouteNode[] = columns
		.sort(([a], [b]) => a - b)
		.flatMap(([column, siblings]) =>
			siblings.map((node, index) => {
				const yOffset = topologySiblingOffset(index, siblings.length);

				return {
					id: node.id,
					name: compactTopologyLabel(node.label),
					label: node.label,
					address: node.address,
					hostname: node.hostname,
					kind: node.kind,
					hopIndex: node.hopIndex,
					hopLabel: topologyHopLabel(node),
					seenCount: node.seenCount,
					avgRttMs: node.avgRttMs,
					lossPercent: node.lossPercent,
					x: topologyPaddingX + column * topologyColumnGap,
					y: centerY + yOffset * topologyRowGap,
					severity: topologySeverity(node.lossPercent, node.avgRttMs)
				};
			})
		);
	const knownNodeIds = new Set(routeNodes.map(node => node.id));
	const routeNodeById = new Map(routeNodes.map(node => [node.id, node]));
	const routeEdges: TopologyRouteEdge[] = edges
		.filter(edge => knownNodeIds.has(edge.source) && knownNodeIds.has(edge.target))
		.map(edge => {
			const sourceNode = routeNodeById.get(edge.source);
			const targetNode = routeNodeById.get(edge.target);
			const seenRatio = edge.seenCount / maxSeen;

			return {
				source: edge.source,
				target: edge.target,
				sourceLabel: sourceNode?.label ?? edge.source,
				targetLabel: targetNode?.label ?? edge.target,
				seenCount: edge.seenCount,
				avgRttMs: edge.avgRttMs,
				lossPercent: edge.lossPercent,
				x1: sourceNode?.x ?? 0,
				y1: sourceNode?.y ?? 0,
				x2: targetNode?.x ?? 0,
				y2: targetNode?.y ?? 0,
				color: topologyTone(edge.lossPercent, edge.avgRttMs),
				width: 1.5 + Math.min(2.5, seenRatio * 2.5),
				opacity: 0.34 + Math.min(0.48, seenRatio)
			};
		});
	const routeStartX = routeNodes.length ? Math.min(...routeNodes.map(node => node.x)) : 0;
	const routeEndX = routeNodes.length ? Math.max(...routeNodes.map(node => node.x)) : 0;

	return { nodes: routeNodes, edges: routeEdges, viewWidth, viewHeight, routeStartX, routeEndX };
}

function topologyNodeTitle(node: TopologyRouteNode) {
	const primaryName = node.hostname || node.label;
	const secondaryName = node.hostname && node.label !== node.hostname ? node.label : null;

	return [
		primaryName,
		secondaryName,
		node.address,
		node.hopLabel,
		`${insightT("map.seen")}: ${formatCount(node.seenCount)}`,
		`${insightT("map.avgRtt")}: ${formatMs(node.avgRttMs)}`,
		`${insightT("map.loss")}: ${formatPercent(node.lossPercent)}`
	]
		.filter(Boolean)
		.join("\n");
}

function topologyEdgeTitle(edge: TopologyRouteEdge) {
	return [
		`${edge.sourceLabel} -> ${edge.targetLabel}`,
		`${insightT("map.seen")}: ${formatCount(edge.seenCount)}`,
		`${insightT("map.avgRtt")}: ${formatMs(edge.avgRttMs)}`,
		`${insightT("map.loss")}: ${formatPercent(edge.lossPercent)}`
	].join("\n");
}

function topologyAriaLabel(value: string) {
	return value.replace(/\s*\n\s*/g, ", ");
}

function clampNumber(value: number, min: number, max: number) {
	return Math.min(Math.max(value, min), max);
}

function topologyDetailPosition(x: number, y: number, layout: TopologyRouteLayout): Pick<TopologyHoverDetail, "x" | "y" | "placement"> {
	const maxX = Math.max(topologyDetailInsetX, layout.viewWidth - topologyDetailInsetX);
	const hasRoomBelow = y + topologyDetailGap + topologyDetailEstimatedHeight <= layout.viewHeight;
	const placement: TopologyDetailPlacement = hasRoomBelow ? "below" : "above";

	return {
		x: clampNumber(x, topologyDetailInsetX, maxX),
		y,
		placement
	};
}

function topologyNodeDetailTone(node: TopologyRouteNode): TopologyDetailTone {
	if (node.kind === "probe") {
		return "agent";
	}
	return node.severity;
}

function topologyNodeDetail(node: TopologyRouteNode, layout: TopologyRouteLayout): TopologyHoverDetail {
	const position = topologyDetailPosition(node.x, node.y, layout);
	const title = node.hostname || node.label;
	const subtitle = node.hostname && node.label !== node.hostname ? node.label : undefined;

	return {
		id: `node:${node.id}`,
		title,
		subtitle,
		x: position.x,
		y: position.y,
		placement: position.placement,
		tone: topologyNodeDetailTone(node),
		rows: [
			...(node.address ? [{ label: insightT("map.address"), value: node.address }] : []),
			{ label: insightT("map.seen"), value: formatCount(node.seenCount) },
			{ label: insightT("map.avgRtt"), value: formatMs(node.avgRttMs) },
			{ label: insightT("map.loss"), value: formatPercent(node.lossPercent) }
		]
	};
}

function topologyEdgeDetail(edge: TopologyRouteEdge, layout: TopologyRouteLayout): TopologyHoverDetail {
	const position = topologyDetailPosition((edge.x1 + edge.x2) / 2, (edge.y1 + edge.y2) / 2, layout);

	return {
		id: `edge:${edge.source}->${edge.target}`,
		title: `${edge.sourceLabel} -> ${edge.targetLabel}`,
		x: position.x,
		y: position.y,
		placement: position.placement,
		tone: topologySeverity(edge.lossPercent, edge.avgRttMs),
		rows: [
			{ label: insightT("map.seen"), value: formatCount(edge.seenCount) },
			{ label: insightT("map.avgRtt"), value: formatMs(edge.avgRttMs) },
			{ label: insightT("map.loss"), value: formatPercent(edge.lossPercent) }
		]
	};
}

export function RouteTopologyMap({ nodes, edges }: { nodes: RouteTopologyNode[]; edges: RouteTopologyEdge[] }) {
	const { t } = useTranslation("insight");
	const shellRef = useRef<HTMLDivElement>(null);
	const viewportRef = useRef<HTMLDivElement>(null);
	const [activeDetail, setActiveDetail] = useState<TopologyHoverDetail | null>(null);
	const layout = topologyRouteLayout(nodes, edges);
	const style: TopologyMapStyle = {
		"--ns-topology-width": `${layout.viewWidth}px`,
		"--ns-topology-height": `${layout.viewHeight}px`
	};
	const clearActiveDetail = () => setActiveDetail(null);
	const showActiveDetail = (detail: TopologyHoverDetail) => {
		const shell = shellRef.current;
		const viewport = viewportRef.current;

		if (!shell || !viewport) {
			setActiveDetail(detail);
			return;
		}

		const viewportRect = viewport.getBoundingClientRect();
		const viewportHeight = window.innerHeight || document.documentElement.clientHeight;
		const screenY = viewportRect.top + detail.y;
		const hasRoomAbove = screenY - topologyDetailGap - topologyDetailEstimatedHeight >= 0;
		const hasRoomBelow = screenY + topologyDetailGap + topologyDetailEstimatedHeight <= viewportHeight;
		const placement = hasRoomAbove && (!hasRoomBelow || screenY > viewportHeight / 2) ? "above" : "below";
		const maxVisibleX = Math.max(topologyDetailInsetX, viewport.clientWidth - topologyDetailInsetX);
		const visibleX = clampNumber(detail.x - viewport.scrollLeft, topologyDetailInsetX, maxVisibleX);

		setActiveDetail({
			...detail,
			x: viewport.offsetLeft + visibleX,
			y: viewport.offsetTop + detail.y,
			placement
		});
	};

	return (
		<div className={styles.topologyShell} ref={shellRef}>
			<TopologyLegend />
			<div className={styles.topologyViewport} ref={viewportRef} onScroll={clearActiveDetail}>
				<div className={styles.topologyMap} style={style}>
					<svg className={styles.topologySvg} viewBox={`0 0 ${layout.viewWidth} ${layout.viewHeight}`} role="img" aria-label={t("map.aggregatedTopology")}>
						<TopologyEdgeLayer
							edges={layout.edges}
							activeDetail={activeDetail}
							onClearDetail={clearActiveDetail}
							onShowDetail={edge => showActiveDetail(topologyEdgeDetail(edge, layout))}
							getEdgeTitle={topologyEdgeTitle}
							ariaLabel={topologyAriaLabel}
						/>
					</svg>
					<TopologyNodeLayer
						nodes={layout.nodes}
						activeDetail={activeDetail}
						onClearDetail={clearActiveDetail}
						onShowDetail={node => showActiveDetail(topologyNodeDetail(node, layout))}
						getNodeTitle={topologyNodeTitle}
						ariaLabel={topologyAriaLabel}
					/>
				</div>
			</div>
			{activeDetail ? <TopologyDetailCard detail={activeDetail} /> : null}
		</div>
	);
}

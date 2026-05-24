import { classNames } from "@/shared/utils/classNames";
import { useRef, useState, type CSSProperties } from "react";
import styles from "./RouteTopologyMap.module.css";

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

interface TopologyRouteNode {
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

interface TopologyRouteEdge {
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

interface TopologyHoverDetail {
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

type TopologyNodeStyle = CSSProperties & {
	"--ns-topology-node-x"?: string;
	"--ns-topology-node-y"?: string;
};

type TopologyDetailStyle = CSSProperties & {
	"--ns-topology-detail-x"?: string;
	"--ns-topology-detail-y"?: string;
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
		return "#ff453a";
	}

	if (typeof avgRttMs === "number" && avgRttMs >= 100) {
		return "#ff9f0a";
	}

	return "#ff7a1a";
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

function topologySiblingOffset(index: number) {
	if (index === 0) {
		return 0;
	}

	const distance = Math.ceil(index / 2);
	return index % 2 === 1 ? -distance : distance;
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

function topologyRouteLayout(nodes: RouteTopologyNode[], edges: RouteTopologyEdge[]): TopologyRouteLayout {
	const maxHop = Math.max(0, ...nodes.map(node => node.hopIndex ?? 0));
	const sortedNodes = [...nodes].sort((a, b) => topologyColumn(a, maxHop) - topologyColumn(b, maxHop) || b.seenCount - a.seenCount || a.label.localeCompare(b.label));
	const nodeColumns = new Map<number, RouteTopologyNode[]>();
	const maxSeen = Math.max(1, ...nodes.map(node => node.seenCount), ...edges.map(edge => edge.seenCount));

	for (const node of sortedNodes) {
		const column = topologyColumn(node, maxHop);
		nodeColumns.set(column, [...(nodeColumns.get(column) ?? []), node]);
	}

	for (const [column, siblings] of nodeColumns) {
		nodeColumns.set(
			column,
			[...siblings].sort((a, b) => topologyKindRank(a.kind) - topologyKindRank(b.kind) || b.seenCount - a.seenCount || a.label.localeCompare(b.label))
		);
	}

	const columns = [...nodeColumns.entries()].sort(([a], [b]) => a - b);
	const maxColumn = Math.max(1, ...columns.map(([column]) => column));
	const maxOffset = Math.max(0, ...columns.map(([, siblings]) => Math.ceil((siblings.length - 1) / 2)));
	const viewWidth = Math.max(topologyMinWidth, topologyPaddingX * 2 + maxColumn * topologyColumnGap);
	const viewHeight = Math.max(topologyMinHeight, 176 + maxOffset * topologyRowGap * 2);
	const centerY = viewHeight / 2 - 14;
	const routeNodes: TopologyRouteNode[] = columns
		.sort(([a], [b]) => a - b)
		.flatMap(([column, siblings]) =>
			siblings.map((node, index) => {
				const yOffset = topologySiblingOffset(index);

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

	return [primaryName, secondaryName, node.address, node.hopLabel, `seen ${formatCount(node.seenCount)}`, `avg ${formatMs(node.avgRttMs)}`, `loss ${formatPercent(node.lossPercent)}`]
		.filter(Boolean)
		.join("\n");
}

function topologyEdgeTitle(edge: TopologyRouteEdge) {
	return [`${edge.sourceLabel} -> ${edge.targetLabel}`, `seen ${formatCount(edge.seenCount)}`, `avg ${formatMs(edge.avgRttMs)}`, `loss ${formatPercent(edge.lossPercent)}`].join("\n");
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
	if (node.kind === "destination") {
		return "destination";
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
			...(node.address ? [{ label: "address", value: node.address }] : []),
			{ label: "seen", value: formatCount(node.seenCount) },
			{ label: "avg rtt", value: formatMs(node.avgRttMs) },
			{ label: "loss", value: formatPercent(node.lossPercent) }
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
			{ label: "seen", value: formatCount(edge.seenCount) },
			{ label: "avg rtt", value: formatMs(edge.avgRttMs) },
			{ label: "loss", value: formatPercent(edge.lossPercent) }
		]
	};
}

function TopologyDetailCard({ detail }: { detail: TopologyHoverDetail }) {
	const style: TopologyDetailStyle = {
		"--ns-topology-detail-x": `${detail.x}px`,
		"--ns-topology-detail-y": `${detail.y}px`
	};

	return (
		<div className={classNames(styles.topologyDetail, styles[`topologyDetail${detail.tone}`])} style={style} data-placement={detail.placement} id="topology-detail-card">
			<strong>{detail.title}</strong>
			{detail.subtitle ? <span className={styles.topologyDetailSubtitle}>{detail.subtitle}</span> : null}
			<dl>
				{detail.rows.map(row => (
					<div key={`${detail.id}:${row.label}`}>
						<dt>{row.label}</dt>
						<dd>{row.value}</dd>
					</div>
				))}
			</dl>
		</div>
	);
}

export function RouteTopologyMap({ nodes, edges }: { nodes: RouteTopologyNode[]; edges: RouteTopologyEdge[] }) {
	const shellRef = useRef<HTMLDivElement>(null);
	const viewportRef = useRef<HTMLDivElement>(null);
	const [activeDetail, setActiveDetail] = useState<TopologyHoverDetail | null>(null);
	const layout = topologyRouteLayout(nodes, edges);
	const centerY = layout.viewHeight / 2 - 14;
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
			<div className={styles.topologyLegend} aria-hidden="true">
				<span data-tone="agent">agent</span>
				<span data-tone="normal">normal</span>
				<span data-tone="warning">high rtt</span>
				<span data-tone="critical">loss</span>
				<span data-tone="destination">destination</span>
			</div>
			<div className={styles.topologyViewport} ref={viewportRef} onScroll={clearActiveDetail}>
				<div className={styles.topologyMap} style={style}>
					<svg className={styles.topologySvg} viewBox={`0 0 ${layout.viewWidth} ${layout.viewHeight}`} role="img" aria-label="Aggregated route topology">
						<line className={styles.topologyCenterLine} x1={layout.routeStartX} x2={layout.routeEndX} y1={centerY} y2={centerY} />
						{layout.edges.map(edge => {
							const edgeTitle = topologyEdgeTitle(edge);

							return (
								<g key={`${edge.source}->${edge.target}`}>
									<line className={styles.topologyEdge} x1={edge.x1} x2={edge.x2} y1={edge.y1} y2={edge.y2} stroke={edge.color} strokeWidth={edge.width} opacity={edge.opacity} />
									<line
										aria-label={topologyAriaLabel(edgeTitle)}
										aria-describedby={activeDetail?.id === `edge:${edge.source}->${edge.target}` ? "topology-detail-card" : undefined}
										className={styles.topologyEdgeHit}
										role="graphics-symbol"
										tabIndex={0}
										x1={edge.x1}
										x2={edge.x2}
										y1={edge.y1}
										y2={edge.y2}
										onBlur={clearActiveDetail}
										onFocus={() => showActiveDetail(topologyEdgeDetail(edge, layout))}
										onPointerEnter={() => showActiveDetail(topologyEdgeDetail(edge, layout))}
										onPointerLeave={clearActiveDetail}
									/>
								</g>
							);
						})}
					</svg>
					{layout.nodes.map(node => {
						const nodeStyle: TopologyNodeStyle = {
							"--ns-topology-node-x": `${node.x}px`,
							"--ns-topology-node-y": `${node.y}px`
						};
						const nodeTitle = topologyNodeTitle(node);

						return (
							<div
								className={classNames(styles.topologyNode, styles[`topologyNode${node.kind}`], styles[`topologyNode${node.severity}`])}
								style={nodeStyle}
								tabIndex={0}
								aria-describedby={activeDetail?.id === `node:${node.id}` ? "topology-detail-card" : undefined}
								aria-label={topologyAriaLabel(nodeTitle)}
								key={node.id}
								onBlur={clearActiveDetail}
								onFocus={() => showActiveDetail(topologyNodeDetail(node, layout))}
								onPointerEnter={() => showActiveDetail(topologyNodeDetail(node, layout))}
								onPointerLeave={clearActiveDetail}
							>
								<span className={styles.topologyNodeDot} />
								<span className={styles.topologyNodeLabel}>
									<strong>{node.name}</strong>
									<span>{node.hopLabel}</span>
								</span>
							</div>
						);
					})}
				</div>
			</div>
			{activeDetail ? <TopologyDetailCard detail={activeDetail} /> : null}
		</div>
	);
}

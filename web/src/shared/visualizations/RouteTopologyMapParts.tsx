import { classNames } from "@/shared/utils/classNames";
import type { CSSProperties } from "react";
import { useTranslation } from "react-i18next";
import type { TopologyHoverDetail, TopologyRouteEdge, TopologyRouteNode } from "./RouteTopologyMap";
import styles from "./RouteTopologyMap.module.css";

type TopologyNodeStyle = CSSProperties & {
	"--ns-topology-node-x"?: string;
	"--ns-topology-node-y"?: string;
};

type TopologyDetailStyle = CSSProperties & {
	"--ns-topology-detail-x"?: string;
	"--ns-topology-detail-y"?: string;
};

export function TopologyLegend() {
	const { t } = useTranslation("insight");
	return (
		<div className={styles.topologyLegend} aria-hidden="true">
			<span data-tone="agent">{t("legend.agent")}</span>
			<span data-tone="normal">{t("legend.normal")}</span>
			<span data-tone="warning">{t("legend.highRtt")}</span>
			<span data-tone="critical">{t("legend.loss")}</span>
			<span data-tone="destination">{t("legend.destination")}</span>
		</div>
	);
}

export function TopologyEdgeLayer({
	edges,
	activeDetail,
	onShowDetail,
	onClearDetail,
	getEdgeTitle,
	ariaLabel
}: {
	edges: TopologyRouteEdge[];
	activeDetail: TopologyHoverDetail | null;
	onShowDetail: (edge: TopologyRouteEdge) => void;
	onClearDetail: () => void;
	getEdgeTitle: (edge: TopologyRouteEdge) => string;
	ariaLabel: (value: string) => string;
}) {
	return (
		<>
			{edges.map(edge => {
				const edgeTitle = getEdgeTitle(edge);

				return (
					<g key={`${edge.source}->${edge.target}`}>
						<line className={styles.topologyEdge} x1={edge.x1} x2={edge.x2} y1={edge.y1} y2={edge.y2} stroke={edge.color} strokeWidth={edge.width} opacity={edge.opacity} />
						<line
							aria-label={ariaLabel(edgeTitle)}
							aria-describedby={activeDetail?.id === `edge:${edge.source}->${edge.target}` ? "topology-detail-card" : undefined}
							className={styles.topologyEdgeHit}
							role="graphics-symbol"
							tabIndex={0}
							x1={edge.x1}
							x2={edge.x2}
							y1={edge.y1}
							y2={edge.y2}
							onBlur={onClearDetail}
							onFocus={() => onShowDetail(edge)}
							onPointerEnter={() => onShowDetail(edge)}
							onPointerLeave={onClearDetail}
						/>
					</g>
				);
			})}
		</>
	);
}

export function TopologyNodeLayer({
	nodes,
	activeDetail,
	onShowDetail,
	onClearDetail,
	getNodeTitle,
	ariaLabel
}: {
	nodes: TopologyRouteNode[];
	activeDetail: TopologyHoverDetail | null;
	onShowDetail: (node: TopologyRouteNode) => void;
	onClearDetail: () => void;
	getNodeTitle: (node: TopologyRouteNode) => string;
	ariaLabel: (value: string) => string;
}) {
	return (
		<>
			{nodes.map(node => {
				const nodeStyle: TopologyNodeStyle = {
					"--ns-topology-node-x": `${node.x}px`,
					"--ns-topology-node-y": `${node.y}px`
				};
				const nodeTitle = getNodeTitle(node);

				return (
					<div
						className={classNames(styles.topologyNode, styles[`topologyNode${node.kind}`], styles[`topologyNode${node.severity}`])}
						style={nodeStyle}
						tabIndex={0}
						aria-describedby={activeDetail?.id === `node:${node.id}` ? "topology-detail-card" : undefined}
						aria-label={ariaLabel(nodeTitle)}
						key={node.id}
						onBlur={onClearDetail}
						onFocus={() => onShowDetail(node)}
						onPointerEnter={() => onShowDetail(node)}
						onPointerLeave={onClearDetail}
					>
						<span className={styles.topologyNodeDot} />
						<span className={styles.topologyNodeLabel}>
							<strong>{node.name}</strong>
							<span>{node.hopLabel}</span>
						</span>
					</div>
				);
			})}
		</>
	);
}

export function TopologyDetailCard({ detail }: { detail: TopologyHoverDetail }) {
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

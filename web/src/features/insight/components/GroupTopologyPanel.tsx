import { RouteTopologyMap, type RouteTopologyEdge, type RouteTopologyNode } from "@/shared/visualizations/RouteTopologyMap";
import { BodyCopy, Panel, Spinner } from "@netstamp/ui";
import { useTranslation } from "react-i18next";

interface GroupTopologyPanelProps {
	title: string;
	nodes: RouteTopologyNode[];
	edges: RouteTopologyEdge[];
	isLoading: boolean;
}

export function GroupTopologyPanel({ title, nodes, edges, isLoading }: GroupTopologyPanelProps) {
	const { t } = useTranslation("insight");
	const hasTopology = nodes.length > 0 && edges.length > 0;

	return (
		<Panel tone="deep" title={title}>
			{hasTopology ? (
				<RouteTopologyMap nodes={nodes} edges={edges} />
			) : isLoading ? (
				<Spinner label={t("panel.loadingRouteGraph")} layout="panel" size="lg" />
			) : (
				<BodyCopy>{t("panel.noTopology")}</BodyCopy>
			)}
		</Panel>
	);
}

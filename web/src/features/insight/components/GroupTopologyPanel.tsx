import { BodyCopy } from "@/shared/components/BodyCopy";
import { RouteTopologyMap, type RouteTopologyEdge, type RouteTopologyNode } from "@/shared/visualizations/RouteTopologyMap";
import { Panel } from "@netstamp/ui";

interface GroupTopologyPanelProps {
	title: string;
	nodes: RouteTopologyNode[];
	edges: RouteTopologyEdge[];
	isLoading: boolean;
}

export function GroupTopologyPanel({ title, nodes, edges, isLoading }: GroupTopologyPanelProps) {
	const hasTopology = nodes.length > 0 && edges.length > 0;

	return (
		<Panel tone="deep" eyebrow="Traceroute topology" title={title}>
			{hasTopology ? (
				<RouteTopologyMap nodes={nodes} edges={edges} />
			) : (
				<BodyCopy>{isLoading ? "Loading aggregated route graph." : "No traceroute topology is available for the selected scope and time range."}</BodyCopy>
			)}
		</Panel>
	);
}

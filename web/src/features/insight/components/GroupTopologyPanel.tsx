import { RouteTopologyMap, type RouteTopologyEdge, type RouteTopologyNode } from "@/shared/visualizations/RouteTopologyMap";
import { BodyCopy, Panel, Spinner } from "@netstamp/ui";

interface GroupTopologyPanelProps {
	title: string;
	nodes: RouteTopologyNode[];
	edges: RouteTopologyEdge[];
	isLoading: boolean;
}

export function GroupTopologyPanel({ title, nodes, edges, isLoading }: GroupTopologyPanelProps) {
	const hasTopology = nodes.length > 0 && edges.length > 0;

	return (
		<Panel tone="deep" title={title}>
			{hasTopology ? (
				<RouteTopologyMap nodes={nodes} edges={edges} />
			) : isLoading ? (
				<Spinner label="Loading route graph" layout="panel" size="lg" />
			) : (
				<BodyCopy>No traceroute topology is available for the selected scope and time range.</BodyCopy>
			)}
		</Panel>
	);
}

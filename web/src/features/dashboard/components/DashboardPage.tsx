import { mapApiChecks } from "@/features/checks/api/checkAdapters";
import { mapApiProbes } from "@/features/probes/api/probeAdapters";
import { type Navigate } from "@/routes/routeTypes";
import { projectQueries } from "@/shared/api/queries";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { BodyCopy } from "@/shared/components/BodyCopy";
import { FleetMatrix } from "@/shared/components/FleetMatrix";
import { NetworkMap } from "@/shared/components/NetworkMap";
import { PageStack } from "@/shared/components/PageStack";
import { ResponsiveGrid } from "@/shared/components/ResponsiveGrid";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { Button, MetricCard, Panel } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import styles from "./DashboardPage.module.css";

interface DashboardPageProps {
	navigate: Navigate;
}

export function DashboardPage({ navigate }: DashboardPageProps) {
	const { projectRef } = useCurrentProject();
	const [selectedProbeId, setSelectedProbeId] = useState("");
	const probesQuery = useQuery({
		...projectQueries.probes(projectRef || ""),
		enabled: Boolean(projectRef),
		select: data => mapApiProbes(data.probes)
	});
	const probes = probesQuery.data ?? [];
	const checksQuery = useQuery({
		...projectQueries.checks(projectRef || ""),
		enabled: Boolean(projectRef),
		select: data => mapApiChecks(data.checks, probes)
	});
	const checks = checksQuery.data ?? [];
	const onlineProbes = probes.filter(probe => probe.status === "Online").length;
	const activeChecks = checks.length;
	const positionedProbes = probes.filter(probe => probe.coordinates);
	const activeProbeId = positionedProbes.some(probe => probe.id === selectedProbeId) ? selectedProbeId : (positionedProbes[0]?.id ?? "");
	const mapTitle = projectRef ? `${positionedProbes.length} positioned / ${probes.length} probes` : "No project selected";
	const mapCopy = projectRef
		? positionedProbes.length > 0
			? "Live probe geography for the selected project. Markers appear after a probe has stored coordinates."
			: probes.length > 0
				? "Probe records exist, but none have coordinates yet. Add locations from the probe detail page to place them on the map."
				: "Create a probe with coordinates to populate the fleet map."
		: "Select a project to view its probe fleet on the world map.";

	return (
		<PageStack>
			<ScreenHeader
				title="Dashboard"
				actions={
					<>
						<Button variant="secondary" onClick={() => navigate("newProbe")}>
							New probe wizard
						</Button>
						<Button onClick={() => navigate("checks")}>Create check</Button>
					</>
				}
			/>

			<ResponsiveGrid>
				<MetricCard label="Probes Online" value={`${onlineProbes}/${probes.length}`} detail="fleet" tone="success" />
				<MetricCard label="Active Checks" value={String(activeChecks)} detail="scheduled" tone="accent" />
			</ResponsiveGrid>

			<Panel className={styles.mapPanel} tone="deep" eyebrow="Probe geography" title={mapTitle}>
				<BodyCopy className={styles.mapCopy}>{mapCopy}</BodyCopy>
				<NetworkMap probes={probes} selectedId={activeProbeId} onSelect={setSelectedProbeId} mode="fleet" className={styles.worldMap} />
			</Panel>

			<Panel tone="glass" eyebrow="Fleet bitmap" title={`${probes.length} probes, ${onlineProbes} lit`}>
				<FleetMatrix total={Math.max(probes.length, 1)} online={onlineProbes} />
			</Panel>
		</PageStack>
	);
}

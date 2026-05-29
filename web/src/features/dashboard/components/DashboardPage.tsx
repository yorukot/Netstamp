import { mapApiChecks } from "@/features/checks/api/checkAdapters";
import { mapApiProbes } from "@/features/probes/api/probeAdapters";
import { type Navigate } from "@/routes/routeTypes";
import { projectQueries } from "@/shared/api/queries";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { NetworkMap } from "@/shared/components/NetworkMap";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { Button, MetricCard } from "@netstamp/ui";
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

	return (
		<PageStack className={styles.dashboard}>
			<ScreenHeader title="Dashboard" actions={<Button onClick={() => navigate("newProbe")}>New Probe</Button>} />

			<div className={styles.metrics}>
				<MetricCard className={styles.metric} label="Probes Online" value={`${onlineProbes}/${probes.length}`} />
				<MetricCard className={styles.metric} label="Active Checks" value={String(activeChecks)} />
			</div>

			<NetworkMap probes={probes} selectedId={activeProbeId} onSelect={setSelectedProbeId} mode="fleet" className={styles.worldMap} />
		</PageStack>
	);
}

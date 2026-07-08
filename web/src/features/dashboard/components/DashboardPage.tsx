import { mapApiChecks } from "@/features/checks/api/checkAdapters";
import { mapApiProbes } from "@/features/probes/api/probeAdapters";
import { projectQueries } from "@/shared/api/queries";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { PageStack } from "@/shared/components/PageStack";
import { classNames } from "@/shared/utils/classNames";
import { NetworkMap } from "@/shared/visualizations/NetworkMap";
import { EmptyState, MetricTile, Panel } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import styles from "./DashboardPage.module.css";

const dashboardFleetFitPadding = { top: 44, right: 48, bottom: 56, left: 48 };
const dashboardFleetMaxZoom = 5.4;

function percentage(part: number, total: number) {
	if (!total) {
		return "0%";
	}

	return `${Math.round((part / total) * 100)}%`;
}

export function DashboardPage() {
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
	const offlineProbes = probes.filter(probe => probe.status === "Offline").length;
	const drainingProbes = probes.filter(probe => probe.status === "Draining").length;
	const activeChecks = checks.length;
	const positionedProbes = probes.filter(probe => probe.coordinates);
	const activeProbeId = positionedProbes.some(probe => probe.id === selectedProbeId) ? selectedProbeId : (positionedProbes[0]?.id ?? "");
	const activeProbe = probes.find(probe => probe.id === activeProbeId) ?? positionedProbes[0] ?? probes[0] ?? null;
	const visibleProbes = probes.slice(0, 6);
	const checkTypeSummary = Array.from(
		checks.reduce((summary, check) => {
			summary.set(check.type, (summary.get(check.type) ?? 0) + 1);
			return summary;
		}, new Map<string, number>())
	);
	const metrics = [
		{ label: "Probes Online", value: `${onlineProbes}/${probes.length}`, detail: "online", tone: "success", meta: `${offlineProbes} offline` },
		{ label: "Map Coverage", value: percentage(positionedProbes.length, probes.length), detail: "coverage", tone: "accent", meta: `${positionedProbes.length} located` },
		{ label: "Active Checks", value: String(activeChecks), detail: "checks", tone: "neutral", meta: checkTypeSummary.map(([type, count]) => `${count} ${type}`).join(" / ") || "none" },
		{ label: "Draining", value: String(drainingProbes), detail: "maintenance", tone: "warning", meta: "maintenance" }
	] as const;

	return (
		<PageStack className={styles.dashboard}>
			<header className={styles.dashboardHeader}>
				<div className={styles.titleBlock}>
					<h1>Dashboard</h1>
				</div>
			</header>

			<div className={styles.sections}>
				<Panel className={styles.overviewSection} title="Fleet" padded={false} bodyClassName={styles.metricsContent}>
					{metrics.map(metric => (
						<MetricTile className={styles.metricTile} key={metric.label} label={metric.label} value={metric.value} description={metric.meta} detail={metric.detail} tone={metric.tone} />
					))}
				</Panel>

				<Panel className={styles.mapSection} title="Network Map" padded={false} bodyClassName={styles.mapContent}>
					<NetworkMap
						probes={probes}
						selectedId={activeProbeId}
						onSelect={setSelectedProbeId}
						mode="fleet"
						theme="dark"
						fleetFitPadding={dashboardFleetFitPadding}
						fleetMaxZoom={dashboardFleetMaxZoom}
						isLoading={probesQuery.isPending}
						loadingLabel="Loading probes"
						className={styles.worldMap}
					/>
					<div className={styles.mapReadout}>
						<span>selected probe</span>
						<strong>{activeProbe?.name ?? "No probe"}</strong>
						<small>{activeProbe?.location ?? "coordinates unavailable"}</small>
					</div>
				</Panel>

				<Panel className={styles.registrySection} title="Probe Registry" padded={false} bodyClassName={styles.listContent}>
					{visibleProbes.length ? (
						<ul className={styles.entityList}>
							{visibleProbes.map(probe => (
								<li key={probe.id}>
									<div>
										<strong>{probe.name}</strong>
										<span>{probe.location}</span>
									</div>
									<span className={classNames(styles.status, probe.status === "Online" && styles.statusOnline, probe.status === "Offline" && styles.statusOffline)}>{probe.status}</span>
								</li>
							))}
						</ul>
					) : (
						<EmptyState title="No probes registered" description="Create a probe to start collecting fleet telemetry." />
					)}
				</Panel>
			</div>
		</PageStack>
	);
}

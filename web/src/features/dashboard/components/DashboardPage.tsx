import { mapApiChecks } from "@/features/checks/api/checkAdapters";
import { mapApiProbes } from "@/features/probes/api/probeAdapters";
import type { ProbeStatus } from "@/features/probes/data/probes";
import { projectQueries } from "@/shared/api/queries";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { PageStack } from "@/shared/components/PageStack";
import { NetworkMap } from "@/shared/visualizations/NetworkMap";
import { Badge, EmptyState, MetricTile, Panel, type BadgeTone } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import styles from "./DashboardPage.module.css";

const dashboardFleetFitPadding = { top: 44, right: 48, bottom: 56, left: 48 };
const dashboardFleetMaxZoom = 5.4;
const probeStatusTones: Record<ProbeStatus, BadgeTone> = {
	Online: "success",
	Draining: "warning",
	Offline: "critical"
};

function percentage(part: number, total: number) {
	if (!total) {
		return "0%";
	}

	return `${Math.round((part / total) * 100)}%`;
}

export function DashboardPage() {
	const { t } = useTranslation(["dashboard", "probes"]);
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
		{
			label: t("dashboard:metrics.probesOnline"),
			value: `${onlineProbes}/${probes.length}`,
			detail: t("dashboard:metrics.online"),
			tone: "success",
			meta: t("dashboard:metrics.offlineCount", { count: offlineProbes })
		},
		{
			label: t("dashboard:metrics.mapCoverage"),
			value: percentage(positionedProbes.length, probes.length),
			detail: t("dashboard:metrics.coverage"),
			tone: "accent",
			meta: t("dashboard:metrics.locatedCount", { count: positionedProbes.length })
		},
		{
			label: t("dashboard:metrics.activeChecks"),
			value: String(activeChecks),
			detail: t("dashboard:metrics.checks"),
			tone: "neutral",
			meta: checkTypeSummary.map(([type, count]) => `${count} ${type}`).join(" / ") || t("dashboard:metrics.none")
		},
		{ label: t("dashboard:metrics.draining"), value: String(drainingProbes), detail: t("dashboard:metrics.maintenance"), tone: "warning", meta: t("dashboard:metrics.maintenance") }
	] as const;
	const statusLabel = (status: ProbeStatus) => t(`probes:status.${status.toLowerCase() as "online" | "draining" | "offline"}`);

	return (
		<PageStack className={styles.dashboard}>
			<header className={styles.dashboardHeader}>
				<div className={styles.titleBlock}>
					<h1>{t("dashboard:title")}</h1>
				</div>
			</header>

			<div className={styles.sections}>
				<Panel className={styles.overviewSection} title={t("dashboard:fleet")} padded={false} bodyClassName={styles.metricsContent}>
					{metrics.map(metric => (
						<MetricTile className={styles.metricTile} key={metric.label} label={metric.label} value={metric.value} description={metric.meta} detail={metric.detail} tone={metric.tone} />
					))}
				</Panel>

				<Panel className={styles.mapSection} title={t("dashboard:networkMap")} padded={false} bodyClassName={styles.mapContent}>
					<NetworkMap
						probes={probes}
						selectedId={activeProbeId}
						onSelect={setSelectedProbeId}
						mode="fleet"
						theme="dark"
						fleetFitPadding={dashboardFleetFitPadding}
						fleetMaxZoom={dashboardFleetMaxZoom}
						isLoading={probesQuery.isPending}
						loadingLabel={t("probes:loading")}
						className={styles.worldMap}
					/>
					<div className={styles.mapReadout}>
						<span>{t("dashboard:selectedProbe")}</span>
						<strong>{activeProbe?.name ?? t("dashboard:noProbe")}</strong>
						<small>{activeProbe?.location ?? t("dashboard:coordinatesUnavailable")}</small>
					</div>
				</Panel>

				<Panel className={styles.registrySection} title={t("dashboard:probeRegistry")} padded={false} bodyClassName={styles.listContent}>
					{visibleProbes.length ? (
						<ul className={styles.entityList}>
							{visibleProbes.map(probe => (
								<li key={probe.id}>
									<div>
										<strong>{probe.name}</strong>
										<span className={styles.probeLocation}>{probe.location}</span>
									</div>
									<Badge tone={probeStatusTones[probe.status]}>{statusLabel(probe.status)}</Badge>
								</li>
							))}
						</ul>
					) : (
						<EmptyState title={t("dashboard:emptyTitle")} description={t("dashboard:emptyDescription")} />
					)}
				</Panel>
			</div>
		</PageStack>
	);
}

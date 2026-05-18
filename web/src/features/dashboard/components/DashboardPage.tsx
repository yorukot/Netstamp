import { type Navigate } from "@/routes/routeTypes";
import { mapApiProbes } from "@/features/probes/api/probeAdapters";
import { projectQueries, systemQueries } from "@/shared/api/queries";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { BodyCopy } from "@/shared/components/BodyCopy";
import { FleetMatrix } from "@/shared/components/FleetMatrix";
import { PageStack } from "@/shared/components/PageStack";
import { ResponsiveGrid } from "@/shared/components/ResponsiveGrid";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { Badge, Button, MetricCard, Panel, Surface, type BadgeTone } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import styles from "./DashboardPage.module.css";

interface DashboardPageProps {
	navigate: Navigate;
}

export function DashboardPage({ navigate }: DashboardPageProps) {
	const { projectRef } = useCurrentProject();
	const probesQuery = useQuery({
		...projectQueries.probes(projectRef || ""),
		enabled: Boolean(projectRef),
		select: data => mapApiProbes(data.probes)
	});
	const checksQuery = useQuery({
		...projectQueries.checks(projectRef || ""),
		enabled: Boolean(projectRef)
	});
	const rootQuery = useQuery(systemQueries.root());
	const healthQuery = useQuery(systemQueries.health());
	const probes = probesQuery.data ?? [];
	const onlineProbes = probes.filter(probe => probe.status === "Online").length;
	const activeChecks = checksQuery.data?.checks?.length ?? 0;

	return (
		<PageStack>
			<ScreenHeader
				eyebrow="Controller overview"
				title="Dashboard"
				copy="A premium active-measurement cockpit for probe health, scheduled checks, stream semantics, and recent path anomalies."
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
				<MetricCard label="API Status" value={healthQuery.data?.status || "unknown"} detail={rootQuery.data?.message || "controller"} tone={healthQuery.data?.status === "ok" ? "success" : "warning"} />
			</ResponsiveGrid>

			<ResponsiveGrid>
				<Panel tone="glass" eyebrow="Fleet bitmap" title={`${probes.length} probes, ${onlineProbes} lit`}>
					<FleetMatrix total={Math.max(probes.length, 1)} online={onlineProbes} />
				</Panel>
				<Panel tone="glass" eyebrow="Anomalies" title="Recent system events">
					<div className={styles.feed}>
						<Event title="Packet loss threshold exceeded" copy="nyc-vps-03 → api.netstamp.io exceeded 18% loss for 5m." tone="critical" />
						<Event title="Path hash changed from previous run" copy="fra-bm-02 observed transit shift at hop 9." tone="warning" />
						<Event title="Controller stream connected" copy="100 probes streaming normalized result payloads." tone="success" />
					</div>
				</Panel>
			</ResponsiveGrid>
		</PageStack>
	);
}

interface EventProps {
	title: string;
	copy: string;
	tone: BadgeTone;
}

function Event({ title, copy, tone }: EventProps) {
	return (
		<Surface as="article" className={styles.event} tone="flat" cut="md" padding="sm">
			<Badge tone={tone}>{title}</Badge>
			<BodyCopy>{copy}</BodyCopy>
		</Surface>
	);
}

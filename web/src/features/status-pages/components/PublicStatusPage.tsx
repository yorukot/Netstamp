import { checkTypeLabel, formatDateTime, formatMetric, publicStatusChartOption, severityTone, statusLabel, statusTone } from "@/features/status-pages/api/statusPageAdapters";
import { ApiError } from "@/shared/api/client";
import { publicStatusQueries } from "@/shared/api/queries";
import type { ApiPublicStatusPublicElement, ApiPublicStatusPublicResponse } from "@/shared/api/types";
import { ChartPanel } from "@/shared/visualizations/ChartPanel";
import { Badge, LoadingState, Panel } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "react-router-dom";
import styles from "./PublicStatusPage.module.css";

type PublicIncident = ApiPublicStatusPublicResponse["incidents"]["active"][number];

export function PublicStatusPage() {
	const { slug = "" } = useParams();
	const statusQuery = useQuery({
		...publicStatusQueries.detail(slug, { includeCharts: true }),
		enabled: Boolean(slug),
		select: data => data as ApiPublicStatusPublicResponse
	});

	if (statusQuery.isPending) {
		return (
			<main className={styles.page}>
				<div className={styles.shell}>
					<LoadingState label="Loading status page" detail="Fetching the current public status snapshot." />
				</div>
			</main>
		);
	}

	if (statusQuery.error) {
		const notFound = statusQuery.error instanceof ApiError && statusQuery.error.status === 404;
		return (
			<main className={styles.page}>
				<div className={styles.shell}>
					<Panel tone="deep" title={notFound ? "Status page not found" : "Status page unavailable"}>
						<p className={styles.muted}>{notFound ? "This public status page is disabled or does not exist." : "The controller could not return this status page right now."}</p>
					</Panel>
				</div>
			</main>
		);
	}

	const data = statusQuery.data;

	return (
		<main className={styles.page}>
			<div className={styles.shell}>
				<header className={styles.header}>
					<div className={styles.brand}>Netstamp Status</div>
					<div className={styles.headerGrid}>
						<div className={styles.headerCopy}>
							<Badge tone={statusTone(data.page.status)}>{statusLabel(data.page.status)}</Badge>
							<h1>{data.page.title}</h1>
							{data.page.description ? <p>{data.page.description}</p> : null}
						</div>
						<div className={styles.generated}>
							<span>Generated</span>
							<strong>{formatDateTime(data.generatedAt)}</strong>
						</div>
					</div>
				</header>

				<IncidentSection incidents={data.incidents.active} />

				<section className={styles.elements} aria-label="Status checks">
					{data.elements.length ? data.elements.map(element => <PublicElement key={element.id} element={element} />) : <div className={styles.empty}>No public status checks are configured.</div>}
				</section>
			</div>
		</main>
	);
}

function IncidentSection({ incidents }: { incidents: PublicIncident[] }) {
	if (!incidents.length) {
		return null;
	}

	return (
		<section className={styles.incidents} aria-label="Incidents">
			<Panel tone="deep" title="Open incidents">
				<div className={styles.incidentList}>
					{incidents.map(incident => (
						<IncidentRow key={incident.id} incident={incident} />
					))}
				</div>
			</Panel>
		</section>
	);
}

function IncidentRow({ incident }: { incident: PublicIncident }) {
	return (
		<div className={styles.incidentRow}>
			<div>
				<div className={styles.incidentTitle}>
					<Badge tone={severityTone(incident.severity)}>{incident.severity}</Badge>
					<strong>{incident.checkTitle}</strong>
				</div>
				<div className={styles.incidentMeta}>
					<span>{incident.status}</span>
					<span>Opened {formatDateTime(incident.openedAt)}</span>
					{incident.resolvedAt ? <span>Resolved {formatDateTime(incident.resolvedAt)}</span> : null}
				</div>
			</div>
			{incident.summary ? (
				<div className={styles.incidentSummary}>
					{incident.summary.metric ? <span>{incident.summary.metric}</span> : null}
					{typeof incident.summary.value === "number" ? <strong>{incident.summary.value.toFixed(1)}</strong> : null}
				</div>
			) : null}
		</div>
	);
}

function PublicElement({ element }: { element: ApiPublicStatusPublicElement }) {
	if (element.kind === "folder") {
		return (
			<div className={styles.folder}>
				<div className={styles.folderHeader}>
					<div>
						<h2>{element.title}</h2>
						{element.description ? <p>{element.description}</p> : null}
					</div>
					<Badge tone={statusTone(element.status)}>{statusLabel(element.status)}</Badge>
				</div>
				<div className={styles.folderChildren}>
					{element.children?.map(child => (
						<PublicElement key={child.id} element={child} />
					))}
				</div>
			</div>
		);
	}

	return (
		<article className={styles.check}>
			<div className={styles.checkMain}>
				<div className={styles.checkCopy}>
					<div className={styles.checkTitle}>
						<Badge tone={statusTone(element.status)}>{statusLabel(element.status)}</Badge>
						<h3>{element.title}</h3>
					</div>
					<div className={styles.checkMeta}>
						<span>{checkTypeLabel(element.type)}</span>
						{element.target ? <span>{element.target}</span> : null}
						{element.latestStartedAt ? <span>Latest {formatDateTime(element.latestStartedAt)}</span> : null}
					</div>
					{element.description ? <p>{element.description}</p> : null}
				</div>
				<div className={styles.assignmentStats}>
					<span>{element.successfulAssignments ?? 0} ok</span>
					<span>{element.failingAssignments ?? 0} failing</span>
					<span>{element.staleAssignments ?? 0} stale</span>
				</div>
			</div>
			<Metrics element={element} />
			{element.chart?.series.length ? <ChartPanel className={styles.chart} option={publicStatusChartOption(element)} height="12rem" /> : null}
		</article>
	);
}

function Metrics({ element }: { element: ApiPublicStatusPublicElement }) {
	const metrics = [
		{ label: "Latency", value: formatMetric(element.metrics?.latencyAvgMs, "ms") },
		{ label: "Loss", value: formatMetric(element.metrics?.lossPercent, "%") },
		{ label: "Connect", value: formatMetric(element.metrics?.connectAvgMs, "ms") },
		{ label: "Failure", value: formatMetric(element.metrics?.failurePercent, "%") }
	].filter(metric => metric.value !== "-");

	if (!metrics.length) {
		return null;
	}

	return (
		<div className={styles.metrics}>
			{metrics.map(metric => (
				<div className={styles.metric} key={metric.label}>
					<span>{metric.label}</span>
					<strong>{metric.value}</strong>
				</div>
			))}
		</div>
	);
}

import { checkTypeLabel, formatDateTime, formatMetric, publicStatusChartOption, severityTone, statusLabel, statusTone } from "@/features/status-pages/api/statusPageAdapters";
import { ApiError } from "@/shared/api/client";
import { publicStatusQueries } from "@/shared/api/queries";
import type { ApiPublicStatusPublicElement, ApiPublicStatusPublicResponse } from "@/shared/api/types";
import { ChartPanel } from "@/shared/visualizations/ChartPanel";
import { Badge, LoadingState, Panel, type BadgeTone } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "react-router-dom";
import styles from "./PublicStatusPage.module.css";

type PublicIncident = ApiPublicStatusPublicResponse["incidents"]["active"][number];
type PublicAssignment = NonNullable<ApiPublicStatusPublicElement["assignments"]>[number];

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
					{data.elements.length ? data.elements.map(element => <PublicElement key={element.id} element={element} />) : <div className={styles.empty}>No public status elements are configured.</div>}
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
						{element.type ? <span>{checkTypeLabel(element.type)}</span> : <span>{element.assignmentCount ?? element.assignments?.length ?? 0} assignments</span>}
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
			<AssignmentRows assignments={element.assignments ?? []} />
			{element.chart?.series.length ? <ChartPanel className={styles.chart} option={publicStatusChartOption(element)} height="12rem" /> : null}
		</article>
	);
}

function AssignmentRows({ assignments }: { assignments: PublicAssignment[] }) {
	if (!assignments.length) {
		return null;
	}

	return (
		<div className={styles.publicAssignmentRows}>
			{assignments.map(assignment => (
				<div key={assignment.assignmentId} className={styles.publicAssignmentRow}>
					<div className={styles.publicAssignmentCopy}>
						<strong>{assignment.checkTitle}</strong>
						<span>
							{checkTypeLabel(assignment.type)} / {assignment.target} / {assignment.probeName}
						</span>
						{assignment.latestStartedAt ? <span>Latest {formatDateTime(assignment.latestStartedAt)}</span> : null}
					</div>
					<div className={styles.publicAssignmentStatus}>
						<Badge tone={latestStatusTone(assignment.latestStatus)}>{latestStatusLabel(assignment.latestStatus)}</Badge>
						<AssignmentMetrics assignment={assignment} />
					</div>
				</div>
			))}
		</div>
	);
}

function latestStatusTone(status: string | undefined): BadgeTone {
	switch (status) {
		case "successful":
			return "success";
		case "partial":
			return "warning";
		case "timeout":
		case "error":
			return "critical";
		default:
			return "neutral";
	}
}

function latestStatusLabel(status: string | undefined) {
	switch (status) {
		case "successful":
			return "Ok";
		case "partial":
			return "Partial";
		case "timeout":
			return "Timeout";
		case "error":
			return "Error";
		default:
			return "No result";
	}
}

function AssignmentMetrics({ assignment }: { assignment: PublicAssignment }) {
	const metrics =
		assignment.type === "ping"
			? [
					{ label: "Latency", value: formatMetric(assignment.metrics?.latencyAvgMs, "ms") },
					{ label: "Loss", value: formatMetric(assignment.metrics?.lossPercent, "%") }
				]
			: [
					{ label: "Connect", value: formatMetric(assignment.metrics?.connectAvgMs, "ms") },
					{ label: "Failure", value: formatMetric(assignment.metrics?.failurePercent, "%") }
				];

	return (
		<div className={styles.publicAssignmentMetrics}>
			{metrics.map(metric => (
				<span key={metric.label}>
					{metric.label} {metric.value}
				</span>
			))}
		</div>
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

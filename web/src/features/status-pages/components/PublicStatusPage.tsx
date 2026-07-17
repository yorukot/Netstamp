import { checkTypeLabel, formatDateTime, formatMetric, publicStatusChartOption, severityTone, statusLabel, statusTone } from "@/features/status-pages/api/statusPageAdapters";
import { hasApiProblemCode } from "@/shared/api/client";
import { publicStatusQueries } from "@/shared/api/queries";
import type {
	ApiPublicStatusElementChartResponse,
	ApiPublicStatusElementDailyStatusResponse,
	ApiPublicStatusElementsResponse,
	ApiPublicStatusIncidentsResponse,
	ApiPublicStatusPublicElement,
	ApiPublicStatusSummaryResponse
} from "@/shared/api/types";
import { ChartPanel } from "@/shared/visualizations/ChartPanel";
import netstampLogo from "@netstamp/brand/assets/netstamp-logo-light.svg";
import { Badge, Panel, Spinner, type BadgeTone } from "@netstamp/ui";
import { CaretDownIcon } from "@phosphor-icons/react/dist/csr/CaretDown";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useRef, useState } from "react";
import { useParams } from "react-router-dom";
import styles from "./PublicStatusPage.module.css";

type PublicIncident = ApiPublicStatusIncidentsResponse["incidents"]["active"][number];
type PublicAssignment = NonNullable<ApiPublicStatusPublicElement["assignments"]>[number];
type DailyStatusDay = ApiPublicStatusElementDailyStatusResponse["days"][number];

export function PublicStatusPage() {
	const { slug = "" } = useParams();
	const summaryQuery = useQuery({
		...publicStatusQueries.summary(slug),
		enabled: Boolean(slug),
		select: data => data as ApiPublicStatusSummaryResponse
	});
	const elementsQuery = useQuery({
		...publicStatusQueries.elements(slug),
		enabled: Boolean(slug),
		select: data => data as ApiPublicStatusElementsResponse
	});
	const incidentsQuery = useQuery({
		...publicStatusQueries.incidents(slug),
		enabled: Boolean(slug),
		select: data => data as ApiPublicStatusIncidentsResponse
	});

	if (summaryQuery.isPending) {
		return (
			<main className={styles.page}>
				<div className={styles.shell}>
					<Spinner label="Loading status page" layout="panel" size="lg" />
				</div>
			</main>
		);
	}

	if (summaryQuery.error || !summaryQuery.data) {
		const notFound = hasApiProblemCode(summaryQuery.error, "PUBLIC_STATUS_PAGE_NOT_FOUND");
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

	const summary = summaryQuery.data;
	const activeIncidents = incidentsQuery.data?.incidents.active ?? [];
	const resolvedIncidents = incidentsQuery.data?.incidents.recentResolved ?? [];

	return (
		<main className={styles.page} data-status={summary.page.status}>
			<div className={styles.shell}>
				<header className={styles.hero}>
					<div className={styles.banner} role="img" aria-label="Network telemetry paths between global monitoring locations" />
					<div className={styles.heroBody}>
						<div className={styles.brandRow}>
							<img src={netstampLogo} alt="Netstamp" />
							<span>Public status</span>
						</div>
						<div className={styles.titleRow}>
							<div className={styles.headerCopy}>
								<h1>{summary.page.title}</h1>
								{summary.page.description ? <p>{summary.page.description}</p> : null}
							</div>
							<div className={styles.generated}>
								<span>Last checked</span>
								<strong>{formatDateTime(summary.generatedAt)}</strong>
							</div>
						</div>
					</div>
				</header>

				<section className={styles.overallStatus} aria-label="Current overall status">
					<span className={styles.statusMarker} aria-hidden="true" />
					<div>
						<strong>{overallStatusTitle(summary.page.status)}</strong>
						<span>{overallStatusSummary(summary.page.status, activeIncidents.length)}</span>
					</div>
					<Badge tone={statusTone(summary.page.status)}>{statusLabel(summary.page.status)}</Badge>
				</section>

				<IncidentSection activeIncidents={activeIncidents} resolvedIncidents={resolvedIncidents} isLoading={incidentsQuery.isPending} hasError={Boolean(incidentsQuery.error)} />
				<ElementSection slug={slug} elements={elementsQuery.data?.elements ?? []} isLoading={elementsQuery.isPending} hasError={Boolean(elementsQuery.error)} />

				<footer className={styles.footer}>
					<p>Measurements are collected by configured Netstamp probes. Status reflects observed availability and is not an independent SLA certification.</p>
					<span>Updated {formatDateTime(summary.generatedAt)}</span>
				</footer>
			</div>
		</main>
	);
}

function overallStatusTitle(status: ApiPublicStatusSummaryResponse["page"]["status"]) {
	switch (status) {
		case "operational":
			return "All systems operational";
		case "degraded":
			return "Some systems are degraded";
		case "down":
			return "Service interruption detected";
		default:
			return "Status is being evaluated";
	}
}

function overallStatusSummary(status: ApiPublicStatusSummaryResponse["page"]["status"], activeIncidentCount: number) {
	if (activeIncidentCount > 0) {
		return `${activeIncidentCount} active ${activeIncidentCount === 1 ? "incident" : "incidents"}`;
	}
	return status === "operational" ? "No active incidents reported" : "Live measurements are shown below";
}

function IncidentSection({
	activeIncidents,
	resolvedIncidents,
	isLoading,
	hasError
}: {
	activeIncidents: PublicIncident[];
	resolvedIncidents: PublicIncident[];
	isLoading: boolean;
	hasError: boolean;
}) {
	if (isLoading) {
		return (
			<section className={styles.incidents} aria-label="Incidents">
				<Panel tone="deep" title="Open incidents">
					<Spinner label="Loading incidents" layout="compact" size="lg" />
				</Panel>
			</section>
		);
	}

	if (hasError) {
		return (
			<section className={styles.incidents} aria-label="Incidents">
				<Panel tone="deep" title="Open incidents">
					<p className={styles.muted}>Incidents are unavailable right now.</p>
				</Panel>
			</section>
		);
	}

	if (!activeIncidents.length && !resolvedIncidents.length) {
		return null;
	}

	return (
		<section className={styles.incidents} aria-label="Incidents">
			{activeIncidents.length ? (
				<Panel tone="deep" title="Active incidents" summary="Current service interruptions and ongoing investigation updates.">
					<div className={styles.incidentList}>
						{activeIncidents.map(incident => (
							<IncidentRow key={incident.id} incident={incident} />
						))}
					</div>
				</Panel>
			) : null}
			{resolvedIncidents.length ? (
				<details className={styles.resolvedIncidents}>
					<summary>
						<span>Resolved incident history</span>
						<Badge tone="neutral">{resolvedIncidents.length}</Badge>
					</summary>
					<div className={styles.incidentList}>
						{resolvedIncidents.map(incident => (
							<IncidentRow key={incident.id} incident={incident} />
						))}
					</div>
				</details>
			) : null}
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

function ElementSection({ slug, elements, isLoading, hasError }: { slug: string; elements: ApiPublicStatusPublicElement[]; isLoading: boolean; hasError: boolean }) {
	if (isLoading) {
		return (
			<section className={styles.elements} aria-label="Status checks">
				<Spinner label="Loading status checks" layout="panel" size="lg" />
			</section>
		);
	}

	if (hasError) {
		return (
			<section className={styles.elements} aria-label="Status checks">
				<Panel tone="deep" title="Status checks unavailable">
					<p className={styles.muted}>The current status checks could not be loaded.</p>
				</Panel>
			</section>
		);
	}

	return (
		<section className={styles.elements} aria-label="Status checks">
			{elements.length ? elements.map(element => <PublicElement key={element.id} slug={slug} element={element} />) : <div className={styles.empty}>No public status elements are configured.</div>}
		</section>
	);
}

function PublicElement({ slug, element }: { slug: string; element: ApiPublicStatusPublicElement }) {
	if (element.kind === "folder") {
		return (
			<section className={styles.folder} aria-labelledby={`status-group-${element.id}`}>
				<div className={styles.folderHeader}>
					<div>
						<h2 id={`status-group-${element.id}`}>{element.title}</h2>
						{element.description ? <p>{element.description}</p> : null}
					</div>
					<Badge tone={statusTone(element.status)}>{statusLabel(element.status)}</Badge>
				</div>
				<div className={styles.folderChildren}>
					{element.children?.map(child => (
						<PublicElement key={child.id} slug={slug} element={child} />
					))}
				</div>
			</section>
		);
	}

	return <ExpandableStatusRow slug={slug} element={element} />;
}

function ExpandableStatusRow({ slug, element }: { slug: string; element: ApiPublicStatusPublicElement }) {
	const [expanded, setExpanded] = useState(false);

	return (
		<article className={styles.check} data-expanded={expanded}>
			<button type="button" className={styles.checkToggle} aria-expanded={expanded} aria-controls={`status-details-${element.id}`} onClick={() => setExpanded(current => !current)}>
				<div className={styles.checkIdentity}>
					<span className={`${styles.serviceState} ${serviceStateClass(element.status)}`} aria-hidden="true" />
					<div className={styles.checkCopy}>
						<h3>{element.title}</h3>
						<div className={styles.checkMeta}>
							{element.type ? <span>{checkTypeLabel(element.type)}</span> : <span>{element.assignmentCount ?? element.assignments?.length ?? 0} viewpoints</span>}
							{element.latestStartedAt ? <span>Checked {formatDateTime(element.latestStartedAt)}</span> : null}
						</div>
					</div>
				</div>
				<div className={styles.checkState}>
					<strong>{statusLabel(element.status)}</strong>
					<CaretDownIcon aria-hidden="true" focusable="false" />
				</div>
			</button>
			<LazyPublicElementDailyStatus slug={slug} element={element} />
			<div id={`status-details-${element.id}`} className={styles.checkDetails} hidden={!expanded}>
				{element.description ? <p className={styles.checkDescription}>{element.description}</p> : null}
				<div className={styles.assignmentStats}>
					<span>{element.successfulAssignments ?? 0} operational</span>
					<span>{element.failingAssignments ?? 0} failing</span>
					<span>{element.staleAssignments ?? 0} stale</span>
				</div>
				<Metrics element={element} />
				<AssignmentRows assignments={element.assignments ?? []} />
				{element.chart?.series.length ? <ChartPanel className={styles.chart} option={publicStatusChartOption(element)} height="12rem" /> : null}
				{!element.chart?.series.length && element.chartMode === "compact" ? <LazyPublicElementChart slug={slug} element={element} /> : null}
			</div>
		</article>
	);
}

function serviceStateClass(status: ApiPublicStatusPublicElement["status"]) {
	switch (status) {
		case "operational":
			return styles.serviceStateOperational;
		case "degraded":
			return styles.serviceStateDegraded;
		case "down":
			return styles.serviceStateDown;
		default:
			return styles.serviceStateUnknown;
	}
}

function LazyPublicElementDailyStatus({ slug, element }: { slug: string; element: ApiPublicStatusPublicElement }) {
	const { ref, inView } = useInView<HTMLDivElement>("200px");
	const filters = { range: "30d" as const };
	const dailyStatusQuery = useQuery({
		...publicStatusQueries.elementDailyStatus(slug, element.id, filters),
		enabled: Boolean(slug) && inView,
		select: data => data as ApiPublicStatusElementDailyStatusResponse
	});
	const days = dailyStatusQuery.data?.days ?? [];

	return (
		<div ref={ref} className={styles.dailyStatus}>
			{dailyStatusQuery.isPending || (dailyStatusQuery.isFetching && !dailyStatusQuery.data) ? (
				<div className={styles.dailyStatusPlaceholder}>
					<Spinner label="Loading 30 days" size="sm" />
				</div>
			) : null}
			{dailyStatusQuery.error ? <div className={styles.dailyStatusPlaceholder}>Daily status unavailable</div> : null}
			{days.length ? (
				<>
					<div className={styles.dailyStatusBars} aria-label={`${element.title} 30 day status`}>
						{days.map(day => (
							<span key={day.date} className={`${styles.dailyStatusBar} ${dailyStatusBarClass(day.status)}`} title={dailyStatusTitle(day)} />
						))}
					</div>
					<div className={styles.dailyStatusMeta}>
						<span>{formatDateLabel(days[0]?.date)}</span>
						<span>{formatDateLabel(days[days.length - 1]?.date)}</span>
					</div>
				</>
			) : null}
		</div>
	);
}

function LazyPublicElementChart({ slug, element }: { slug: string; element: ApiPublicStatusPublicElement }) {
	const { ref, inView } = useInView<HTMLDivElement>("200px");
	const filters = element.chartRange ? { range: element.chartRange } : {};
	const chartQuery = useQuery({
		...publicStatusQueries.elementChart(slug, element.id, filters),
		enabled: Boolean(slug) && inView,
		select: data => data as ApiPublicStatusElementChartResponse
	});
	const chart = chartQuery.data?.chart;

	return (
		<div ref={ref} className={styles.chartSlot}>
			{chartQuery.isPending || (chartQuery.isFetching && !chartQuery.data) ? (
				<div className={styles.chartPlaceholder}>
					<Spinner label="Loading chart" size="lg" />
				</div>
			) : null}
			{chartQuery.error ? <div className={styles.chartPlaceholder}>Chart unavailable</div> : null}
			{chart?.series.length ? <ChartPanel className={styles.chart} option={publicStatusChartOption({ ...element, chart })} height="12rem" /> : null}
		</div>
	);
}

function useInView<T extends Element>(rootMargin: string) {
	const ref = useRef<T | null>(null);
	const [inView, setInView] = useState(() => typeof IntersectionObserver === "undefined");

	useEffect(() => {
		const node = ref.current;
		if (!node || inView || typeof IntersectionObserver === "undefined") {
			return;
		}
		const observer = new IntersectionObserver(
			entries => {
				if (entries.some(entry => entry.isIntersecting)) {
					setInView(true);
					observer.disconnect();
				}
			},
			{ rootMargin }
		);
		observer.observe(node);
		return () => observer.disconnect();
	}, [inView, rootMargin]);

	return { ref, inView };
}

function dailyStatusBarClass(status: DailyStatusDay["status"]) {
	switch (status) {
		case "operational":
			return styles.dailyStatusBarOperational;
		case "degraded":
			return styles.dailyStatusBarDegraded;
		case "down":
			return styles.dailyStatusBarDown;
		default:
			return styles.dailyStatusBarUnknown;
	}
}

function dailyStatusTitle(day: DailyStatusDay) {
	const incidentText = day.incidentCount === 1 ? "1 incident" : `${day.incidentCount} incidents`;
	return `${formatDateLabel(day.date)} / ${statusLabel(day.status)} / ${incidentText}`;
}

function formatDateLabel(value: string | undefined) {
	if (!value) {
		return "-";
	}
	const date = new Date(`${value}T00:00:00Z`);
	if (Number.isNaN(date.getTime())) {
		return value;
	}
	return date.toLocaleDateString([], { month: "short", day: "2-digit", timeZone: "UTC" });
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
			: assignment.type === "http"
				? [
						{ label: "Total", value: formatMetric(assignment.metrics?.latencyAvgMs, "ms") },
						{ label: "Failure", value: formatMetric(assignment.metrics?.failurePercent, "%") }
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
		{ label: element.type === "http" ? "Total" : "Latency", value: formatMetric(element.metrics?.latencyAvgMs, "ms") },
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

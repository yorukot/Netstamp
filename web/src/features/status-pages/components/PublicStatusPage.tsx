import { checkTypeLabel, formatDateTime, formatMetric, publicStatusChartOption, severityTone, statusLabel, statusTone } from "@/features/status-pages/api/statusPageAdapters";
import { ApiError } from "@/shared/api/client";
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
import { Badge, LoadingState, Panel, type BadgeTone } from "@netstamp/ui";
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
					<LoadingState label="Loading status page" detail="Fetching the current public status snapshot." />
				</div>
			</main>
		);
	}

	if (summaryQuery.error || !summaryQuery.data) {
		const notFound = summaryQuery.error instanceof ApiError && summaryQuery.error.status === 404;
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

	return (
		<main className={styles.page}>
			<div className={styles.shell}>
				<header className={styles.header}>
					<div className={styles.brand}>Netstamp Status</div>
					<div className={styles.headerGrid}>
						<div className={styles.headerCopy}>
							<Badge tone={statusTone(summary.page.status)}>{statusLabel(summary.page.status)}</Badge>
							<h1>{summary.page.title}</h1>
							{summary.page.description ? <p>{summary.page.description}</p> : null}
						</div>
						<div className={styles.generated}>
							<span>Generated</span>
							<strong>{formatDateTime(summary.generatedAt)}</strong>
						</div>
					</div>
				</header>

				<IncidentSection incidents={incidentsQuery.data?.incidents.active ?? []} isLoading={incidentsQuery.isPending} hasError={Boolean(incidentsQuery.error)} />
				<ElementSection slug={slug} elements={elementsQuery.data?.elements ?? []} isLoading={elementsQuery.isPending} hasError={Boolean(elementsQuery.error)} />
			</div>
		</main>
	);
}

function IncidentSection({ incidents, isLoading, hasError }: { incidents: PublicIncident[]; isLoading: boolean; hasError: boolean }) {
	if (isLoading) {
		return (
			<section className={styles.incidents} aria-label="Incidents">
				<Panel tone="deep" title="Open incidents">
					<LoadingState label="Loading incidents" detail="Fetching active public incidents." />
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

function ElementSection({ slug, elements, isLoading, hasError }: { slug: string; elements: ApiPublicStatusPublicElement[]; isLoading: boolean; hasError: boolean }) {
	if (isLoading) {
		return (
			<section className={styles.elements} aria-label="Status checks">
				<LoadingState label="Loading status checks" detail="Fetching current public check status." />
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
						<PublicElement key={child.id} slug={slug} element={child} />
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
			<LazyPublicElementDailyStatus slug={slug} element={element} />
			<AssignmentRows assignments={element.assignments ?? []} />
			{element.chart?.series.length ? <ChartPanel className={styles.chart} option={publicStatusChartOption(element)} height="12rem" /> : null}
			{!element.chart?.series.length && element.chartMode === "compact" ? <LazyPublicElementChart slug={slug} element={element} /> : null}
		</article>
	);
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
			{dailyStatusQuery.isPending || (dailyStatusQuery.isFetching && !dailyStatusQuery.data) ? <div className={styles.dailyStatusPlaceholder}>Loading 30 days</div> : null}
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
			{chartQuery.isPending || (chartQuery.isFetching && !chartQuery.data) ? <div className={styles.chartPlaceholder}>Loading chart</div> : null}
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

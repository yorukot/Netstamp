import { useSession } from "@/features/auth/session/SessionContext";
import { checkTypeLabel, formatDateTime, formatMetric, publicStatusChartOption, severityTone, statusLabel, statusTone } from "@/features/status-pages/api/statusPageAdapters";
import { pathForRoute, pathForStatusPageEditor } from "@/routes/routePaths";
import { hasApiProblemCode } from "@/shared/api/client";
import { projectQueries, publicStatusQueries } from "@/shared/api/queries";
import type {
	ApiPublicStatusEditorContextResponse,
	ApiPublicStatusElementChartResponse,
	ApiPublicStatusElementDailyStatusResponse,
	ApiPublicStatusElementsResponse,
	ApiPublicStatusIncidentsResponse,
	ApiPublicStatusPublicElement,
	ApiPublicStatusSummaryResponse
} from "@/shared/api/types";
import { ChartPanel } from "@/shared/visualizations/ChartPanel";
import netstampLogoDark from "@netstamp/brand/assets/netstamp-logo-dark.svg";
import netstampLogoLight from "@netstamp/brand/assets/netstamp-logo-light.svg";
import { Badge, Button, Spinner, type BadgeTone } from "@netstamp/ui";
import { CaretDownIcon } from "@phosphor-icons/react/dist/csr/CaretDown";
import { useQueries, useQuery } from "@tanstack/react-query";
import { lazy, Suspense, useEffect, useMemo, useRef, useState } from "react";
import { Link, useParams } from "react-router-dom";
import styles from "./PublicStatusPage.module.css";

type PublicIncident = ApiPublicStatusIncidentsResponse["incidents"]["active"][number];
type PublicAssignment = NonNullable<ApiPublicStatusPublicElement["assignments"]>[number];
type DailyStatusDay = ApiPublicStatusElementDailyStatusResponse["days"][number];
type PublicPageTheme = ApiPublicStatusSummaryResponse["page"]["theme"];

const NetworkMap = lazy(() => import("@/shared/visualizations/NetworkMap").then(module => ({ default: module.NetworkMap })));

export function PublicStatusPage() {
	const { slug = "" } = useParams();
	const { session } = useSession();
	const editorContextQuery = useQuery({
		...publicStatusQueries.editorContext(slug),
		enabled: Boolean(slug && session),
		select: data => data as ApiPublicStatusEditorContextResponse
	});
	const useLegacyEditorLookup = Boolean(session && hasApiProblemCode(editorContextQuery.error, "ROUTE_NOT_FOUND"));
	const legacyProjectsQuery = useQuery({
		...projectQueries.list(),
		enabled: useLegacyEditorLookup
	});
	const legacyProjectRefs = (legacyProjectsQuery.data?.projects ?? []).map(project => project.slug || project.id);
	const legacyStatusPageQueries = useQueries({
		queries: legacyProjectRefs.map(projectRef => ({
			...projectQueries.statusPages(projectRef),
			enabled: useLegacyEditorLookup
		}))
	});
	let legacyEditorContext: ApiPublicStatusEditorContextResponse | undefined;
	for (const [index, query] of legacyStatusPageQueries.entries()) {
		const page = query.data?.pages.find(candidate => candidate.slug === slug);
		if (page) {
			legacyEditorContext = { projectRef: legacyProjectRefs[index] || "", pageId: page.id };
			break;
		}
	}
	const editorContext = editorContextQuery.data ?? legacyEditorContext;
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
	const mapTheme = usePublicStatusTheme(summaryQuery.data?.page.theme);

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
					<section className={styles.messageSurface}>
						<h1>{notFound ? "Status page not found" : "Status page unavailable"}</h1>
						<p className={styles.muted}>{notFound ? "This public status page is disabled or does not exist." : "The controller could not return this status page right now."}</p>
					</section>
				</div>
			</main>
		);
	}

	const summary = summaryQuery.data;
	const activeIncidents = incidentsQuery.data?.incidents.active ?? [];
	const resolvedIncidents = summary.page.showIncidentHistory ? (incidentsQuery.data?.incidents.recentResolved ?? []) : [];
	const defaultFooter = "Measurements are collected by configured Netstamp probes. Status reflects observed availability and is not an independent SLA certification.";

	return (
		<main className={`${styles.page} ns-status-page`} data-status={summary.page.status}>
			{summary.page.customCss ? <style>{summary.page.customCss}</style> : null}
			<header className={`${styles.hero} ns-status-hero`}>
				{summary.page.bannerImageUrl ? (
					<img className={`${styles.banner} ${styles.bannerImage} ns-status-banner`} src={summary.page.bannerImageUrl} alt="" />
				) : (
					<div className={`${styles.banner} ns-status-banner`} aria-hidden="true" />
				)}
				<div className={styles.heroBody}>
					<div className={styles.brandRow}>
						<img src={mapTheme === "dark" ? netstampLogoLight : netstampLogoDark} alt="Netstamp" />
						<span>Public status</span>
					</div>
					<div className={styles.titleRow}>
						<div className={styles.headerCopy}>
							<h1>{summary.page.title}</h1>
							{summary.page.description ? <p>{summary.page.description}</p> : null}
						</div>
						{summary.page.showGeneratedAt || session ? (
							<div className={styles.heroActions}>
								{summary.page.showGeneratedAt ? (
									<div className={styles.generated}>
										<span>Last checked</span>
										<strong>{formatDateTime(summary.generatedAt)}</strong>
									</div>
								) : null}
								{session ? (
									<div className={styles.sessionActions}>
										<Button asChild variant="ghost" size="sm">
											<Link to={pathForRoute("dashboard", { projectRef: editorContext?.projectRef })}>Go to dashboard</Link>
										</Button>
										{editorContext ? (
											<Button asChild variant="outline" size="sm">
												<Link to={pathForStatusPageEditor(editorContext.projectRef, editorContext.pageId)}>Edit status page</Link>
											</Button>
										) : editorContextQuery.isPending || (useLegacyEditorLookup && (legacyProjectsQuery.isPending || legacyStatusPageQueries.some(query => query.isPending))) ? (
											<Button type="button" variant="outline" size="sm" disabled>
												Checking edit access
											</Button>
										) : null}
									</div>
								) : null}
							</div>
						) : null}
					</div>
				</div>
			</header>

			<div className={styles.shell}>
				<section className={`${styles.overallStatus} ns-status-overall`} aria-label="Current overall status">
					<span className={styles.statusMarker} aria-hidden="true" />
					<div>
						<strong>{overallStatusTitle(summary.page.status)}</strong>
						<span>{overallStatusSummary(summary.page.status, activeIncidents.length)}</span>
					</div>
					<Badge tone={statusTone(summary.page.status)}>{statusLabel(summary.page.status)}</Badge>
				</section>

				<IncidentSection activeIncidents={activeIncidents} resolvedIncidents={resolvedIncidents} isLoading={incidentsQuery.isPending} hasError={Boolean(incidentsQuery.error)} />
				<ElementSection slug={slug} elements={elementsQuery.data?.elements ?? []} isLoading={elementsQuery.isPending} hasError={Boolean(elementsQuery.error)} mapTheme={mapTheme} />

				<footer className={`${styles.footer} ns-status-footer`}>
					<p>{summary.page.footerText || defaultFooter}</p>
					{summary.page.showGeneratedAt ? <span>Updated {formatDateTime(summary.generatedAt)}</span> : null}
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
				<div className={styles.incidentPanel}>
					<h2>Open incidents</h2>
					<Spinner label="Loading incidents" layout="compact" size="lg" />
				</div>
			</section>
		);
	}

	if (hasError) {
		return (
			<section className={styles.incidents} aria-label="Incidents">
				<div className={styles.incidentPanel}>
					<h2>Open incidents</h2>
					<p className={styles.muted}>Incidents are unavailable right now.</p>
				</div>
			</section>
		);
	}

	if (!activeIncidents.length && !resolvedIncidents.length) {
		return null;
	}

	return (
		<section className={styles.incidents} aria-label="Incidents">
			{activeIncidents.length ? (
				<div className={styles.incidentPanel}>
					<div className={styles.incidentHeading}>
						<h2>Active incidents</h2>
						<p>Current service interruptions and ongoing investigation updates.</p>
					</div>
					<div className={styles.incidentList}>
						{activeIncidents.map(incident => (
							<IncidentRow key={incident.id} incident={incident} />
						))}
					</div>
				</div>
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

function ElementSection({
	slug,
	elements,
	isLoading,
	hasError,
	mapTheme
}: {
	slug: string;
	elements: ApiPublicStatusPublicElement[];
	isLoading: boolean;
	hasError: boolean;
	mapTheme: "light" | "dark";
}) {
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
				<div className={styles.messageSurface}>
					<h2>Status checks unavailable</h2>
					<p className={styles.muted}>The current status checks could not be loaded.</p>
				</div>
			</section>
		);
	}

	return (
		<section className={styles.elements} aria-label="Status checks">
			{elements.length ? (
				elements.map(element => <PublicElement key={element.id} slug={slug} element={element} mapTheme={mapTheme} />)
			) : (
				<div className={styles.empty}>No public status elements are configured.</div>
			)}
		</section>
	);
}

function PublicElement({ slug, element, mapTheme }: { slug: string; element: ApiPublicStatusPublicElement; mapTheme: "light" | "dark" }) {
	if (element.kind === "folder") {
		return (
			<section className={`${styles.folder} ns-status-group`} aria-labelledby={`status-group-${element.id}`}>
				<div className={styles.folderHeader}>
					<div>
						<h2 id={`status-group-${element.id}`}>{element.title}</h2>
						{element.description ? <p>{element.description}</p> : null}
					</div>
					<Badge tone={statusTone(element.status)}>{statusLabel(element.status)}</Badge>
				</div>
				<div className={styles.folderChildren}>
					{element.children?.map(child => (
						<PublicElement key={child.id} slug={slug} element={child} mapTheme={mapTheme} />
					))}
				</div>
			</section>
		);
	}

	return <ExpandableStatusRow slug={slug} element={element} mapTheme={mapTheme} />;
}

function ExpandableStatusRow({ slug, element, mapTheme }: { slug: string; element: ApiPublicStatusPublicElement; mapTheme: "light" | "dark" }) {
	const [expanded, setExpanded] = useState(false);
	const showChart = element.displayMode === "latency" || element.chartMode === "compact";

	return (
		<article className={`${styles.check} ns-status-block`} data-expanded={expanded}>
			<button type="button" className={styles.checkToggle} aria-expanded={expanded} aria-controls={`status-details-${element.id}`} onClick={() => setExpanded(current => !current)}>
				<div className={styles.checkSummary}>
					<div className={styles.checkIdentity}>
						<span className={`${styles.serviceState} ${serviceStateClass(element.status)}`} aria-hidden="true" />
						<div className={styles.checkCopy}>
							<h3>{element.title}</h3>
							<div className={styles.checkMeta}>
								{element.type ? <span>{checkTypeLabel(element.type)}</span> : <span>{element.assignmentCount ?? element.assignments?.length ?? 0} viewpoints</span>}
								<span>{displayModeLabel(element.displayMode)}</span>
								{element.latestStartedAt ? <span>Checked {formatDateTime(element.latestStartedAt)}</span> : null}
							</div>
						</div>
					</div>
					<div className={styles.checkState}>
						<strong>{statusLabel(element.status)}</strong>
						<CaretDownIcon aria-hidden="true" focusable="false" />
					</div>
				</div>
				<LazyPublicElementDailyStatus slug={slug} element={element} />
			</button>
			<div id={`status-details-${element.id}`} className={styles.checkDetailsRegion} aria-hidden={!expanded}>
				<div className={styles.checkDetailsClip}>
					<div className={styles.checkDetails}>
						{element.description ? <p className={styles.checkDescription}>{element.description}</p> : null}
						<div className={styles.assignmentStats}>
							<span>{element.successfulAssignments ?? 0} operational</span>
							<span>{element.failingAssignments ?? 0} failing</span>
							<span>{element.staleAssignments ?? 0} stale</span>
						</div>
						<Metrics element={element} />
						<AssignmentRows assignments={element.assignments ?? []} />
						{expanded && element.displayMode === "map" ? <AssignmentMap assignments={element.assignments ?? []} theme={mapTheme} /> : null}
						{showChart && element.chart?.series.length ? <ChartPanel className={styles.chart} option={publicStatusChartOption(element)} height="12rem" /> : null}
						{showChart && !element.chart?.series.length ? <LazyPublicElementChart slug={slug} element={element} /> : null}
					</div>
				</div>
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
			{assignments.map((assignment, index) => {
				const context = [checkTypeLabel(assignment.type), assignment.target, assignment.probeName, assignment.probeLocationName].filter(Boolean);
				return (
					<div key={`${assignment.checkTitle}-${assignment.probeName ?? assignment.probeLocationName ?? index}`} className={styles.publicAssignmentRow}>
						<div className={styles.publicAssignmentCopy}>
							<strong>{assignment.checkTitle}</strong>
							{context.length ? <span>{context.join(" / ")}</span> : null}
							{assignment.latestStartedAt ? <span>Latest {formatDateTime(assignment.latestStartedAt)}</span> : null}
						</div>
						<div className={styles.publicAssignmentStatus}>
							<Badge tone={latestStatusTone(assignment.latestStatus)}>{latestStatusLabel(assignment.latestStatus)}</Badge>
							<AssignmentMetrics assignment={assignment} />
						</div>
					</div>
				);
			})}
		</div>
	);
}

function AssignmentMap({ assignments, theme }: { assignments: PublicAssignment[]; theme: "light" | "dark" }) {
	const markers = useMemo(
		() =>
			assignments.flatMap((assignment, index) => {
				if (typeof assignment.latitude !== "number" || typeof assignment.longitude !== "number") {
					return [];
				}
				return [
					{
						id: `public-probe-${index}`,
						name: assignment.probeName ?? assignment.probeLocationName ?? `Viewpoint ${index + 1}`,
						coordinates: [assignment.longitude, assignment.latitude] as [number, number],
						status: assignment.latestStatus === "error" || assignment.latestStatus === "timeout" ? "offline" : "online"
					}
				];
			}),
		[assignments]
	);

	if (!markers.length) {
		return <div className={styles.mapUnavailable}>No public probe locations are available for this block.</div>;
	}

	return (
		<Suspense fallback={<Spinner label="Loading probe map" layout="panel" size="lg" />}>
			<NetworkMap className={styles.probeMap} probes={markers} selectedId="" mode="fleet" theme={theme} />
		</Suspense>
	);
}

function displayModeLabel(mode: ApiPublicStatusPublicElement["displayMode"]) {
	switch (mode) {
		case "history":
			return "30-day history";
		case "latency":
			return "Latency";
		case "map":
			return "Probe map";
		default:
			return "Live status";
	}
}

function usePublicStatusTheme(theme: PublicPageTheme | undefined) {
	const preferredTheme = () => (typeof window !== "undefined" && window.matchMedia("(prefers-color-scheme: light)").matches ? "light" : "dark");
	const [resolvedTheme, setResolvedTheme] = useState<"light" | "dark">(() => (theme === "light" || theme === "dark" ? theme : preferredTheme()));

	useEffect(() => {
		if (!theme) {
			return;
		}

		const root = document.documentElement;
		const previousTheme = root.dataset.theme;
		const preference = window.matchMedia("(prefers-color-scheme: light)");
		const applyTheme = () => {
			const nextTheme = theme === "auto" ? (preference.matches ? "light" : "dark") : theme;
			root.dataset.theme = nextTheme;
			setResolvedTheme(nextTheme);
		};

		applyTheme();
		if (theme === "auto") {
			preference.addEventListener("change", applyTheme);
		}

		return () => {
			preference.removeEventListener("change", applyTheme);
			if (previousTheme) {
				root.dataset.theme = previousTheme;
			} else {
				delete root.dataset.theme;
			}
		};
	}, [theme]);

	return resolvedTheme;
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

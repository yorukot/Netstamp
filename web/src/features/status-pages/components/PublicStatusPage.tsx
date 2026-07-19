import { useSession } from "@/features/auth/session/SessionContext";
import { checkTypeLabel, formatDateTime, formatMetric, publicStatusChartOption, severityTone, statusLabel, statusTone } from "@/features/status-pages/api/statusPageAdapters";
import { currentLocale } from "@/i18n";
import { pathForRoute, pathForStatusPageEditor } from "@/routes/routePaths";
import { hasApiProblemCode } from "@/shared/api/client";
import { publicStatusQueries } from "@/shared/api/queries";
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
import { useQuery } from "@tanstack/react-query";
import type { TFunction } from "i18next";
import { lazy, Suspense, useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useParams } from "react-router-dom";
import styles from "./PublicStatusPage.module.css";

type PublicIncident = ApiPublicStatusIncidentsResponse["incidents"]["active"][number];
type PublicAssignment = NonNullable<ApiPublicStatusPublicElement["assignments"]>[number];
type DailyStatusDay = ApiPublicStatusElementDailyStatusResponse["days"][number];
type PublicPageTheme = ApiPublicStatusSummaryResponse["page"]["theme"];
type StatusT = TFunction<"status">;

const NetworkMap = lazy(() => import("@/shared/visualizations/NetworkMap").then(module => ({ default: module.NetworkMap })));

export function PublicStatusPage() {
	const { t } = useTranslation("status");
	const { slug = "" } = useParams();
	const { session } = useSession();
	const editorContextQuery = useQuery({
		...publicStatusQueries.editorContext(slug),
		enabled: Boolean(slug && session),
		select: data => data as ApiPublicStatusEditorContextResponse
	});
	const editorContext = editorContextQuery.data;
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
					<Spinner label={t("public.loading")} layout="panel" size="lg" />
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
						<h1>{notFound ? t("public.notFound") : t("public.unavailable")}</h1>
						<p className={styles.muted}>{notFound ? t("public.notFoundDescription") : t("public.unavailableDescription")}</p>
					</section>
				</div>
			</main>
		);
	}

	const summary = summaryQuery.data;
	const activeIncidents = incidentsQuery.data?.incidents.active ?? [];
	const resolvedIncidents = summary.page.showIncidentHistory ? (incidentsQuery.data?.incidents.recentResolved ?? []) : [];
	const defaultFooter = t("public.defaultFooter");

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
						<span>{t("public.publicStatus")}</span>
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
										<span>{t("public.lastChecked")}</span>
										<strong>{formatDateTime(summary.generatedAt)}</strong>
									</div>
								) : null}
								{session ? (
									<div className={styles.sessionActions}>
										<Button asChild variant="ghost" size="sm">
											<Link to={pathForRoute("dashboard", { projectRef: editorContext?.projectRef })}>{t("public.dashboard")}</Link>
										</Button>
										{editorContext ? (
											<Button asChild size="sm">
												<Link to={pathForStatusPageEditor(editorContext.projectRef, editorContext.pageId)}>{t("public.edit")}</Link>
											</Button>
										) : editorContextQuery.isPending ? (
											<Button type="button" variant="outline" size="sm" disabled>
												{t("public.checkingEditAccess")}
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
				<section className={`${styles.overallStatus} ns-status-overall`} aria-label={t("public.overallAria")}>
					<span className={styles.statusMarker} aria-hidden="true" />
					<div>
						<strong>{overallStatusTitle(summary.page.status, t)}</strong>
						<span>{overallStatusSummary(summary.page.status, activeIncidents.length, t)}</span>
					</div>
					<Badge tone={statusTone(summary.page.status)}>{statusLabel(summary.page.status)}</Badge>
				</section>

				<IncidentSection activeIncidents={activeIncidents} resolvedIncidents={resolvedIncidents} isLoading={incidentsQuery.isPending} hasError={Boolean(incidentsQuery.error)} />
				<ElementSection slug={slug} elements={elementsQuery.data?.elements ?? []} isLoading={elementsQuery.isPending} hasError={Boolean(elementsQuery.error)} mapTheme={mapTheme} />

				<footer className={`${styles.footer} ns-status-footer`}>
					<p>{summary.page.footerText || defaultFooter}</p>
					{summary.page.showGeneratedAt ? <span>{t("public.updated", { date: formatDateTime(summary.generatedAt) })}</span> : null}
				</footer>
			</div>
		</main>
	);
}

function overallStatusTitle(status: ApiPublicStatusSummaryResponse["page"]["status"], t: StatusT) {
	switch (status) {
		case "operational":
			return t("public.overall.operational");
		case "degraded":
			return t("public.overall.degraded");
		case "down":
			return t("public.overall.down");
		default:
			return t("public.overall.unknown");
	}
}

function overallStatusSummary(status: ApiPublicStatusSummaryResponse["page"]["status"], activeIncidentCount: number, t: StatusT) {
	if (activeIncidentCount > 0) {
		return t("public.overall.activeIncidents", { count: activeIncidentCount });
	}
	return status === "operational" ? t("public.overall.none") : t("public.overall.live");
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
	const { t } = useTranslation("status");
	if (isLoading) {
		return (
			<section className={styles.incidents} aria-label={t("public.incidents")}>
				<div className={styles.incidentPanel}>
					<h2>{t("public.openIncidents")}</h2>
					<Spinner label={t("public.loadingIncidents")} layout="compact" size="lg" />
				</div>
			</section>
		);
	}

	if (hasError) {
		return (
			<section className={styles.incidents} aria-label={t("public.incidents")}>
				<div className={styles.incidentPanel}>
					<h2>{t("public.openIncidents")}</h2>
					<p className={styles.muted}>{t("public.incidentsUnavailable")}</p>
				</div>
			</section>
		);
	}

	if (!activeIncidents.length && !resolvedIncidents.length) {
		return null;
	}

	return (
		<section className={styles.incidents} aria-label={t("public.incidents")}>
			{activeIncidents.length ? (
				<div className={styles.incidentPanel}>
					<div className={styles.incidentHeading}>
						<h2>{t("public.activeIncidents")}</h2>
						<p>{t("public.activeDescription")}</p>
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
						<span>{t("public.resolvedHistory")}</span>
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
	const { t } = useTranslation("status");
	return (
		<div className={styles.incidentRow}>
			<div>
				<div className={styles.incidentTitle}>
					<Badge tone={severityTone(incident.severity)}>{incident.severity}</Badge>
					<strong>{incident.checkTitle}</strong>
				</div>
				<div className={styles.incidentMeta}>
					<span>{incident.status}</span>
					<span>{t("public.opened", { date: formatDateTime(incident.openedAt) })}</span>
					{incident.resolvedAt ? <span>{t("public.resolvedAt", { date: formatDateTime(incident.resolvedAt) })}</span> : null}
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
	const { t } = useTranslation("status");
	if (isLoading) {
		return (
			<section className={styles.elements} aria-label={t("public.checksAria")}>
				<Spinner label={t("public.loadingChecks")} layout="panel" size="lg" />
			</section>
		);
	}

	if (hasError) {
		return (
			<section className={styles.elements} aria-label={t("public.checksAria")}>
				<div className={styles.messageSurface}>
					<h2>{t("public.checksUnavailable")}</h2>
					<p className={styles.muted}>{t("public.checksUnavailableDescription")}</p>
				</div>
			</section>
		);
	}

	return (
		<section className={styles.elements} aria-label={t("public.checksAria")}>
			{elements.length ? elements.map(element => <PublicElement key={element.id} slug={slug} element={element} mapTheme={mapTheme} />) : <div className={styles.empty}>{t("public.noElements")}</div>}
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
	const { t } = useTranslation("status");
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
								{element.type ? <span>{checkTypeLabel(element.type)}</span> : <span>{t("public.viewpointCount", { count: element.assignmentCount ?? element.assignments?.length ?? 0 })}</span>}
								<span>{displayModeLabel(element.displayMode, t)}</span>
								{element.latestStartedAt ? <span>{t("public.checked", { date: formatDateTime(element.latestStartedAt) })}</span> : null}
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
							<span>{t("public.operationalCount", { count: element.successfulAssignments ?? 0 })}</span>
							<span>{t("public.failingCount", { count: element.failingAssignments ?? 0 })}</span>
							<span>{t("public.staleCount", { count: element.staleAssignments ?? 0 })}</span>
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
	const { t } = useTranslation("status");
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
					<Spinner label={t("public.loadingDays")} size="sm" />
				</div>
			) : null}
			{dailyStatusQuery.error ? <div className={styles.dailyStatusPlaceholder}>{t("public.dailyUnavailable")}</div> : null}
			{days.length ? (
				<>
					<div className={styles.dailyStatusBars} aria-label={t("public.dailyAria", { title: element.title })}>
						{days.map(day => (
							<span key={day.date} className={`${styles.dailyStatusBar} ${dailyStatusBarClass(day.status)}`} title={dailyStatusTitle(day, t)} />
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
	const { t } = useTranslation("status");
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
					<Spinner label={t("public.loadingChart")} size="lg" />
				</div>
			) : null}
			{chartQuery.error ? <div className={styles.chartPlaceholder}>{t("public.chartUnavailable")}</div> : null}
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

function dailyStatusTitle(day: DailyStatusDay, t: StatusT) {
	const incidentText = t("public.incidentCount", { count: day.incidentCount });
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
	return date.toLocaleDateString(currentLocale(), { month: "short", day: "2-digit", timeZone: "UTC" });
}

function AssignmentRows({ assignments }: { assignments: PublicAssignment[] }) {
	const { t } = useTranslation("status");
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
							{assignment.latestStartedAt ? <span>{t("public.latest", { date: formatDateTime(assignment.latestStartedAt) })}</span> : null}
						</div>
						<div className={styles.publicAssignmentStatus}>
							<Badge tone={latestStatusTone(assignment.latestStatus)}>{latestStatusLabel(assignment.latestStatus, t)}</Badge>
							<AssignmentMetrics assignment={assignment} />
						</div>
					</div>
				);
			})}
		</div>
	);
}

function AssignmentMap({ assignments, theme }: { assignments: PublicAssignment[]; theme: "light" | "dark" }) {
	const { t } = useTranslation("status");
	const markers = useMemo(
		() =>
			assignments.flatMap((assignment, index) => {
				if (typeof assignment.latitude !== "number" || typeof assignment.longitude !== "number") {
					return [];
				}
				return [
					{
						id: `public-probe-${index}`,
						name: assignment.probeName ?? assignment.probeLocationName ?? t("public.viewpoint", { number: index + 1 }),
						coordinates: [assignment.longitude, assignment.latitude] as [number, number],
						status: assignment.latestStatus === "error" || assignment.latestStatus === "timeout" ? "offline" : "online"
					}
				];
			}),
		[assignments, t]
	);

	if (!markers.length) {
		return <div className={styles.mapUnavailable}>{t("public.noLocations")}</div>;
	}

	return (
		<Suspense fallback={<Spinner label={t("public.loadingMap")} layout="panel" size="lg" />}>
			<NetworkMap className={styles.probeMap} probes={markers} selectedId="" mode="fleet" theme={theme} />
		</Suspense>
	);
}

function displayModeLabel(mode: ApiPublicStatusPublicElement["displayMode"], t: StatusT) {
	switch (mode) {
		case "history":
			return t("public.display.history");
		case "latency":
			return t("public.display.latency");
		case "map":
			return t("public.display.map");
		default:
			return t("public.display.status");
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

function latestStatusLabel(status: string | undefined, t: StatusT) {
	switch (status) {
		case "successful":
			return t("public.results.successful");
		case "partial":
			return t("public.results.partial");
		case "timeout":
			return t("public.results.timeout");
		case "error":
			return t("public.results.error");
		default:
			return t("public.results.none");
	}
}

function AssignmentMetrics({ assignment }: { assignment: PublicAssignment }) {
	const { t } = useTranslation("status");
	const metrics =
		assignment.type === "ping"
			? [
					{ label: t("public.metrics.latency"), value: formatMetric(assignment.metrics?.latencyAvgMs, "ms") },
					{ label: t("public.metrics.loss"), value: formatMetric(assignment.metrics?.lossPercent, "%") }
				]
			: assignment.type === "http"
				? [
						{ label: t("public.metrics.total"), value: formatMetric(assignment.metrics?.latencyAvgMs, "ms") },
						{ label: t("public.metrics.failure"), value: formatMetric(assignment.metrics?.failurePercent, "%") }
					]
				: [
						{ label: t("public.metrics.connect"), value: formatMetric(assignment.metrics?.connectAvgMs, "ms") },
						{ label: t("public.metrics.failure"), value: formatMetric(assignment.metrics?.failurePercent, "%") }
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
	const { t } = useTranslation("status");
	const metrics = [
		{ label: element.type === "http" ? t("public.metrics.total") : t("public.metrics.latency"), value: formatMetric(element.metrics?.latencyAvgMs, "ms") },
		{ label: t("public.metrics.loss"), value: formatMetric(element.metrics?.lossPercent, "%") },
		{ label: t("public.metrics.connect"), value: formatMetric(element.metrics?.connectAvgMs, "ms") },
		{ label: t("public.metrics.failure"), value: formatMetric(element.metrics?.failurePercent, "%") }
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

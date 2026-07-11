import { queryOptions } from "@tanstack/react-query";
import { apiClient, readApiData } from "../client";
import { apiQueryKeys } from "../queryKeys";
import type {
	AlertIncidentFilters,
	AlertRuleFilters,
	HttpInsightFilters,
	HttpSeriesFilters,
	HttpSeriesResponse,
	LatestResultsFilters,
	PingInsightFilters,
	PingSeriesFilters,
	PingSeriesResponse,
	ProjectAssignmentFilters,
	PublicStatusChartFilters,
	PublicStatusDailyStatusFilters,
	PublicStatusIncidentsFilters,
	TcpInsightFilters,
	TcpSeriesFilters,
	TcpSeriesResponse,
	TracerouteInsightFilters,
	TracerouteRunsFilters,
	TracerouteTopologyFilters
} from "../types";

const defaultPingSeries = "latency_avg,latency_min,latency_max,loss_percent";
const defaultTCPSeries = "connect_avg,connect_min,connect_max,failure_percent";
const defaultHTTPSeries = "dns_avg,connect_avg,tls_avg,ttfb_avg,total_avg,failure_percent";

export const projectQueries = {
	list: () =>
		queryOptions({
			queryKey: apiQueryKeys.projects.list(),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/projects", { signal })),
			staleTime: 2 * 60 * 1000
		}),
	detail: (ref: string) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.detail(ref),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/projects/{ref}", { params: { path: { ref } }, signal })),
			staleTime: 2 * 60 * 1000
		}),
	checks: (ref: string) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.checks(ref),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/projects/{ref}/checks", { params: { path: { ref } }, signal })),
			staleTime: 30 * 1000
		}),
	checkDetail: (ref: string, checkId: string) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.checkDetail(ref, checkId),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/projects/{ref}/checks/{check_id}", { params: { path: { ref, check_id: checkId } }, signal })),
			staleTime: 30 * 1000
		}),
	assignments: (ref: string, filters: ProjectAssignmentFilters = {}) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.assignments(ref, filters),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/projects/{ref}/assignments", { params: { path: { ref }, query: filters }, signal })),
			staleTime: 30 * 1000
		}),
	alertRules: (ref: string, filters: AlertRuleFilters = {}) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.alertRules(ref, filters),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/projects/{ref}/alerts/rules", { params: { path: { ref }, query: filters }, signal })),
			staleTime: 30 * 1000
		}),
	alertIncidents: (ref: string, filters: AlertIncidentFilters = {}) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.alertIncidents(ref, filters),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/projects/{ref}/alerts/incidents", { params: { path: { ref }, query: filters }, signal })),
			staleTime: 15 * 1000
		}),
	alertIncidentDetail: (ref: string, incidentId: string) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.alertIncidentDetail(ref, incidentId),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/projects/{ref}/alerts/incidents/{incident_id}", { params: { path: { ref, incident_id: incidentId } }, signal })),
			staleTime: 15 * 1000
		}),
	notifications: (ref: string) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.notifications(ref),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/projects/{ref}/alerts/notifications", { params: { path: { ref } }, signal })),
			staleTime: 30 * 1000
		}),
	statusPages: (ref: string) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.statusPages(ref),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/projects/{ref}/status-pages", { params: { path: { ref } }, signal })),
			staleTime: 30 * 1000
		}),
	statusPageDetail: (ref: string, pageId: string) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.statusPageDetail(ref, pageId),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/projects/{ref}/status-pages/{page_id}", { params: { path: { ref, page_id: pageId } }, signal })),
			staleTime: 30 * 1000
		}),
	labels: (ref: string) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.labels(ref),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/projects/{ref}/labels", { params: { path: { ref } }, signal })),
			staleTime: 5 * 60 * 1000
		}),
	members: (ref: string) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.members(ref),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/projects/{ref}/members", { params: { path: { ref } }, signal })),
			staleTime: 60 * 1000
		}),
	invites: (ref: string) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.invites(ref),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/projects/{ref}/invites", { params: { path: { ref } }, signal })),
			staleTime: 60 * 1000
		}),
	currentUserInvites: () =>
		queryOptions({
			queryKey: apiQueryKeys.projects.currentUserInvites(),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/me/project-invites", { signal })),
			staleTime: 60 * 1000
		}),
	latestResults: (ref: string, filters: LatestResultsFilters = {}) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.latestResults(ref, filters),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/projects/{ref}/results/latest", { params: { path: { ref }, query: filters }, signal })),
			staleTime: 15 * 1000
		}),
	pingSeries: (ref: string, probeId: string, checkId: string, filters: PingSeriesFilters = {}) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.pingSeries(ref, probeId, checkId, filters),
			queryFn: ({ signal }) =>
				readApiData(
					apiClient.GET("/projects/{ref}/results/ping/series", {
						params: { path: { ref }, query: { probeId, checkId, series: defaultPingSeries, maxDataPoints: 600, ...filters } },
						signal
					})
				) as Promise<PingSeriesResponse>,
			staleTime: 30 * 1000
		}),
	pingInsight: (ref: string, probeId: string, checkId: string, filters: PingInsightFilters = {}) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.pingInsight(ref, probeId, checkId, filters),
			queryFn: ({ signal }) =>
				readApiData(
					apiClient.GET("/projects/{ref}/results/ping/insight", {
						params: { path: { ref }, query: { probeId, checkId, maxDataPoints: 600, ...filters } },
						signal
					})
				),
			staleTime: 30 * 1000
		}),
	tcpInsight: (ref: string, probeId: string, checkId: string, filters: TcpInsightFilters = {}) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.tcpInsight(ref, probeId, checkId, filters),
			queryFn: ({ signal }) =>
				readApiData(
					apiClient.GET("/projects/{ref}/results/tcp/insight", {
						params: { path: { ref }, query: { probeId, checkId, maxDataPoints: 600, ...filters } },
						signal
					})
				),
			staleTime: 30 * 1000
		}),
	tcpSeries: (ref: string, probeId: string, checkId: string, filters: TcpSeriesFilters = {}) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.tcpSeries(ref, probeId, checkId, filters),
			queryFn: ({ signal }) =>
				readApiData(
					apiClient.GET("/projects/{ref}/results/tcp/series", {
						params: { path: { ref }, query: { probeId, checkId, series: defaultTCPSeries, maxDataPoints: 600, ...filters } },
						signal
					})
				) as Promise<TcpSeriesResponse>,
			staleTime: 30 * 1000
		}),
	httpInsight: (ref: string, probeId: string, checkId: string, filters: HttpInsightFilters = {}) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.httpInsight(ref, probeId, checkId, filters),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/projects/{ref}/results/http/insight", { params: { path: { ref }, query: { probeId, checkId, maxDataPoints: 600, ...filters } }, signal })),
			staleTime: 30 * 1000
		}),
	httpSeries: (ref: string, probeId: string, checkId: string, filters: HttpSeriesFilters = {}) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.httpSeries(ref, probeId, checkId, filters),
			queryFn: ({ signal }) =>
				readApiData(
					apiClient.GET("/projects/{ref}/results/http/series", { params: { path: { ref }, query: { probeId, checkId, series: defaultHTTPSeries, maxDataPoints: 600, ...filters } }, signal })
				) as Promise<HttpSeriesResponse>,
			staleTime: 30 * 1000
		}),
	probes: (ref: string) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.probes(ref),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/projects/{ref}/probes", { params: { path: { ref } }, signal })),
			staleTime: 30 * 1000
		}),
	probeDetail: (ref: string, probeId: string) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.probeDetail(ref, probeId),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/projects/{ref}/probes/{probe_id}", { params: { path: { ref, probe_id: probeId } }, signal })),
			staleTime: 30 * 1000
		}),
	tracerouteRuns: (ref: string, probeId: string, checkId: string, filters: TracerouteRunsFilters = {}) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.tracerouteRuns(ref, probeId, checkId, filters),
			queryFn: ({ signal }) =>
				readApiData(
					apiClient.GET("/projects/{ref}/results/traceroute/runs", {
						params: { path: { ref }, query: { probeId, checkId, limit: 100, ...filters } },
						signal
					})
				),
			staleTime: 30 * 1000
		}),
	tracerouteInsight: (ref: string, probeId: string, checkId: string, filters: TracerouteInsightFilters = {}) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.tracerouteInsight(ref, probeId, checkId, filters),
			queryFn: ({ signal }) =>
				readApiData(
					apiClient.GET("/projects/{ref}/results/traceroute/insight", {
						params: { path: { ref }, query: { probeId, checkId, maxDataPoints: 600, ...filters } },
						signal
					})
				),
			staleTime: 30 * 1000
		}),
	tracerouteTopology: (ref: string, filters: TracerouteTopologyFilters = {}) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.tracerouteTopology(ref, filters),
			queryFn: ({ signal }) =>
				readApiData(
					apiClient.GET("/projects/{ref}/results/traceroute/topology", {
						params: { path: { ref }, query: { limit: 100, ...filters } },
						signal
					})
				),
			staleTime: 30 * 1000
		})
};

export const publicStatusQueries = {
	summary: (slug: string) =>
		queryOptions({
			queryKey: apiQueryKeys.publicStatus.summary(slug),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/public/status-pages/{slug}/summary", { params: { path: { slug } }, signal })),
			retry: false,
			staleTime: 30 * 1000
		}),
	elements: (slug: string) =>
		queryOptions({
			queryKey: apiQueryKeys.publicStatus.elements(slug),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/public/status-pages/{slug}/elements", { params: { path: { slug } }, signal })),
			retry: false,
			staleTime: 30 * 1000
		}),
	incidents: (slug: string, filters: PublicStatusIncidentsFilters = {}) =>
		queryOptions({
			queryKey: apiQueryKeys.publicStatus.incidents(slug, filters),
			queryFn: ({ signal }) =>
				readApiData(
					apiClient.GET("/public/status-pages/{slug}/incidents", {
						params: { path: { slug }, query: filters },
						signal
					})
				),
			retry: false,
			staleTime: 30 * 1000
		}),
	elementChart: (slug: string, elementId: string, filters: PublicStatusChartFilters = {}) =>
		queryOptions({
			queryKey: apiQueryKeys.publicStatus.elementChart(slug, elementId, filters),
			queryFn: ({ signal }) =>
				readApiData(
					apiClient.GET("/public/status-pages/{slug}/elements/{element_id}/chart", {
						params: { path: { slug, element_id: elementId }, query: filters },
						signal
					})
				),
			retry: false,
			staleTime: 30 * 1000
		}),
	elementDailyStatus: (slug: string, elementId: string, filters: PublicStatusDailyStatusFilters = {}) =>
		queryOptions({
			queryKey: apiQueryKeys.publicStatus.elementDailyStatus(slug, elementId, filters),
			queryFn: ({ signal }) =>
				readApiData(
					apiClient.GET("/public/status-pages/{slug}/elements/{element_id}/daily-status", {
						params: { path: { slug, element_id: elementId }, query: filters },
						signal
					})
				),
			retry: false,
			staleTime: 30 * 1000
		})
};

import { queryOptions } from "@tanstack/react-query";
import { apiClient, readApiData } from "./client";
import { apiQueryKeys } from "./queryKeys";
import type { MeasurementFilters, PingInsightFilters, PingSeriesFilters, ProjectAssignmentFilters, TracerouteRunsFilters, TracerouteTopologyFilters } from "./types";

export const systemQueries = {
	root: () =>
		queryOptions({
			queryKey: apiQueryKeys.system.root(),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/", { signal })),
			staleTime: 5 * 60 * 1000
		}),
	health: () =>
		queryOptions({
			queryKey: apiQueryKeys.system.health(),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/healthz", { signal })),
			staleTime: 30 * 1000
		})
};

export const authQueries = {
	me: () =>
		queryOptions({
			queryKey: apiQueryKeys.auth.me(),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/auth/me", { signal })),
			retry: false,
			staleTime: 5 * 60 * 1000
		})
};

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
	measurements: (ref: string, filters: MeasurementFilters = {}) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.measurements(ref, filters),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/projects/{ref}/measurements", { params: { path: { ref }, query: filters }, signal })),
			staleTime: 15 * 1000
		}),
	pingSeries: (ref: string, probeId: string, checkId: string, filters: PingSeriesFilters = {}) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.pingSeries(ref, probeId, checkId, filters),
			queryFn: ({ signal }) =>
				readApiData(
					apiClient.GET("/projects/{ref}/results/ping/series", {
						params: { path: { ref }, query: { probeId, checkId, maxDataPoints: 120, ...filters } },
						signal
					})
				),
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

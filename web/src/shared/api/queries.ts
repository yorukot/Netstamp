import { queryOptions } from "@tanstack/react-query";
import { apiClient, readApiData } from "./client";
import { apiQueryKeys } from "./queryKeys";

type PingSeriesMetric = "rttAvgMs" | "lossPercent" | "successRate";

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
	pingSeries: (ref: string, probeId: string, checkId: string, metric: PingSeriesMetric = "rttAvgMs") =>
		queryOptions({
			queryKey: apiQueryKeys.projects.pingSeries(ref, probeId, checkId, metric),
			queryFn: ({ signal }) =>
				readApiData(
					apiClient.GET("/projects/{ref}/results/ping/series", {
						params: { path: { ref }, query: { probeId, checkId, metric, maxDataPoints: 120 } },
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
		})
};

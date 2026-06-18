import { queryOptions } from "@tanstack/react-query";
import { apiClient, readApiData } from "../client";
import { apiQueryKeys } from "../queryKeys";

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

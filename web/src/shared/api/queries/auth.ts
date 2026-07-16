import { queryOptions } from "@tanstack/react-query";
import { apiClient, readApiData } from "../client";
import { apiQueryKeys } from "../queryKeys";

export const authQueries = {
	me: () =>
		queryOptions({
			queryKey: apiQueryKeys.auth.me(),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/auth/me", { signal })),
			retry: false,
			staleTime: 5 * 60 * 1000
		}),
	sessions: () =>
		queryOptions({
			queryKey: apiQueryKeys.auth.sessions(),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/auth/sessions", { signal }))
		}),
	apiTokens: () =>
		queryOptions({
			queryKey: apiQueryKeys.auth.apiTokens(),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/auth/api-tokens", { signal }))
		})
};

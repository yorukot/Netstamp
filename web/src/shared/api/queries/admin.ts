import { queryOptions } from "@tanstack/react-query";
import { apiClient, readApiData } from "../client";
import { apiQueryKeys } from "../queryKeys";

export const adminQueries = {
	settings: () =>
		queryOptions({
			queryKey: apiQueryKeys.admin.settings(),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/admin/settings", { signal })),
			staleTime: 30 * 1000
		}),
	systemAdmins: () =>
		queryOptions({
			queryKey: apiQueryKeys.admin.systemAdmins(),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/admin/system-admins", { signal })),
			staleTime: 30 * 1000
		}),
	users: () =>
		queryOptions({
			queryKey: apiQueryKeys.admin.users(),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/admin/users", { signal })),
			staleTime: 30 * 1000
		})
};

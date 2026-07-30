import { queryOptions } from "@tanstack/react-query";
import {
	readVersionedAdminSettings,
	type AdminAccessSettings,
	type AdminGitHubProviderSettings,
	type AdminGoogleProviderSettings,
	type AdminOIDCProviderSettings,
	type AdminSMTPSettings
} from "../adminSettings";
import { apiClient, readApiData } from "../client";
import { apiQueryKeys } from "../queryKeys";

export const adminQueries = {
	accessSettings: () =>
		queryOptions({
			queryKey: apiQueryKeys.admin.accessSettings(),
			queryFn: ({ signal }) => readVersionedAdminSettings<AdminAccessSettings>(apiClient.GET("/admin/settings/access", { signal })),
			staleTime: 30 * 1000
		}),
	smtpSettings: () =>
		queryOptions({
			queryKey: apiQueryKeys.admin.smtpSettings(),
			queryFn: ({ signal }) => readVersionedAdminSettings<AdminSMTPSettings>(apiClient.GET("/admin/settings/smtp", { signal })),
			staleTime: 30 * 1000
		}),
	oidcSettings: () =>
		queryOptions({
			queryKey: apiQueryKeys.admin.providerSettings("oidc"),
			queryFn: ({ signal }) => readVersionedAdminSettings<AdminOIDCProviderSettings>(apiClient.GET("/admin/settings/authentication-providers/oidc", { signal })),
			staleTime: 30 * 1000
		}),
	googleSettings: () =>
		queryOptions({
			queryKey: apiQueryKeys.admin.providerSettings("google"),
			queryFn: ({ signal }) => readVersionedAdminSettings<AdminGoogleProviderSettings>(apiClient.GET("/admin/settings/authentication-providers/google", { signal })),
			staleTime: 30 * 1000
		}),
	githubSettings: () =>
		queryOptions({
			queryKey: apiQueryKeys.admin.providerSettings("github"),
			queryFn: ({ signal }) => readVersionedAdminSettings<AdminGitHubProviderSettings>(apiClient.GET("/admin/settings/authentication-providers/github", { signal })),
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

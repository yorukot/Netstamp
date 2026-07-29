import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
	readAdminSettingsValidation,
	readVersionedAdminSettings,
	type AdminAccessSettings,
	type AdminAccessSettingsPatch,
	type AdminGitHubProviderSettings,
	type AdminGitHubProviderSettingsPatch,
	type AdminGoogleProviderSettings,
	type AdminGoogleProviderSettingsPatch,
	type AdminOIDCProviderSettings,
	type AdminOIDCProviderSettingsPatch,
	type AdminSMTPSettings,
	type AdminSMTPSettingsPatch
} from "../adminSettings";
import { apiClient, readApiData, readEmptyApiResponse } from "../client";
import { apiQueryKeys } from "../queryKeys";
import type { ApiAdminDataExport, GrantSystemAdminInput, SetManagedUserPasswordInput, UpdateManagedUserInput } from "../types";
import { requireWritableAccess } from "./shared";

interface VersionedSettingsMutation<TPatch> {
	body: TPatch;
	etag: string;
}

export function updateAdminAccessSettings({ body, etag }: VersionedSettingsMutation<AdminAccessSettingsPatch>) {
	requireWritableAccess();
	return readVersionedAdminSettings<AdminAccessSettings>(apiClient.PATCH("/admin/settings/access", { body, params: { header: { "If-Match": etag } } }));
}

export function updateAdminSMTPSettings({ body, etag }: VersionedSettingsMutation<AdminSMTPSettingsPatch>) {
	requireWritableAccess();
	return readVersionedAdminSettings<AdminSMTPSettings>(apiClient.PATCH("/admin/settings/smtp", { body, params: { header: { "If-Match": etag } } }));
}

export function updateAdminOIDCSettings({ body, etag }: VersionedSettingsMutation<AdminOIDCProviderSettingsPatch>) {
	requireWritableAccess();
	return readVersionedAdminSettings<AdminOIDCProviderSettings>(apiClient.PATCH("/admin/settings/authentication-providers/oidc", { body, params: { header: { "If-Match": etag } } }));
}

export function updateAdminGoogleSettings({ body, etag }: VersionedSettingsMutation<AdminGoogleProviderSettingsPatch>) {
	requireWritableAccess();
	return readVersionedAdminSettings<AdminGoogleProviderSettings>(apiClient.PATCH("/admin/settings/authentication-providers/google", { body, params: { header: { "If-Match": etag } } }));
}

export function updateAdminGitHubSettings({ body, etag }: VersionedSettingsMutation<AdminGitHubProviderSettingsPatch>) {
	requireWritableAccess();
	return readVersionedAdminSettings<AdminGitHubProviderSettings>(apiClient.PATCH("/admin/settings/authentication-providers/github", { body, params: { header: { "If-Match": etag } } }));
}

export function validateAdminOIDCSettings({ body, etag }: VersionedSettingsMutation<AdminOIDCProviderSettingsPatch>) {
	requireWritableAccess();
	return readAdminSettingsValidation(apiClient.POST("/admin/settings/authentication-providers/oidc/validate", { body, params: { header: { "If-Match": etag } } }));
}

export function validateAdminGoogleSettings({ body, etag }: VersionedSettingsMutation<AdminGoogleProviderSettingsPatch>) {
	requireWritableAccess();
	return readAdminSettingsValidation(apiClient.POST("/admin/settings/authentication-providers/google/validate", { body, params: { header: { "If-Match": etag } } }));
}

export function validateAdminGitHubSettings({ body, etag }: VersionedSettingsMutation<AdminGitHubProviderSettingsPatch>) {
	requireWritableAccess();
	return readAdminSettingsValidation(apiClient.POST("/admin/settings/authentication-providers/github/validate", { body, params: { header: { "If-Match": etag } } }));
}

export function testAdminSMTP(etag: string) {
	requireWritableAccess();
	return readAdminSettingsValidation(apiClient.POST("/admin/settings/smtp/test", { params: { header: { "If-Match": etag } } }));
}

export function grantSystemAdmin(body: GrantSystemAdminInput) {
	requireWritableAccess();
	return readApiData(apiClient.POST("/admin/system-admins", { body }));
}

export function revokeSystemAdmin(userId: string) {
	requireWritableAccess();
	return readEmptyApiResponse(apiClient.DELETE("/admin/system-admins/{user_id}", { params: { path: { user_id: userId } } }));
}

export function updateManagedUser(userId: string, body: UpdateManagedUserInput) {
	requireWritableAccess();
	return readApiData(apiClient.PATCH("/admin/users/{user_id}", { params: { path: { user_id: userId } }, body }));
}

export function setManagedUserPassword(userId: string, body: SetManagedUserPasswordInput) {
	requireWritableAccess();
	return readApiData(apiClient.POST("/admin/users/{user_id}/password", { params: { path: { user_id: userId } }, body }));
}

export function clearManagedUserPassword(userId: string) {
	requireWritableAccess();
	return readApiData(apiClient.DELETE("/admin/users/{user_id}/password", { params: { path: { user_id: userId } } }));
}

export function exportAdminData() {
	requireWritableAccess();
	return readApiData(apiClient.GET("/admin/data-export")) as Promise<ApiAdminDataExport>;
}

export function importAdminData(body: ApiAdminDataExport) {
	requireWritableAccess();
	return readApiData(apiClient.POST("/admin/data-import", { body: body as never }));
}

export function useUpdateAdminAccessSettingsMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: updateAdminAccessSettings,
		onSuccess: data => {
			queryClient.setQueryData(apiQueryKeys.admin.accessSettings(), data);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.auth.me() });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.auth.methods() });
		}
	});
}

export function useUpdateAdminSMTPSettingsMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: updateAdminSMTPSettings,
		onSuccess: data => {
			queryClient.setQueryData(apiQueryKeys.admin.smtpSettings(), data);
		}
	});
}

export function useUpdateAdminOIDCSettingsMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: updateAdminOIDCSettings,
		onSuccess: data => {
			queryClient.setQueryData(apiQueryKeys.admin.providerSettings("oidc"), data);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.auth.methods() });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.auth.sudo() });
		}
	});
}

export function useUpdateAdminGoogleSettingsMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: updateAdminGoogleSettings,
		onSuccess: data => {
			queryClient.setQueryData(apiQueryKeys.admin.providerSettings("google"), data);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.auth.methods() });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.auth.sudo() });
		}
	});
}

export function useUpdateAdminGitHubSettingsMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: updateAdminGitHubSettings,
		onSuccess: data => {
			queryClient.setQueryData(apiQueryKeys.admin.providerSettings("github"), data);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.auth.methods() });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.auth.sudo() });
		}
	});
}

export function useValidateAdminOIDCSettingsMutation() {
	return useMutation({ mutationFn: validateAdminOIDCSettings });
}

export function useValidateAdminGoogleSettingsMutation() {
	return useMutation({ mutationFn: validateAdminGoogleSettings });
}

export function useValidateAdminGitHubSettingsMutation() {
	return useMutation({ mutationFn: validateAdminGitHubSettings });
}

export function useTestAdminSMTPMutation() {
	return useMutation({ mutationFn: testAdminSMTP });
}

export function useGrantSystemAdminMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: grantSystemAdmin,
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.admin.systemAdmins() });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.admin.users() });
		}
	});
}

export function useRevokeSystemAdminMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: revokeSystemAdmin,
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.admin.systemAdmins() });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.admin.users() });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.auth.me() });
		}
	});
}

export function useUpdateManagedUserMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: ({ userId, body }: { userId: string; body: UpdateManagedUserInput }) => updateManagedUser(userId, body),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.admin.users() });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.admin.systemAdmins() });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.auth.me() });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.all });
		}
	});
}

export function useSetManagedUserPasswordMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: ({ userId, body }: { userId: string; body: SetManagedUserPasswordInput }) => setManagedUserPassword(userId, body),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.admin.users() });
		}
	});
}

export function useClearManagedUserPasswordMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: clearManagedUserPassword,
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.admin.users() });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.auth.me() });
		}
	});
}

export function useExportAdminDataMutation() {
	return useMutation({
		mutationFn: exportAdminData
	});
}

export function useImportAdminDataMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: importAdminData,
		onSuccess: () => {
			queryClient.invalidateQueries();
		}
	});
}

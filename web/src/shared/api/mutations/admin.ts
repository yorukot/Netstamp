import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
	readAdminSettings,
	readAdminSettingsValidation,
	type AdminAccessSettings,
	type AdminAccessSettingsPatch,
	type AdminUpdateSettings,
	type AdminUpdateSettingsPatch,
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
import type { GrantSystemAdminInput, SetManagedUserPasswordInput, UpdateManagedUserInput } from "../types";
import { requireWritableAccess } from "./shared";

export const updateAdminAccessSettings = (body: AdminAccessSettingsPatch) => {
	requireWritableAccess();
	return readAdminSettings<AdminAccessSettings>(apiClient.PATCH("/admin/settings/access", { body }));
};

export const updateAdminUpdateSettings = (body: AdminUpdateSettingsPatch) => {
	requireWritableAccess();
	return readAdminSettings<AdminUpdateSettings>(apiClient.PATCH("/admin/settings/updates", { body }));
};

export const updateAdminSMTPSettings = (body: AdminSMTPSettingsPatch) => {
	requireWritableAccess();
	return readAdminSettings<AdminSMTPSettings>(apiClient.PATCH("/admin/settings/smtp", { body }));
};

export const updateAdminOIDCSettings = (body: AdminOIDCProviderSettingsPatch) => {
	requireWritableAccess();
	return readAdminSettings<AdminOIDCProviderSettings>(apiClient.PATCH("/admin/settings/authentication-providers/oidc", { body }));
};

export const updateAdminGoogleSettings = (body: AdminGoogleProviderSettingsPatch) => {
	requireWritableAccess();
	return readAdminSettings<AdminGoogleProviderSettings>(apiClient.PATCH("/admin/settings/authentication-providers/google", { body }));
};

export const updateAdminGitHubSettings = (body: AdminGitHubProviderSettingsPatch) => {
	requireWritableAccess();
	return readAdminSettings<AdminGitHubProviderSettings>(apiClient.PATCH("/admin/settings/authentication-providers/github", { body }));
};

export const validateAdminOIDCSettings = (body: AdminOIDCProviderSettingsPatch) => {
	requireWritableAccess();
	return readAdminSettingsValidation(apiClient.POST("/admin/settings/authentication-providers/oidc/validate", { body }));
};

export const validateAdminGoogleSettings = (body: AdminGoogleProviderSettingsPatch) => {
	requireWritableAccess();
	return readAdminSettingsValidation(apiClient.POST("/admin/settings/authentication-providers/google/validate", { body }));
};

export const validateAdminGitHubSettings = (body: AdminGitHubProviderSettingsPatch) => {
	requireWritableAccess();
	return readAdminSettingsValidation(apiClient.POST("/admin/settings/authentication-providers/github/validate", { body }));
};

export const testAdminSMTP = () => {
	requireWritableAccess();
	return readAdminSettingsValidation(apiClient.POST("/admin/settings/smtp/test", {}));
};

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

export function useUpdateAdminUpdateSettingsMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: updateAdminUpdateSettings,
		onSuccess: data => {
			queryClient.setQueryData(apiQueryKeys.admin.updateSettings(), data);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.admin.updateStatus() });
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

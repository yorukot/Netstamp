import { useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient, readApiData, readEmptyApiResponse } from "../client";
import { apiQueryKeys } from "../queryKeys";
import type { ApiAdminDataExport, GrantSystemAdminInput, SetManagedUserPasswordInput, UpdateAdminSettingsInput, UpdateManagedUserInput } from "../types";
import { requireWritableAccess } from "./shared";

export function updateAdminSettings(body: UpdateAdminSettingsInput) {
	requireWritableAccess();
	return readApiData(apiClient.PATCH("/admin/settings", { body }));
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

export function useUpdateAdminSettingsMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: updateAdminSettings,
		onSuccess: data => {
			queryClient.setQueryData(apiQueryKeys.admin.settings(), data);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.auth.me() });
		}
	});
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

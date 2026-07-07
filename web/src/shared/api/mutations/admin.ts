import { useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient, readApiData, readEmptyApiResponse } from "../client";
import { apiQueryKeys } from "../queryKeys";
import type { GrantSystemAdminInput, UpdateAdminSettingsInput } from "../types";
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
		}
	});
}

export function useRevokeSystemAdminMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: revokeSystemAdmin,
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.admin.systemAdmins() });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.auth.me() });
		}
	});
}

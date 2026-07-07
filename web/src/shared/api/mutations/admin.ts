import { useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient, readApiData } from "../client";
import { apiQueryKeys } from "../queryKeys";
import type { UpdateAdminSettingsInput } from "../types";
import { requireWritableAccess } from "./shared";

export function updateAdminSettings(body: UpdateAdminSettingsInput) {
	requireWritableAccess();
	return readApiData(apiClient.PATCH("/admin/settings", { body }));
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

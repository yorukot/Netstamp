import { useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient, readApiData, readEmptyApiResponse } from "../client";
import { apiQueryKeys } from "../queryKeys";
import type { ChangeCurrentUserEmailInput, ChangeCurrentUserPasswordInput, UpdateCurrentUserInput } from "../types";
import { requireWritableAccess } from "./shared";

export function updateCurrentUser(body: UpdateCurrentUserInput) {
	requireWritableAccess();
	return readApiData(apiClient.PATCH("/users/me", { body }));
}

export function changeCurrentUserEmail(body: ChangeCurrentUserEmailInput) {
	requireWritableAccess();
	return readApiData(apiClient.POST("/users/me/email-change", { body }));
}

export function changeCurrentUserPassword(body: ChangeCurrentUserPasswordInput) {
	requireWritableAccess();
	return readEmptyApiResponse(apiClient.POST("/users/me/password-change", { body }));
}

export function deactivateCurrentUser() {
	requireWritableAccess();
	return readEmptyApiResponse(apiClient.POST("/users/me/deactivation"));
}

export function useUpdateCurrentUserMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: updateCurrentUser,
		onSuccess: data => {
			queryClient.setQueryData(apiQueryKeys.auth.me(), { authenticated: true, user: data.user });
		}
	});
}

export function useChangeCurrentUserEmailMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: changeCurrentUserEmail,
		onSuccess: data => {
			queryClient.setQueryData(apiQueryKeys.auth.me(), { authenticated: true, user: data.user });
		}
	});
}

export function useChangeCurrentUserPasswordMutation() {
	return useMutation({
		mutationFn: changeCurrentUserPassword
	});
}

export function useDeactivateCurrentUserMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: deactivateCurrentUser,
		onSuccess: () => {
			queryClient.clear();
		}
	});
}

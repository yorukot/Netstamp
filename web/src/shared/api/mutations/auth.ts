import { useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient, clearCSRFToken, readApiData, readEmptyApiResponse } from "../client";
import { apiQueryKeys } from "../queryKeys";
import type { ConfirmEmailVerificationInput, ConfirmPasswordResetInput, CreateEmailVerificationInput, CreatePasswordResetInput, LoginInput, RegisterInput } from "../types";
import { requireWritableAccess } from "./shared";

export function loginUser(body: LoginInput) {
	return readApiData(apiClient.POST("/auth/login", { body }));
}

export function logoutUser() {
	return readEmptyApiResponse(apiClient.POST("/auth/logout"));
}

export function registerUser(body: RegisterInput) {
	requireWritableAccess();
	return readApiData(apiClient.POST("/auth/register", { body }));
}

export function createPasswordReset(body: CreatePasswordResetInput) {
	requireWritableAccess();
	return readEmptyApiResponse(apiClient.POST("/auth/password-resets", { body }));
}

export function confirmPasswordReset(body: ConfirmPasswordResetInput) {
	requireWritableAccess();
	return readEmptyApiResponse(apiClient.PATCH("/auth/password-resets", { body }));
}

export function createEmailVerification(body: CreateEmailVerificationInput) {
	requireWritableAccess();
	return readEmptyApiResponse(apiClient.POST("/auth/email-verifications", { body }));
}

export function confirmEmailVerification(body: ConfirmEmailVerificationInput) {
	requireWritableAccess();
	return readEmptyApiResponse(apiClient.PATCH("/auth/email-verifications", { body }));
}

export function useLoginMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: loginUser,
		onSuccess: data => {
			queryClient.setQueryData(apiQueryKeys.auth.me(), { authenticated: true, user: data.user });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.all });
		}
	});
}

export function useRegisterMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: registerUser,
		onSuccess: data => {
			if ("user" in data) {
				queryClient.setQueryData(apiQueryKeys.auth.me(), { authenticated: true, user: data.user });
				queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.all });
			}
		}
	});
}

export function useCreatePasswordResetMutation() {
	return useMutation({
		mutationFn: createPasswordReset
	});
}

export function useConfirmPasswordResetMutation() {
	return useMutation({
		mutationFn: confirmPasswordReset
	});
}

export function useCreateEmailVerificationMutation() {
	return useMutation({
		mutationFn: createEmailVerification
	});
}

export function useConfirmEmailVerificationMutation() {
	return useMutation({
		mutationFn: confirmEmailVerification
	});
}

export function useLogoutMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: logoutUser,
		onSettled: () => {
			clearCSRFToken();
			queryClient.removeQueries({ queryKey: apiQueryKeys.auth.all });
			queryClient.removeQueries({ queryKey: apiQueryKeys.projects.all });
		}
	});
}

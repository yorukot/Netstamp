import { useMutation, useQueryClient, type QueryClient } from "@tanstack/react-query";
import { apiClient, clearCSRFToken, readApiData, readEmptyApiResponse } from "../client";
import { apiQueryKeys } from "../queryKeys";
import type { ConfirmEmailVerificationInput, ConfirmPasswordResetInput, CreateApiTokenInput, CreateEmailVerificationInput, CreatePasswordResetInput, LoginInput, RegisterInput } from "../types";
import { requireWritableAccess } from "./shared";

function clearAuthenticatedClientState(queryClient: QueryClient) {
	clearCSRFToken();
	queryClient.setQueryData(apiQueryKeys.auth.me(), null);
	queryClient.removeQueries({ queryKey: apiQueryKeys.auth.sessions() });
	queryClient.removeQueries({ queryKey: apiQueryKeys.auth.apiTokens() });
	queryClient.removeQueries({ queryKey: apiQueryKeys.projects.all });
}

export function loginUser(body: LoginInput) {
	return readApiData(apiClient.POST("/auth/login", { body }));
}

export function logoutUser() {
	return readEmptyApiResponse(apiClient.POST("/auth/logout"));
}

export function revokeAuthSession(sessionId: string) {
	requireWritableAccess();
	return readEmptyApiResponse(
		apiClient.DELETE("/auth/sessions/{session_id}", {
			params: { path: { session_id: sessionId } }
		})
	);
}

export function revokeAllAuthSessions() {
	requireWritableAccess();
	return readEmptyApiResponse(apiClient.DELETE("/auth/sessions"));
}

export function createAPIToken(body: CreateApiTokenInput) {
	requireWritableAccess();
	return readApiData(apiClient.POST("/auth/api-tokens", { body }));
}

export function revokeAPIToken(tokenId: string) {
	requireWritableAccess();
	return readEmptyApiResponse(apiClient.DELETE("/auth/api-tokens/{token_id}", { params: { path: { token_id: tokenId } } }));
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
		onSuccess: () => {
			clearAuthenticatedClientState(queryClient);
		}
	});
}

export function useRevokeAuthSessionMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: revokeAuthSession,
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.auth.sessions() });
		}
	});
}

export function useRevokeAllAuthSessionsMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: revokeAllAuthSessions,
		onSuccess: () => {
			clearAuthenticatedClientState(queryClient);
		}
	});
}

export function useCreateAPITokenMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: createAPIToken,
		onSuccess: () => queryClient.invalidateQueries({ queryKey: apiQueryKeys.auth.apiTokens() })
	});
}

export function useRevokeAPITokenMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: revokeAPIToken,
		onSuccess: () => queryClient.invalidateQueries({ queryKey: apiQueryKeys.auth.apiTokens() })
	});
}

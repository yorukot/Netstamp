import { useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient, readApiData, readEmptyApiResponse } from "../client";
import { apiQueryKeys } from "../queryKeys";
import type { CreateAlertRuleInput, CreateNotificationInput, UpdateAlertRuleInput, UpdateNotificationInput } from "../types";
import { mutationToastOptions, requireProjectRef, requireWritableAccess, type AppMutationOptions } from "./shared";

export function createProjectAlertRule(ref: string, body: CreateAlertRuleInput) {
	requireWritableAccess();
	return readApiData(apiClient.POST("/projects/{ref}/alerts/rules", { params: { path: { ref } }, body }));
}

export function updateProjectAlertRule(ref: string, ruleId: string, body: UpdateAlertRuleInput) {
	requireWritableAccess();
	return readApiData(apiClient.PATCH("/projects/{ref}/alerts/rules/{rule_id}", { params: { path: { ref, rule_id: ruleId } }, body }));
}

export function deleteProjectAlertRule(ref: string, ruleId: string) {
	requireWritableAccess();
	return readEmptyApiResponse(apiClient.DELETE("/projects/{ref}/alerts/rules/{rule_id}", { params: { path: { ref, rule_id: ruleId } } }));
}

export function createProjectNotification(ref: string, body: CreateNotificationInput) {
	requireWritableAccess();
	return readApiData(apiClient.POST("/projects/{ref}/alerts/notifications", { params: { path: { ref } }, body }));
}

export function updateProjectNotification(ref: string, notificationId: string, body: UpdateNotificationInput) {
	requireWritableAccess();
	return readApiData(apiClient.PATCH("/projects/{ref}/alerts/notifications/{notification_id}", { params: { path: { ref, notification_id: notificationId } }, body }));
}

export function deleteProjectNotification(ref: string, notificationId: string) {
	requireWritableAccess();
	return readEmptyApiResponse(apiClient.DELETE("/projects/{ref}/alerts/notifications/{notification_id}", { params: { path: { ref, notification_id: notificationId } } }));
}

export function testProjectNotification(ref: string, notificationId: string) {
	requireWritableAccess();
	return readApiData(apiClient.POST("/projects/{ref}/alerts/notifications/{notification_id}/test", { params: { path: { ref, notification_id: notificationId } } }));
}

export function useCreateProjectAlertRuleMutation(projectRef: string | null | undefined, options?: AppMutationOptions) {
	const queryClient = useQueryClient();

	return useMutation({
		...mutationToastOptions(options),
		mutationFn: (body: CreateAlertRuleInput) => createProjectAlertRule(requireProjectRef(projectRef), body),
		onSuccess: () => {
			const ref = requireProjectRef(projectRef);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.alertsRoot(ref) });
		}
	});
}

export function useUpdateProjectAlertRuleMutation(projectRef: string | null | undefined, options?: AppMutationOptions) {
	const queryClient = useQueryClient();

	return useMutation({
		...mutationToastOptions(options),
		mutationFn: ({ ruleId, body }: { ruleId: string; body: UpdateAlertRuleInput }) => updateProjectAlertRule(requireProjectRef(projectRef), ruleId, body),
		onSuccess: () => {
			const ref = requireProjectRef(projectRef);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.alertsRoot(ref) });
		}
	});
}

export function useDeleteProjectAlertRuleMutation(projectRef: string | null | undefined, options?: AppMutationOptions) {
	const queryClient = useQueryClient();

	return useMutation({
		...mutationToastOptions(options),
		mutationFn: (ruleId: string) => deleteProjectAlertRule(requireProjectRef(projectRef), ruleId),
		onSuccess: () => {
			const ref = requireProjectRef(projectRef);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.alertsRoot(ref) });
		}
	});
}

export function useCreateProjectNotificationMutation(projectRef: string | null | undefined, options?: AppMutationOptions) {
	const queryClient = useQueryClient();

	return useMutation({
		...mutationToastOptions(options),
		mutationFn: (body: CreateNotificationInput) => createProjectNotification(requireProjectRef(projectRef), body),
		onSuccess: () => {
			const ref = requireProjectRef(projectRef);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.alertsRoot(ref) });
		}
	});
}

export function useUpdateProjectNotificationMutation(projectRef: string | null | undefined, options?: AppMutationOptions) {
	const queryClient = useQueryClient();

	return useMutation({
		...mutationToastOptions(options),
		mutationFn: ({ notificationId, body }: { notificationId: string; body: UpdateNotificationInput }) => updateProjectNotification(requireProjectRef(projectRef), notificationId, body),
		onSuccess: () => {
			const ref = requireProjectRef(projectRef);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.alertsRoot(ref) });
		}
	});
}

export function useDeleteProjectNotificationMutation(projectRef: string | null | undefined, options?: AppMutationOptions) {
	const queryClient = useQueryClient();

	return useMutation({
		...mutationToastOptions(options),
		mutationFn: (notificationId: string) => deleteProjectNotification(requireProjectRef(projectRef), notificationId),
		onSuccess: () => {
			const ref = requireProjectRef(projectRef);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.alertsRoot(ref) });
		}
	});
}

export function useTestProjectNotificationMutation(projectRef: string | null | undefined, options?: AppMutationOptions) {
	return useMutation({
		...mutationToastOptions(options),
		mutationFn: (notificationId: string) => testProjectNotification(requireProjectRef(projectRef), notificationId)
	});
}

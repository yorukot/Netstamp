import { useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient, readApiData, readEmptyApiResponse } from "../client";
import { apiQueryKeys } from "../queryKeys";
import type { CreateCheckInput, UpdateCheckInput } from "../types";
import { mutationToastOptions, requireProjectRef, requireWritableAccess, type AppMutationOptions } from "./shared";

export class BatchCheckDeleteError extends Error {
	readonly failedIds: string[];
	readonly succeededIds: string[];

	constructor(succeededIds: string[], failedIds: string[]) {
		super(`${failedIds.length} check delete ${failedIds.length === 1 ? "request" : "requests"} failed.`);
		this.name = "BatchCheckDeleteError";
		this.failedIds = failedIds;
		this.succeededIds = succeededIds;
	}
}

export function createProjectCheck(ref: string, body: CreateCheckInput) {
	requireWritableAccess();
	return readApiData(apiClient.POST("/projects/{ref}/checks", { params: { path: { ref } }, body }));
}

export function updateProjectCheck(ref: string, checkId: string, body: UpdateCheckInput) {
	requireWritableAccess();
	return readApiData(apiClient.PATCH("/projects/{ref}/checks/{check_id}", { params: { path: { ref, check_id: checkId } }, body }));
}

export function deleteProjectCheck(ref: string, checkId: string) {
	requireWritableAccess();
	return readEmptyApiResponse(apiClient.DELETE("/projects/{ref}/checks/{check_id}", { params: { path: { ref, check_id: checkId } } }));
}

export async function deleteProjectChecks(ref: string, checkIds: string[]) {
	requireWritableAccess();
	const results = await Promise.allSettled(
		checkIds.map(async checkId => {
			await deleteProjectCheck(ref, checkId);
			return checkId;
		})
	);
	const succeededIds: string[] = [];
	const failedIds: string[] = [];

	for (let index = 0; index < results.length; index++) {
		const result = results[index];
		const checkId = checkIds[index];
		if (!checkId) {
			continue;
		}
		if (result?.status === "fulfilled") {
			succeededIds.push(checkId);
		} else {
			failedIds.push(checkId);
		}
	}

	if (failedIds.length) {
		throw new BatchCheckDeleteError(succeededIds, failedIds);
	}

	return { failedIds, succeededIds };
}

export function useCreateProjectCheckMutation(projectRef: string | null | undefined) {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (body: CreateCheckInput) => createProjectCheck(requireProjectRef(projectRef), body),
		onSuccess: data => {
			const ref = requireProjectRef(projectRef);
			queryClient.setQueryData(apiQueryKeys.projects.checkDetail(ref, data.check.id), data);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.assignmentsRoot(ref) });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.checks(ref) });
		}
	});
}

export function useUpdateProjectCheckMutation(projectRef: string | null | undefined) {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: ({ checkId, body }: { checkId: string; body: UpdateCheckInput }) => updateProjectCheck(requireProjectRef(projectRef), checkId, body),
		onSuccess: data => {
			const ref = requireProjectRef(projectRef);
			queryClient.setQueryData(apiQueryKeys.projects.checkDetail(ref, data.check.id), data);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.assignmentsRoot(ref) });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.checks(ref) });
		}
	});
}

export function useDeleteProjectCheckMutation(projectRef: string | null | undefined) {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (checkId: string) => deleteProjectCheck(requireProjectRef(projectRef), checkId),
		onSuccess: (_data, checkId) => {
			const ref = requireProjectRef(projectRef);
			queryClient.removeQueries({ queryKey: apiQueryKeys.projects.checkDetail(ref, checkId) });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.assignmentsRoot(ref) });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.checks(ref) });
		}
	});
}

export function useDeleteProjectChecksMutation(projectRef: string | null | undefined, options?: AppMutationOptions) {
	const queryClient = useQueryClient();

	return useMutation({
		...mutationToastOptions(options),
		mutationFn: (checkIds: string[]) => deleteProjectChecks(requireProjectRef(projectRef), checkIds),
		onSettled: (_data, error, checkIds) => {
			const ref = requireProjectRef(projectRef);
			const removedIds = error instanceof BatchCheckDeleteError ? error.succeededIds : (checkIds ?? []);
			for (const checkId of removedIds) {
				queryClient.removeQueries({ queryKey: apiQueryKeys.projects.checkDetail(ref, checkId) });
			}
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.assignmentsRoot(ref) });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.checks(ref) });
		}
	});
}

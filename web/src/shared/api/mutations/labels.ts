import { useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient, readApiData, readEmptyApiResponse } from "../client";
import { apiQueryKeys } from "../queryKeys";
import type { CreateLabelInput, UpdateLabelInput } from "../types";
import { requireProjectRef, requireWritableAccess } from "./shared";

type SaveProjectLabelVariables = { labelId?: string; body: CreateLabelInput };

export function createProjectLabel(ref: string, body: CreateLabelInput) {
	requireWritableAccess();
	return readApiData(apiClient.POST("/projects/{ref}/labels", { params: { path: { ref } }, body }));
}

export function updateProjectLabel(ref: string, labelId: string, body: UpdateLabelInput) {
	requireWritableAccess();
	return readApiData(apiClient.PATCH("/projects/{ref}/labels/{label_id}", { params: { path: { ref, label_id: labelId } }, body }));
}

export function deleteProjectLabel(ref: string, labelId: string) {
	requireWritableAccess();
	return readEmptyApiResponse(apiClient.DELETE("/projects/{ref}/labels/{label_id}", { params: { path: { ref, label_id: labelId } } }));
}

export function useSaveProjectLabelMutation(projectRef: string | null | undefined) {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (variables: SaveProjectLabelVariables) => {
			const ref = requireProjectRef(projectRef);

			return variables.labelId ? updateProjectLabel(ref, variables.labelId, variables.body) : createProjectLabel(ref, variables.body);
		},
		onSuccess: () => {
			const ref = requireProjectRef(projectRef);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.assignmentsRoot(ref) });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.labels(ref) });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.probes(ref) });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.checks(ref) });
		}
	});
}

export function useDeleteProjectLabelMutation(projectRef: string | null | undefined) {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (labelId: string) => deleteProjectLabel(requireProjectRef(projectRef), labelId),
		onSuccess: () => {
			const ref = requireProjectRef(projectRef);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.assignmentsRoot(ref) });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.labels(ref) });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.probes(ref) });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.checks(ref) });
		}
	});
}

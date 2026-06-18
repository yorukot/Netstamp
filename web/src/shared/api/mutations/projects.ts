import { useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient, readApiData, readEmptyApiResponse } from "../client";
import { apiQueryKeys } from "../queryKeys";
import type { CreateProjectInput, UpdateProjectInput } from "../types";
import { mutationToastOptions, projectCacheRef, requireProjectRef, requireWritableAccess, type AppMutationOptions } from "./shared";

function cacheCreatedProject(queryClient: ReturnType<typeof useQueryClient>, data: Awaited<ReturnType<typeof createProject>>) {
	queryClient.setQueryData(apiQueryKeys.projects.detail(projectCacheRef(data.project)), data);
	queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.list() });
}

export function createProject(body: CreateProjectInput) {
	requireWritableAccess();
	return readApiData(apiClient.POST("/projects", { body }));
}

export function updateProject(ref: string, body: UpdateProjectInput) {
	requireWritableAccess();
	return readApiData(apiClient.PATCH("/projects/{ref}", { params: { path: { ref } }, body }));
}

export function deleteProject(ref: string) {
	requireWritableAccess();
	return readEmptyApiResponse(apiClient.DELETE("/projects/{ref}", { params: { path: { ref } } }));
}

export function useCreateProjectMutation(options?: AppMutationOptions) {
	const queryClient = useQueryClient();

	return useMutation({
		...mutationToastOptions(options),
		mutationFn: createProject,
		onSuccess: data => cacheCreatedProject(queryClient, data)
	});
}

export function useUpdateProjectMutation(projectRef: string | null | undefined) {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (body: UpdateProjectInput) => updateProject(requireProjectRef(projectRef), body),
		onSuccess: data => {
			const ref = requireProjectRef(projectRef);
			queryClient.setQueryData(apiQueryKeys.projects.detail(ref), data);
			queryClient.setQueryData(apiQueryKeys.projects.detail(projectCacheRef(data.project)), data);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.list() });
		}
	});
}

export function useDeleteProjectMutation(projectRef: string | null | undefined) {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: () => deleteProject(requireProjectRef(projectRef)),
		onSuccess: () => {
			queryClient.removeQueries({ queryKey: apiQueryKeys.projects.detail(requireProjectRef(projectRef)) });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.list() });
		}
	});
}

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient, readApiData } from "../client";
import { apiQueryKeys } from "../queryKeys";
import type { CreatePublicStatusPageInput, UpdatePublicStatusPageInput } from "../types";
import { mutationToastOptions, requireProjectRef, requireWritableAccess, type AppMutationOptions } from "./shared";

type UpdatePublicStatusPageVariables = { pageId: string; previousSlug?: string; body: UpdatePublicStatusPageInput };

export function createProjectPublicStatusPage(ref: string, body: CreatePublicStatusPageInput) {
	requireWritableAccess();
	return readApiData(apiClient.POST("/projects/{ref}/status-pages", { params: { path: { ref } }, body }));
}

export function updateProjectPublicStatusPage(ref: string, pageId: string, body: UpdatePublicStatusPageInput) {
	requireWritableAccess();
	return readApiData(apiClient.PATCH("/projects/{ref}/status-pages/{page_id}", { params: { path: { ref, page_id: pageId } }, body }));
}

export function useCreatePublicStatusPageMutation(projectRef: string | null | undefined, options?: AppMutationOptions) {
	const queryClient = useQueryClient();

	return useMutation({
		...mutationToastOptions(options),
		mutationFn: (body: CreatePublicStatusPageInput) => createProjectPublicStatusPage(requireProjectRef(projectRef), body),
		onSuccess: data => {
			const ref = requireProjectRef(projectRef);
			queryClient.setQueryData(apiQueryKeys.projects.statusPageDetail(ref, data.page.id), data);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.statusPages(ref) });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.publicStatus.pageRoot(data.page.slug) });
		}
	});
}

export function useUpdatePublicStatusPageMutation(projectRef: string | null | undefined, options?: AppMutationOptions) {
	const queryClient = useQueryClient();

	return useMutation({
		...mutationToastOptions(options),
		mutationFn: ({ pageId, body }: UpdatePublicStatusPageVariables) => updateProjectPublicStatusPage(requireProjectRef(projectRef), pageId, body),
		onSuccess: (data, variables) => {
			const ref = requireProjectRef(projectRef);
			queryClient.setQueryData(apiQueryKeys.projects.statusPageDetail(ref, data.page.id), data);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.statusPages(ref) });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.publicStatus.pageRoot(data.page.slug) });
			if (variables.previousSlug && variables.previousSlug !== data.page.slug) {
				queryClient.invalidateQueries({ queryKey: apiQueryKeys.publicStatus.pageRoot(variables.previousSlug) });
			}
		}
	});
}

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient, readApiData, readEmptyApiResponse } from "../client";
import { apiQueryKeys } from "../queryKeys";
import type { CreatePublicStatusElementInput, CreatePublicStatusPageInput, UpdatePublicStatusElementInput, UpdatePublicStatusPageInput } from "../types";
import { mutationToastOptions, requireProjectRef, requireWritableAccess, type AppMutationOptions } from "./shared";

type UpdatePublicStatusPageVariables = { pageId: string; previousSlug?: string; body: UpdatePublicStatusPageInput };
type CreatePublicStatusElementVariables = { pageId: string; body: CreatePublicStatusElementInput };
type UpdatePublicStatusElementVariables = { pageId: string; elementId: string; body: UpdatePublicStatusElementInput };
type DeletePublicStatusElementVariables = { pageId: string; elementId: string };

export function createProjectPublicStatusPage(ref: string, body: CreatePublicStatusPageInput) {
	requireWritableAccess();
	return readApiData(apiClient.POST("/projects/{ref}/status-pages", { params: { path: { ref } }, body }));
}

export function updateProjectPublicStatusPage(ref: string, pageId: string, body: UpdatePublicStatusPageInput) {
	requireWritableAccess();
	return readApiData(apiClient.PATCH("/projects/{ref}/status-pages/{page_id}", { params: { path: { ref, page_id: pageId } }, body }));
}

export function deleteProjectPublicStatusPage(ref: string, pageId: string) {
	requireWritableAccess();
	return readEmptyApiResponse(apiClient.DELETE("/projects/{ref}/status-pages/{page_id}", { params: { path: { ref, page_id: pageId } } }));
}

export function createProjectPublicStatusElement(ref: string, pageId: string, body: CreatePublicStatusElementInput) {
	requireWritableAccess();
	return readApiData(apiClient.POST("/projects/{ref}/status-pages/{page_id}/elements", { params: { path: { ref, page_id: pageId } }, body }));
}

export function updateProjectPublicStatusElement(ref: string, pageId: string, elementId: string, body: UpdatePublicStatusElementInput) {
	requireWritableAccess();
	return readApiData(apiClient.PATCH("/projects/{ref}/status-pages/{page_id}/elements/{element_id}", { params: { path: { ref, page_id: pageId, element_id: elementId } }, body }));
}

export function deleteProjectPublicStatusElement(ref: string, pageId: string, elementId: string) {
	requireWritableAccess();
	return readEmptyApiResponse(apiClient.DELETE("/projects/{ref}/status-pages/{page_id}/elements/{element_id}", { params: { path: { ref, page_id: pageId, element_id: elementId } } }));
}

export function useCreatePublicStatusPageMutation(projectRef: string | null | undefined, options?: AppMutationOptions) {
	const queryClient = useQueryClient();

	return useMutation({
		...mutationToastOptions(options),
		mutationFn: (body: CreatePublicStatusPageInput) => createProjectPublicStatusPage(requireProjectRef(projectRef), body),
		onSuccess: data => {
			const ref = requireProjectRef(projectRef);
			queryClient.setQueryData(apiQueryKeys.projects.statusPageDetail(ref, data.page.id), { page: data.page, elements: [] });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.statusPagesRoot(ref) });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.publicStatus.detailRoot(data.page.slug) });
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
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.statusPagesRoot(ref) });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.statusPageDetail(ref, data.page.id) });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.publicStatus.detailRoot(data.page.slug) });
			if (variables.previousSlug && variables.previousSlug !== data.page.slug) {
				queryClient.invalidateQueries({ queryKey: apiQueryKeys.publicStatus.detailRoot(variables.previousSlug) });
			}
		}
	});
}

export function useDeletePublicStatusPageMutation(projectRef: string | null | undefined, options?: AppMutationOptions) {
	const queryClient = useQueryClient();

	return useMutation({
		...mutationToastOptions(options),
		mutationFn: ({ pageId }: { pageId: string; slug?: string }) => deleteProjectPublicStatusPage(requireProjectRef(projectRef), pageId),
		onSuccess: (_data, variables) => {
			const ref = requireProjectRef(projectRef);
			queryClient.removeQueries({ queryKey: apiQueryKeys.projects.statusPageDetail(ref, variables.pageId) });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.statusPagesRoot(ref) });
			if (variables.slug) {
				queryClient.invalidateQueries({ queryKey: apiQueryKeys.publicStatus.detailRoot(variables.slug) });
			}
		}
	});
}

export function useCreatePublicStatusElementMutation(projectRef: string | null | undefined, options?: AppMutationOptions) {
	const queryClient = useQueryClient();

	return useMutation({
		...mutationToastOptions(options),
		mutationFn: ({ pageId, body }: CreatePublicStatusElementVariables) => createProjectPublicStatusElement(requireProjectRef(projectRef), pageId, body),
		onSuccess: (_data, variables) => {
			const ref = requireProjectRef(projectRef);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.statusPageDetail(ref, variables.pageId) });
		}
	});
}

export function useUpdatePublicStatusElementMutation(projectRef: string | null | undefined, options?: AppMutationOptions) {
	const queryClient = useQueryClient();

	return useMutation({
		...mutationToastOptions(options),
		mutationFn: ({ pageId, elementId, body }: UpdatePublicStatusElementVariables) => updateProjectPublicStatusElement(requireProjectRef(projectRef), pageId, elementId, body),
		onSuccess: (_data, variables) => {
			const ref = requireProjectRef(projectRef);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.statusPageDetail(ref, variables.pageId) });
		}
	});
}

export function useDeletePublicStatusElementMutation(projectRef: string | null | undefined, options?: AppMutationOptions) {
	const queryClient = useQueryClient();

	return useMutation({
		...mutationToastOptions(options),
		mutationFn: ({ pageId, elementId }: DeletePublicStatusElementVariables) => deleteProjectPublicStatusElement(requireProjectRef(projectRef), pageId, elementId),
		onSuccess: (_data, variables) => {
			const ref = requireProjectRef(projectRef);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.statusPageDetail(ref, variables.pageId) });
		}
	});
}

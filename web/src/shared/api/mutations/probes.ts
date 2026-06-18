import { useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient, readApiData, readEmptyApiResponse } from "../client";
import { apiQueryKeys } from "../queryKeys";
import type { CreateProbeInput, SelectorPreviewInput, UpdateProbeInput } from "../types";
import { mutationToastOptions, requireProjectRef, requireWritableAccess, type AppMutationOptions } from "./shared";

export function createProjectProbe(ref: string, body: CreateProbeInput) {
	requireWritableAccess();
	return readApiData(apiClient.POST("/projects/{ref}/probes", { params: { path: { ref } }, body }));
}

export function updateProjectProbe(ref: string, probeId: string, body: UpdateProbeInput) {
	requireWritableAccess();
	return readApiData(apiClient.PATCH("/projects/{ref}/probes/{probe_id}", { params: { path: { ref, probe_id: probeId } }, body }));
}

export function deleteProjectProbe(ref: string, probeId: string) {
	requireWritableAccess();
	return readEmptyApiResponse(apiClient.DELETE("/projects/{ref}/probes/{probe_id}", { params: { path: { ref, probe_id: probeId } } }));
}

export function rotateProjectProbeSecret(ref: string, probeId: string) {
	requireWritableAccess();
	return readApiData(apiClient.POST("/projects/{ref}/probes/{probe_id}/secret-rotations", { params: { path: { ref, probe_id: probeId } } }));
}

export function previewProjectSelector(ref: string, body: SelectorPreviewInput) {
	return readApiData(apiClient.POST("/projects/{ref}/selector-previews", { params: { path: { ref } }, body }));
}

export function useCreateProjectProbeMutation(projectRef: string | null | undefined) {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (body: CreateProbeInput) => createProjectProbe(requireProjectRef(projectRef), body),
		onSuccess: data => {
			const ref = requireProjectRef(projectRef);
			queryClient.setQueryData(apiQueryKeys.projects.probeDetail(ref, data.probe.id), data);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.probes(ref) });
		}
	});
}

export function useUpdateProjectProbeMutation(projectRef: string | null | undefined, options?: AppMutationOptions) {
	const queryClient = useQueryClient();

	return useMutation({
		...mutationToastOptions(options),
		mutationFn: ({ probeId, body }: { probeId: string; body: UpdateProbeInput }) => updateProjectProbe(requireProjectRef(projectRef), probeId, body),
		onSuccess: data => {
			const ref = requireProjectRef(projectRef);
			queryClient.setQueryData(apiQueryKeys.projects.probeDetail(ref, data.probe.id), data);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.probes(ref) });
		}
	});
}

export function useDeleteProjectProbeMutation(projectRef: string | null | undefined) {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (probeId: string) => deleteProjectProbe(requireProjectRef(projectRef), probeId),
		onSuccess: (_data, probeId) => {
			const ref = requireProjectRef(projectRef);
			queryClient.removeQueries({ queryKey: apiQueryKeys.projects.probeDetail(ref, probeId) });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.probes(ref) });
		}
	});
}

export function useRotateProjectProbeSecretMutation(projectRef: string | null | undefined) {
	return useMutation({
		mutationFn: (probeId: string) => rotateProjectProbeSecret(requireProjectRef(projectRef), probeId)
	});
}

export function usePreviewProjectSelectorMutation(projectRef: string | null | undefined) {
	return useMutation({
		mutationFn: (body: SelectorPreviewInput) => previewProjectSelector(requireProjectRef(projectRef), body)
	});
}

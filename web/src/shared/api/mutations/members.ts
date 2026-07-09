import { useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient, readApiData, readEmptyApiResponse } from "../client";
import { apiQueryKeys } from "../queryKeys";
import type { ApiMember, ApiProjectInvite, CreateProjectInviteInput, ProjectMemberRole } from "../types";
import { mutationToastOptions, projectCacheRef, requireProjectRef, requireWritableAccess, type AppMutationOptions } from "./shared";

type ProjectMembersCache = { members: ApiMember[] };
type ProjectInvitesCache = { invites: ApiProjectInvite[] };

function cacheCreatedProjectInvite(queryClient: ReturnType<typeof useQueryClient>, ref: string, data: Awaited<ReturnType<typeof createProjectInvite>>) {
	queryClient.setQueryData<ProjectInvitesCache | undefined>(apiQueryKeys.projects.invites(ref), current =>
		current ? { ...current, invites: [data.invite, ...current.invites.filter(invite => invite.id !== data.invite.id)] } : current
	);
	queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.invites(ref) });
}

export function createProjectInvite(ref: string, body: CreateProjectInviteInput) {
	requireWritableAccess();
	return readApiData(apiClient.POST("/projects/{ref}/invites", { params: { path: { ref } }, body }));
}

export function cancelProjectInvite(ref: string, inviteId: string) {
	requireWritableAccess();
	return readEmptyApiResponse(apiClient.DELETE("/projects/{ref}/invites/{invite_id}", { params: { path: { ref, invite_id: inviteId } } }));
}

export function removeProjectMember(ref: string, userId: string) {
	requireWritableAccess();
	return readEmptyApiResponse(apiClient.DELETE("/projects/{ref}/members/{user_id}", { params: { path: { ref, user_id: userId } } }));
}

export function updateProjectMemberRole(ref: string, userId: string, role: ProjectMemberRole) {
	requireWritableAccess();
	return readApiData(apiClient.PATCH("/projects/{ref}/members/{user_id}", { params: { path: { ref, user_id: userId } }, body: { role } }));
}

export function acceptProjectInvite(inviteId: string) {
	requireWritableAccess();
	return readApiData(apiClient.POST("/me/project-invites/{invite_id}/accept", { params: { path: { invite_id: inviteId } } }));
}

export function rejectProjectInvite(inviteId: string) {
	requireWritableAccess();
	return readApiData(apiClient.POST("/me/project-invites/{invite_id}/reject", { params: { path: { invite_id: inviteId } } }));
}

export function useCreateProjectInviteMutation(projectRef: string | null | undefined) {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (body: CreateProjectInviteInput) => createProjectInvite(requireProjectRef(projectRef), body),
		onSuccess: data => {
			const ref = requireProjectRef(projectRef);
			cacheCreatedProjectInvite(queryClient, ref, data);
		}
	});
}

export function useCreateProjectInviteForRefMutation(options?: AppMutationOptions) {
	const queryClient = useQueryClient();

	return useMutation({
		...mutationToastOptions(options),
		mutationFn: ({ projectRef, body }: { projectRef: string; body: CreateProjectInviteInput }) => createProjectInvite(projectRef, body),
		onSuccess: (data, variables) => cacheCreatedProjectInvite(queryClient, variables.projectRef, data)
	});
}

export function useCancelProjectInviteMutation(projectRef: string | null | undefined) {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (inviteId: string) => cancelProjectInvite(requireProjectRef(projectRef), inviteId),
		onSuccess: (_data, inviteId) => {
			const ref = requireProjectRef(projectRef);
			queryClient.setQueryData<ProjectInvitesCache | undefined>(apiQueryKeys.projects.invites(ref), data =>
				data ? { ...data, invites: data.invites.filter(invite => invite.id !== inviteId) } : data
			);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.currentUserInvites() });
		}
	});
}

export function useAcceptProjectInviteMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: acceptProjectInvite,
		onSuccess: data => {
			const projectRef = projectCacheRef(data.invite.project);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.currentUserInvites() });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.list() });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.detail(projectRef) });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.members(projectRef) });
		}
	});
}

export function useRejectProjectInviteMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: rejectProjectInvite,
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.currentUserInvites() });
		}
	});
}

export function useRemoveProjectMemberMutation(projectRef: string | null | undefined) {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (userId: string) => removeProjectMember(requireProjectRef(projectRef), userId),
		onSuccess: (_data, userId) => {
			const ref = requireProjectRef(projectRef);
			queryClient.setQueryData<ProjectMembersCache | undefined>(apiQueryKeys.projects.members(ref), data =>
				data ? { ...data, members: data.members.filter(member => member.userId !== userId) } : data
			);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.list() });
		}
	});
}

export function useUpdateProjectMemberRoleMutation(projectRef: string | null | undefined) {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: ({ userId, role }: { userId: string; role: ProjectMemberRole }) => updateProjectMemberRole(requireProjectRef(projectRef), userId, role),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.members(requireProjectRef(projectRef)) });
		}
	});
}

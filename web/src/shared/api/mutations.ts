import { useMutation, useQueryClient } from "@tanstack/react-query";
import { ApiError, apiClient, readApiData, readEmptyApiResponse } from "./client";
import { apiQueryKeys } from "./queryKeys";
import type {
	ApiMember,
	ApiProjectInvite,
	ChangeCurrentUserEmailInput,
	ChangeCurrentUserPasswordInput,
	CreateCheckInput,
	CreateLabelInput,
	CreateProbeInput,
	CreateProjectInput,
	CreateProjectInviteInput,
	LoginInput,
	ProjectMemberRole,
	RegisterInput,
	SelectorPreviewInput,
	UpdateCheckInput,
	UpdateCurrentUserInput,
	UpdateLabelInput,
	UpdateProbeInput,
	UpdateProjectInput
} from "./types";

type SaveProjectLabelVariables = { labelId?: string; body: CreateLabelInput };
type ProjectMembersCache = { members: ApiMember[] };
type ProjectInvitesCache = { invites: ApiProjectInvite[] };

function requireProjectRef(projectRef: string | null | undefined) {
	if (!projectRef) {
		throw new ApiError("No project selected.", 400);
	}

	return projectRef;
}

function projectCacheRef(project: { id: string; slug?: string }) {
	return project.slug || project.id;
}

export function loginUser(body: LoginInput) {
	return readApiData(apiClient.POST("/auth/login", { body }));
}

export function logoutUser() {
	return readEmptyApiResponse(apiClient.POST("/auth/logout"));
}

export function registerUser(body: RegisterInput) {
	return readApiData(apiClient.POST("/auth/register", { body }));
}

export function createProject(body: CreateProjectInput) {
	return readApiData(apiClient.POST("/projects", { body }));
}

export function updateProject(ref: string, body: UpdateProjectInput) {
	return readApiData(apiClient.PATCH("/projects/{ref}", { params: { path: { ref } }, body }));
}

export function deleteProject(ref: string) {
	return readEmptyApiResponse(apiClient.DELETE("/projects/{ref}", { params: { path: { ref } } }));
}

export function createProjectProbe(ref: string, body: CreateProbeInput) {
	return readApiData(apiClient.POST("/projects/{ref}/probes", { params: { path: { ref } }, body }));
}

export function updateProjectProbe(ref: string, probeId: string, body: UpdateProbeInput) {
	return readApiData(apiClient.PATCH("/projects/{ref}/probes/{probe_id}", { params: { path: { ref, probe_id: probeId } }, body }));
}

export function deleteProjectProbe(ref: string, probeId: string) {
	return readEmptyApiResponse(apiClient.DELETE("/projects/{ref}/probes/{probe_id}", { params: { path: { ref, probe_id: probeId } } }));
}

export function createProjectCheck(ref: string, body: CreateCheckInput) {
	return readApiData(apiClient.POST("/projects/{ref}/checks", { params: { path: { ref } }, body }));
}

export function updateProjectCheck(ref: string, checkId: string, body: UpdateCheckInput) {
	return readApiData(apiClient.PATCH("/projects/{ref}/checks/{check_id}", { params: { path: { ref, check_id: checkId } }, body }));
}

export function deleteProjectCheck(ref: string, checkId: string) {
	return readEmptyApiResponse(apiClient.DELETE("/projects/{ref}/checks/{check_id}", { params: { path: { ref, check_id: checkId } } }));
}

export function createProjectLabel(ref: string, body: CreateLabelInput) {
	return readApiData(apiClient.POST("/projects/{ref}/labels", { params: { path: { ref } }, body }));
}

export function updateProjectLabel(ref: string, labelId: string, body: UpdateLabelInput) {
	return readApiData(apiClient.PATCH("/projects/{ref}/labels/{label_id}", { params: { path: { ref, label_id: labelId } }, body }));
}

export function deleteProjectLabel(ref: string, labelId: string) {
	return readEmptyApiResponse(apiClient.DELETE("/projects/{ref}/labels/{label_id}", { params: { path: { ref, label_id: labelId } } }));
}

export function createProjectInvite(ref: string, body: CreateProjectInviteInput) {
	return readApiData(apiClient.POST("/projects/{ref}/invites", { params: { path: { ref } }, body }));
}

export function removeProjectMember(ref: string, userId: string) {
	return readEmptyApiResponse(apiClient.DELETE("/projects/{ref}/members/{user_id}", { params: { path: { ref, user_id: userId } } }));
}

export function updateProjectMemberRole(ref: string, userId: string, role: ProjectMemberRole) {
	return readApiData(apiClient.PATCH("/projects/{ref}/members/{user_id}", { params: { path: { ref, user_id: userId } }, body: { role } }));
}

export function acceptProjectInvite(inviteId: string) {
	return readApiData(apiClient.POST("/me/project-invites/{invite_id}/accept", { params: { path: { invite_id: inviteId } } }));
}

export function rejectProjectInvite(inviteId: string) {
	return readApiData(apiClient.POST("/me/project-invites/{invite_id}/reject", { params: { path: { invite_id: inviteId } } }));
}

export function rotateProjectProbeSecret(ref: string, probeId: string) {
	return readApiData(apiClient.POST("/projects/{ref}/probes/{probe_id}/secret-rotations", { params: { path: { ref, probe_id: probeId } } }));
}

export function previewProjectSelector(ref: string, body: SelectorPreviewInput) {
	return readApiData(apiClient.POST("/projects/{ref}/selector-previews", { params: { path: { ref } }, body }));
}

export function updateCurrentUser(body: UpdateCurrentUserInput) {
	return readApiData(apiClient.PATCH("/users/me", { body }));
}

export function changeCurrentUserEmail(body: ChangeCurrentUserEmailInput) {
	return readApiData(apiClient.POST("/users/me/email-change", { body }));
}

export function changeCurrentUserPassword(body: ChangeCurrentUserPasswordInput) {
	return readEmptyApiResponse(apiClient.POST("/users/me/password-change", { body }));
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
			queryClient.setQueryData(apiQueryKeys.auth.me(), { authenticated: true, user: data.user });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.all });
		}
	});
}

export function useLogoutMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: logoutUser,
		onSettled: () => {
			queryClient.removeQueries({ queryKey: apiQueryKeys.auth.all });
			queryClient.removeQueries({ queryKey: apiQueryKeys.projects.all });
		}
	});
}

export function useCreateProjectMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: createProject,
		onSuccess: data => {
			queryClient.setQueryData(apiQueryKeys.projects.detail(projectCacheRef(data.project)), data);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.list() });
		}
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

export function useUpdateProjectProbeMutation(projectRef: string | null | undefined) {
	const queryClient = useQueryClient();

	return useMutation({
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

export function useCreateProjectInviteMutation(projectRef: string | null | undefined) {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (body: CreateProjectInviteInput) => createProjectInvite(requireProjectRef(projectRef), body),
		onSuccess: data => {
			const ref = requireProjectRef(projectRef);
			queryClient.setQueryData<ProjectInvitesCache | undefined>(apiQueryKeys.projects.invites(ref), current =>
				current ? { ...current, invites: [data.invite, ...current.invites.filter(invite => invite.id !== data.invite.id)] } : current
			);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.invites(ref) });
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

export function useUpdateCurrentUserMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: updateCurrentUser,
		onSuccess: data => {
			queryClient.setQueryData(apiQueryKeys.auth.me(), { authenticated: true, user: data.user });
		}
	});
}

export function useChangeCurrentUserEmailMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: changeCurrentUserEmail,
		onSuccess: data => {
			queryClient.setQueryData(apiQueryKeys.auth.me(), { authenticated: true, user: data.user });
		}
	});
}

export function useChangeCurrentUserPasswordMutation() {
	return useMutation({
		mutationFn: changeCurrentUserPassword
	});
}

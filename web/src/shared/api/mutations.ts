import { useMutation, useQueryClient } from "@tanstack/react-query";
import { demoMode } from "../config/features";
import { ApiError, apiClient, readApiData, readEmptyApiResponse } from "./client";
import { apiQueryKeys } from "./queryKeys";
import type {
	ApiMember,
	ApiProjectInvite,
	ChangeCurrentUserEmailInput,
	ChangeCurrentUserPasswordInput,
	CreateAlertRuleInput,
	CreateCheckInput,
	CreateLabelInput,
	CreateNotificationInput,
	CreateProbeInput,
	CreateProjectInput,
	CreateProjectInviteInput,
	LoginInput,
	ProjectMemberRole,
	RegisterInput,
	SelectorPreviewInput,
	UpdateAlertRuleInput,
	UpdateCheckInput,
	UpdateCurrentUserInput,
	UpdateLabelInput,
	UpdateNotificationInput,
	UpdateProbeInput,
	UpdateProjectInput
} from "./types";

type SaveProjectLabelVariables = { labelId?: string; body: CreateLabelInput };
type ProjectMembersCache = { members: ApiMember[] };
type ProjectInvitesCache = { invites: ApiProjectInvite[] };

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

function requireProjectRef(projectRef: string | null | undefined) {
	if (!projectRef) {
		throw new ApiError("No project selected.", 400);
	}

	return projectRef;
}

function projectCacheRef(project: { id: string; slug?: string }) {
	return project.slug || project.id;
}

function cacheCreatedProject(queryClient: ReturnType<typeof useQueryClient>, data: Awaited<ReturnType<typeof createProject>>) {
	queryClient.setQueryData(apiQueryKeys.projects.detail(projectCacheRef(data.project)), data);
	queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.list() });
}

function cacheCreatedProjectInvite(queryClient: ReturnType<typeof useQueryClient>, ref: string, data: Awaited<ReturnType<typeof createProjectInvite>>) {
	queryClient.setQueryData<ProjectInvitesCache | undefined>(apiQueryKeys.projects.invites(ref), current =>
		current ? { ...current, invites: [data.invite, ...current.invites.filter(invite => invite.id !== data.invite.id)] } : current
	);
	queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.invites(ref) });
}

function requireWritableAccess() {
	if (demoMode) {
		throw new ApiError("Demo mode is read-only.", 403);
	}
}

export function loginUser(body: LoginInput) {
	return readApiData(apiClient.POST("/auth/login", { body }));
}

export function logoutUser() {
	return readEmptyApiResponse(apiClient.POST("/auth/logout"));
}

export function registerUser(body: RegisterInput) {
	requireWritableAccess();
	return readApiData(apiClient.POST("/auth/register", { body }));
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

export function createProjectInvite(ref: string, body: CreateProjectInviteInput) {
	requireWritableAccess();
	return readApiData(apiClient.POST("/projects/{ref}/invites", { params: { path: { ref } }, body }));
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

export function rotateProjectProbeSecret(ref: string, probeId: string) {
	requireWritableAccess();
	return readApiData(apiClient.POST("/projects/{ref}/probes/{probe_id}/secret-rotations", { params: { path: { ref, probe_id: probeId } } }));
}

export function previewProjectSelector(ref: string, body: SelectorPreviewInput) {
	return readApiData(apiClient.POST("/projects/{ref}/selector-previews", { params: { path: { ref } }, body }));
}

export function updateCurrentUser(body: UpdateCurrentUserInput) {
	requireWritableAccess();
	return readApiData(apiClient.PATCH("/users/me", { body }));
}

export function changeCurrentUserEmail(body: ChangeCurrentUserEmailInput) {
	requireWritableAccess();
	return readApiData(apiClient.POST("/users/me/email-change", { body }));
}

export function changeCurrentUserPassword(body: ChangeCurrentUserPasswordInput) {
	requireWritableAccess();
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

export function useDeleteProjectChecksMutation(projectRef: string | null | undefined) {
	const queryClient = useQueryClient();

	return useMutation({
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

export function useCreateProjectAlertRuleMutation(projectRef: string | null | undefined) {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (body: CreateAlertRuleInput) => createProjectAlertRule(requireProjectRef(projectRef), body),
		onSuccess: () => {
			const ref = requireProjectRef(projectRef);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.alertsRoot(ref) });
		}
	});
}

export function useUpdateProjectAlertRuleMutation(projectRef: string | null | undefined) {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: ({ ruleId, body }: { ruleId: string; body: UpdateAlertRuleInput }) => updateProjectAlertRule(requireProjectRef(projectRef), ruleId, body),
		onSuccess: () => {
			const ref = requireProjectRef(projectRef);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.alertsRoot(ref) });
		}
	});
}

export function useDeleteProjectAlertRuleMutation(projectRef: string | null | undefined) {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (ruleId: string) => deleteProjectAlertRule(requireProjectRef(projectRef), ruleId),
		onSuccess: () => {
			const ref = requireProjectRef(projectRef);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.alertsRoot(ref) });
		}
	});
}

export function useCreateProjectNotificationMutation(projectRef: string | null | undefined) {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (body: CreateNotificationInput) => createProjectNotification(requireProjectRef(projectRef), body),
		onSuccess: () => {
			const ref = requireProjectRef(projectRef);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.alertsRoot(ref) });
		}
	});
}

export function useUpdateProjectNotificationMutation(projectRef: string | null | undefined) {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: ({ notificationId, body }: { notificationId: string; body: UpdateNotificationInput }) => updateProjectNotification(requireProjectRef(projectRef), notificationId, body),
		onSuccess: () => {
			const ref = requireProjectRef(projectRef);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.alertsRoot(ref) });
		}
	});
}

export function useDeleteProjectNotificationMutation(projectRef: string | null | undefined) {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (notificationId: string) => deleteProjectNotification(requireProjectRef(projectRef), notificationId),
		onSuccess: () => {
			const ref = requireProjectRef(projectRef);
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.alertsRoot(ref) });
		}
	});
}

export function useTestProjectNotificationMutation(projectRef: string | null | undefined) {
	return useMutation({
		mutationFn: (notificationId: string) => testProjectNotification(requireProjectRef(projectRef), notificationId)
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
			cacheCreatedProjectInvite(queryClient, ref, data);
		}
	});
}

export function useCreateProjectInviteForRefMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: ({ projectRef, body }: { projectRef: string; body: CreateProjectInviteInput }) => createProjectInvite(projectRef, body),
		onSuccess: (data, variables) => cacheCreatedProjectInvite(queryClient, variables.projectRef, data)
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

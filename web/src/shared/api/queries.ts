import { queryOptions } from "@tanstack/react-query";
import { apiClient, readApiData, readEmptyApiResponse } from "./client";
import { apiQueryKeys } from "./queryKeys";
import type {
	AddMemberInput,
	CreateCheckInput,
	CreateLabelInput,
	CreateProbeInput,
	CreateProjectInput,
	LoginInput,
	ProjectMemberRole,
	RegisterInput,
	UpdateCheckInput,
	UpdateLabelInput,
	UpdateProbeInput,
	UpdateProjectInput
} from "./types";

export const systemQueries = {
	root: () =>
		queryOptions({
			queryKey: apiQueryKeys.system.root(),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/", { signal })),
			staleTime: 5 * 60 * 1000
		}),
	health: () =>
		queryOptions({
			queryKey: apiQueryKeys.system.health(),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/healthz", { signal })),
			staleTime: 30 * 1000
		})
};

export const authQueries = {
	me: () =>
		queryOptions({
			queryKey: apiQueryKeys.auth.me(),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/auth/me", { signal })),
			retry: false,
			staleTime: 5 * 60 * 1000
		})
};

export const projectQueries = {
	list: () =>
		queryOptions({
			queryKey: apiQueryKeys.projects.list(),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/projects", { signal })),
			staleTime: 2 * 60 * 1000
		}),
	detail: (ref: string) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.detail(ref),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/projects/{ref}", { params: { path: { ref } }, signal })),
			staleTime: 2 * 60 * 1000
		}),
	checks: (ref: string) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.checks(ref),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/projects/{ref}/checks", { params: { path: { ref } }, signal })),
			staleTime: 30 * 1000
		}),
	checkDetail: (ref: string, checkId: string) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.checkDetail(ref, checkId),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/projects/{ref}/checks/{check_id}", { params: { path: { ref, check_id: checkId } }, signal })),
			staleTime: 30 * 1000
		}),
	labels: (ref: string) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.labels(ref),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/projects/{ref}/labels", { params: { path: { ref } }, signal })),
			staleTime: 5 * 60 * 1000
		}),
	members: (ref: string) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.members(ref),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/projects/{ref}/members", { params: { path: { ref } }, signal })),
			staleTime: 60 * 1000
		}),
	pingSeries: (ref: string, probeId: string, checkId: string, metric = "rttAvgMs") =>
		queryOptions({
			queryKey: apiQueryKeys.projects.pingSeries(ref, probeId, checkId, metric),
			queryFn: ({ signal }) =>
				readApiData(
					apiClient.GET("/projects/{ref}/results/ping/series", {
						params: { path: { ref }, query: { probeId, checkId, metric, maxDataPoints: 120 } },
						signal
					})
				),
			staleTime: 30 * 1000
		}),
	probes: (ref: string) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.probes(ref),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/projects/{ref}/probes", { params: { path: { ref } }, signal })),
			staleTime: 30 * 1000
		}),
	probeDetail: (ref: string, probeId: string) =>
		queryOptions({
			queryKey: apiQueryKeys.projects.probeDetail(ref, probeId),
			queryFn: ({ signal }) => readApiData(apiClient.GET("/projects/{ref}/probes/{probe_id}", { params: { path: { ref, probe_id: probeId } }, signal })),
			staleTime: 30 * 1000
		})
};

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

export function addProjectMember(ref: string, body: AddMemberInput) {
	return readApiData(apiClient.POST("/projects/{ref}/members", { params: { path: { ref } }, body }));
}

export function removeProjectMember(ref: string, userId: string) {
	return readEmptyApiResponse(apiClient.DELETE("/projects/{ref}/members/{user_id}", { params: { path: { ref, user_id: userId } } }));
}

export function updateProjectMemberRole(ref: string, userId: string, role: ProjectMemberRole) {
	return readApiData(apiClient.PATCH("/projects/{ref}/members/{user_id}", { params: { path: { ref, user_id: userId } }, body: { role } }));
}

export function rotateProjectProbeSecret(ref: string, probeId: string) {
	return readApiData(apiClient.POST("/projects/{ref}/probes/{probe_id}/secret-rotations", { params: { path: { ref, probe_id: probeId } } }));
}

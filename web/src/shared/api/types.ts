import type { components } from "./openapi";

export type AddMemberInput = components["schemas"]["AddProjectMemberRequest"];
export type ApiCheck = components["schemas"]["Check"];
export type ApiLabel = components["schemas"]["Label"];
export type ApiMeasurement = components["schemas"]["Measurement"];
export type ApiMember = components["schemas"]["ProjectMember"];
export type ApiProbe = components["schemas"]["Probe"];
export type ApiProject = components["schemas"]["Project"];
export type ApiProjectAssignment = components["schemas"]["ProjectAssignment"];
export type ChangeCurrentUserEmailInput = components["schemas"]["ChangeCurrentUserEmailRequest"];
export type ChangeCurrentUserPasswordInput = components["schemas"]["ChangeCurrentUserPasswordRequest"];
export type CreateCheckInput = components["schemas"]["CreateCheckRequest"];
export type CreateLabelInput = components["schemas"]["CreateLabelRequest"];
export type CreateProbeInput = components["schemas"]["CreateProbeRequest"];
export type CreateProjectInput = components["schemas"]["CreateProjectRequest"];
export type LoginInput = components["schemas"]["LoginUserRequest"];
export type MeasurementStatus = components["parameters"]["MeasurementQuery.status"];
export type MeasurementType = components["parameters"]["MeasurementQuery.type"];
export type PingSeriesMetric = components["parameters"]["PingSeriesQuery.metric"];
export type ProjectMemberRole = NonNullable<components["schemas"]["UpdateProjectMemberRoleRequest"]["role"]>;
export type RegisterInput = components["schemas"]["RegisterUserRequest"];
export type SelectorPreviewInput = components["schemas"]["SelectorPreviewRequest"];
export type UpdateCheckInput = components["schemas"]["UpdateCheckRequest"];
export type UpdateCurrentUserInput = components["schemas"]["UpdateCurrentUserRequest"];
export type UpdateLabelInput = components["schemas"]["UpdateLabelRequest"];
export type UpdateProbeInput = components["schemas"]["UpdateProbeRequest"];
export type UpdateProjectInput = components["schemas"]["UpdateProjectRequest"];
export type UserResponse = components["schemas"]["User"];

export interface ProjectAssignmentFilters {
	probeId?: string;
	checkId?: string;
}

export interface MeasurementFilters {
	probeId?: string;
	checkId?: string;
	type?: MeasurementType;
	status?: MeasurementStatus;
	from?: number;
	to?: number;
	limit?: number;
	cursor?: number;
}

export interface PingSeriesFilters {
	from?: number;
	to?: number;
	metric?: PingSeriesMetric;
	maxDataPoints?: number;
}

export interface TracerouteRunsFilters {
	from?: number;
	to?: number;
	limit?: number;
	cursor?: number;
}

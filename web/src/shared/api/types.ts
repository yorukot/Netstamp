import type { components } from "./openapi";

export type ApiCheck = components["schemas"]["Check"];
export type ApiLabel = components["schemas"]["Label"];
export type ApiMeasurement = components["schemas"]["Measurement"];
export type ApiMember = components["schemas"]["ProjectMember"];
export type ApiProbe = components["schemas"]["Probe"];
export type ApiProject = components["schemas"]["Project"];
export type ApiProjectAssignment = components["schemas"]["ProjectAssignment"];
export type ApiProjectInvite = components["schemas"]["ProjectInvite"];
export type ApiPublicPage = components["schemas"]["PublicPage"];
export type ApiPublicPageCheck = components["schemas"]["PublicPageCheck"];
export type ApiPublicPageFolder = components["schemas"]["PublicPageFolder"];
export type ApiPublicPingPair = components["schemas"]["PublicPingPair"];
export type ApiSelector = components["schemas"]["Selector"];
export type ChangeCurrentUserEmailInput = components["schemas"]["ChangeCurrentUserEmailRequest"];
export type ChangeCurrentUserPasswordInput = components["schemas"]["ChangeCurrentUserPasswordRequest"];
export type CreateCheckInput = components["schemas"]["CreateCheckRequest"];
export type CreateLabelInput = components["schemas"]["CreateLabelRequest"];
export type CreateProbeInput = components["schemas"]["CreateProbeRequest"];
export type CreatePublicPageFolderInput = components["schemas"]["CreatePublicPageFolderRequest"];
export type CreatePublicPageInput = components["schemas"]["CreatePublicPageRequest"];
export type CreateProjectInviteInput = components["schemas"]["CreateProjectInviteRequest"];
export type CreateProjectInput = components["schemas"]["CreateProjectRequest"];
export type LoginInput = components["schemas"]["LoginUserRequest"];
export type MeasurementStatus = components["parameters"]["MeasurementQuery.status"];
export type MeasurementType = components["parameters"]["MeasurementQuery.type"];
export type PingInsightResponse = components["schemas"]["PingInsightResponse"];
export type ApiSeries = Omit<components["schemas"]["Series"], "labels"> & {
	labels: Record<string, string>;
};
export type PingSeriesKey = "latency_avg" | "latency_min" | "latency_max" | "loss_percent";
export type PingSeriesParameter = components["parameters"]["PingSeriesQuery.series"];
export type PingSeriesResponse = Omit<components["schemas"]["PingSeriesResponse"], "series"> & {
	series: Partial<Record<PingSeriesKey, ApiSeries>> & Record<string, ApiSeries | undefined>;
};
export type ProjectMemberRole = NonNullable<components["schemas"]["UpdateProjectMemberRoleRequest"]["role"]>;
export type PublicPingInsightResponse = components["schemas"]["PublicPingInsightResponse"];
export type RegisterInput = components["schemas"]["RegisterUserRequest"];
export type SelectorPreviewInput = components["schemas"]["SelectorPreviewRequest"];
export type SetPublicPageFolderChecksInput = components["schemas"]["SetPublicPageFolderChecksRequest"];
export type TcpInsightResponse = components["schemas"]["TcpInsightResponse"];
export type TcpSeriesKey = "connect_avg" | "connect_min" | "connect_max" | "failure_percent";
export type TcpSeriesParameter = components["parameters"]["TcpSeriesQuery.series"];
export type TcpSeriesResponse = Omit<components["schemas"]["TcpSeriesResponse"], "series"> & {
	series: Partial<Record<TcpSeriesKey, ApiSeries>> & Record<string, ApiSeries | undefined>;
};
export type TracerouteHop = components["schemas"]["TracerouteHop"];
export type TracerouteInsightResponse = components["schemas"]["TracerouteInsightResponse"];
export type TracerouteResult = components["schemas"]["TracerouteResult"];
export type TracerouteTopologyEdge = components["schemas"]["TracerouteTopologyEdge"];
export type TracerouteTopologyNode = components["schemas"]["TracerouteTopologyNode"];
export type UpdateCheckInput = components["schemas"]["UpdateCheckRequest"];
export type UpdateCurrentUserInput = components["schemas"]["UpdateCurrentUserRequest"];
export type UpdateLabelInput = components["schemas"]["UpdateLabelRequest"];
export type UpdateProbeInput = components["schemas"]["UpdateProbeRequest"];
export type UpdateProjectInput = components["schemas"]["UpdateProjectRequest"];
export type UpdatePublicPageFolderInput = components["schemas"]["UpdatePublicPageFolderRequest"];
export type UpdatePublicPageInput = components["schemas"]["UpdatePublicPageRequest"];
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
	series?: PingSeriesParameter;
	maxDataPoints?: number;
}

export interface PingInsightFilters {
	from?: number;
	to?: number;
	maxDataPoints?: number;
}

export interface TcpInsightFilters {
	from?: number;
	to?: number;
	maxDataPoints?: number;
}

export interface TcpSeriesFilters {
	from?: number;
	to?: number;
	series?: TcpSeriesParameter;
	maxDataPoints?: number;
}

export interface PublicPingInsightFilters {
	from?: number;
	to?: number;
	maxDataPoints?: number;
}

export interface TracerouteRunsFilters {
	from?: number;
	to?: number;
	limit?: number;
	cursor?: number;
}

export interface TracerouteInsightFilters {
	from?: number;
	to?: number;
	maxDataPoints?: number;
}

export interface TracerouteTopologyFilters {
	probeId?: string;
	checkId?: string;
	from?: number;
	to?: number;
	limit?: number;
}

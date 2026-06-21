import type { components } from "./openapi";

export type ApiCheck = components["schemas"]["Check"];
export type ApiAlertIncident = components["schemas"]["AlertIncident"];
export type ApiAlertRule = components["schemas"]["AlertRule"];
export type ApiLabel = components["schemas"]["Label"];
export type ApiLatestResult = components["schemas"]["LatestResult"];
export type ApiMember = components["schemas"]["ProjectMember"];
export type ApiNotification = components["schemas"]["Notification"];
export type ApiProbe = components["schemas"]["Probe"];
export type ApiProject = components["schemas"]["Project"];
export type ApiProjectAssignment = components["schemas"]["ProjectAssignment"];
export type ApiProjectInvite = components["schemas"]["ProjectInvite"];
export type ApiPublicStatusElement = components["schemas"]["PublicStatusElement"];
export type ApiPublicStatusPage = components["schemas"]["PublicStatusPage"];
export type ApiPublicStatusPublicElement = Omit<components["schemas"]["PublicStatusPublicElement"], "chart" | "children"> & {
	chart?: Omit<components["schemas"]["PublicStatusChart"], "series"> & {
		series: ApiSeries[];
	};
	children?: ApiPublicStatusPublicElement[];
};
export type ApiPublicStatusPublicResponse = Omit<components["schemas"]["PublicStatusPublicResponse"], "elements"> & {
	elements: ApiPublicStatusPublicElement[];
};
export type ApiSelector = components["schemas"]["Selector"];
export type CreateAlertRuleInput = components["schemas"]["CreateAlertRuleRequest"];
export type ChangeCurrentUserEmailInput = components["schemas"]["ChangeCurrentUserEmailRequest"];
export type ChangeCurrentUserPasswordInput = components["schemas"]["ChangeCurrentUserPasswordRequest"];
export type CreateCheckInput = components["schemas"]["CreateCheckRequest"];
export type CreateLabelInput = components["schemas"]["CreateLabelRequest"];
export type CreateNotificationInput = components["schemas"]["CreateNotificationRequest"];
export type CreateProbeInput = components["schemas"]["CreateProbeRequest"];
export type CreateProjectInviteInput = components["schemas"]["CreateProjectInviteRequest"];
export type CreateProjectInput = components["schemas"]["CreateProjectRequest"];
export type CreatePublicStatusElementInput = components["schemas"]["CreatePublicStatusElementRequest"];
export type CreatePublicStatusPageInput = components["schemas"]["CreatePublicStatusPageRequest"];
export type LoginInput = components["schemas"]["LoginUserRequest"];
export type LatestResultType = components["parameters"]["LatestResultsQuery.type"];
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
export type RegisterInput = components["schemas"]["RegisterUserRequest"];
export type SelectorPreviewInput = components["schemas"]["SelectorPreviewRequest"];
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
export type UpdateAlertRuleInput = components["schemas"]["UpdateAlertRuleRequest"];
export type UpdateCurrentUserInput = components["schemas"]["UpdateCurrentUserRequest"];
export type UpdateLabelInput = components["schemas"]["UpdateLabelRequest"];
export type UpdateNotificationInput = components["schemas"]["UpdateNotificationRequest"];
export type UpdateProbeInput = components["schemas"]["UpdateProbeRequest"];
export type UpdateProjectInput = components["schemas"]["UpdateProjectRequest"];
export type UpdatePublicStatusElementInput = components["schemas"]["UpdatePublicStatusElementRequest"];
export type UpdatePublicStatusPageInput = components["schemas"]["UpdatePublicStatusPageRequest"];
export type UserResponse = components["schemas"]["User"];

export type PublicStatusChartMode = ApiPublicStatusPage["defaultChartMode"];
export type PublicStatusChartRange = ApiPublicStatusPage["defaultChartRange"];
export type PublicStatusElementChartMode = ApiPublicStatusElement["chartMode"];
export type PublicStatusElementKind = ApiPublicStatusElement["kind"];
export type PublicStatusState = components["schemas"]["PublicStatusPageSummary"]["status"];

export interface ProjectAssignmentFilters {
	probeId?: string;
	checkId?: string;
}

export interface AlertRuleFilters {
	status?: "enabled" | "disabled";
	checkType?: "ping" | "tcp" | "traceroute";
}

export interface AlertIncidentFilters {
	status?: "open" | "acknowledged" | "resolved";
	limit?: number;
}

export interface LatestResultsFilters {
	probeId?: string;
	checkId?: string;
	type?: LatestResultType;
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

export interface PublicStatusFilters {
	includeCharts?: boolean;
	range?: PublicStatusChartRange;
}

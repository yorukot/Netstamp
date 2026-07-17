import "@typespec/http";

using Http;

namespace Netstamp;

alias PublicStatusChartMode = "inherit" | "off" | "compact";
alias PublicStatusPageChartMode = "off" | "compact";
alias PublicStatusChartRange = "24h" | "7d" | "30d";
alias PublicStatusDailyStatusRange = "30d";
alias PublicStatusElementKind = "folder" | "assignment_group";
alias PublicStatusElementDisplayMode = "status" | "history" | "latency" | "map";
alias PublicStatusAssignmentSelectionMode = "all_check" | "selected_assignments";
alias PublicStatusState = "operational" | "degraded" | "down" | "unknown";
alias PublicStatusTheme = "light" | "dark" | "auto";
alias PublicIncidentStatus = "open" | "acknowledged" | "resolved";

model PublicStatusPageIdPathParam {
	@path
	page_id: uuid;
}

model PublicStatusElementIdPathParam {
	@path
	element_id: uuid;
}

model PublicStatusSlugPathParam {
	@path
	@minLength(1)
	@maxLength(64)
	@pattern("^[a-z0-9-]+$")
	slug: string;
}

model PublicStatusPage {
	id: uuid;
	projectId: uuid;

	@minLength(1)
	@maxLength(64)
	@pattern("^[a-z0-9-]+$")
	slug: string;

	@minLength(1)
	@maxLength(128)
	title: string;

	@minLength(1)
	@maxLength(1024)
	description?: string;

	enabled: boolean;

	@minLength(1)
	@maxLength(2048)
	footerText?: string;

	@minLength(1)
	@maxLength(2048)
	bannerImageUrl?: string;

	theme: PublicStatusTheme;
	showTargets: boolean;
	showProbeNames: boolean;
	showProbeLocations: boolean;
	showIncidentHistory: boolean;
	showGeneratedAt: boolean;

	@minLength(1)
	@maxLength(65536)
	customCss?: string;

	defaultChartMode: PublicStatusPageChartMode;
	defaultChartRange: PublicStatusChartRange;
	createdAt: utcDateTime;
	updatedAt: utcDateTime;
}

model PublicStatusPageSummary {
	id: uuid;

	@minLength(1)
	@maxLength(64)
	@pattern("^[a-z0-9-]+$")
	slug: string;

	@minLength(1)
	@maxLength(128)
	title: string;

	@minLength(1)
	@maxLength(1024)
	description?: string;

	status: PublicStatusState;

	@minLength(1)
	@maxLength(2048)
	footerText?: string;

	@minLength(1)
	@maxLength(2048)
	bannerImageUrl?: string;

	theme: PublicStatusTheme;
	showIncidentHistory: boolean;
	showGeneratedAt: boolean;

	@minLength(1)
	@maxLength(65536)
	customCss?: string;

	defaultChartMode: PublicStatusPageChartMode;
	defaultChartRange: PublicStatusChartRange;
	updatedAt: utcDateTime;
}

model PublicStatusElement {
	id: uuid;
	publicPageId: uuid;
	parentElementId?: uuid;
	kind: PublicStatusElementKind;
	checkId?: uuid;
	assignmentSelectionMode?: PublicStatusAssignmentSelectionMode;
	assignmentIds: uuid[];

	@minLength(1)
	@maxLength(1024)
	title?: string;

	@minLength(1)
	@maxLength(1024)
	description?: string;

	@minValue(0)
	sortOrder: int32;

	displayMode: PublicStatusElementDisplayMode;
	chartMode: PublicStatusChartMode;
	chartRange?: PublicStatusChartRange;
	checkName?: string;
	checkType?: CheckType;
	checkTarget?: string;
	checkIntervalSeconds?: int32;
	createdAt: utcDateTime;
	updatedAt: utcDateTime;
}

model PublicStatusMetrics {
	@minValue(0)
	latencyAvgMs?: float64;

	@minValue(0)
	@maxValue(100)
	lossPercent?: float64;

	@minValue(0)
	connectAvgMs?: float64;

	@minValue(0)
	@maxValue(100)
	failurePercent?: float64;
}

model PublicStatusPublicAssignment {
	checkTitle: string;
	type: CheckType;
	target?: CheckTarget;
	probeName?: string;
	probeLocationName?: string;
	latitude?: float64;
	longitude?: float64;
	latestStartedAt?: utcDateTime;
	latestStatus?: LatestResultStatus;
	metrics?: PublicStatusMetrics;
}

model PublicStatusChart {
	range: PublicStatusChartRange;
	series: Series[];
}

model PublicStatusPublicElement {
	id: uuid;
	kind: PublicStatusElementKind;

	@minLength(1)
	title: string;

	@minLength(1)
	@maxLength(1024)
	description?: string;

	type?: CheckType;
	target?: string;
	status: PublicStatusState;
	latestStartedAt?: utcDateTime;
	latestStatus?: LatestResultStatus;
	displayMode: PublicStatusElementDisplayMode;
	chartMode?: PublicStatusPageChartMode;
	chartRange?: PublicStatusChartRange;

	@minValue(0)
	assignmentCount?: int32;

	@minValue(0)
	successfulAssignments?: int32;

	@minValue(0)
	failingAssignments?: int32;

	@minValue(0)
	staleAssignments?: int32;

	metrics?: PublicStatusMetrics;
	chart?: PublicStatusChart;
	assignments?: PublicStatusPublicAssignment[];
	children?: PublicStatusPublicElement[];
}

model PublicStatusIncidentSummary {
	metric?: string;
	value?: float64;
	threshold?: float64;
	windowSeconds?: int32;
}

model PublicStatusIncident {
	id: uuid;
	checkTitle: string;
	status: PublicIncidentStatus;
	severity: AlertSeverity;
	openedAt: utcDateTime;
	resolvedAt?: utcDateTime;
	lastTriggeredAt: utcDateTime;
	summary?: PublicStatusIncidentSummary;
}

model PublicStatusIncidents {
	active: PublicStatusIncident[];
	recentResolved: PublicStatusIncident[];
}

model PublicStatusPageResponse {
	page: PublicStatusPage;
}

model PublicStatusPageListResponse {
	pages: PublicStatusPage[];
}

model PublicStatusPageDetailResponse {
	page: PublicStatusPage;
	elements: PublicStatusElement[];
}

model PublicStatusEditorContextResponse {
	projectRef: string;
	pageId: uuid;
}

model PublicStatusElementResponse {
	element: PublicStatusElement;
}

model PublicStatusSummaryResponse {
	page: PublicStatusPageSummary;
	generatedAt: utcDateTime;
}

model PublicStatusElementsResponse {
	elements: PublicStatusPublicElement[];
	generatedAt: utcDateTime;
}

model PublicStatusIncidentsResponse {
	incidents: PublicStatusIncidents;
	generatedAt: utcDateTime;
}

model PublicStatusElementChartResponse {
	chart?: PublicStatusChart;
	generatedAt: utcDateTime;
}

model PublicStatusDailyStatusDay {
	@format("date")
	date: string;

	status: PublicStatusState;

	@minValue(0)
	incidentCount: int32;

	severity?: AlertSeverity;
}

model PublicStatusElementDailyStatusResponse {
	range: PublicStatusDailyStatusRange;
	days: PublicStatusDailyStatusDay[];
	generatedAt: utcDateTime;
}

model CreatePublicStatusPageRequest {
	@minLength(1)
	@maxLength(64)
	@pattern("^[a-z0-9-]+$")
	slug: string;

	@minLength(1)
	@maxLength(128)
	title: string;

	@minLength(1)
	@maxLength(1024)
	description?: string;

	enabled?: boolean = true;

	@minLength(1)
	@maxLength(2048)
	footerText?: string;

	@minLength(1)
	@maxLength(2048)
	bannerImageUrl?: string;

	theme?: PublicStatusTheme = "auto";
	showTargets?: boolean = false;
	showProbeNames?: boolean = false;
	showProbeLocations?: boolean = false;
	showIncidentHistory?: boolean = true;
	showGeneratedAt?: boolean = true;

	@minLength(1)
	@maxLength(65536)
	customCss?: string;

	defaultChartMode?: PublicStatusPageChartMode = "off";
	defaultChartRange?: PublicStatusChartRange = "24h";
}

@doc("Full replacement patch for public status page settings.")
model UpdatePublicStatusPageRequest is CreatePublicStatusPageRequest;

model CreatePublicStatusElementRequest {
	parentElementId?: uuid;
	kind: PublicStatusElementKind;
	checkId?: uuid;
	assignmentSelectionMode?: PublicStatusAssignmentSelectionMode;
	assignmentIds?: uuid[];

	@minLength(1)
	@maxLength(1024)
	title?: string;

	@minLength(1)
	@maxLength(1024)
	description?: string;

	@minValue(0)
	sortOrder: int32;

	displayMode?: PublicStatusElementDisplayMode = "status";
	chartMode?: PublicStatusChartMode = "inherit";
	chartRange?: PublicStatusChartRange;
}

@doc("Full replacement patch for a public status page element.")
model UpdatePublicStatusElementRequest is CreatePublicStatusElementRequest;

model PublicStatusIncidentsQuery {
	@query
	@minValue(1)
	@maxValue(200)
	limit?: int32;
}

model PublicStatusChartQuery {
	@query
	range?: PublicStatusChartRange;
}

model PublicStatusDailyStatusQuery {
	@query
	range?: PublicStatusDailyStatusRange;
}

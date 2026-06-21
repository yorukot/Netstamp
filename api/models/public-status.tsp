import "@typespec/http";

using Http;

namespace Netstamp;

alias PublicStatusChartMode = "inherit" | "off" | "compact";
alias PublicStatusPageChartMode = "off" | "compact";
alias PublicStatusChartRange = "24h" | "7d" | "30d";
alias PublicStatusElementKind = "folder" | "check";
alias PublicStatusState = "operational" | "degraded" | "down" | "unknown";
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

	@minLength(1)
	@maxLength(1024)
	title?: string;

	@minLength(1)
	@maxLength(1024)
	description?: string;

	@minValue(0)
	sortOrder: int32;

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

model PublicStatusChart {
	range: PublicStatusChartRange;
	series: Series[];
}

model PublicStatusPublicElement {
	id: uuid;
	kind: PublicStatusElementKind;
	checkId?: uuid;

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
	checkId: uuid;
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

model PublicStatusElementResponse {
	element: PublicStatusElement;
}

model PublicStatusPublicResponse {
	page: PublicStatusPageSummary;
	elements: PublicStatusPublicElement[];
	incidents: PublicStatusIncidents;
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
	defaultChartMode?: PublicStatusPageChartMode = "off";
	defaultChartRange?: PublicStatusChartRange = "24h";
}

@doc("Full replacement patch for public status page settings.")
model UpdatePublicStatusPageRequest is CreatePublicStatusPageRequest;

model CreatePublicStatusElementRequest {
	parentElementId?: uuid;
	kind: PublicStatusElementKind;
	checkId?: uuid;

	@minLength(1)
	@maxLength(1024)
	title?: string;

	@minLength(1)
	@maxLength(1024)
	description?: string;

	@minValue(0)
	sortOrder: int32;

	chartMode?: PublicStatusChartMode = "inherit";
	chartRange?: PublicStatusChartRange;
}

@doc("Full replacement patch for a public status page element.")
model UpdatePublicStatusElementRequest is CreatePublicStatusElementRequest;

model PublicStatusQuery {
	@query
	includeCharts?: boolean = true;

	@query
	range?: PublicStatusChartRange;
}

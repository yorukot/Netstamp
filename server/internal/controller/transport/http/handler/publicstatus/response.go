package publicstatus

import (
	"encoding/json"
	"time"

	apppublic "github.com/yorukot/netstamp/internal/controller/application/publicstatus"
	domainpublic "github.com/yorukot/netstamp/internal/domain/publicstatus"
)

type pageBody struct {
	ID                string                  `json:"id"`
	ProjectID         string                  `json:"projectId,omitempty"`
	Slug              string                  `json:"slug"`
	Title             string                  `json:"title"`
	Description       *string                 `json:"description,omitempty"`
	Enabled           bool                    `json:"enabled"`
	DefaultChartMode  domainpublic.ChartMode  `json:"defaultChartMode"`
	DefaultChartRange domainpublic.ChartRange `json:"defaultChartRange"`
	CreatedAt         time.Time               `json:"createdAt,omitempty"`
	UpdatedAt         time.Time               `json:"updatedAt"`
}

type elementBody struct {
	ID                      string                                `json:"id"`
	PublicPageID            string                                `json:"publicPageId"`
	ParentElementID         *string                               `json:"parentElementId,omitempty"`
	Kind                    domainpublic.ElementKind              `json:"kind"`
	CheckID                 *string                               `json:"checkId,omitempty"`
	AssignmentSelectionMode *domainpublic.AssignmentSelectionMode `json:"assignmentSelectionMode,omitempty"`
	AssignmentIDs           []string                              `json:"assignmentIds"`
	Title                   *string                               `json:"title,omitempty"`
	Description             *string                               `json:"description,omitempty"`
	SortOrder               int32                                 `json:"sortOrder"`
	ChartMode               domainpublic.ChartMode                `json:"chartMode"`
	ChartRange              *domainpublic.ChartRange              `json:"chartRange,omitempty"`
	CheckName               *string                               `json:"checkName,omitempty"`
	CheckType               any                                   `json:"checkType,omitempty"`
	CheckTarget             *string                               `json:"checkTarget,omitempty"`
	CheckIntervalSeconds    *int32                                `json:"checkIntervalSeconds,omitempty"`
	CreatedAt               time.Time                             `json:"createdAt"`
	UpdatedAt               time.Time                             `json:"updatedAt"`
}

type pageResponseBody struct {
	Page pageBody `json:"page"`
}

type pageListResponseBody struct {
	Pages []pageBody `json:"pages"`
}

type pageDetailResponseBody struct {
	Page     pageBody      `json:"page"`
	Elements []elementBody `json:"elements"`
}

type elementResponseBody struct {
	Element elementBody `json:"element"`
}

type publicStatusSummaryResponseBody struct {
	Page        publicPageBody `json:"page"`
	GeneratedAt time.Time      `json:"generatedAt"`
}

type publicStatusElementsResponseBody struct {
	Elements    []publicElementBody `json:"elements"`
	GeneratedAt time.Time           `json:"generatedAt"`
}

type publicStatusIncidentsResponseBody struct {
	Incidents   publicIncidentsBody `json:"incidents"`
	GeneratedAt time.Time           `json:"generatedAt"`
}

type publicStatusElementChartResponseBody struct {
	Chart       *chartBody `json:"chart,omitempty"`
	GeneratedAt time.Time  `json:"generatedAt"`
}

type publicStatusElementDailyStatusResponseBody struct {
	Range       domainpublic.ChartRange `json:"range"`
	Days        []dailyStatusDayBody    `json:"days"`
	GeneratedAt time.Time               `json:"generatedAt"`
}

type dailyStatusDayBody struct {
	Date          string              `json:"date"`
	Status        domainpublic.Status `json:"status"`
	IncidentCount int32               `json:"incidentCount"`
	Severity      *string             `json:"severity,omitempty"`
}

type publicPageBody struct {
	ID                string                  `json:"id"`
	Slug              string                  `json:"slug"`
	Title             string                  `json:"title"`
	Description       *string                 `json:"description,omitempty"`
	Status            domainpublic.Status     `json:"status"`
	DefaultChartMode  domainpublic.ChartMode  `json:"defaultChartMode"`
	DefaultChartRange domainpublic.ChartRange `json:"defaultChartRange"`
	UpdatedAt         time.Time               `json:"updatedAt"`
}

type publicElementBody struct {
	ID                    string                   `json:"id"`
	Kind                  domainpublic.ElementKind `json:"kind"`
	CheckID               *string                  `json:"checkId,omitempty"`
	Title                 string                   `json:"title"`
	Description           *string                  `json:"description,omitempty"`
	Type                  any                      `json:"type,omitempty"`
	Target                *string                  `json:"target,omitempty"`
	Status                domainpublic.Status      `json:"status"`
	LatestStartedAt       *time.Time               `json:"latestStartedAt,omitempty"`
	LatestStatus          *string                  `json:"latestStatus,omitempty"`
	ChartMode             domainpublic.ChartMode   `json:"chartMode,omitempty"`
	ChartRange            domainpublic.ChartRange  `json:"chartRange,omitempty"`
	AssignmentCount       int32                    `json:"assignmentCount,omitempty"`
	SuccessfulAssignments int32                    `json:"successfulAssignments,omitempty"`
	FailingAssignments    int32                    `json:"failingAssignments,omitempty"`
	StaleAssignments      int32                    `json:"staleAssignments,omitempty"`
	Metrics               *metricsBody             `json:"metrics,omitempty"`
	Chart                 *chartBody               `json:"chart,omitempty"`
	Assignments           []publicAssignmentBody   `json:"assignments,omitempty"`
	Children              []publicElementBody      `json:"children,omitempty"`
}

type metricsBody struct {
	LatencyAvgMs   *float64 `json:"latencyAvgMs,omitempty"`
	LossPercent    *float64 `json:"lossPercent,omitempty"`
	ConnectAvgMs   *float64 `json:"connectAvgMs,omitempty"`
	FailurePercent *float64 `json:"failurePercent,omitempty"`
}

type publicAssignmentBody struct {
	AssignmentID      string       `json:"assignmentId"`
	CheckID           string       `json:"checkId"`
	CheckTitle        string       `json:"checkTitle"`
	Type              any          `json:"type"`
	Target            string       `json:"target"`
	ProbeID           string       `json:"probeId"`
	ProbeName         string       `json:"probeName"`
	ProbeLocationName *string      `json:"probeLocationName,omitempty"`
	LatestStartedAt   *time.Time   `json:"latestStartedAt,omitempty"`
	LatestStatus      *string      `json:"latestStatus,omitempty"`
	Metrics           *metricsBody `json:"metrics,omitempty"`
}

type chartBody struct {
	Range  domainpublic.ChartRange `json:"range"`
	Series []seriesBody            `json:"series"`
}

type seriesBody struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
	Unit   string            `json:"unit"`
	Points []seriesPointBody `json:"points"`
}

type seriesPointBody [2]any

type publicIncidentsBody struct {
	Active         []incidentBody `json:"active"`
	RecentResolved []incidentBody `json:"recentResolved"`
}

type incidentBody struct {
	ID              string           `json:"id"`
	CheckID         string           `json:"checkId"`
	CheckTitle      string           `json:"checkTitle"`
	Status          string           `json:"status"`
	Severity        string           `json:"severity"`
	OpenedAt        time.Time        `json:"openedAt"`
	ResolvedAt      *time.Time       `json:"resolvedAt,omitempty"`
	LastTriggeredAt time.Time        `json:"lastTriggeredAt"`
	Summary         *incidentSummary `json:"summary,omitempty"`
}

type incidentSummary struct {
	Metric        *string  `json:"metric,omitempty"`
	Value         *float64 `json:"value,omitempty"`
	Threshold     *float64 `json:"threshold,omitempty"`
	WindowSeconds *int32   `json:"windowSeconds,omitempty"`
}

func newPageBody(page domainpublic.Page, includeProjectID bool) pageBody {
	body := pageBody{
		ID:                page.ID,
		Slug:              page.Slug,
		Title:             page.Title,
		Description:       page.Description,
		Enabled:           page.Enabled,
		DefaultChartMode:  page.DefaultChartMode,
		DefaultChartRange: page.DefaultChartRange,
		CreatedAt:         page.CreatedAt,
		UpdatedAt:         page.UpdatedAt,
	}
	if includeProjectID {
		body.ProjectID = page.ProjectID
	}
	return body
}

func newElementBody(element domainpublic.Element) elementBody {
	body := elementBody{
		ID:                      element.ID,
		PublicPageID:            element.PublicPageID,
		ParentElementID:         element.ParentElementID,
		Kind:                    element.Kind,
		CheckID:                 element.CheckID,
		AssignmentSelectionMode: element.AssignmentSelectionMode,
		AssignmentIDs:           append([]string{}, element.AssignmentIDs...),
		Title:                   element.Title,
		Description:             element.Description,
		SortOrder:               element.SortOrder,
		ChartMode:               element.ChartMode,
		ChartRange:              element.ChartRange,
		CheckName:               element.CheckName,
		CheckTarget:             element.CheckTarget,
		CheckIntervalSeconds:    element.CheckIntervalSeconds,
		CreatedAt:               element.CreatedAt,
		UpdatedAt:               element.UpdatedAt,
	}
	if element.CheckType != nil {
		body.CheckType = *element.CheckType
	}
	return body
}

func newPageListResponse(pages []domainpublic.Page) pageListResponseBody {
	body := pageListResponseBody{Pages: make([]pageBody, 0, len(pages))}
	for _, page := range pages {
		body.Pages = append(body.Pages, newPageBody(page, true))
	}
	return body
}

func newPageDetailResponse(detail apppublic.PageDetail) pageDetailResponseBody {
	body := pageDetailResponseBody{
		Page:     newPageBody(detail.Page, true),
		Elements: make([]elementBody, 0, len(detail.Elements)),
	}
	for _, element := range detail.Elements {
		body.Elements = append(body.Elements, newElementBody(element))
	}
	return body
}

func newPublicStatusSummaryResponse(summary apppublic.PublicSummary) publicStatusSummaryResponseBody {
	return publicStatusSummaryResponseBody{
		Page:        newPublicPageBody(summary.Page, summary.Status),
		GeneratedAt: summary.GeneratedAt,
	}
}

func newPublicStatusElementsResponse(elements apppublic.PublicElements) publicStatusElementsResponseBody {
	return publicStatusElementsResponseBody{
		Elements:    newPublicElementBodies(elements.Elements),
		GeneratedAt: elements.GeneratedAt,
	}
}

func newPublicStatusIncidentsResponse(incidents apppublic.PublicIncidents) publicStatusIncidentsResponseBody {
	return publicStatusIncidentsResponseBody{
		Incidents: publicIncidentsBody{
			Active:         newIncidentBodies(incidents.ActiveIncidents),
			RecentResolved: newIncidentBodies(incidents.ResolvedIncidents),
		},
		GeneratedAt: incidents.GeneratedAt,
	}
}

func newPublicStatusElementChartResponse(chart apppublic.PublicElementChart) publicStatusElementChartResponseBody {
	return publicStatusElementChartResponseBody{
		Chart:       newChartBody(chart.Chart),
		GeneratedAt: chart.GeneratedAt,
	}
}

func newPublicStatusElementDailyStatusResponse(dailyStatus apppublic.PublicElementDailyStatus) publicStatusElementDailyStatusResponseBody {
	body := publicStatusElementDailyStatusResponseBody{
		Range:       dailyStatus.Range,
		Days:        make([]dailyStatusDayBody, 0, len(dailyStatus.Days)),
		GeneratedAt: dailyStatus.GeneratedAt,
	}
	for _, day := range dailyStatus.Days {
		body.Days = append(body.Days, dailyStatusDayBody{
			Date:          day.Date.Format("2006-01-02"),
			Status:        day.Status,
			IncidentCount: day.IncidentCount,
			Severity:      day.Severity,
		})
	}
	return body
}

func newPublicPageBody(page domainpublic.Page, status domainpublic.Status) publicPageBody {
	return publicPageBody{
		ID:                page.ID,
		Slug:              page.Slug,
		Title:             page.Title,
		Description:       page.Description,
		Status:            status,
		DefaultChartMode:  page.DefaultChartMode,
		DefaultChartRange: page.DefaultChartRange,
		UpdatedAt:         page.UpdatedAt,
	}
}

func newPublicElementBodies(elements []domainpublic.RenderedElement) []publicElementBody {
	values := make([]publicElementBody, 0, len(elements))
	for _, element := range elements {
		values = append(values, newPublicElementBody(element))
	}
	return values
}

func newPublicElementBody(element domainpublic.RenderedElement) publicElementBody {
	body := publicElementBody{
		ID:                    element.ID,
		Kind:                  element.Kind,
		CheckID:               element.CheckID,
		Title:                 publicElementTitle(element),
		Description:           elementDescription(element),
		Target:                element.CheckTarget,
		Status:                element.Status,
		LatestStartedAt:       element.LatestStartedAt,
		LatestStatus:          element.LatestStatus,
		ChartMode:             element.ResolvedChartMode,
		ChartRange:            element.ResolvedChartRange,
		AssignmentCount:       element.AssignmentCount,
		SuccessfulAssignments: element.SuccessfulAssignments,
		FailingAssignments:    element.FailingAssignments,
		StaleAssignments:      element.StaleAssignments,
		Metrics:               newMetricsBody(element.Metrics),
		Chart:                 newChartBody(element.Chart),
		Assignments:           newPublicAssignmentBodies(element.Assignments),
		Children:              newPublicElementBodies(element.Children),
	}
	if element.CheckType != nil {
		body.Type = *element.CheckType
	}
	return body
}

func publicElementTitle(element domainpublic.RenderedElement) string {
	if element.Title != nil {
		return *element.Title
	}
	if element.CheckName != nil {
		return *element.CheckName
	}
	if len(element.Assignments) == 1 {
		return element.Assignments[0].CheckName
	}
	if element.Kind == domainpublic.ElementKindFolder {
		return "Untitled folder"
	}
	return "Assignment group"
}

func elementDescription(element domainpublic.RenderedElement) *string {
	if element.Description != nil {
		return element.Description
	}
	return element.CheckDescription
}

func newPublicAssignmentBodies(assignments []domainpublic.Assignment) []publicAssignmentBody {
	values := make([]publicAssignmentBody, 0, len(assignments))
	for _, assignment := range assignments {
		latestStartedAt, latestStatus := latestAssignmentFields(assignment)
		values = append(values, publicAssignmentBody{
			AssignmentID:      assignment.AssignmentID,
			CheckID:           assignment.CheckID,
			CheckTitle:        assignment.CheckName,
			Type:              assignment.CheckType,
			Target:            assignment.CheckTarget,
			ProbeID:           assignment.ProbeID,
			ProbeName:         assignment.ProbeName,
			ProbeLocationName: assignment.ProbeLocationName,
			LatestStartedAt:   latestStartedAt,
			LatestStatus:      latestStatus,
			Metrics:           assignmentMetricsBody(assignment),
		})
	}
	return values
}

func latestAssignmentFields(assignment domainpublic.Assignment) (*time.Time, *string) {
	if assignment.LatestStatus == "" {
		return nil, nil
	}
	startedAt := assignment.LatestStartedAt
	status := assignment.LatestStatus
	return &startedAt, &status
}

func assignmentMetricsBody(assignment domainpublic.Assignment) *metricsBody {
	if assignment.LatestStatus == "" {
		return nil
	}
	switch assignment.CheckType {
	case "ping":
		lossPercent := assignment.LossPercent
		return &metricsBody{LatencyAvgMs: assignment.LatencyAvgMs, LossPercent: &lossPercent}
	case "tcp":
		return &metricsBody{ConnectAvgMs: assignment.ConnectAvgMs, FailurePercent: assignment.FailurePercent}
	case "http":
		return &metricsBody{LatencyAvgMs: assignment.LatencyAvgMs, FailurePercent: assignment.FailurePercent}
	default:
		return nil
	}
}

func newMetricsBody(metrics *domainpublic.Metrics) *metricsBody {
	if metrics == nil {
		return nil
	}
	return &metricsBody{
		LatencyAvgMs:   metrics.LatencyAvgMs,
		LossPercent:    metrics.LossPercent,
		ConnectAvgMs:   metrics.ConnectAvgMs,
		FailurePercent: metrics.FailurePercent,
	}
}

func newChartBody(chart *domainpublic.Chart) *chartBody {
	if chart == nil {
		return nil
	}
	body := chartBody{Range: chart.Range, Series: make([]seriesBody, 0, len(chart.Series))}
	for _, series := range chart.Series {
		body.Series = append(body.Series, seriesBody{
			Name:   series.Name,
			Labels: series.Labels,
			Unit:   series.Unit,
			Points: seriesPoints(series.Points),
		})
	}
	return &body
}

func seriesPoints(points []domainpublic.SeriesPoint) []seriesPointBody {
	values := make([]seriesPointBody, 0, len(points))
	for _, point := range points {
		values = append(values, seriesPointBody{point.TimestampMs, point.Value})
	}
	return values
}

func newIncidentBodies(incidents []domainpublic.Incident) []incidentBody {
	values := make([]incidentBody, 0, len(incidents))
	for _, incident := range incidents {
		values = append(values, incidentBody{
			ID:              incident.ID,
			CheckID:         incident.CheckID,
			CheckTitle:      incident.CheckName,
			Status:          incident.Status,
			Severity:        incident.Severity,
			OpenedAt:        incident.OpenedAt,
			ResolvedAt:      incident.ResolvedAt,
			LastTriggeredAt: incident.LastTriggeredAt,
			Summary:         publicIncidentSummary(incident),
		})
	}
	return values
}

func publicIncidentSummary(incident domainpublic.Incident) *incidentSummary {
	var raw struct {
		Metric        *string  `json:"metric"`
		Value         *float64 `json:"value"`
		Threshold     *float64 `json:"threshold"`
		WindowSeconds *int32   `json:"windowSeconds"`
	}
	if len(incident.LastSummary) > 0 {
		if err := json.Unmarshal(incident.LastSummary, &raw); err != nil {
			raw = struct {
				Metric        *string  `json:"metric"`
				Value         *float64 `json:"value"`
				Threshold     *float64 `json:"threshold"`
				WindowSeconds *int32   `json:"windowSeconds"`
			}{}
		}
	}
	if raw.Value == nil {
		raw.Value = incident.LastValue
	}
	if raw.Metric == nil && raw.Value == nil && raw.Threshold == nil && raw.WindowSeconds == nil {
		return nil
	}
	return &incidentSummary{
		Metric:        raw.Metric,
		Value:         raw.Value,
		Threshold:     raw.Threshold,
		WindowSeconds: raw.WindowSeconds,
	}
}

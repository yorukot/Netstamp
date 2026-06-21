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
	ID                   string                   `json:"id"`
	PublicPageID         string                   `json:"publicPageId"`
	ParentElementID      *string                  `json:"parentElementId,omitempty"`
	Kind                 domainpublic.ElementKind `json:"kind"`
	CheckID              *string                  `json:"checkId,omitempty"`
	Title                *string                  `json:"title,omitempty"`
	Description          *string                  `json:"description,omitempty"`
	SortOrder            int32                    `json:"sortOrder"`
	ChartMode            domainpublic.ChartMode   `json:"chartMode"`
	ChartRange           *domainpublic.ChartRange `json:"chartRange,omitempty"`
	CheckName            *string                  `json:"checkName,omitempty"`
	CheckType            any                      `json:"checkType,omitempty"`
	CheckTarget          *string                  `json:"checkTarget,omitempty"`
	CheckIntervalSeconds *int32                   `json:"checkIntervalSeconds,omitempty"`
	CreatedAt            time.Time                `json:"createdAt"`
	UpdatedAt            time.Time                `json:"updatedAt"`
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

type publicStatusPageResponseBody struct {
	Page        publicPageBody      `json:"page"`
	Elements    []publicElementBody `json:"elements"`
	Incidents   publicIncidentsBody `json:"incidents"`
	GeneratedAt time.Time           `json:"generatedAt"`
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
	AssignmentCount       int32                    `json:"assignmentCount,omitempty"`
	SuccessfulAssignments int32                    `json:"successfulAssignments,omitempty"`
	FailingAssignments    int32                    `json:"failingAssignments,omitempty"`
	StaleAssignments      int32                    `json:"staleAssignments,omitempty"`
	Metrics               *metricsBody             `json:"metrics,omitempty"`
	Chart                 *chartBody               `json:"chart,omitempty"`
	Children              []publicElementBody      `json:"children,omitempty"`
}

type metricsBody struct {
	LatencyAvgMs   *float64 `json:"latencyAvgMs,omitempty"`
	LossPercent    *float64 `json:"lossPercent,omitempty"`
	ConnectAvgMs   *float64 `json:"connectAvgMs,omitempty"`
	FailurePercent *float64 `json:"failurePercent,omitempty"`
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
		ID:                   element.ID,
		PublicPageID:         element.PublicPageID,
		ParentElementID:      element.ParentElementID,
		Kind:                 element.Kind,
		CheckID:              element.CheckID,
		Title:                element.Title,
		Description:          element.Description,
		SortOrder:            element.SortOrder,
		ChartMode:            element.ChartMode,
		ChartRange:           element.ChartRange,
		CheckName:            element.CheckName,
		CheckTarget:          element.CheckTarget,
		CheckIntervalSeconds: element.CheckIntervalSeconds,
		CreatedAt:            element.CreatedAt,
		UpdatedAt:            element.UpdatedAt,
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

func newPublicStatusPageResponse(rendered domainpublic.RenderedPage) publicStatusPageResponseBody {
	return publicStatusPageResponseBody{
		Page: publicPageBody{
			ID:                rendered.Page.ID,
			Slug:              rendered.Page.Slug,
			Title:             rendered.Page.Title,
			Description:       rendered.Page.Description,
			Status:            rendered.Status,
			DefaultChartMode:  rendered.Page.DefaultChartMode,
			DefaultChartRange: rendered.Page.DefaultChartRange,
			UpdatedAt:         rendered.Page.UpdatedAt,
		},
		Elements: newPublicElementBodies(rendered.Elements),
		Incidents: publicIncidentsBody{
			Active:         newIncidentBodies(rendered.ActiveIncidents),
			RecentResolved: newIncidentBodies(rendered.ResolvedIncidents),
		},
		GeneratedAt: rendered.GeneratedAt,
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
	title := ""
	switch {
	case element.Title != nil:
		title = *element.Title
	case element.CheckName != nil:
		title = *element.CheckName
	}
	body := publicElementBody{
		ID:                    element.ID,
		Kind:                  element.Kind,
		CheckID:               element.CheckID,
		Title:                 title,
		Description:           elementDescription(element),
		Target:                element.CheckTarget,
		Status:                element.Status,
		LatestStartedAt:       element.LatestStartedAt,
		LatestStatus:          element.LatestStatus,
		AssignmentCount:       element.AssignmentCount,
		SuccessfulAssignments: element.SuccessfulAssignments,
		FailingAssignments:    element.FailingAssignments,
		StaleAssignments:      element.StaleAssignments,
		Metrics:               newMetricsBody(element.Metrics),
		Chart:                 newChartBody(element.Chart),
		Children:              newPublicElementBodies(element.Children),
	}
	if element.CheckType != nil {
		body.Type = *element.CheckType
	}
	return body
}

func elementDescription(element domainpublic.RenderedElement) *string {
	if element.Description != nil {
		return element.Description
	}
	return element.CheckDescription
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

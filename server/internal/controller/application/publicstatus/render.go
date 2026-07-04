package publicstatus

import (
	"context"
	"sort"
	"time"

	domainpublic "github.com/yorukot/netstamp/internal/domain/publicstatus"
)

const dailyStatusDayCount = 30

func groupAssignmentsByElement(assignments []domainpublic.Assignment) map[string][]domainpublic.Assignment {
	values := make(map[string][]domainpublic.Assignment)
	for _, assignment := range assignments {
		values[assignment.ElementID] = append(values[assignment.ElementID], assignment)
	}
	return values
}

func splitIncidents(incidents []domainpublic.Incident) ([]domainpublic.Incident, []domainpublic.Incident) {
	active := make([]domainpublic.Incident, 0)
	resolved := make([]domainpublic.Incident, 0)
	for _, incident := range incidents {
		switch incident.Status {
		case "open", "acknowledged":
			active = append(active, incident)
		case "resolved":
			resolved = append(resolved, incident)
		}
	}
	return active, resolved
}

func activeIncidentSeverityByElement(incidents []domainpublic.Incident, assignmentsByElement map[string][]domainpublic.Assignment) map[string]string {
	values := make(map[string]string)
	for elementID, assignments := range assignmentsByElement {
		for _, incident := range incidents {
			if !incidentMatchesAssignments(incident, assignments) {
				continue
			}
			if incident.Severity == "critical" {
				values[elementID] = incident.Severity
				break
			}
			if _, exists := values[elementID]; !exists {
				values[elementID] = incident.Severity
			}
		}
	}
	return values
}

func incidentMatchesAssignments(incident domainpublic.Incident, assignments []domainpublic.Assignment) bool {
	for _, assignment := range assignments {
		if assignment.CheckID != incident.CheckID {
			continue
		}
		if incident.ProbeID == nil || *incident.ProbeID == assignment.ProbeID {
			return true
		}
	}
	return false
}

func incidentBasedDailyStatus(assignments []domainpublic.Assignment, incidents []domainpublic.Incident, now time.Time) []domainpublic.DailyStatusDay {
	start := dayStartUTC(now).AddDate(0, 0, -(dailyStatusDayCount - 1))
	days := make([]domainpublic.DailyStatusDay, 0, dailyStatusDayCount)
	for i := range dailyStatusDayCount {
		dayStart := start.AddDate(0, 0, i)
		dayEnd := dayStart.AddDate(0, 0, 1)
		day := domainpublic.DailyStatusDay{
			Date:   dayStart,
			Status: domainpublic.StatusOperational,
		}
		if len(assignments) == 0 {
			day.Status = domainpublic.StatusUnknown
			days = append(days, day)
			continue
		}
		for _, incident := range incidents {
			if !incidentMatchesAssignments(incident, assignments) || !incidentOverlapsDay(incident, dayStart, dayEnd, now) {
				continue
			}
			day.IncidentCount++
			day.Severity = highestIncidentSeverity(day.Severity, incident.Severity)
		}
		day.Status = dailyStatusFromSeverity(day.Severity)
		days = append(days, day)
	}
	return days
}

func dayStartUTC(value time.Time) time.Time {
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func incidentOverlapsDay(incident domainpublic.Incident, dayStart, dayEnd, now time.Time) bool {
	incidentEnd := now
	if incident.ResolvedAt != nil {
		incidentEnd = *incident.ResolvedAt
	}
	return incident.OpenedAt.Before(dayEnd) && incidentEnd.After(dayStart)
}

func highestIncidentSeverity(current *string, next string) *string {
	if current == nil || incidentSeverityRank(next) > incidentSeverityRank(*current) {
		value := next
		return &value
	}
	return current
}

func incidentSeverityRank(severity string) int {
	switch severity {
	case "critical":
		return 3
	case "warning":
		return 2
	case "info":
		return 1
	default:
		return 0
	}
}

func dailyStatusFromSeverity(severity *string) domainpublic.Status {
	if severity == nil {
		return domainpublic.StatusOperational
	}
	if *severity == "critical" {
		return domainpublic.StatusDown
	}
	return domainpublic.StatusDegraded
}

func (s *Service) renderElements(
	ctx context.Context,
	page domainpublic.Page,
	elements []domainpublic.Element,
	assignmentsByElement map[string][]domainpublic.Assignment,
	activeSeverityByElement map[string]string,
	now time.Time,
	includeCharts bool,
	chartRangeOverride *domainpublic.ChartRange,
) []domainpublic.RenderedElement {
	childrenByParent := make(map[string][]domainpublic.Element)
	root := make([]domainpublic.Element, 0)
	for _, element := range elements {
		if element.ParentElementID == nil {
			root = append(root, element)
			continue
		}
		childrenByParent[*element.ParentElementID] = append(childrenByParent[*element.ParentElementID], element)
	}
	sortElements(root)
	return s.renderElementList(ctx, page, root, childrenByParent, assignmentsByElement, activeSeverityByElement, now, includeCharts, chartRangeOverride, page.DefaultChartMode, page.DefaultChartRange)
}

func (s *Service) renderElementList(
	ctx context.Context,
	page domainpublic.Page,
	elements []domainpublic.Element,
	childrenByParent map[string][]domainpublic.Element,
	assignmentsByElement map[string][]domainpublic.Assignment,
	activeSeverityByElement map[string]string,
	now time.Time,
	includeCharts bool,
	chartRangeOverride *domainpublic.ChartRange,
	parentChartMode domainpublic.ChartMode,
	parentChartRange domainpublic.ChartRange,
) []domainpublic.RenderedElement {
	rendered := make([]domainpublic.RenderedElement, 0, len(elements))
	for _, element := range elements {
		mode := resolveChartMode(parentChartMode, element.ChartMode)
		chartRange := resolveChartRange(parentChartRange, element.ChartRange)
		if chartRangeOverride != nil {
			chartRange = *chartRangeOverride
		}
		item := domainpublic.RenderedElement{
			Element:            element,
			ResolvedChartMode:  mode,
			ResolvedChartRange: chartRange,
		}
		if element.Kind == domainpublic.ElementKindFolder {
			children := childrenByParent[element.ID]
			sortElements(children)
			item.Children = s.renderElementList(ctx, page, children, childrenByParent, assignmentsByElement, activeSeverityByElement, now, includeCharts, chartRangeOverride, mode, chartRange)
			item.Status = rollupStatus(item.Children)
			rendered = append(rendered, item)
			continue
		}
		if element.Kind == domainpublic.ElementKindAssignmentGroup {
			elementAssignments := assignmentsByElement[element.ID]
			summary := checkStatusSummary(elementAssignments, activeSeverityByElement[element.ID], now)
			item.Status = summary.status
			item.LatestStartedAt = summary.latestStartedAt
			item.LatestStatus = summary.latestStatus
			item.AssignmentCount = summary.assignmentCount
			item.SuccessfulAssignments = summary.successfulAssignments
			item.FailingAssignments = summary.failingAssignments
			item.StaleAssignments = summary.staleAssignments
			item.Metrics = summary.metrics
			item.Assignments = elementAssignments
			if includeCharts && mode == domainpublic.ChartModeCompact {
				item.Chart = s.chartForElement(ctx, page, elementAssignments, chartRange, now)
			}
		}
		if item.Status == "" {
			item.Status = domainpublic.StatusUnknown
		}
		rendered = append(rendered, item)
	}
	return rendered
}

func sortElements(elements []domainpublic.Element) {
	sort.SliceStable(elements, func(i, j int) bool {
		if elements[i].SortOrder != elements[j].SortOrder {
			return elements[i].SortOrder < elements[j].SortOrder
		}
		if !elements[i].CreatedAt.Equal(elements[j].CreatedAt) {
			return elements[i].CreatedAt.Before(elements[j].CreatedAt)
		}
		return elements[i].ID < elements[j].ID
	})
}

func resolveChartMode(parentMode, currentMode domainpublic.ChartMode) domainpublic.ChartMode {
	if currentMode == domainpublic.ChartModeInherit {
		return parentMode
	}
	return currentMode
}

func resolveChartRange(parentRange domainpublic.ChartRange, currentRange *domainpublic.ChartRange) domainpublic.ChartRange {
	if currentRange == nil {
		return parentRange
	}
	return *currentRange
}

type statusSummary struct {
	status                domainpublic.Status
	latestStartedAt       *time.Time
	latestStatus          *string
	assignmentCount       int32
	successfulAssignments int32
	failingAssignments    int32
	staleAssignments      int32
	metrics               *domainpublic.Metrics
}

const maxStatusSummaryCount int32 = 1<<31 - 1

func checkStatusSummary(assignments []domainpublic.Assignment, activeSeverity string, now time.Time) statusSummary {
	summary := statusSummary{}
	var nonStale int32
	var partial bool
	metric := metricAccumulator{}

	for _, assignment := range assignments {
		summary.assignmentCount = incrementStatusCount(summary.assignmentCount)
		if isStaleAssignment(assignment, now) {
			summary.staleAssignments = incrementStatusCount(summary.staleAssignments)
			continue
		}
		nonStale = incrementStatusCount(nonStale)
		summary.recordLatest(assignment)
		partial = summary.recordOutcome(assignment.LatestStatus) || partial
		metric.add(assignment)
	}

	if summary.assignmentCount == 0 {
		summary.status = domainpublic.StatusUnknown
		return summary
	}
	summary.metrics = metric.metrics()
	summary.status = statusFromSummary(activeSeverity, nonStale, summary.failingAssignments, partial)
	return summary
}

func incrementStatusCount(value int32) int32 {
	if value == maxStatusSummaryCount {
		return value
	}
	return value + 1
}

func isStaleAssignment(assignment domainpublic.Assignment, now time.Time) bool {
	if assignment.LatestStatus == "" {
		return true
	}
	staleBefore := now.Add(-time.Duration(assignment.IntervalSeconds) * 3 * time.Second)
	return assignment.LatestStartedAt.Before(staleBefore)
}

func (s *statusSummary) recordLatest(assignment domainpublic.Assignment) {
	if s.latestStartedAt != nil && !assignment.LatestStartedAt.After(*s.latestStartedAt) {
		return
	}
	startedAt := assignment.LatestStartedAt
	status := assignment.LatestStatus
	s.latestStartedAt = &startedAt
	s.latestStatus = &status
}

func (s *statusSummary) recordOutcome(latestStatus string) bool {
	if latestStatus == "successful" {
		s.successfulAssignments = incrementStatusCount(s.successfulAssignments)
		return false
	}
	s.failingAssignments = incrementStatusCount(s.failingAssignments)
	return latestStatus == "partial"
}

func statusFromSummary(activeSeverity string, nonStale, failingAssignments int32, partial bool) domainpublic.Status {
	switch {
	case activeSeverity == "critical":
		return domainpublic.StatusDown
	case activeSeverity == "warning" || activeSeverity == "info":
		return domainpublic.StatusDegraded
	case nonStale == 0:
		return domainpublic.StatusUnknown
	case failingAssignments == 0:
		return domainpublic.StatusOperational
	case failingAssignments == nonStale && !partial:
		return domainpublic.StatusDown
	default:
		return domainpublic.StatusDegraded
	}
}

type metricAccumulator struct {
	latencySum float64
	latencyN   int
	lossSum    float64
	lossN      int
	connectSum float64
	connectN   int
	failureSum float64
	failureN   int
}

func (m *metricAccumulator) add(assignment domainpublic.Assignment) {
	if assignment.LatencyAvgMs != nil {
		m.latencySum += *assignment.LatencyAvgMs
		m.latencyN++
	}
	m.lossSum += assignment.LossPercent
	m.lossN++
	if assignment.ConnectAvgMs != nil {
		m.connectSum += *assignment.ConnectAvgMs
		m.connectN++
	}
	if assignment.FailurePercent != nil {
		m.failureSum += *assignment.FailurePercent
		m.failureN++
	}
}

func (m metricAccumulator) metrics() *domainpublic.Metrics {
	metrics := domainpublic.Metrics{
		LatencyAvgMs:   average(m.latencySum, m.latencyN),
		LossPercent:    average(m.lossSum, m.lossN),
		ConnectAvgMs:   average(m.connectSum, m.connectN),
		FailurePercent: average(m.failureSum, m.failureN),
	}
	if metrics.LatencyAvgMs == nil && metrics.LossPercent == nil && metrics.ConnectAvgMs == nil && metrics.FailurePercent == nil {
		return nil
	}
	return &metrics
}

func average(sum float64, count int) *float64 {
	if count == 0 {
		return nil
	}
	value := sum / float64(count)
	return &value
}

func rollupStatus(elements []domainpublic.RenderedElement) domainpublic.Status {
	if len(elements) == 0 {
		return domainpublic.StatusUnknown
	}
	allUnknown := true
	status := domainpublic.StatusOperational
	for _, element := range elements {
		if element.Status == domainpublic.StatusDown {
			return domainpublic.StatusDown
		}
		if element.Status != domainpublic.StatusUnknown {
			allUnknown = false
		}
		if element.Status == domainpublic.StatusDegraded {
			status = domainpublic.StatusDegraded
		}
	}
	if allUnknown {
		return domainpublic.StatusUnknown
	}
	return status
}

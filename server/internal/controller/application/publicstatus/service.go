package publicstatus

import (
	"context"
	"sort"
	"time"

	"github.com/yorukot/netstamp/internal/controller/application/pingquery"
	"github.com/yorukot/netstamp/internal/controller/application/tcpquery"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainpublic "github.com/yorukot/netstamp/internal/domain/publicstatus"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
)

const (
	defaultIncidentLimit int32 = 50
	publicMaxDataPoints  int32 = 600
)

type Service struct {
	repo          Repository
	projectAccess ProjectAccess
	pings         PingSeriesRepository
	tcps          TCPSeriesRepository
}

func NewService(repo Repository, projectAccess ProjectAccess, pings PingSeriesRepository, tcps TCPSeriesRepository) *Service {
	return &Service{repo: repo, projectAccess: projectAccess, pings: pings, tcps: tcps}
}

func (s *Service) ListPages(ctx context.Context, input ListPagesInput) ([]domainpublic.Page, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListPages(ctx, project.ID)
}

func (s *Service) GetPage(ctx context.Context, input GetPageInput) (PageDetail, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return PageDetail{}, err
	}
	pageID, err := domainpublic.VNPageID(input.PageID)
	if err != nil {
		return PageDetail{}, invalidField("pageId", err.Error(), input.PageID)
	}
	page, err := s.repo.GetPage(ctx, project.ID, pageID)
	if err != nil {
		return PageDetail{}, err
	}
	elements, err := s.repo.ListElements(ctx, page.ID)
	if err != nil {
		return PageDetail{}, err
	}
	return PageDetail{Page: page, Elements: elements}, nil
}

func (s *Service) CreatePage(ctx context.Context, input CreatePageInput) (domainpublic.Page, error) {
	project, err := s.loadWritableProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainpublic.Page{}, err
	}
	page, err := normalizeCreatePageInput(project.ID, input)
	if err != nil {
		return domainpublic.Page{}, err
	}
	return s.repo.CreatePage(ctx, page)
}

func (s *Service) UpdatePage(ctx context.Context, input UpdatePageInput) (domainpublic.Page, error) {
	project, err := s.loadWritableProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainpublic.Page{}, err
	}
	page, err := normalizeUpdatePageInput(project.ID, input)
	if err != nil {
		return domainpublic.Page{}, err
	}
	return s.repo.UpdatePage(ctx, page)
}

func (s *Service) DeletePage(ctx context.Context, input DeletePageInput) error {
	project, err := s.loadWritableProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return err
	}
	pageID, err := domainpublic.VNPageID(input.PageID)
	if err != nil {
		return invalidField("pageId", err.Error(), input.PageID)
	}
	return s.repo.DeletePage(ctx, project.ID, pageID)
}

func (s *Service) CreateElement(ctx context.Context, input CreateElementInput) (domainpublic.Element, error) {
	return s.saveElement(ctx, input.ProjectRef, input.CurrentUserID, func(projectID string) (domainpublic.Element, error) {
		return normalizeCreateElementInput(projectID, input.PageID, input)
	}, s.repo.CreateElement)
}

func (s *Service) UpdateElement(ctx context.Context, input UpdateElementInput) (domainpublic.Element, error) {
	return s.saveElement(ctx, input.ProjectRef, input.CurrentUserID, func(projectID string) (domainpublic.Element, error) {
		return normalizeUpdateElementInput(projectID, input.PageID, input)
	}, s.repo.UpdateElement)
}

func (s *Service) DeleteElement(ctx context.Context, input DeleteElementInput) error {
	project, err := s.loadWritableProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return err
	}
	pageID, err := domainpublic.VNPageID(input.PageID)
	if err != nil {
		return invalidField("pageId", err.Error(), input.PageID)
	}
	elementID, err := domainpublic.VNElementID(input.ElementID)
	if err != nil {
		return invalidField("elementId", err.Error(), input.ElementID)
	}
	return s.repo.DeleteElement(ctx, project.ID, pageID, elementID)
}

func (s *Service) GetPublicPage(ctx context.Context, input PublicPageInput) (domainpublic.RenderedPage, error) {
	now := publicNow(input.Now)
	page, err := s.loadPublicPage(ctx, input.Slug)
	if err != nil {
		return domainpublic.RenderedPage{}, err
	}
	root, activeIncidents, resolvedIncidents, err := s.renderPublicPageData(ctx, page, now, input.IncludeCharts, input.Range)
	if err != nil {
		return domainpublic.RenderedPage{}, err
	}

	return domainpublic.RenderedPage{
		Page:              page,
		Status:            rollupStatus(root),
		Elements:          root,
		ActiveIncidents:   activeIncidents,
		ResolvedIncidents: resolvedIncidents,
		GeneratedAt:       now,
	}, nil
}

func (s *Service) GetPublicSummary(ctx context.Context, input PublicSummaryInput) (PublicSummary, error) {
	now := publicNow(input.Now)
	page, err := s.loadPublicPage(ctx, input.Slug)
	if err != nil {
		return PublicSummary{}, err
	}
	root, _, _, err := s.renderPublicPageData(ctx, page, now, false, nil)
	if err != nil {
		return PublicSummary{}, err
	}
	return PublicSummary{Page: page, Status: rollupStatus(root), GeneratedAt: now}, nil
}

func (s *Service) GetPublicElements(ctx context.Context, input PublicElementsInput) (PublicElements, error) {
	now := publicNow(input.Now)
	page, err := s.loadPublicPage(ctx, input.Slug)
	if err != nil {
		return PublicElements{}, err
	}
	root, _, _, err := s.renderPublicPageData(ctx, page, now, false, nil)
	if err != nil {
		return PublicElements{}, err
	}
	return PublicElements{Elements: root, GeneratedAt: now}, nil
}

func (s *Service) GetPublicIncidents(ctx context.Context, input PublicIncidentsInput) (PublicIncidents, error) {
	now := publicNow(input.Now)
	page, err := s.loadPublicPage(ctx, input.Slug)
	if err != nil {
		return PublicIncidents{}, err
	}
	incidents, err := s.repo.ListIncidents(ctx, page.ID, input.Limit)
	if err != nil {
		return PublicIncidents{}, err
	}
	activeIncidents, resolvedIncidents := splitIncidents(incidents)
	return PublicIncidents{ActiveIncidents: activeIncidents, ResolvedIncidents: resolvedIncidents, GeneratedAt: now}, nil
}

func (s *Service) GetPublicElementChart(ctx context.Context, input PublicElementChartInput) (PublicElementChart, error) {
	now := publicNow(input.Now)
	page, err := s.loadPublicPage(ctx, input.Slug)
	if err != nil {
		return PublicElementChart{}, err
	}
	elementID, err := domainpublic.VNElementID(input.ElementID)
	if err != nil {
		return PublicElementChart{}, invalidField("elementId", err.Error(), input.ElementID)
	}
	elements, err := s.repo.ListElements(ctx, page.ID)
	if err != nil {
		return PublicElementChart{}, err
	}
	element, ok := findPublicElement(elements, elementID)
	if !ok {
		return PublicElementChart{}, domainpublic.ErrElementNotFound
	}
	mode, chartRange := resolvedElementChartSettings(page, elements, element)
	if input.Range != nil {
		chartRange = *input.Range
	}
	if element.Kind != domainpublic.ElementKindAssignmentGroup || mode != domainpublic.ChartModeCompact {
		return PublicElementChart{GeneratedAt: now}, nil
	}
	assignments, err := s.repo.ListAssignments(ctx, page.ID)
	if err != nil {
		return PublicElementChart{}, err
	}
	chart := s.chartForElement(ctx, page, groupAssignmentsByElement(assignments)[element.ID], chartRange, now)
	return PublicElementChart{Chart: chart, GeneratedAt: now}, nil
}

func publicNow(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}
	return value.UTC()
}

func (s *Service) loadPublicPage(ctx context.Context, rawSlug string) (domainpublic.Page, error) {
	slug, err := domainpublic.VNSlug(rawSlug)
	if err != nil {
		return domainpublic.Page{}, invalidField("slug", err.Error(), rawSlug)
	}
	return s.repo.GetPageBySlug(ctx, slug)
}

func (s *Service) renderPublicPageData(
	ctx context.Context,
	page domainpublic.Page,
	now time.Time,
	includeCharts bool,
	chartRangeOverride *domainpublic.ChartRange,
) ([]domainpublic.RenderedElement, []domainpublic.Incident, []domainpublic.Incident, error) {
	elements, err := s.repo.ListElements(ctx, page.ID)
	if err != nil {
		return nil, nil, nil, err
	}
	assignments, err := s.repo.ListAssignments(ctx, page.ID)
	if err != nil {
		return nil, nil, nil, err
	}
	incidents, err := s.repo.ListIncidents(ctx, page.ID, defaultIncidentLimit)
	if err != nil {
		return nil, nil, nil, err
	}

	activeIncidents, resolvedIncidents := splitIncidents(incidents)
	assignmentsByElement := groupAssignmentsByElement(assignments)
	activeSeverityByElement := activeIncidentSeverityByElement(activeIncidents, assignmentsByElement)
	root := s.renderElements(ctx, page, elements, assignmentsByElement, activeSeverityByElement, now, includeCharts, chartRangeOverride)
	return root, activeIncidents, resolvedIncidents, nil
}

func findPublicElement(elements []domainpublic.Element, elementID string) (domainpublic.Element, bool) {
	for _, element := range elements {
		if element.ID == elementID {
			return element, true
		}
	}
	return domainpublic.Element{}, false
}

func resolvedElementChartSettings(page domainpublic.Page, elements []domainpublic.Element, target domainpublic.Element) (domainpublic.ChartMode, domainpublic.ChartRange) {
	byID := make(map[string]domainpublic.Element, len(elements))
	for _, element := range elements {
		byID[element.ID] = element
	}

	path := []domainpublic.Element{target}
	for parentID := target.ParentElementID; parentID != nil; {
		parent, ok := byID[*parentID]
		if !ok {
			break
		}
		path = append(path, parent)
		parentID = parent.ParentElementID
	}

	mode := page.DefaultChartMode
	chartRange := page.DefaultChartRange
	for i := len(path) - 1; i >= 0; i-- {
		mode = resolveChartMode(mode, path[i].ChartMode)
		chartRange = resolveChartRange(chartRange, path[i].ChartRange)
	}
	return mode, chartRange
}

func (s *Service) validateElementReferences(ctx context.Context, element domainpublic.Element) error {
	if _, err := s.repo.GetPage(ctx, element.ProjectID, element.PublicPageID); err != nil {
		return err
	}
	if element.ID != "" {
		current, err := s.repo.GetElement(ctx, element.ProjectID, element.PublicPageID, element.ID)
		if err != nil {
			return err
		}
		if current.Kind != element.Kind {
			return invalidField("kind", "element kind cannot be changed", element.Kind)
		}
	}
	if element.ParentElementID != nil {
		parent, err := s.repo.GetElement(ctx, element.ProjectID, element.PublicPageID, *element.ParentElementID)
		if err != nil {
			return err
		}
		if err := validateParent(parent, element.ID); err != nil {
			return err
		}
	}
	if element.Kind == domainpublic.ElementKindAssignmentGroup {
		if err := s.validateAssignmentGroupScope(ctx, element); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) loadProject(ctx context.Context, projectRef, userID string) (domainproject.Project, error) {
	return s.projectAccess.GetProjectForUser(ctx, projectRef, userID)
}

func (s *Service) loadWritableProject(ctx context.Context, projectRef, userID string) (domainproject.Project, error) {
	project, err := s.loadProject(ctx, projectRef, userID)
	if err != nil {
		return domainproject.Project{}, err
	}
	err = s.requireProjectWrite(ctx, project.ID, userID)
	if err != nil {
		return domainproject.Project{}, err
	}
	return project, nil
}

func (s *Service) requireProjectWrite(ctx context.Context, projectID, userID string) error {
	role, err := s.projectAccess.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		return err
	}
	if !domainproject.Can(role, domainproject.ActionUpdateProject) {
		return ErrForbidden
	}
	return nil
}

func (s *Service) saveElement(
	ctx context.Context,
	projectRef string,
	userID string,
	normalize func(projectID string) (domainpublic.Element, error),
	save func(context.Context, domainpublic.Element) (domainpublic.Element, error),
) (domainpublic.Element, error) {
	project, err := s.loadWritableProject(ctx, projectRef, userID)
	if err != nil {
		return domainpublic.Element{}, err
	}
	element, err := normalize(project.ID)
	if err != nil {
		return domainpublic.Element{}, err
	}
	err = s.validateElementReferences(ctx, element)
	if err != nil {
		return domainpublic.Element{}, err
	}
	return save(ctx, element)
}

func (s *Service) validateAssignmentGroupScope(ctx context.Context, element domainpublic.Element) error {
	if element.AssignmentSelectionMode == nil {
		return invalidField("assignmentSelectionMode", "must be provided for assignment groups", nil)
	}
	switch *element.AssignmentSelectionMode {
	case domainpublic.AssignmentSelectionModeAllCheck:
		if element.CheckID == nil {
			return invalidField("checkId", "must be provided for all-check assignment groups", nil)
		}
		ok, err := s.repo.HasAssignableCheck(ctx, element.ProjectID, *element.CheckID)
		if err != nil {
			return err
		}
		if !ok {
			return invalidField("checkId", "check must be an active ping or tcp check", *element.CheckID)
		}
	case domainpublic.AssignmentSelectionModeSelectedAssignments:
		count, err := s.repo.CountAssignableAssignments(ctx, element.ProjectID, element.AssignmentIDs)
		if err != nil {
			return err
		}
		if count != int64(len(element.AssignmentIDs)) {
			return invalidField("assignmentIds", "assignments must be active ping or tcp assignments in this project", element.AssignmentIDs)
		}
	}
	return nil
}

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

func (s *Service) chartForElement(ctx context.Context, page domainpublic.Page, assignments []domainpublic.Assignment, chartRange domainpublic.ChartRange, now time.Time) *domainpublic.Chart {
	from := chartFrom(now, chartRange)
	var series []domainpublic.Series
	for _, assignment := range assignments {
		switch assignment.CheckType {
		case domaincheck.TypePing:
			series = append(series, s.pingChartSeries(ctx, page.ProjectID, assignment, assignment.CheckID, from, now)...)
		case domaincheck.TypeTCP:
			series = append(series, s.tcpChartSeries(ctx, page.ProjectID, assignment, assignment.CheckID, from, now)...)
		}
	}
	if len(series) == 0 {
		return nil
	}
	return &domainpublic.Chart{Range: chartRange, Series: series}
}

func chartFrom(now time.Time, chartRange domainpublic.ChartRange) time.Time {
	switch chartRange {
	case domainpublic.ChartRange30d:
		return now.Add(-30 * 24 * time.Hour)
	case domainpublic.ChartRange7d:
		return now.Add(-7 * 24 * time.Hour)
	default:
		return now.Add(-24 * time.Hour)
	}
}

type chartSeriesScope struct {
	projectID string
	probeID   string
	checkID   string
	checkName string
	from      time.Time
	to        time.Time
	probeName string
}

func (s *Service) pingChartSeries(ctx context.Context, projectID string, assignment domainpublic.Assignment, checkID string, from, to time.Time) []domainpublic.Series {
	if s.pings == nil {
		return nil
	}
	scope := newChartSeriesScope(projectID, assignment, checkID, from, to)
	reader := pingChartReader{repo: s.pings}
	return readChartSeries(ctx, scope, reader.count, reader.list, reader.series)
}

func (s *Service) tcpChartSeries(ctx context.Context, projectID string, assignment domainpublic.Assignment, checkID string, from, to time.Time) []domainpublic.Series {
	if s.tcps == nil {
		return nil
	}
	scope := newChartSeriesScope(projectID, assignment, checkID, from, to)
	reader := tcpChartReader{repo: s.tcps}
	return readChartSeries(ctx, scope, reader.count, reader.list, reader.series)
}

func newChartSeriesScope(projectID string, assignment domainpublic.Assignment, checkID string, from, to time.Time) chartSeriesScope {
	return chartSeriesScope{
		projectID: projectID,
		probeID:   assignment.ProbeID,
		checkID:   checkID,
		checkName: assignment.CheckName,
		from:      from,
		to:        to,
		probeName: assignment.ProbeName,
	}
}

func readChartSeries[D any](
	ctx context.Context,
	scope chartSeriesScope,
	count func(context.Context, chartSeriesScope) (int64, error),
	list func(context.Context, chartSeriesScope, int64) (map[string]D, error),
	series func(map[string]D, chartSeriesScope) []domainpublic.Series,
) []domainpublic.Series {
	rawPoints, err := count(ctx, scope)
	if err != nil {
		return nil
	}
	values, err := list(ctx, scope, rawPoints)
	if err != nil {
		return nil
	}
	return series(values, scope)
}

type pingChartReader struct {
	repo PingSeriesRepository
}

func (r pingChartReader) count(ctx context.Context, scope chartSeriesScope) (int64, error) {
	return r.repo.CountPingSeriesPoints(ctx, domainping.SeriesPointCountQuery{ProjectID: scope.projectID, ProbeID: scope.probeID, CheckID: scope.checkID, From: scope.from, To: scope.to})
}

func (r pingChartReader) list(ctx context.Context, scope chartSeriesScope, rawPoints int64) (map[string]domainping.SeriesData, error) {
	plan := pingquery.SelectReadPlan(rawPoints, scope.from, scope.to, publicMaxDataPoints)
	return r.repo.ListPingSeries(ctx, domainping.SeriesReadQuery{
		ProjectID:     scope.projectID,
		ProbeID:       scope.probeID,
		CheckID:       scope.checkID,
		From:          scope.from,
		To:            scope.to,
		Series:        []string{"latency_avg"},
		MaxDataPoints: publicMaxDataPoints,
		Mode:          plan.Mode,
	})
}

func (r pingChartReader) series(values map[string]domainping.SeriesData, scope chartSeriesScope) []domainpublic.Series {
	return pingDomainSeries(values, scope.probeName, scope.checkName, scope.checkID, map[string]string{
		"latency_avg": "ms",
	})
}

type tcpChartReader struct {
	repo TCPSeriesRepository
}

func (r tcpChartReader) count(ctx context.Context, scope chartSeriesScope) (int64, error) {
	return r.repo.CountTCPSeriesPoints(ctx, domaintcp.SeriesPointCountQuery{ProjectID: scope.projectID, ProbeID: scope.probeID, CheckID: scope.checkID, From: scope.from, To: scope.to})
}

func (r tcpChartReader) list(ctx context.Context, scope chartSeriesScope, rawPoints int64) (map[string]domaintcp.SeriesData, error) {
	plan := tcpquery.SelectReadPlan(rawPoints, scope.from, scope.to, publicMaxDataPoints)
	return r.repo.ListTCPSeries(ctx, domaintcp.SeriesReadQuery{
		ProjectID:     scope.projectID,
		ProbeID:       scope.probeID,
		CheckID:       scope.checkID,
		From:          scope.from,
		To:            scope.to,
		Series:        []string{"connect_avg"},
		MaxDataPoints: publicMaxDataPoints,
		Mode:          plan.Mode,
	})
}

func (r tcpChartReader) series(values map[string]domaintcp.SeriesData, scope chartSeriesScope) []domainpublic.Series {
	return tcpDomainSeries(values, scope.probeName, scope.checkName, scope.checkID, map[string]string{
		"connect_avg": "ms",
	})
}

func pingDomainSeries(values map[string]domainping.SeriesData, probeName, checkName, checkID string, units map[string]string) []domainpublic.Series {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	series := make([]domainpublic.Series, 0, len(keys))
	for _, key := range keys {
		series = append(series, domainpublic.Series{
			Name: key,
			Labels: map[string]string{
				"checkId":   checkID,
				"checkName": checkName,
				"checkType": "ping",
				"probeName": probeName,
			},
			Unit:   units[key],
			Points: pingSeriesPoints(values[key].Points),
		})
	}
	return series
}

func tcpDomainSeries(values map[string]domaintcp.SeriesData, probeName, checkName, checkID string, units map[string]string) []domainpublic.Series {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	series := make([]domainpublic.Series, 0, len(keys))
	for _, key := range keys {
		series = append(series, domainpublic.Series{
			Name: key,
			Labels: map[string]string{
				"checkId":   checkID,
				"checkName": checkName,
				"checkType": "tcp",
				"probeName": probeName,
			},
			Unit:   units[key],
			Points: tcpSeriesPoints(values[key].Points),
		})
	}
	return series
}

func pingSeriesPoints(points []domainping.SeriesPoint) []domainpublic.SeriesPoint {
	values := make([]domainpublic.SeriesPoint, 0, len(points))
	for _, point := range points {
		values = append(values, domainpublic.SeriesPoint{TimestampMs: point.Timestamp.UTC().UnixMilli(), Value: point.Value})
	}
	return values
}

func tcpSeriesPoints(points []domaintcp.SeriesPoint) []domainpublic.SeriesPoint {
	values := make([]domainpublic.SeriesPoint, 0, len(points))
	for _, point := range points {
		values = append(values, domainpublic.SeriesPoint{TimestampMs: point.Timestamp.UTC().UnixMilli(), Value: point.Value})
	}
	return values
}

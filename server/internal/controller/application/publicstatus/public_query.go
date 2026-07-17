package publicstatus

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel/trace"

	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainpublic "github.com/yorukot/netstamp/internal/domain/publicstatus"
)

const (
	defaultIncidentLimit           = 50
	publicIncidentCacheLimit int32 = 200
)

func (s *Service) GetEditorContext(ctx context.Context, input EditorContextInput) (EditorContext, error) {
	page, err := s.loadPublicPage(ctx, input.Slug)
	if err != nil {
		return EditorContext{}, err
	}
	project, err := s.loadProject(ctx, page.ProjectID, input.CurrentUserID)
	if err != nil {
		return EditorContext{}, err
	}
	role, err := s.projectAccess.GetMemberRole(ctx, project.ID, input.CurrentUserID)
	if err != nil {
		return EditorContext{}, err
	}
	if !domainproject.Can(role, domainproject.ActionUpdateProject) {
		return EditorContext{}, ErrForbidden
	}
	projectRef := project.Slug
	if projectRef == "" {
		projectRef = project.ID
	}
	return EditorContext{ProjectRef: projectRef, PageID: page.ID}, nil
}

func (s *Service) GetPublicSummary(ctx context.Context, input PublicSummaryInput) (PublicSummary, error) {
	ctx, span := startPublicQuery(ctx, "public_status.public.summary.query", "public.summary", input.Slug)
	defer span.End()

	now := publicNow(input.Now)
	snapshot, err := s.publicSnapshot(ctx, input.Slug, now)
	if err != nil {
		return PublicSummary{}, recordPublicQueryFailure(span, PublicStatusReasonPageLookupFailed, err)
	}
	markPublicQuerySuccess(span)
	return PublicSummary{Page: snapshot.page, Status: snapshot.status, GeneratedAt: snapshot.generatedAt}, nil
}

func (s *Service) GetPublicElements(ctx context.Context, input PublicElementsInput) (PublicElements, error) {
	ctx, span := startPublicQuery(ctx, "public_status.public.elements.query", "public.elements", input.Slug)
	defer span.End()

	now := publicNow(input.Now)
	snapshot, err := s.publicSnapshot(ctx, input.Slug, now)
	if err != nil {
		return PublicElements{}, recordPublicQueryFailure(span, PublicStatusReasonElementListFailed, err)
	}
	markPublicQuerySuccess(span)
	return PublicElements{Page: snapshot.page, Elements: snapshot.elements, GeneratedAt: snapshot.generatedAt}, nil
}

func (s *Service) GetPublicIncidents(ctx context.Context, input PublicIncidentsInput) (PublicIncidents, error) {
	ctx, span := startPublicQuery(ctx, "public_status.public.incidents.query", "public.incidents", input.Slug)
	defer span.End()

	now := publicNow(input.Now)
	snapshot, err := s.publicSnapshot(ctx, input.Slug, now)
	if err != nil {
		return PublicIncidents{}, recordPublicQueryFailure(span, PublicStatusReasonIncidentListFailed, err)
	}
	activeIncidents, resolvedIncidents := splitIncidents(limitedIncidents(snapshot.incidents, input.Limit))
	markPublicQuerySuccess(span)
	return PublicIncidents{ActiveIncidents: activeIncidents, ResolvedIncidents: resolvedIncidents, GeneratedAt: snapshot.generatedAt}, nil
}

func (s *Service) GetPublicElementChart(ctx context.Context, input PublicElementChartInput) (PublicElementChart, error) {
	ctx, span := startPublicQuery(ctx, "public_status.public.element_chart.query", "public.element_chart", input.Slug)
	defer span.End()

	now := publicNow(input.Now)
	page, err := s.loadPublicPage(ctx, input.Slug)
	if err != nil {
		return PublicElementChart{}, recordPublicQueryFailure(span, PublicStatusReasonPageLookupFailed, err)
	}
	span.SetAttributes(attrPublicStatusPageID.String(page.ID))
	elementID, err := domainpublic.VNElementID(input.ElementID)
	if err != nil {
		return PublicElementChart{}, recordPublicQueryFailure(span, PublicStatusReasonInvalidInput, invalidField("elementId", err.Error(), input.ElementID))
	}
	span.SetAttributes(attrPublicStatusElementID.String(elementID))
	elements, err := s.repo.ListElements(ctx, page.ID)
	if err != nil {
		return PublicElementChart{}, recordPublicQueryFailure(span, PublicStatusReasonElementListFailed, err)
	}
	element, ok := findPublicElement(elements, elementID)
	if !ok {
		return PublicElementChart{}, recordPublicQueryFailure(span, PublicStatusReasonElementNotFound, domainpublic.ErrElementNotFound)
	}
	mode, chartRange := resolvedElementChartSettings(page, elements, element)
	if input.Range != nil {
		chartRange = *input.Range
	}
	if element.Kind != domainpublic.ElementKindAssignmentGroup || mode != domainpublic.ChartModeCompact {
		markPublicQuerySuccess(span)
		return PublicElementChart{GeneratedAt: now}, nil
	}
	assignments, err := s.repo.ListElementAssignments(ctx, page.ID, element.ID)
	if err != nil {
		return PublicElementChart{}, recordPublicQueryFailure(span, PublicStatusReasonAssignmentListFailed, err)
	}
	chart := s.chartForElement(ctx, page, assignments, chartRange, now)
	markPublicQuerySuccess(span)
	return PublicElementChart{Chart: chart, GeneratedAt: now}, nil
}

func (s *Service) GetPublicElementDailyStatus(ctx context.Context, input PublicElementDailyStatusInput) (PublicElementDailyStatus, error) {
	ctx, span := startPublicQuery(ctx, "public_status.public.element_daily_status.query", "public.element_daily_status", input.Slug)
	defer span.End()

	now := publicNow(input.Now)
	chartRange := domainpublic.ChartRange30d
	if input.Range != nil {
		if *input.Range != domainpublic.ChartRange30d {
			return PublicElementDailyStatus{}, recordPublicQueryFailure(span, PublicStatusReasonInvalidInput, invalidField("range", "must be 30d", *input.Range))
		}
		chartRange = *input.Range
	}
	elementID, err := domainpublic.VNElementID(input.ElementID)
	if err != nil {
		return PublicElementDailyStatus{}, recordPublicQueryFailure(span, PublicStatusReasonInvalidInput, invalidField("elementId", err.Error(), input.ElementID))
	}
	span.SetAttributes(attrPublicStatusElementID.String(elementID))
	snapshot, err := s.publicSnapshot(ctx, input.Slug, now)
	if err != nil {
		return PublicElementDailyStatus{}, recordPublicQueryFailure(span, PublicStatusReasonPageLookupFailed, err)
	}
	element, ok := findRenderedElement(snapshot.elements, elementID)
	if !ok {
		return PublicElementDailyStatus{}, recordPublicQueryFailure(span, PublicStatusReasonElementNotFound, domainpublic.ErrElementNotFound)
	}
	markPublicQuerySuccess(span)
	return PublicElementDailyStatus{
		Range:       chartRange,
		Days:        incidentBasedDailyStatus(collectRenderedAssignments(element), snapshot.incidents, now),
		GeneratedAt: snapshot.generatedAt,
	}, nil
}

func startPublicQuery(ctx context.Context, spanName, action, rawSlug string) (context.Context, trace.Span) {
	return publicStatusTracer.Start(ctx, spanName, trace.WithAttributes(
		attrPublicStatusAction.String(action),
		attrProjectSlug.String(rawSlug),
	))
}

func markPublicQuerySuccess(span trace.Span) {
	span.SetAttributes(attrPublicStatusOutcome.String(string(PublicStatusOutcomeSuccess)))
}

func recordPublicQueryFailure(span trace.Span, fallback PublicStatusReason, err error) error {
	reason := fallback
	switch {
	case errors.Is(err, ErrInvalidInput), errors.Is(err, domainpublic.ErrInvalidInput):
		reason = PublicStatusReasonInvalidInput
	case errors.Is(err, domainpublic.ErrPageNotFound):
		reason = PublicStatusReasonPageNotFound
	case errors.Is(err, domainpublic.ErrElementNotFound):
		reason = PublicStatusReasonElementNotFound
	}

	span.SetAttributes(attrPublicStatusOutcome.String(string(PublicStatusOutcomeFailure)))
	recordSpanError(span, err, reason)
	return err
}

func publicNow(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}
	return value.UTC()
}

func validatePublicSlug(rawSlug string) (string, error) {
	slug, err := domainpublic.VNSlug(rawSlug)
	if err != nil {
		return "", invalidField("slug", err.Error(), rawSlug)
	}
	return slug, nil
}

func (s *Service) loadPublicPage(ctx context.Context, rawSlug string) (domainpublic.Page, error) {
	slug, err := validatePublicSlug(rawSlug)
	if err != nil {
		return domainpublic.Page{}, err
	}
	return s.repo.GetPageBySlug(ctx, slug)
}

func (s *Service) publicSnapshot(ctx context.Context, rawSlug string, now time.Time) (publicSnapshot, error) {
	slug, err := validatePublicSlug(rawSlug)
	if err != nil {
		return publicSnapshot{}, err
	}
	if s.snapshots != nil {
		if snapshot, ok := s.snapshots.get(slug, now); ok {
			return snapshot, nil
		}
	}
	page, err := s.repo.GetPageBySlug(ctx, slug)
	if err != nil {
		return publicSnapshot{}, err
	}
	snapshot, err := s.buildPublicSnapshot(ctx, page, now)
	if err != nil {
		return publicSnapshot{}, err
	}
	if s.snapshots != nil {
		s.snapshots.set(slug, snapshot)
	}
	return snapshot, nil
}

func (s *Service) buildPublicSnapshot(ctx context.Context, page domainpublic.Page, now time.Time) (publicSnapshot, error) {
	elements, err := s.repo.ListElements(ctx, page.ID)
	if err != nil {
		return publicSnapshot{}, err
	}
	assignments, err := s.repo.ListAssignments(ctx, page.ID)
	if err != nil {
		return publicSnapshot{}, err
	}
	incidents, err := s.repo.ListIncidents(ctx, page.ID, publicIncidentCacheLimit)
	if err != nil {
		return publicSnapshot{}, err
	}

	activeIncidents, _ := splitIncidents(incidents)
	assignmentsByElement := groupAssignmentsByElement(assignments)
	activeSeverityByElement := activeIncidentSeverityByElement(activeIncidents, assignmentsByElement)
	root := s.renderElements(ctx, page, elements, assignmentsByElement, activeSeverityByElement, now, false, nil)
	return publicSnapshot{
		page:        page,
		status:      rollupStatus(root),
		elements:    root,
		incidents:   incidents,
		generatedAt: now,
	}, nil
}

func limitedIncidents(incidents []domainpublic.Incident, limit int32) []domainpublic.Incident {
	limitCount := normalizeIncidentLimit(limit)
	if len(incidents) <= limitCount {
		return incidents
	}
	return incidents[:limitCount]
}

func normalizeIncidentLimit(limit int32) int {
	if limit <= 0 || limit > publicIncidentCacheLimit {
		return defaultIncidentLimit
	}
	return int(limit)
}

func findPublicElement(elements []domainpublic.Element, elementID string) (domainpublic.Element, bool) {
	for _, element := range elements {
		if element.ID == elementID {
			return element, true
		}
	}
	return domainpublic.Element{}, false
}

func findRenderedElement(elements []domainpublic.RenderedElement, elementID string) (domainpublic.RenderedElement, bool) {
	for _, element := range elements {
		if element.ID == elementID {
			return element, true
		}
		if child, ok := findRenderedElement(element.Children, elementID); ok {
			return child, true
		}
	}
	return domainpublic.RenderedElement{}, false
}

func collectRenderedAssignments(element domainpublic.RenderedElement) []domainpublic.Assignment {
	assignments := append([]domainpublic.Assignment{}, element.Assignments...)
	for _, child := range element.Children {
		assignments = append(assignments, collectRenderedAssignments(child)...)
	}
	return assignments
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

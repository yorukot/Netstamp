package publicstatus

import (
	"context"
	"errors"
	"testing"
	"time"

	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainpublic "github.com/yorukot/netstamp/internal/domain/publicstatus"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
)

const (
	testProjectID = "11111111-1111-1111-1111-111111111111"
	testUserID    = "22222222-2222-2222-2222-222222222222"
	testPageID    = "33333333-3333-3333-3333-333333333333"
)

func TestPublicSummaryAndElementsRenderOrderedElementsAndRollUpStatus(t *testing.T) {
	now := time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)
	rootCheckID := "44444444-4444-4444-4444-444444444444"
	folderID := "55555555-5555-5555-5555-555555555555"
	childCheckID := "66666666-6666-6666-6666-666666666666"
	checkType := domaincheck.TypePing

	repo := &fakePublicStatusRepository{
		page: testPage(now),
		elements: []domainpublic.Element{
			{
				ID:           folderID,
				PublicPageID: testPageID,
				ProjectID:    testProjectID,
				Kind:         domainpublic.ElementKindFolder,
				Title:        ptr("Core services"),
				SortOrder:    2,
				ChartMode:    domainpublic.ChartModeInherit,
				CreatedAt:    now.Add(-3 * time.Minute),
				UpdatedAt:    now.Add(-3 * time.Minute),
			},
			{
				ID:           rootCheckID,
				PublicPageID: testPageID,
				ProjectID:    testProjectID,
				Kind:         domainpublic.ElementKindAssignmentGroup,
				CheckID:      &rootCheckID,
				CheckName:    ptr("Landing page"),
				CheckType:    &checkType,
				SortOrder:    1,
				ChartMode:    domainpublic.ChartModeInherit,
				CreatedAt:    now.Add(-2 * time.Minute),
				UpdatedAt:    now.Add(-2 * time.Minute),
			},
			{
				ID:              childCheckID,
				PublicPageID:    testPageID,
				ProjectID:       testProjectID,
				ParentElementID: &folderID,
				Kind:            domainpublic.ElementKindAssignmentGroup,
				CheckID:         &childCheckID,
				CheckName:       ptr("API"),
				CheckType:       &checkType,
				SortOrder:       1,
				ChartMode:       domainpublic.ChartModeInherit,
				CreatedAt:       now.Add(-1 * time.Minute),
				UpdatedAt:       now.Add(-1 * time.Minute),
			},
		},
		assignments: []domainpublic.Assignment{
			testAssignment(rootCheckID, "successful", now.Add(-30*time.Second)),
			testAssignment(childCheckID, "error", now.Add(-30*time.Second)),
		},
	}

	service := NewService(repo, nil, nil, nil)
	summary, err := service.GetPublicSummary(context.Background(), PublicSummaryInput{Slug: "main", Now: now})
	if err != nil {
		t.Fatalf("GetPublicSummary returned error: %v", err)
	}
	elements, err := service.GetPublicElements(context.Background(), PublicElementsInput{Slug: "main", Now: now})
	if err != nil {
		t.Fatalf("GetPublicElements returned error: %v", err)
	}

	if summary.Status != domainpublic.StatusDown {
		t.Fatalf("status = %q, want %q", summary.Status, domainpublic.StatusDown)
	}
	if len(elements.Elements) != 2 {
		t.Fatalf("root element count = %d, want 2", len(elements.Elements))
	}
	if elements.Elements[0].ID != rootCheckID {
		t.Fatalf("first root element = %q, want root check %q", elements.Elements[0].ID, rootCheckID)
	}
	if elements.Elements[0].Status != domainpublic.StatusOperational {
		t.Fatalf("root check status = %q, want %q", elements.Elements[0].Status, domainpublic.StatusOperational)
	}
	folder := elements.Elements[1]
	if folder.ID != folderID {
		t.Fatalf("second root element = %q, want folder %q", folder.ID, folderID)
	}
	if folder.Status != domainpublic.StatusDown {
		t.Fatalf("folder status = %q, want %q", folder.Status, domainpublic.StatusDown)
	}
	if len(folder.Children) != 1 || folder.Children[0].ID != childCheckID {
		t.Fatalf("folder children = %#v, want child check %q", folder.Children, childCheckID)
	}
}

func TestPublicSummaryElementsAndIncidentsRespectCriticalIncident(t *testing.T) {
	now := time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)
	checkID := "44444444-4444-4444-4444-444444444444"
	incidentID := "77777777-7777-7777-7777-777777777777"
	resolvedIncidentID := "88888888-8888-8888-8888-888888888888"
	checkType := domaincheck.TypeTCP
	resolvedAt := now.Add(-time.Hour)

	repo := &fakePublicStatusRepository{
		page: testPage(now),
		elements: []domainpublic.Element{
			{
				ID:           checkID,
				PublicPageID: testPageID,
				ProjectID:    testProjectID,
				Kind:         domainpublic.ElementKindAssignmentGroup,
				CheckID:      &checkID,
				CheckName:    ptr("API"),
				CheckType:    &checkType,
				SortOrder:    1,
				ChartMode:    domainpublic.ChartModeInherit,
				CreatedAt:    now.Add(-time.Minute),
				UpdatedAt:    now.Add(-time.Minute),
			},
		},
		assignments: []domainpublic.Assignment{
			testAssignment(checkID, "successful", now.Add(-30*time.Second)),
		},
		incidents: []domainpublic.Incident{
			{
				ID:              incidentID,
				CheckID:         checkID,
				CheckName:       "API",
				Status:          "open",
				Severity:        "critical",
				OpenedAt:        now.Add(-5 * time.Minute),
				LastTriggeredAt: now.Add(-time.Minute),
			},
			{
				ID:              resolvedIncidentID,
				CheckID:         checkID,
				CheckName:       "API",
				Status:          "resolved",
				Severity:        "warning",
				OpenedAt:        now.Add(-2 * time.Hour),
				ResolvedAt:      &resolvedAt,
				LastTriggeredAt: resolvedAt,
			},
		},
	}

	service := NewService(repo, nil, nil, nil)
	summary, err := service.GetPublicSummary(context.Background(), PublicSummaryInput{Slug: "main", Now: now})
	if err != nil {
		t.Fatalf("GetPublicSummary returned error: %v", err)
	}
	elements, err := service.GetPublicElements(context.Background(), PublicElementsInput{Slug: "main", Now: now})
	if err != nil {
		t.Fatalf("GetPublicElements returned error: %v", err)
	}
	incidents, err := service.GetPublicIncidents(context.Background(), PublicIncidentsInput{Slug: "main", Now: now})
	if err != nil {
		t.Fatalf("GetPublicIncidents returned error: %v", err)
	}

	if summary.Status != domainpublic.StatusDown {
		t.Fatalf("status = %q, want %q", summary.Status, domainpublic.StatusDown)
	}
	if elements.Elements[0].Status != domainpublic.StatusDown {
		t.Fatalf("element status = %q, want %q", elements.Elements[0].Status, domainpublic.StatusDown)
	}
	if len(incidents.ActiveIncidents) != 1 || incidents.ActiveIncidents[0].ID != incidentID {
		t.Fatalf("active incidents = %#v, want only %q", incidents.ActiveIncidents, incidentID)
	}
	if len(incidents.ResolvedIncidents) != 1 || incidents.ResolvedIncidents[0].ID != resolvedIncidentID {
		t.Fatalf("resolved incidents = %#v, want only %q", incidents.ResolvedIncidents, resolvedIncidentID)
	}
}

func TestPublicSnapshotReusesCurrentStatusAcrossSplitEndpoints(t *testing.T) {
	now := time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)
	checkID := "44444444-4444-4444-4444-444444444444"
	checkType := domaincheck.TypePing

	repo := &fakePublicStatusRepository{
		page: testPage(now),
		elements: []domainpublic.Element{
			{
				ID:           checkID,
				PublicPageID: testPageID,
				ProjectID:    testProjectID,
				Kind:         domainpublic.ElementKindAssignmentGroup,
				CheckID:      &checkID,
				CheckName:    ptr("API"),
				CheckType:    &checkType,
				SortOrder:    1,
				ChartMode:    domainpublic.ChartModeInherit,
				CreatedAt:    now.Add(-time.Minute),
				UpdatedAt:    now.Add(-time.Minute),
			},
		},
		assignments: []domainpublic.Assignment{
			testAssignment(checkID, "successful", now.Add(-30*time.Second)),
		},
		incidents: []domainpublic.Incident{
			{
				ID:              "77777777-7777-7777-7777-777777777777",
				CheckID:         checkID,
				CheckName:       "API",
				Status:          "resolved",
				Severity:        "warning",
				OpenedAt:        now.Add(-2 * time.Hour),
				LastTriggeredAt: now.Add(-time.Hour),
			},
		},
	}
	service := NewService(repo, nil, nil, nil)

	if _, err := service.GetPublicSummary(context.Background(), PublicSummaryInput{Slug: "main", Now: now}); err != nil {
		t.Fatalf("GetPublicSummary returned error: %v", err)
	}
	if _, err := service.GetPublicElements(context.Background(), PublicElementsInput{Slug: "main", Now: now.Add(time.Second)}); err != nil {
		t.Fatalf("GetPublicElements returned error: %v", err)
	}
	if _, err := service.GetPublicIncidents(context.Background(), PublicIncidentsInput{Slug: "main", Now: now.Add(2 * time.Second)}); err != nil {
		t.Fatalf("GetPublicIncidents returned error: %v", err)
	}

	if repo.getPageBySlugCalls != 1 {
		t.Fatalf("GetPageBySlug calls = %d, want 1", repo.getPageBySlugCalls)
	}
	if repo.listElementsCalls != 1 {
		t.Fatalf("ListElements calls = %d, want 1", repo.listElementsCalls)
	}
	if repo.listAssignmentsCalls != 1 {
		t.Fatalf("ListAssignments calls = %d, want 1", repo.listAssignmentsCalls)
	}
	if repo.listIncidentsCalls != 1 {
		t.Fatalf("ListIncidents calls = %d, want 1", repo.listIncidentsCalls)
	}
}

func TestGetPublicElementsDoesNotReadChartSeries(t *testing.T) {
	now := time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)
	checkID := "44444444-4444-4444-4444-444444444444"
	checkType := domaincheck.TypePing
	page := testPage(now)
	page.DefaultChartMode = domainpublic.ChartModeCompact

	repo := &fakePublicStatusRepository{
		page: page,
		elements: []domainpublic.Element{
			{
				ID:           checkID,
				PublicPageID: testPageID,
				ProjectID:    testProjectID,
				Kind:         domainpublic.ElementKindAssignmentGroup,
				CheckID:      &checkID,
				CheckName:    ptr("API"),
				CheckType:    &checkType,
				SortOrder:    1,
				ChartMode:    domainpublic.ChartModeInherit,
				CreatedAt:    now.Add(-time.Minute),
				UpdatedAt:    now.Add(-time.Minute),
			},
		},
		assignments: []domainpublic.Assignment{
			testAssignment(checkID, "successful", now.Add(-30*time.Second)),
		},
	}
	pings := &fakePingSeriesRepository{}

	elements, err := NewService(repo, nil, pings, nil).GetPublicElements(context.Background(), PublicElementsInput{
		Slug: "main",
		Now:  now,
	})
	if err != nil {
		t.Fatalf("GetPublicElements returned error: %v", err)
	}
	if pings.countCalls != 0 || pings.listCalls != 0 {
		t.Fatalf("chart series was queried for elements endpoint: count=%d list=%d", pings.countCalls, pings.listCalls)
	}
	if len(elements.Elements) != 1 {
		t.Fatalf("element count = %d, want 1", len(elements.Elements))
	}
	if elements.Elements[0].Chart != nil {
		t.Fatalf("elements endpoint returned embedded chart: %#v", elements.Elements[0].Chart)
	}
}

func TestGetPublicElementChartReturnsOnlyTargetElementSeries(t *testing.T) {
	now := time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)
	targetID := "44444444-4444-4444-4444-444444444444"
	otherID := "55555555-5555-5555-5555-555555555555"
	checkType := domaincheck.TypePing
	page := testPage(now)
	page.DefaultChartMode = domainpublic.ChartModeCompact
	page.DefaultChartRange = domainpublic.ChartRange7d
	chartRange := domainpublic.ChartRange30d

	repo := &fakePublicStatusRepository{
		page: page,
		elements: []domainpublic.Element{
			{
				ID:           targetID,
				PublicPageID: testPageID,
				ProjectID:    testProjectID,
				Kind:         domainpublic.ElementKindAssignmentGroup,
				CheckID:      &targetID,
				CheckName:    ptr("API"),
				CheckType:    &checkType,
				SortOrder:    1,
				ChartMode:    domainpublic.ChartModeInherit,
				CreatedAt:    now.Add(-time.Minute),
				UpdatedAt:    now.Add(-time.Minute),
			},
			{
				ID:           otherID,
				PublicPageID: testPageID,
				ProjectID:    testProjectID,
				Kind:         domainpublic.ElementKindAssignmentGroup,
				CheckID:      &otherID,
				CheckName:    ptr("Worker"),
				CheckType:    &checkType,
				SortOrder:    2,
				ChartMode:    domainpublic.ChartModeInherit,
				CreatedAt:    now.Add(-time.Minute),
				UpdatedAt:    now.Add(-time.Minute),
			},
		},
		assignments: []domainpublic.Assignment{
			testAssignment(targetID, "successful", now.Add(-30*time.Second)),
			testAssignment(otherID, "successful", now.Add(-30*time.Second)),
		},
	}
	pings := &fakePingSeriesRepository{pointTime: now.Add(-5 * time.Minute)}

	result, err := NewService(repo, nil, pings, nil).GetPublicElementChart(context.Background(), PublicElementChartInput{
		Slug:      "main",
		ElementID: targetID,
		Range:     &chartRange,
		Now:       now,
	})
	if err != nil {
		t.Fatalf("GetPublicElementChart returned error: %v", err)
	}
	if pings.countCalls != 1 || pings.listCalls != 1 {
		t.Fatalf("chart series calls = count %d list %d, want 1 each", pings.countCalls, pings.listCalls)
	}
	if repo.listAssignmentsCalls != 0 {
		t.Fatalf("ListAssignments calls = %d, want 0", repo.listAssignmentsCalls)
	}
	if repo.listElementAssignmentsCalls != 1 {
		t.Fatalf("ListElementAssignments calls = %d, want 1", repo.listElementAssignmentsCalls)
	}
	if pings.lastList.CheckID != targetID {
		t.Fatalf("chart query check ID = %q, want target %q", pings.lastList.CheckID, targetID)
	}
	if result.Chart == nil {
		t.Fatalf("chart = nil, want data")
	}
	if result.Chart.Range != domainpublic.ChartRange30d {
		t.Fatalf("chart range = %q, want 30d", result.Chart.Range)
	}
	if len(result.Chart.Series) != 1 {
		t.Fatalf("series count = %d, want 1", len(result.Chart.Series))
	}
	if got := result.Chart.Series[0].Labels["checkId"]; got != targetID {
		t.Fatalf("series check label = %q, want target %q", got, targetID)
	}
}

func TestGetPublicElementChartSkipsChartWhenModeOff(t *testing.T) {
	now := time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)
	checkID := "44444444-4444-4444-4444-444444444444"
	checkType := domaincheck.TypePing

	repo := &fakePublicStatusRepository{
		page: testPage(now),
		elements: []domainpublic.Element{
			{
				ID:           checkID,
				PublicPageID: testPageID,
				ProjectID:    testProjectID,
				Kind:         domainpublic.ElementKindAssignmentGroup,
				CheckID:      &checkID,
				CheckName:    ptr("API"),
				CheckType:    &checkType,
				SortOrder:    1,
				ChartMode:    domainpublic.ChartModeInherit,
				CreatedAt:    now.Add(-time.Minute),
				UpdatedAt:    now.Add(-time.Minute),
			},
		},
		assignments: []domainpublic.Assignment{
			testAssignment(checkID, "successful", now.Add(-30*time.Second)),
		},
	}
	pings := &fakePingSeriesRepository{}

	result, err := NewService(repo, nil, pings, nil).GetPublicElementChart(context.Background(), PublicElementChartInput{
		Slug:      "main",
		ElementID: checkID,
		Now:       now,
	})
	if err != nil {
		t.Fatalf("GetPublicElementChart returned error: %v", err)
	}
	if result.Chart != nil {
		t.Fatalf("chart = %#v, want nil", result.Chart)
	}
	if pings.countCalls != 0 || pings.listCalls != 0 {
		t.Fatalf("chart series was queried for off chart: count=%d list=%d", pings.countCalls, pings.listCalls)
	}
}

func TestUpdateElementRejectsKindChange(t *testing.T) {
	now := time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)
	elementID := "55555555-5555-5555-5555-555555555555"
	checkID := "44444444-4444-4444-4444-444444444444"

	repo := &fakePublicStatusRepository{
		page: testPage(now),
		elements: []domainpublic.Element{
			{
				ID:           elementID,
				PublicPageID: testPageID,
				ProjectID:    testProjectID,
				Kind:         domainpublic.ElementKindFolder,
				Title:        ptr("Core services"),
				SortOrder:    1,
				ChartMode:    domainpublic.ChartModeInherit,
			},
		},
	}
	service := NewService(repo, fakeProjectAccess{role: domainproject.RoleAdmin}, nil, nil)

	_, err := service.UpdateElement(context.Background(), UpdateElementInput{
		CurrentUserID:           testUserID,
		ProjectRef:              "main",
		PageID:                  testPageID,
		ElementID:               elementID,
		Kind:                    domainpublic.ElementKindAssignmentGroup,
		CheckID:                 &checkID,
		AssignmentSelectionMode: domainpublic.AssignmentSelectionModeAllCheck,
		SortOrder:               1,
		ChartMode:               domainpublic.ChartModeInherit,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("UpdateElement error = %v, want ErrInvalidInput", err)
	}
	fields, ok := appvalidation.FieldErrors(err)
	if !ok || len(fields) != 1 || fields[0].Field != "kind" {
		t.Fatalf("validation fields = %#v, want kind error", fields)
	}
	if repo.updatedElement != nil {
		t.Fatalf("UpdateElement wrote element despite validation failure")
	}
}

func testPage(now time.Time) domainpublic.Page {
	return domainpublic.Page{
		ID:                testPageID,
		ProjectID:         testProjectID,
		Slug:              "main",
		Title:             "Main status",
		Enabled:           true,
		DefaultChartMode:  domainpublic.ChartModeOff,
		DefaultChartRange: domainpublic.ChartRange24h,
		CreatedByUserID:   testUserID,
		CreatedAt:         now.Add(-time.Hour),
		UpdatedAt:         now.Add(-time.Minute),
	}
}

func testAssignment(checkID, status string, startedAt time.Time) domainpublic.Assignment {
	return domainpublic.Assignment{
		ElementID:       checkID,
		AssignmentID:    "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		CheckID:         checkID,
		CheckName:       "API",
		CheckType:       domaincheck.TypePing,
		CheckTarget:     "example.com",
		IntervalSeconds: 60,
		ProbeID:         "99999999-9999-9999-9999-999999999999",
		ProbeName:       "Tokyo",
		LatestStartedAt: startedAt,
		LatestStatus:    status,
		LossPercent:     0,
	}
}

type fakePublicStatusRepository struct {
	page                        domainpublic.Page
	elements                    []domainpublic.Element
	assignments                 []domainpublic.Assignment
	incidents                   []domainpublic.Incident
	getPageBySlugCalls          int
	listElementsCalls           int
	listAssignmentsCalls        int
	listElementAssignmentsCalls int
	listIncidentsCalls          int
	updatedElement              *domainpublic.Element
}

func (r *fakePublicStatusRepository) ListPages(context.Context, string) ([]domainpublic.Page, error) {
	return []domainpublic.Page{r.page}, nil
}

func (r *fakePublicStatusRepository) GetPage(_ context.Context, projectID, pageID string) (domainpublic.Page, error) {
	if r.page.ProjectID == projectID && r.page.ID == pageID {
		return r.page, nil
	}
	return domainpublic.Page{}, domainpublic.ErrPageNotFound
}

func (r *fakePublicStatusRepository) GetPageBySlug(_ context.Context, slug string) (domainpublic.Page, error) {
	r.getPageBySlugCalls++
	if r.page.Slug == slug && r.page.Enabled {
		return r.page, nil
	}
	return domainpublic.Page{}, domainpublic.ErrPageNotFound
}

func (r *fakePublicStatusRepository) CreatePage(_ context.Context, input domainpublic.Page) (domainpublic.Page, error) {
	return input, nil
}

func (r *fakePublicStatusRepository) UpdatePage(_ context.Context, input domainpublic.Page) (domainpublic.Page, error) {
	return input, nil
}

func (r *fakePublicStatusRepository) DeletePage(context.Context, string, string) error {
	return nil
}

func (r *fakePublicStatusRepository) ListElements(context.Context, string) ([]domainpublic.Element, error) {
	r.listElementsCalls++
	return append([]domainpublic.Element{}, r.elements...), nil
}

func (r *fakePublicStatusRepository) GetElement(_ context.Context, projectID, pageID, elementID string) (domainpublic.Element, error) {
	for _, element := range r.elements {
		if element.ProjectID == projectID && element.PublicPageID == pageID && element.ID == elementID {
			return element, nil
		}
	}
	return domainpublic.Element{}, domainpublic.ErrElementNotFound
}

func (r *fakePublicStatusRepository) CreateElement(_ context.Context, input domainpublic.Element) (domainpublic.Element, error) {
	return input, nil
}

func (r *fakePublicStatusRepository) UpdateElement(_ context.Context, input domainpublic.Element) (domainpublic.Element, error) {
	r.updatedElement = &input
	return input, nil
}

func (r *fakePublicStatusRepository) DeleteElement(context.Context, string, string, string) error {
	return nil
}

func (r *fakePublicStatusRepository) HasAssignableCheck(context.Context, string, string) (bool, error) {
	return true, nil
}

func (r *fakePublicStatusRepository) CountAssignableAssignments(_ context.Context, _ string, assignmentIDs []string) (int64, error) {
	return int64(len(assignmentIDs)), nil
}

func (r *fakePublicStatusRepository) ListAssignments(context.Context, string) ([]domainpublic.Assignment, error) {
	r.listAssignmentsCalls++
	return append([]domainpublic.Assignment{}, r.assignments...), nil
}

func (r *fakePublicStatusRepository) ListElementAssignments(_ context.Context, _ string, elementID string) ([]domainpublic.Assignment, error) {
	r.listElementAssignmentsCalls++
	assignments := make([]domainpublic.Assignment, 0)
	for _, assignment := range r.assignments {
		if assignment.ElementID == elementID {
			assignments = append(assignments, assignment)
		}
	}
	return assignments, nil
}

func (r *fakePublicStatusRepository) ListIncidents(context.Context, string, int32) ([]domainpublic.Incident, error) {
	r.listIncidentsCalls++
	return append([]domainpublic.Incident{}, r.incidents...), nil
}

type fakeProjectAccess struct {
	role domainproject.Role
}

func (a fakeProjectAccess) GetProjectForUser(context.Context, string, string) (domainproject.Project, error) {
	return domainproject.Project{ID: testProjectID, Name: "Main", Slug: "main", CreatedByUserID: testUserID}, nil
}

func (a fakeProjectAccess) GetMemberRole(context.Context, string, string) (domainproject.Role, error) {
	return a.role, nil
}

type fakePingSeriesRepository struct {
	countCalls int
	listCalls  int
	lastCount  domainping.SeriesPointCountQuery
	lastList   domainping.SeriesReadQuery
	pointTime  time.Time
}

func (r *fakePingSeriesRepository) CountPingSeriesPoints(_ context.Context, input domainping.SeriesPointCountQuery) (int64, error) {
	r.countCalls++
	r.lastCount = input
	return 1, nil
}

func (r *fakePingSeriesRepository) ListPingSeries(_ context.Context, input domainping.SeriesReadQuery) (map[string]domainping.SeriesData, error) {
	r.listCalls++
	r.lastList = input
	pointTime := r.pointTime
	if pointTime.IsZero() {
		pointTime = input.To
	}
	return map[string]domainping.SeriesData{
		"latency_avg": {
			Points: []domainping.SeriesPoint{{Timestamp: pointTime, Value: 12.5}},
		},
	}, nil
}

type fakeTCPSeriesRepository struct{}

func (r fakeTCPSeriesRepository) CountTCPSeriesPoints(context.Context, domaintcp.SeriesPointCountQuery) (int64, error) {
	return 0, nil
}

func (r fakeTCPSeriesRepository) ListTCPSeries(context.Context, domaintcp.SeriesReadQuery) (map[string]domaintcp.SeriesData, error) {
	return nil, nil
}

func ptr[T any](value T) *T {
	return &value
}

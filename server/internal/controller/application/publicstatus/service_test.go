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

func TestGetPublicPageRendersOrderedElementsAndRollsUpStatus(t *testing.T) {
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
				Kind:         domainpublic.ElementKindCheck,
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
				Kind:            domainpublic.ElementKindCheck,
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

	rendered, err := NewService(repo, nil, nil, nil).GetPublicPage(context.Background(), PublicPageInput{
		Slug:          "main",
		IncludeCharts: false,
		Now:           now,
	})
	if err != nil {
		t.Fatalf("GetPublicPage returned error: %v", err)
	}

	if rendered.Status != domainpublic.StatusDown {
		t.Fatalf("status = %q, want %q", rendered.Status, domainpublic.StatusDown)
	}
	if len(rendered.Elements) != 2 {
		t.Fatalf("root element count = %d, want 2", len(rendered.Elements))
	}
	if rendered.Elements[0].ID != rootCheckID {
		t.Fatalf("first root element = %q, want root check %q", rendered.Elements[0].ID, rootCheckID)
	}
	if rendered.Elements[0].Status != domainpublic.StatusOperational {
		t.Fatalf("root check status = %q, want %q", rendered.Elements[0].Status, domainpublic.StatusOperational)
	}
	folder := rendered.Elements[1]
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

func TestGetPublicPageCriticalIncidentOverridesSuccessfulAssignments(t *testing.T) {
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
				Kind:         domainpublic.ElementKindCheck,
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

	rendered, err := NewService(repo, nil, nil, nil).GetPublicPage(context.Background(), PublicPageInput{
		Slug:          "main",
		IncludeCharts: false,
		Now:           now,
	})
	if err != nil {
		t.Fatalf("GetPublicPage returned error: %v", err)
	}

	if rendered.Status != domainpublic.StatusDown {
		t.Fatalf("status = %q, want %q", rendered.Status, domainpublic.StatusDown)
	}
	if rendered.Elements[0].Status != domainpublic.StatusDown {
		t.Fatalf("element status = %q, want %q", rendered.Elements[0].Status, domainpublic.StatusDown)
	}
	if len(rendered.ActiveIncidents) != 1 || rendered.ActiveIncidents[0].ID != incidentID {
		t.Fatalf("active incidents = %#v, want only %q", rendered.ActiveIncidents, incidentID)
	}
	if len(rendered.ResolvedIncidents) != 1 || rendered.ResolvedIncidents[0].ID != resolvedIncidentID {
		t.Fatalf("resolved incidents = %#v, want only %q", rendered.ResolvedIncidents, resolvedIncidentID)
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
		CurrentUserID: testUserID,
		ProjectRef:    "main",
		PageID:        testPageID,
		ElementID:     elementID,
		Kind:          domainpublic.ElementKindCheck,
		CheckID:       &checkID,
		SortOrder:     1,
		ChartMode:     domainpublic.ChartModeInherit,
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
		CheckID:         checkID,
		CheckType:       domaincheck.TypePing,
		IntervalSeconds: 60,
		ProbeID:         "99999999-9999-9999-9999-999999999999",
		ProbeName:       "Tokyo",
		LatestStartedAt: startedAt,
		LatestStatus:    status,
		LossPercent:     0,
	}
}

type fakePublicStatusRepository struct {
	page           domainpublic.Page
	elements       []domainpublic.Element
	assignments    []domainpublic.Assignment
	incidents      []domainpublic.Incident
	updatedElement *domainpublic.Element
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

func (r *fakePublicStatusRepository) ListAssignments(context.Context, string) ([]domainpublic.Assignment, error) {
	return append([]domainpublic.Assignment{}, r.assignments...), nil
}

func (r *fakePublicStatusRepository) ListIncidents(context.Context, string, int32) ([]domainpublic.Incident, error) {
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

type unusedPingSeriesRepository struct{}

func (unusedPingSeriesRepository) CountPingSeriesPoints(context.Context, domainping.SeriesPointCountQuery) (int64, error) {
	return 0, nil
}

func (unusedPingSeriesRepository) ListPingSeries(context.Context, domainping.SeriesReadQuery) (map[string]domainping.SeriesData, error) {
	return nil, nil
}

type unusedTCPSeriesRepository struct{}

func (unusedTCPSeriesRepository) CountTCPSeriesPoints(context.Context, domaintcp.SeriesPointCountQuery) (int64, error) {
	return 0, nil
}

func (unusedTCPSeriesRepository) ListTCPSeries(context.Context, domaintcp.SeriesReadQuery) (map[string]domaintcp.SeriesData, error) {
	return nil, nil
}

func ptr[T any](value T) *T {
	return &value
}

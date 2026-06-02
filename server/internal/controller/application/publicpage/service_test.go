package publicpage

import (
	"context"
	"errors"
	"testing"

	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainpublicpage "github.com/yorukot/netstamp/internal/domain/publicpage"
)

const (
	testUserID    = "11111111-1111-1111-1111-111111111111"
	testProjectID = "22222222-2222-2222-2222-222222222222"
	testPageID    = "33333333-3333-3333-3333-333333333333"
	testFolderID  = "44444444-4444-4444-4444-444444444444"
	testChildID   = "55555555-5555-5555-5555-555555555555"
	testCheckID   = "66666666-6666-6666-6666-666666666666"
	testProbeID   = "77777777-7777-7777-7777-777777777777"
)

func TestCreatePageRejectsViewer(t *testing.T) {
	repo := &publicPageRepositoryFake{}
	service := NewService(repo, &projectAccessFake{role: domainproject.RoleViewer}, nil, &publicPageEventRecorderFake{})

	_, err := service.CreatePage(context.Background(), CreatePageInput{
		CurrentUserID: testUserID,
		ProjectRef:    "project",
		Slug:          "public-edge",
		Title:         "Public Edge",
		Enabled:       true,
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
	if repo.createPageCalled {
		t.Fatalf("expected repository create not to be called")
	}
}

func TestUpdateFolderRejectsCycle(t *testing.T) {
	parentID := testFolderID
	repo := &publicPageRepositoryFake{
		folders: []domainpublicpage.Folder{
			{ID: testFolderID, PageID: testPageID, Name: "Root"},
			{ID: testChildID, PageID: testPageID, ParentID: &parentID, Name: "Child"},
		},
	}
	service := NewService(repo, &projectAccessFake{role: domainproject.RoleOwner}, nil, &publicPageEventRecorderFake{})
	newParentID := testChildID

	_, err := service.UpdateFolder(context.Background(), UpdateFolderInput{
		CurrentUserID: testUserID,
		ProjectRef:    "project",
		PageID:        testPageID,
		FolderID:      testFolderID,
		ParentID:      &newParentID,
		ParentIDSet:   true,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
	if repo.updateFolderCalled {
		t.Fatalf("expected repository update not to be called")
	}
}

func TestSetFolderChecksRejectsDuplicateIDs(t *testing.T) {
	repo := &publicPageRepositoryFake{
		folders: []domainpublicpage.Folder{{ID: testFolderID, PageID: testPageID, Name: "Root"}},
	}
	service := NewService(repo, &projectAccessFake{role: domainproject.RoleOwner}, nil, &publicPageEventRecorderFake{})

	_, err := service.SetFolderChecks(context.Background(), SetFolderChecksInput{
		CurrentUserID: testUserID,
		ProjectRef:    "project",
		PageID:        testPageID,
		FolderID:      testFolderID,
		CheckIDs:      []string{testCheckID, testCheckID},
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
	if repo.setFolderChecksCalled {
		t.Fatalf("expected repository set checks not to be called")
	}
}

func TestQueryPublicPingInsightRequiresPublishedPair(t *testing.T) {
	repo := &publicPageRepositoryFake{resolveErr: domainpublicpage.ErrCheckNotPublished}
	pings := &pingInsightRepositoryFake{}
	service := NewService(repo, &projectAccessFake{role: domainproject.RoleOwner}, pings, &publicPageEventRecorderFake{})

	_, err := service.QueryPublicPingInsight(context.Background(), QueryPublicPingInsightInput{
		Slug:    "public-edge",
		ProbeID: testProbeID,
		CheckID: testCheckID,
	})
	if !errors.Is(err, domainpublicpage.ErrCheckNotPublished) {
		t.Fatalf("expected check not published, got %v", err)
	}
	if pings.called {
		t.Fatalf("expected ping repository not to be called")
	}
}

func TestListPagesRecordsTechnicalFailure(t *testing.T) {
	expected := errors.New("database unavailable")
	events := &publicPageEventRecorderFake{}
	service := NewService(&publicPageRepositoryFake{listErr: expected}, &projectAccessFake{role: domainproject.RoleOwner}, nil, events)

	_, err := service.ListPages(context.Background(), ListPagesInput{
		CurrentUserID: testUserID,
		ProjectRef:    "project",
	})
	if !errors.Is(err, expected) {
		t.Fatalf("expected list error, got %v", err)
	}
	if len(events.events) != 1 {
		t.Fatalf("expected one event, got %d", len(events.events))
	}

	event := events.events[0]
	if event.Name != PublicPageEventListFailure {
		t.Fatalf("expected list failure event, got %s", event.Name)
	}
	if event.Reason != PublicPageReasonPageListFailed {
		t.Fatalf("expected page list failed reason, got %s", event.Reason)
	}
	if !errors.Is(event.Err, expected) {
		t.Fatalf("expected event error to preserve cause")
	}
	if event.ProjectID != testProjectID || event.ProjectRef != "project" {
		t.Fatalf("expected project context on event, got id=%q ref=%q", event.ProjectID, event.ProjectRef)
	}
}

type projectAccessFake struct {
	role domainproject.Role
}

func (f *projectAccessFake) GetProjectForUser(context.Context, string, string) (domainproject.Project, error) {
	return domainproject.Project{ID: testProjectID, Slug: "project", Name: "Project"}, nil
}

func (f *projectAccessFake) GetMemberRole(context.Context, string, string) (domainproject.Role, error) {
	return f.role, nil
}

type publicPageRepositoryFake struct {
	createPageCalled      bool
	updateFolderCalled    bool
	setFolderChecksCalled bool
	folders               []domainpublicpage.Folder
	listErr               error
	resolveErr            error
	resolvedPairProjectID string
}

func (f *publicPageRepositoryFake) ListPages(context.Context, string) ([]domainpublicpage.Page, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return nil, nil
}

func (f *publicPageRepositoryFake) GetPageForProject(context.Context, string, string) (domainpublicpage.Page, error) {
	return domainpublicpage.Page{ID: testPageID, ProjectID: testProjectID, Slug: "public-edge", Title: "Public Edge", Enabled: true}, nil
}

func (f *publicPageRepositoryFake) GetEnabledPageBySlug(context.Context, string) (domainpublicpage.Page, error) {
	return domainpublicpage.Page{ID: testPageID, ProjectID: testProjectID, Slug: "public-edge", Title: "Public Edge", Enabled: true}, nil
}

func (f *publicPageRepositoryFake) CreatePage(context.Context, domainpublicpage.Page) (domainpublicpage.Page, error) {
	f.createPageCalled = true
	return domainpublicpage.Page{}, nil
}

func (f *publicPageRepositoryFake) UpdatePage(context.Context, domainpublicpage.PageUpdate) (domainpublicpage.Page, error) {
	return domainpublicpage.Page{}, nil
}

func (f *publicPageRepositoryFake) SoftDeletePage(context.Context, string, string) error {
	return nil
}

func (f *publicPageRepositoryFake) ListFolders(context.Context, string, string) ([]domainpublicpage.Folder, error) {
	return f.folders, nil
}

func (f *publicPageRepositoryFake) ListFolderChecks(context.Context, string, string) ([]domainpublicpage.PublishedCheck, error) {
	return nil, nil
}

func (f *publicPageRepositoryFake) ListPingPairs(context.Context, string, string) ([]domainpublicpage.PingPair, error) {
	return nil, nil
}

func (f *publicPageRepositoryFake) CreateFolder(context.Context, string, domainpublicpage.Folder) (domainpublicpage.Folder, error) {
	return domainpublicpage.Folder{}, nil
}

func (f *publicPageRepositoryFake) UpdateFolder(context.Context, string, domainpublicpage.FolderUpdate) (domainpublicpage.Folder, error) {
	f.updateFolderCalled = true
	return domainpublicpage.Folder{}, nil
}

func (f *publicPageRepositoryFake) DeleteFolder(context.Context, string, string, string) error {
	return nil
}

func (f *publicPageRepositoryFake) SetFolderChecks(context.Context, string, string, string, []string) ([]domainpublicpage.PublishedCheck, error) {
	f.setFolderChecksCalled = true
	return nil, nil
}

func (f *publicPageRepositoryFake) ResolvePublicPingPairProjectID(context.Context, string, string, string) (string, error) {
	if f.resolveErr != nil {
		return "", f.resolveErr
	}
	if f.resolvedPairProjectID != "" {
		return f.resolvedPairProjectID, nil
	}
	return testProjectID, nil
}

type pingInsightRepositoryFake struct {
	called bool
}

func (f *pingInsightRepositoryFake) CountPingSeriesPoints(context.Context, domainping.SeriesPointCountQuery) (int64, int64, error) {
	f.called = true
	return 0, 0, nil
}

func (f *pingInsightRepositoryFake) GetPingInsightSummary(context.Context, domainping.InsightSummaryQuery) (domainping.InsightSummary, error) {
	return domainping.InsightSummary{}, nil
}

type publicPageEventRecorderFake struct {
	events []PublicPageEvent
}

func (f *publicPageEventRecorderFake) RecordPublicPageEvent(_ context.Context, event PublicPageEvent) {
	f.events = append(f.events, event)
}

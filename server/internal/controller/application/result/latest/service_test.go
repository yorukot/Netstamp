package latest

import (
	"context"
	"errors"
	"testing"
	"time"

	resultshared "github.com/yorukot/netstamp/internal/controller/application/result/shared"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainresult "github.com/yorukot/netstamp/internal/domain/result"
)

const (
	testUserID    = "11111111-1111-1111-1111-111111111111"
	testProjectID = "22222222-2222-2222-2222-222222222222"
	testProbeID   = "33333333-3333-3333-3333-333333333333"
	testCheckID   = "44444444-4444-4444-4444-444444444444"
)

func TestQueryMapsLatestResults(t *testing.T) {
	startedAt := time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)
	repo := &recordingRepository{
		results: []domainresult.LatestResult{
			{
				Type:            domainresult.LatestResultTypeTCP,
				ProbeID:         testProbeID,
				CheckID:         testCheckID,
				LatestStartedAt: startedAt,
				LatestStatus:    "successful",
			},
		},
	}
	service := NewService(repo, staticProjectAccess{})

	output, err := service.Query(context.Background(), QueryInput{
		CurrentUserID: testUserID,
		ProjectRef:    "vector-ix",
		ProbeID:       testProbeID,
		CheckID:       testCheckID,
		Type:          "tcp",
	})
	if err != nil {
		t.Fatalf("expected query to succeed: %v", err)
	}

	if repo.got.ProjectID != testProjectID || repo.got.ProbeID != testProbeID || repo.got.CheckID != testCheckID {
		t.Fatalf("unexpected repository input: %#v", repo.got)
	}
	if repo.got.Type == nil || *repo.got.Type != domainresult.LatestResultTypeTCP {
		t.Fatalf("expected tcp type filter, got %#v", repo.got.Type)
	}
	if len(output.Results) != 1 {
		t.Fatalf("expected one result, got %#v", output.Results)
	}
	got := output.Results[0]
	if got.Type != "tcp" || got.ProbeID != testProbeID || got.CheckID != testCheckID || !got.LatestStartedAt.Equal(startedAt) || got.LatestStatus != "successful" {
		t.Fatalf("unexpected output result: %#v", got)
	}
}

func TestQueryRejectsInvalidType(t *testing.T) {
	service := NewService(&recordingRepository{}, staticProjectAccess{})

	_, err := service.Query(context.Background(), QueryInput{
		CurrentUserID: testUserID,
		ProjectRef:    "vector-ix",
		Type:          "dns",
	})
	if !errors.Is(err, resultshared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

type staticProjectAccess struct{}

func (staticProjectAccess) GetProjectForUser(_ context.Context, projectRef, userID string) (domainproject.Project, error) {
	if projectRef != "vector-ix" || userID != testUserID {
		return domainproject.Project{}, domainproject.ErrProjectNotFound
	}
	return domainproject.Project{ID: testProjectID, Slug: "vector-ix"}, nil
}

type recordingRepository struct {
	got     domainresult.LatestResultQuery
	results []domainresult.LatestResult
}

func (r *recordingRepository) ListLatestResults(_ context.Context, input domainresult.LatestResultQuery) (domainresult.LatestResultList, error) {
	r.got = input
	return domainresult.LatestResultList{Results: r.results}, nil
}

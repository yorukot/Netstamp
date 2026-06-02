package tcp

import (
	"context"
	"testing"
	"time"

	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
)

const (
	testUserID    = "11111111-1111-1111-1111-111111111111"
	testProjectID = "22222222-2222-2222-2222-222222222222"
	testProbeID   = "33333333-3333-3333-3333-333333333333"
	testCheckID   = "44444444-4444-4444-4444-444444444444"
)

func TestQueryInsightUsesDefaultsAndMapsBuckets(t *testing.T) {
	now := time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)
	pointTime := now.Add(-time.Hour)
	connectAvg := 42.5
	latestStatus := domaintcp.StatusSuccessful
	latestStartedAt := pointTime
	repo := &recordingInsightRepository{
		result: domaintcp.InsightResult{
			Buckets: []domaintcp.InsightBucket{{
				Timestamp:       pointTime,
				ResultCount:     3,
				ConnectAvgMs:    &connectAvg,
				ConnectMedianMs: &connectAvg,
				SuccessRate:     float64Pointer(100),
			}},
			Summary: domaintcp.InsightSummary{
				TotalResults:    3,
				SuccessfulCount: 3,
				AvgConnectMs:    &connectAvg,
				LatestStatus:    &latestStatus,
				LatestStartedAt: &latestStartedAt,
				LatestConnectMs: &connectAvg,
			},
			Resolution:  domaintcp.InsightResolutionRaw,
			TotalPoints: 3,
		},
	}
	service := NewService(repo, staticProjectAccess{})

	output, err := service.QueryInsight(context.Background(), QueryInsightInput{
		CurrentUserID: testUserID,
		ProjectRef:    "vector-ix",
		ProbeID:       testProbeID,
		CheckID:       testCheckID,
		Now:           now,
	})
	if err != nil {
		t.Fatalf("expected query to succeed: %v", err)
	}

	if repo.got.ProjectID != testProjectID || repo.got.ProbeID != testProbeID || repo.got.CheckID != testCheckID {
		t.Fatalf("unexpected repository identity input: %#v", repo.got)
	}
	if !repo.got.From.Equal(now.Add(-24*time.Hour)) || !repo.got.To.Equal(now) {
		t.Fatalf("unexpected default range: from=%s to=%s", repo.got.From, repo.got.To)
	}
	if output.Query.Resolution != string(domaintcp.InsightResolutionRaw) || output.Query.TotalPoints != 3 {
		t.Fatalf("unexpected query metadata: %#v", output.Query)
	}
	if len(output.Buckets) != 1 || output.Buckets[0].TimestampMs != pointTime.UnixMilli() || output.Buckets[0].ConnectAvgMs == nil || *output.Buckets[0].ConnectAvgMs != connectAvg {
		t.Fatalf("unexpected insight buckets: %#v", output.Buckets)
	}
	if output.Summary.LatestStatus == nil || *output.Summary.LatestStatus != string(domaintcp.StatusSuccessful) || output.Summary.LatestStartedAtMs == nil || *output.Summary.LatestStartedAtMs != pointTime.UnixMilli() {
		t.Fatalf("unexpected summary: %#v", output.Summary)
	}
}

type staticProjectAccess struct{}

func (staticProjectAccess) GetProjectForUser(_ context.Context, projectRef, userID string) (domainproject.Project, error) {
	if projectRef != "vector-ix" || userID != testUserID {
		return domainproject.Project{}, domainproject.ErrProjectNotFound
	}
	return domainproject.Project{ID: testProjectID, Slug: "vector-ix"}, nil
}

type recordingInsightRepository struct {
	got    domaintcp.InsightQuery
	result domaintcp.InsightResult
}

func (r *recordingInsightRepository) ListTCPInsight(_ context.Context, input domaintcp.InsightQuery) (domaintcp.InsightResult, error) {
	r.got = input
	return r.result, nil
}

func float64Pointer(value float64) *float64 {
	return &value
}

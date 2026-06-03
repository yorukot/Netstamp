package tcp

import (
	"context"
	"slices"
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

func TestQuerySeriesUsesDefaultsAndMapsPoints(t *testing.T) {
	now := time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)
	pointTime := now.Add(-time.Hour)
	repo := &recordingInsightRepository{
		rawPoints: 1,
		series: map[string]domaintcp.SeriesData{
			string(SeriesConnectAvg): {
				Points: []domaintcp.SeriesPoint{{Timestamp: pointTime, Value: 42.5}},
			},
		},
	}
	service := NewService(repo, staticProjectAccess{})

	output, err := service.QuerySeries(context.Background(), QuerySeriesInput{
		CurrentUserID: testUserID,
		ProjectRef:    "vector-ix",
		ProbeID:       testProbeID,
		CheckID:       testCheckID,
		Now:           now,
	})
	if err != nil {
		t.Fatalf("expected query to succeed: %v", err)
	}

	if repo.gotCount.ProjectID != testProjectID || repo.gotCount.ProbeID != testProbeID || repo.gotCount.CheckID != testCheckID {
		t.Fatalf("unexpected count identity input: %#v", repo.gotCount)
	}
	if !repo.gotSeries.From.Equal(now.Add(-24*time.Hour)) || !repo.gotSeries.To.Equal(now) {
		t.Fatalf("unexpected default range: from=%s to=%s", repo.gotSeries.From, repo.gotSeries.To)
	}
	wantSeries := []string{
		string(SeriesConnectAvg),
		string(SeriesConnectMin),
		string(SeriesConnectMax),
		string(SeriesFailurePercent),
	}
	if !slices.Equal(repo.gotSeries.Series, wantSeries) {
		t.Fatalf("expected default series %#v, got %#v", wantSeries, repo.gotSeries.Series)
	}
	if repo.gotSeries.Mode != domaintcp.SeriesReadModeRaw {
		t.Fatalf("expected raw read mode, got %s", repo.gotSeries.Mode)
	}
	if output.Meta.Resolution != string(domaintcp.SeriesResolutionRaw) || output.Meta.Source != string(domaintcp.SeriesSourceRaw) || output.Meta.TotalPoints != 1 {
		t.Fatalf("unexpected query sampling metadata: %#v", output.Meta)
	}
	if got := output.Series[string(SeriesConnectAvg)].Points[0]; got.TimestampMs != pointTime.UnixMilli() || got.Value != 42.5 {
		t.Fatalf("unexpected mapped point: %#v", got)
	}
}

func TestQueryInsightUsesRollupPastRawRetention(t *testing.T) {
	now := time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)
	from := now.Add(-72*time.Hour - time.Millisecond)
	fromMs := from.UnixMilli()
	toMs := now.UnixMilli()
	averageConnect := 42.5
	maxConnect := 82.5
	failurePercent := 1.25
	successRate := 98.75
	repo := &recordingInsightRepository{
		rawPoints: 0,
		summary: domaintcp.InsightSummary{
			TotalResults:     12,
			AverageConnectMs: &averageConnect,
			MaxConnectMs:     &maxConnect,
			FailurePercent:   &failurePercent,
			SuccessRate:      &successRate,
			Samples:          12,
		},
	}
	service := NewService(repo, staticProjectAccess{})

	output, err := service.QueryInsight(context.Background(), QueryInsightInput{
		CurrentUserID: testUserID,
		ProjectRef:    "vector-ix",
		ProbeID:       testProbeID,
		CheckID:       testCheckID,
		FromMs:        &fromMs,
		ToMs:          &toMs,
		Now:           now,
	})
	if err != nil {
		t.Fatalf("expected query to succeed: %v", err)
	}

	if repo.gotSummary.Source != domaintcp.SeriesSourceAggregate {
		t.Fatalf("expected aggregate summary source, got %s", repo.gotSummary.Source)
	}
	if output.Meta.Resolution != string(domaintcp.SeriesResolutionOneMinute) || output.Meta.Source != string(domaintcp.SeriesSourceAggregate) || output.Meta.TotalPoints != 12 {
		t.Fatalf("unexpected query metadata: %#v", output.Meta)
	}
	if output.Summary.AverageConnectMs == nil || *output.Summary.AverageConnectMs != averageConnect || output.Summary.MaxConnectMs == nil || *output.Summary.MaxConnectMs != maxConnect || output.Summary.Samples != 12 {
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
	gotCount   domaintcp.SeriesPointCountQuery
	gotSeries  domaintcp.SeriesReadQuery
	gotSummary domaintcp.InsightSummaryQuery
	rawPoints  int64
	series     map[string]domaintcp.SeriesData
	summary    domaintcp.InsightSummary
}

func (r *recordingInsightRepository) CountTCPSeriesPoints(_ context.Context, input domaintcp.SeriesPointCountQuery) (int64, error) {
	r.gotCount = input
	return r.rawPoints, nil
}

func (r *recordingInsightRepository) ListTCPSeries(_ context.Context, input domaintcp.SeriesReadQuery) (map[string]domaintcp.SeriesData, error) {
	r.gotSeries = input
	return r.series, nil
}

func (r *recordingInsightRepository) GetTCPInsightSummary(_ context.Context, input domaintcp.InsightSummaryQuery) (domaintcp.InsightSummary, error) {
	r.gotSummary = input
	return r.summary, nil
}

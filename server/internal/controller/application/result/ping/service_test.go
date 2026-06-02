package ping

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	resultshared "github.com/yorukot/netstamp/internal/controller/application/result/shared"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
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
	repo := &recordingSeriesRepository{
		rawPoints: 1,
		series: map[string]domainping.SeriesData{
			string(SeriesLatencyAvg): {
				Points: []domainping.SeriesPoint{{Timestamp: pointTime, Value: 42.5}},
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
		string(SeriesLatencyAvg),
		string(SeriesLatencyMin),
		string(SeriesLatencyMax),
		string(SeriesLossPercent),
	}
	if !slices.Equal(repo.gotSeries.Series, wantSeries) {
		t.Fatalf("expected default series %#v, got %#v", wantSeries, repo.gotSeries.Series)
	}
	if repo.gotSeries.Mode != domainping.SeriesReadModeRaw {
		t.Fatalf("expected raw read mode, got %s", repo.gotSeries.Mode)
	}
	if output.Meta.Resolution != string(domainping.SeriesResolutionRaw) || output.Meta.Source != string(domainping.SeriesSourceRaw) || output.Meta.TotalPoints != 1 {
		t.Fatalf("unexpected query sampling metadata: %#v", output.Meta)
	}
	if got := output.Series[string(SeriesLatencyAvg)].Points[0]; got.TimestampMs != pointTime.UnixMilli() || got.Value != 42.5 {
		t.Fatalf("unexpected mapped point: %#v", got)
	}
}

func TestQuerySeriesRejectsInvalidSeries(t *testing.T) {
	service := NewService(&recordingSeriesRepository{}, staticProjectAccess{})

	_, err := service.QuerySeries(context.Background(), QuerySeriesInput{
		CurrentUserID: testUserID,
		ProjectRef:    "vector-ix",
		ProbeID:       testProbeID,
		CheckID:       testCheckID,
		Series:        "median",
		Now:           time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC),
	})
	if !errors.Is(err, resultshared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestQueryInsightUsesDefaultsAndMapsSummary(t *testing.T) {
	now := time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)
	rttAvg := 42.5
	rttMax := 82.5
	lossPercent := 1.25
	successRate := 98.5
	repo := &recordingSeriesRepository{
		rawPoints: 3,
		summary: domainping.InsightSummary{
			AverageRttMs: &rttAvg,
			MaxRttMs:     &rttMax,
			LossPercent:  &lossPercent,
			SuccessRate:  &successRate,
			Samples:      12,
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

	if repo.gotSummary.Source != domainping.SeriesSourceRaw {
		t.Fatalf("expected raw summary source, got %s", repo.gotSummary.Source)
	}
	if output.Meta.Resolution != string(domainping.SeriesResolutionRaw) || output.Meta.Source != string(domainping.SeriesSourceRaw) || output.Meta.TotalPoints != 3 {
		t.Fatalf("unexpected query metadata: %#v", output.Meta)
	}
	if output.Summary.AverageRttMs == nil || *output.Summary.AverageRttMs != rttAvg || output.Summary.MaxRttMs == nil || *output.Summary.MaxRttMs != rttMax || output.Summary.Samples != 12 {
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

type recordingSeriesRepository struct {
	gotCount     domainping.SeriesPointCountQuery
	gotSeries    domainping.SeriesReadQuery
	gotSummary   domainping.InsightSummaryQuery
	rawPoints    int64
	rollupPoints int64
	series       map[string]domainping.SeriesData
	summary      domainping.InsightSummary
}

func (r *recordingSeriesRepository) CountPingSeriesPoints(_ context.Context, input domainping.SeriesPointCountQuery) (int64, int64, error) {
	r.gotCount = input
	return r.rawPoints, r.rollupPoints, nil
}

func (r *recordingSeriesRepository) ListPingSeries(_ context.Context, input domainping.SeriesReadQuery) (map[string]domainping.SeriesData, error) {
	r.gotSeries = input
	return r.series, nil
}

func (r *recordingSeriesRepository) GetPingInsightSummary(_ context.Context, input domainping.InsightSummaryQuery) (domainping.InsightSummary, error) {
	r.gotSummary = input
	return r.summary, nil
}

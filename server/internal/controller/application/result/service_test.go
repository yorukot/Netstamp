package result

import (
	"context"
	"errors"
	"testing"
	"time"

	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

const (
	testUserID    = "11111111-1111-1111-1111-111111111111"
	testProjectID = "22222222-2222-2222-2222-222222222222"
	testProbeID   = "33333333-3333-3333-3333-333333333333"
	testCheckID   = "44444444-4444-4444-4444-444444444444"
)

func TestQueryPingSeriesUsesDefaultsAndMapsPoints(t *testing.T) {
	now := time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)
	pointTime := now.Add(-time.Hour)
	pings := &recordingPingSeriesRepository{
		result: domainping.SeriesResult{
			Points: []domainping.SeriesPoint{{
				Timestamp: pointTime,
				Value:     42.5,
			}},
			Resolution:  domainping.SeriesResolutionRaw,
			TotalPoints: 1,
		},
	}
	service := NewService(pings, staticProjectAccess{})

	output, err := service.QueryPingSeries(context.Background(), QueryPingSeriesInput{
		CurrentUserID: testUserID,
		ProjectRef:    "vector-ix",
		ProbeID:       testProbeID,
		CheckID:       testCheckID,
		Now:           now,
	})
	if err != nil {
		t.Fatalf("expected query to succeed: %v", err)
	}

	if pings.got.ProjectID != testProjectID || pings.got.ProbeID != testProbeID || pings.got.CheckID != testCheckID {
		t.Fatalf("unexpected repository identity input: %#v", pings.got)
	}
	if !pings.got.From.Equal(now.Add(-24*time.Hour)) || !pings.got.To.Equal(now) {
		t.Fatalf("unexpected default range: from=%s to=%s", pings.got.From, pings.got.To)
	}
	if pings.got.Metric != string(PingMetricRTTAvgMS) {
		t.Fatalf("expected default metric %q, got %q", PingMetricRTTAvgMS, pings.got.Metric)
	}
	if pings.got.MaxDataPoints != defaultMaxDataPoint {
		t.Fatalf("expected default max data points %d, got %d", defaultMaxDataPoint, pings.got.MaxDataPoints)
	}
	if output.Query.FromMs != now.Add(-24*time.Hour).UnixMilli() || output.Query.ToMs != now.UnixMilli() {
		t.Fatalf("unexpected output query metadata: %#v", output.Query)
	}
	if len(output.Series) != 1 || len(output.Series[0].Points) != 1 {
		t.Fatalf("expected one series with one point, got %#v", output.Series)
	}
	if got := output.Series[0].Points[0]; got.TimestampMs != pointTime.UnixMilli() || got.Value != 42.5 {
		t.Fatalf("unexpected mapped point: %#v", got)
	}
	if output.Query.Resolution != string(domainping.SeriesResolutionRaw) || output.Query.TotalPoints != 1 {
		t.Fatalf("unexpected query sampling metadata: %#v", output.Query)
	}
}

func TestQueryPingSeriesRejectsInvalidMetric(t *testing.T) {
	service := NewService(&recordingPingSeriesRepository{}, staticProjectAccess{})

	_, err := service.QueryPingSeries(context.Background(), QueryPingSeriesInput{
		CurrentUserID: testUserID,
		ProjectRef:    "vector-ix",
		ProbeID:       testProbeID,
		CheckID:       testCheckID,
		Metric:        "median",
		Now:           time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC),
	})
	if !errors.Is(err, ErrInvalidInput) {
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

type recordingPingSeriesRepository struct {
	got    domainping.SeriesQuery
	result domainping.SeriesResult
}

func (r *recordingPingSeriesRepository) ListPingSeries(_ context.Context, input domainping.SeriesQuery) (domainping.SeriesResult, error) {
	r.got = input
	return r.result, nil
}

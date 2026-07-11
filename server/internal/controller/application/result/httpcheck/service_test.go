package httpcheck

import (
	"context"
	"slices"
	"testing"
	"time"

	domainhttp "github.com/yorukot/netstamp/internal/domain/httpcheck"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

const (
	testUserID    = "11111111-1111-1111-1111-111111111111"
	testProjectID = "22222222-2222-2222-2222-222222222222"
	testProbeID   = "33333333-3333-3333-3333-333333333333"
	testCheckID   = "44444444-4444-4444-4444-444444444444"
)

func TestQuerySeriesUsesDefaultsAndMapsPoints(t *testing.T) {
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	pointTime := now.Add(-time.Hour)
	repo := &recordingRepository{
		rawPoints: 1,
		series: map[string]domainhttp.SeriesData{
			string(SeriesTotalAvg): {
				Points: []domainhttp.SeriesPoint{{Timestamp: pointTime, Value: 42.5}},
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

	wantSeries := []string{
		string(SeriesDNSAvg),
		string(SeriesConnectAvg),
		string(SeriesTLSAvg),
		string(SeriesTTFBAvg),
		string(SeriesTotalAvg),
		string(SeriesFailurePercent),
	}
	if !slices.Equal(repo.gotSeries.Series, wantSeries) {
		t.Fatalf("expected default series %#v, got %#v", wantSeries, repo.gotSeries.Series)
	}
	if repo.gotSeries.Mode != domainhttp.SeriesReadModeRaw {
		t.Fatalf("expected raw read mode, got %s", repo.gotSeries.Mode)
	}
	if output.Meta.Resolution != string(domainhttp.SeriesResolutionRaw) || output.Meta.Source != string(domainhttp.SeriesSourceRaw) || output.Meta.TotalPoints != 1 {
		t.Fatalf("unexpected query metadata: %#v", output.Meta)
	}
	if got := output.Series[string(SeriesTotalAvg)].Points[0]; got.TimestampMs != pointTime.UnixMilli() || got.Value != 42.5 {
		t.Fatalf("unexpected mapped point: %#v", got)
	}
}

func TestQueryInsightUsesRollupPastRawRetention(t *testing.T) {
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	fromMs := now.Add(-72*time.Hour - time.Millisecond).UnixMilli()
	toMs := now.UnixMilli()
	averageTotal := 42.5
	certificateDays := 14.25
	repo := &recordingRepository{
		summary: domainhttp.InsightSummary{
			TotalResults:             12,
			AverageTotalMs:           &averageTotal,
			CertificateDaysRemaining: &certificateDays,
			Samples:                  12,
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

	if repo.gotSummary.Source != domainhttp.SeriesSourceAggregate {
		t.Fatalf("expected aggregate summary source, got %s", repo.gotSummary.Source)
	}
	if output.Meta.Resolution != string(domainhttp.SeriesResolutionOneMinute) || output.Meta.Source != string(domainhttp.SeriesSourceAggregate) || output.Meta.TotalPoints != 12 {
		t.Fatalf("unexpected query metadata: %#v", output.Meta)
	}
	if output.Summary.AverageTotalMs == nil || *output.Summary.AverageTotalMs != averageTotal || output.Summary.CertificateDaysRemaining == nil || *output.Summary.CertificateDaysRemaining != certificateDays {
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

type recordingRepository struct {
	gotCount   domainhttp.SeriesPointCountQuery
	gotSeries  domainhttp.SeriesReadQuery
	gotSummary domainhttp.InsightSummaryQuery
	rawPoints  int64
	series     map[string]domainhttp.SeriesData
	summary    domainhttp.InsightSummary
}

func (r *recordingRepository) CountHTTPSeriesPoints(_ context.Context, input domainhttp.SeriesPointCountQuery) (int64, error) {
	r.gotCount = input
	return r.rawPoints, nil
}

func (r *recordingRepository) ListHTTPSeries(_ context.Context, input domainhttp.SeriesReadQuery) (map[string]domainhttp.SeriesData, error) {
	r.gotSeries = input
	return r.series, nil
}

func (r *recordingRepository) GetHTTPInsightSummary(_ context.Context, input domainhttp.InsightSummaryQuery) (domainhttp.InsightSummary, error) {
	r.gotSummary = input
	return r.summary, nil
}

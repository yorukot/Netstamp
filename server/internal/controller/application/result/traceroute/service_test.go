package traceroute

import (
	"context"
	"testing"
	"time"

	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

const (
	serviceTestUserID    = "11111111-1111-1111-1111-111111111111"
	serviceTestProjectID = "22222222-2222-2222-2222-222222222222"
	serviceTestProbeID   = "33333333-3333-3333-3333-333333333333"
	serviceTestCheckID   = "44444444-4444-4444-4444-444444444444"
)

func TestQueryRunsPassesRawCutoff(t *testing.T) {
	now := time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)
	repo := &recordingRunsRepository{}
	service := NewService(repo, staticProjectAccess{})

	_, err := service.QueryRuns(context.Background(), QueryRunsInput{
		CurrentUserID: serviceTestUserID,
		ProjectRef:    "vector-ix",
		ProbeID:       serviceTestProbeID,
		CheckID:       serviceTestCheckID,
		Now:           now,
	})
	if err != nil {
		t.Fatalf("expected query to succeed: %v", err)
	}

	if !repo.gotRuns.RawCutoff.Equal(now.Add(-rawRetentionWindow)) {
		t.Fatalf("unexpected raw cutoff: got %s want %s", repo.gotRuns.RawCutoff, now.Add(-rawRetentionWindow))
	}
}

func TestQueryInsightPassesRawCutoff(t *testing.T) {
	now := time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)
	repo := &recordingRunsRepository{}
	service := NewService(repo, staticProjectAccess{})

	_, err := service.QueryInsight(context.Background(), QueryInsightInput{
		CurrentUserID: serviceTestUserID,
		ProjectRef:    "vector-ix",
		ProbeID:       serviceTestProbeID,
		CheckID:       serviceTestCheckID,
		Now:           now,
	})
	if err != nil {
		t.Fatalf("expected query to succeed: %v", err)
	}

	if !repo.gotInsight.RawCutoff.Equal(now.Add(-rawRetentionWindow)) {
		t.Fatalf("unexpected raw cutoff: got %s want %s", repo.gotInsight.RawCutoff, now.Add(-rawRetentionWindow))
	}
}

func TestQueryTopologyPassesRawCutoff(t *testing.T) {
	now := time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)
	repo := &recordingRunsRepository{}
	service := NewService(repo, staticProjectAccess{})

	_, err := service.QueryTopology(context.Background(), QueryTopologyInput{
		CurrentUserID: serviceTestUserID,
		ProjectRef:    "vector-ix",
		ProbeID:       serviceTestProbeID,
		CheckID:       serviceTestCheckID,
		Now:           now,
	})
	if err != nil {
		t.Fatalf("expected query to succeed: %v", err)
	}

	if !repo.gotTopology.RawCutoff.Equal(now.Add(-rawRetentionWindow)) {
		t.Fatalf("unexpected raw cutoff: got %s want %s", repo.gotTopology.RawCutoff, now.Add(-rawRetentionWindow))
	}
}

type staticProjectAccess struct{}

func (staticProjectAccess) GetProjectForUser(_ context.Context, projectRef, userID string) (domainproject.Project, error) {
	if projectRef != "vector-ix" || userID != serviceTestUserID {
		return domainproject.Project{}, domainproject.ErrProjectNotFound
	}
	return domainproject.Project{ID: serviceTestProjectID, Slug: "vector-ix"}, nil
}

type recordingRunsRepository struct {
	gotRuns     domaintraceroute.RunQuery
	gotInsight  domaintraceroute.InsightQuery
	gotTopology domaintraceroute.TopologyQuery
}

func (r *recordingRunsRepository) ListTracerouteRuns(_ context.Context, input domaintraceroute.RunQuery) (domaintraceroute.RunResult, error) {
	r.gotRuns = input
	return domaintraceroute.RunResult{}, nil
}

func (r *recordingRunsRepository) ListTracerouteInsight(_ context.Context, input domaintraceroute.InsightQuery) (domaintraceroute.InsightResult, error) {
	r.gotInsight = input
	return domaintraceroute.InsightResult{}, nil
}

func (r *recordingRunsRepository) ListTracerouteTopologyRuns(_ context.Context, input domaintraceroute.TopologyQuery) (domaintraceroute.TopologyRunResult, error) {
	r.gotTopology = input
	return domaintraceroute.TopologyRunResult{}, nil
}

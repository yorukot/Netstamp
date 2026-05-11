package reporter

import (
	"strconv"
	"testing"

	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

func TestBatchesSplitsResultsAtMaxBatchSize(t *testing.T) {
	results := make([]domainprobe.Result, MaxBatchSize+1)
	for i := range results {
		results[i] = domainprobe.Result{AssignmentID: strconv.Itoa(i + 1)}
	}

	batches := New().Batches("probe-1", results)
	if len(batches) != 2 {
		t.Fatalf("expected two batches, got %d", len(batches))
	}
	if len(batches[0].Results) != MaxBatchSize {
		t.Fatalf("expected first batch size %d, got %d", MaxBatchSize, len(batches[0].Results))
	}
	if len(batches[1].Results) != 1 {
		t.Fatalf("expected second batch size 1, got %d", len(batches[1].Results))
	}
	if batches[0].ProbeID != "probe-1" || batches[1].ProbeID != "probe-1" {
		t.Fatalf("expected probe id on every batch, got %#v", batches)
	}
}

func TestBatchesReturnsNilForNoResults(t *testing.T) {
	if batches := New().Batches("probe-1", nil); batches != nil {
		t.Fatalf("expected nil batches, got %#v", batches)
	}
}

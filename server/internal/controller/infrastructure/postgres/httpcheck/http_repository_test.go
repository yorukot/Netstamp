package pghttpcheck

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
)

func TestMapLatestHTTPResultsPreservesTLSMetadata(t *testing.T) {
	startedAt := time.Date(2026, 7, 18, 9, 30, 0, 0, time.UTC)
	finishedAt := startedAt.Add(125 * time.Millisecond)
	tlsVersion := "TLS 1.3"
	certificateNotAfter := startedAt.Add(30 * 24 * time.Hour)
	rows := []sqlc.ListLatestHTTPResultsRow{{
		ProbeID:             uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		CheckID:             uuid.MustParse("44444444-4444-4444-4444-444444444444"),
		StartedAt:           startedAt,
		FinishedAt:          finishedAt,
		DurationMs:          125,
		Status:              sqlc.HttpStatusSuccessful,
		TlsVersion:          &tlsVersion,
		CertificateNotAfter: &certificateNotAfter,
	}}

	got := mapLatestHTTPResults(rows)
	if len(got.Results) != 1 {
		t.Fatalf("expected one result, got %#v", got.Results)
	}
	result := got.Results[0]
	if result.Result.TLSVersion == nil || *result.Result.TLSVersion != tlsVersion || result.Result.CertificateNotAfter == nil || !result.Result.CertificateNotAfter.Equal(certificateNotAfter) {
		t.Fatalf("unexpected mapped result: %#v", result)
	}
}

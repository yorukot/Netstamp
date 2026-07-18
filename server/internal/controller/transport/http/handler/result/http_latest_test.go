package result

import (
	"testing"
	"time"

	appresult "github.com/yorukot/netstamp/internal/controller/application/result"
	domainhttp "github.com/yorukot/netstamp/internal/domain/httpcheck"
)

func TestNewQueryLatestHTTPResultsBodyMapsTLSMetadata(t *testing.T) {
	startedAt := time.Date(2026, 7, 18, 9, 30, 0, 0, time.UTC)
	finishedAt := startedAt.Add(125 * time.Millisecond)
	tlsDuration := 35.5
	tlsVersion := "TLS 1.3"
	cipher := "TLS_AES_128_GCM_SHA256"
	certificateNotAfter := startedAt.Add(30 * 24 * time.Hour)

	body := newQueryLatestHTTPResultsBody(appresult.LatestHTTPResultsOutput{Results: []appresult.LatestHTTPResult{{
		ProbeID: "33333333-3333-3333-3333-333333333333",
		CheckID: "44444444-4444-4444-4444-444444444444",
		Result: domainhttp.Result{
			StartedAt:           startedAt,
			FinishedAt:          finishedAt,
			DurationMs:          125,
			Status:              domainhttp.StatusSuccessful,
			TLSDurationMs:       &tlsDuration,
			TLSVersion:          &tlsVersion,
			TLSCipherSuite:      &cipher,
			CertificateNotAfter: &certificateNotAfter,
		},
	}}})

	if len(body.Results) != 1 {
		t.Fatalf("expected one result, got %#v", body.Results)
	}
	got := body.Results[0]
	if got.Result.Status != "successful" || got.Result.TLSDurationMs == nil || *got.Result.TLSDurationMs != tlsDuration || got.Result.TLSVersion == nil || *got.Result.TLSVersion != tlsVersion || got.Result.CertificateNotAfter == nil || !got.Result.CertificateNotAfter.Equal(certificateNotAfter) {
		t.Fatalf("unexpected latest HTTP response body: %#v", got)
	}
}

package publicstatus

import (
	"testing"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainpublic "github.com/yorukot/netstamp/internal/domain/publicstatus"
)

func TestAssignmentMetricsBodyMapsHTTPTotalLatency(t *testing.T) {
	totalMs := 125.5
	failurePercent := 100.0
	metrics := assignmentMetricsBody(domainpublic.Assignment{
		CheckType:      domaincheck.TypeHTTP,
		LatestStatus:   "error",
		LatencyAvgMs:   &totalMs,
		FailurePercent: &failurePercent,
	})

	if metrics == nil || metrics.LatencyAvgMs == nil || *metrics.LatencyAvgMs != totalMs {
		t.Fatalf("expected HTTP total latency, got %#v", metrics)
	}
	if metrics.FailurePercent == nil || *metrics.FailurePercent != failurePercent {
		t.Fatalf("expected HTTP failure percent, got %#v", metrics)
	}
	if metrics.ConnectAvgMs != nil || metrics.LossPercent != nil {
		t.Fatalf("unexpected TCP or ping metrics in HTTP response: %#v", metrics)
	}
}

func TestPublicAssignmentBodyRedactsOperationalDetailsByDefault(t *testing.T) {
	location := "Taipei"
	latitude := 25.033
	longitude := 121.5654
	assignment := domainpublic.Assignment{
		CheckName:         "Public API",
		CheckType:         domaincheck.TypeHTTP,
		CheckTarget:       "https://internal.example.com/health",
		ProbeName:         "edge-tpe-01",
		ProbeLocationName: &location,
		ProbeLatitude:     &latitude,
		ProbeLongitude:    &longitude,
	}

	body := newPublicAssignmentBodies([]domainpublic.Assignment{assignment}, domainpublic.Page{})[0]
	if body.Target != nil || body.ProbeName != nil || body.ProbeLocationName != nil || body.Latitude != nil || body.Longitude != nil {
		t.Fatalf("expected target and probe details to be redacted, got %#v", body)
	}
}

func TestPublicAssignmentBodyExposesOnlyEnabledDetails(t *testing.T) {
	location := "Taipei"
	latitude := 25.033
	longitude := 121.5654
	assignment := domainpublic.Assignment{
		CheckName:         "Public API",
		CheckType:         domaincheck.TypeHTTP,
		CheckTarget:       "https://status.example.com/health",
		ProbeName:         "edge-tpe-01",
		ProbeLocationName: &location,
		ProbeLatitude:     &latitude,
		ProbeLongitude:    &longitude,
	}
	page := domainpublic.Page{
		ShowTargets:        true,
		ShowProbeNames:     false,
		ShowProbeLocations: true,
	}

	body := newPublicAssignmentBodies([]domainpublic.Assignment{assignment}, page)[0]
	if body.Target == nil || *body.Target != assignment.CheckTarget {
		t.Fatalf("expected target to be public, got %#v", body.Target)
	}
	if body.ProbeName != nil {
		t.Fatalf("expected probe name to remain private, got %#v", body.ProbeName)
	}
	if body.ProbeLocationName == nil || *body.ProbeLocationName != location || body.Latitude == nil || body.Longitude == nil {
		t.Fatalf("expected location details to be public, got %#v", body)
	}
}

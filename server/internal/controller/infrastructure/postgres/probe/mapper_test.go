package pgprobe

import (
	"testing"

	"github.com/google/uuid"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
)

func TestMapAssignmentIncludesCheckNameAndTCPConfig(t *testing.T) {
	port := int32(8443)
	timeoutMs := int32(1200)
	ipFamily := sqlc.IpFamilyInet6

	assignment := mapAssignment(sqlc.ListActiveAssignmentsForProbeRow{
		AssignmentID:    uuid.MustParse("0ac05ca4-0df0-445a-ac33-ed46e9595ccc"),
		ProjectID:       uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		ProbeID:         uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		CheckID:         uuid.MustParse("5880599f-5539-4466-848b-d57b9c7e1d4c"),
		ProbeName:       "Pan Tencent Cloud",
		CheckName:       "HTTPS Connect",
		ProbeInternalID: 10,
		CheckInternalID: 20,
		CheckVersion:    "check-v1",
		SelectorVersion: "selector-v1",
		CheckType:       sqlc.CheckTypeTcp,
		Target:          "example.com",
		IntervalSeconds: 30,
		TcpPort:         &port,
		TcpTimeoutMs:    &timeoutMs,
		TcpIpFamily:     &ipFamily,
	})

	if assignment.Check == nil {
		t.Fatal("expected mapped assignment check")
	}
	if assignment.Probe != nil {
		t.Fatalf("expected runtime assignment probe to stay omitted, got %#v", assignment.Probe)
	}
	if assignment.Check.Name != "HTTPS Connect" {
		t.Fatalf("expected check name, got %q", assignment.Check.Name)
	}
	if assignment.Check.Type != domaincheck.TypeTCP {
		t.Fatalf("expected tcp check type, got %q", assignment.Check.Type)
	}
	if assignment.Check.TCPConfig == nil {
		t.Fatal("expected tcp config")
	}
	if assignment.Check.TCPConfig.Port != port || assignment.Check.TCPConfig.TimeoutMs != timeoutMs {
		t.Fatalf("unexpected tcp config: %#v", assignment.Check.TCPConfig)
	}
	if assignment.Check.TCPConfig.IPFamily == nil || *assignment.Check.TCPConfig.IPFamily != domainnetwork.IPFamilyInet6 {
		t.Fatalf("unexpected tcp ip family: %#v", assignment.Check.TCPConfig.IPFamily)
	}
}

func TestMapAssignmentForProbeChecksIncludesProbeName(t *testing.T) {
	assignment := mapAssignmentForProbeChecks(sqlc.ListActiveAssignmentsForProbeChecksRow{
		AssignmentID:    uuid.MustParse("0ac05ca4-0df0-445a-ac33-ed46e9595ccc"),
		ProjectID:       uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		ProbeID:         uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		CheckID:         uuid.MustParse("5880599f-5539-4466-848b-d57b9c7e1d4c"),
		ProbeName:       "Pan Tencent Cloud",
		CheckName:       "HTTPS Connect",
		ProbeInternalID: 10,
		CheckInternalID: 20,
		CheckVersion:    "check-v1",
		SelectorVersion: "selector-v1",
		CheckType:       sqlc.CheckTypeTcp,
		Target:          "example.com",
		IntervalSeconds: 30,
	})

	if assignment.Probe == nil {
		t.Fatal("expected mapped assignment probe")
	}
	if assignment.Probe.Name != "Pan Tencent Cloud" {
		t.Fatalf("expected probe name, got %q", assignment.Probe.Name)
	}
	if assignment.Check == nil || assignment.Check.Name != "HTTPS Connect" {
		t.Fatalf("expected check name, got %#v", assignment.Check)
	}
}

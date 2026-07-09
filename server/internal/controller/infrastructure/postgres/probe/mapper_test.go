package pgprobe

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

func TestMapProbeStatusFieldsCalculatesOnlineUptime(t *testing.T) {
	onlineSince := time.Now().Add(-2 * time.Minute)

	status := mapProbeStatusFields(
		uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		sqlc.ProbeStateOnline,
		nil,
		&onlineSince,
		nil,
		nil,
		nil,
		nil,
		nil,
		time.Now(),
	)

	if status.State != domainprobe.StateOnline {
		t.Fatalf("expected online state, got %q", status.State)
	}
	if status.OnlineSince == nil || !status.OnlineSince.Equal(onlineSince) {
		t.Fatalf("expected online since to be preserved, got %#v", status.OnlineSince)
	}
	if status.UptimeSeconds == nil {
		t.Fatal("expected uptime seconds")
	}
	if *status.UptimeSeconds < 119 || *status.UptimeSeconds > 121 {
		t.Fatalf("expected uptime around 120 seconds, got %d", *status.UptimeSeconds)
	}
}

func TestMapProbeStatusFieldsOmitsOfflineUptime(t *testing.T) {
	onlineSince := time.Now().Add(-2 * time.Minute)

	status := mapProbeStatusFields(
		uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		sqlc.ProbeStateOffline,
		nil,
		&onlineSince,
		nil,
		nil,
		nil,
		nil,
		nil,
		time.Now(),
	)

	if status.State != domainprobe.StateOffline {
		t.Fatalf("expected offline state, got %q", status.State)
	}
	if status.OnlineSince != nil {
		t.Fatalf("expected offline status to omit online since, got %#v", status.OnlineSince)
	}
	if status.UptimeSeconds != nil {
		t.Fatalf("expected offline status to omit uptime, got %#v", status.UptimeSeconds)
	}
}

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

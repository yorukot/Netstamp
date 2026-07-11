package pgalert

import (
	"testing"

	"github.com/google/uuid"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
)

func TestAddIncidentSummariesRedactsHTTPQuery(t *testing.T) {
	var incident domainalert.Incident
	addIncidentSummaries(
		&incident,
		uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		"Probe",
		uuid.MustParse("5880599f-5539-4466-848b-d57b9c7e1d4c"),
		"HTTP check",
		sqlc.CheckTypeHttp,
		"https://example.com/health?token=secret",
	)

	if incident.Check == nil || incident.Check.Target != "https://example.com/health" {
		t.Fatalf("expected query-free HTTP target, got %#v", incident.Check)
	}
}

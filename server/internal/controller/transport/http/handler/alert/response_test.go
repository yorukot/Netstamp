package alert

import (
	"testing"

	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
)

func TestIncidentResponseIncludesResolutionReason(t *testing.T) {
	t.Parallel()

	reason := domainalert.IncidentResolutionReasonTargetNoLongerEvaluated
	response := incidentResponse(domainalert.Incident{ResolutionReason: &reason})
	if response.ResolutionReason == nil || *response.ResolutionReason != reason {
		t.Fatalf("resolution reason = %v, want %q", response.ResolutionReason, reason)
	}
}

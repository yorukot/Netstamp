package pgassignment

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainselector "github.com/yorukot/netstamp/internal/domain/selector"
)

func TestMatchingProbeIDsReturnsOnlyEnabledMatchingProbes(t *testing.T) {
	tokyoProbeID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	osakaProbeID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	disabledProbeID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	probes := []activeProbeLabels{
		{
			probeID: tokyoProbeID,
			enabled: true,
			labels:  []domainlabel.Label{{Key: "region", Value: "tokyo"}},
		},
		{
			probeID: osakaProbeID,
			enabled: true,
			labels:  []domainlabel.Label{{Key: "region", Value: "osaka"}},
		},
		{
			probeID: disabledProbeID,
			enabled: false,
			labels:  []domainlabel.Label{{Key: "region", Value: "tokyo"}},
		},
	}

	selector := mustParseSelector(t, json.RawMessage(`{"label":{"key":"region","op":"eq","value":"tokyo"}}`))
	got := matchingProbeIDs(selector, probes)
	if len(got) != 1 || got[0] != tokyoProbeID {
		t.Fatalf("expected only enabled tokyo probe, got %#v", got)
	}
}

func TestMatchingProbeIDsMatchAllSelector(t *testing.T) {
	firstProbeID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	secondProbeID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	probes := []activeProbeLabels{
		{probeID: firstProbeID, enabled: true, labels: []domainlabel.Label{{Key: "region", Value: "tokyo"}}},
		{probeID: secondProbeID, enabled: true},
	}

	got := matchingProbeIDs(mustParseSelector(t, json.RawMessage(`{}`)), probes)
	if len(got) != 2 || got[0] != firstProbeID || got[1] != secondProbeID {
		t.Fatalf("expected all enabled probes, got %#v", got)
	}
}

func mustParseSelector(t *testing.T, raw json.RawMessage) domainselector.Selector {
	t.Helper()

	selector, err := domainselector.Parse(raw)
	if err != nil {
		t.Fatalf("parse selector: %v", err)
	}

	return selector
}

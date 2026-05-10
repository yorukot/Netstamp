package pgcheck

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainselector "github.com/yorukot/netstamp/internal/domain/selector"
)

func TestMatchingProbeIDs(t *testing.T) {
	tokyoProbeID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	osakaProbeID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	unlabeledProbeID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	probes := []activeProbeLabels{
		{
			probeID: tokyoProbeID,
			labels: []domainlabel.Label{
				{Key: "region", Value: "tokyo"},
				{Key: "env", Value: "prod"},
			},
		},
		{
			probeID: osakaProbeID,
			labels: []domainlabel.Label{
				{Key: "region", Value: "osaka"},
				{Key: "env", Value: "prod"},
			},
		},
		{
			probeID: unlabeledProbeID,
		},
	}

	selector := mustParseSelector(t, json.RawMessage(`{"label":{"key":"region","op":"eq","value":"tokyo"}}`))
	got := matchingProbeIDs(selector, probes)
	if len(got) != 1 || got[0] != tokyoProbeID {
		t.Fatalf("expected only tokyo probe, got %#v", got)
	}
}

func TestMatchingProbeIDsMatchAll(t *testing.T) {
	firstProbeID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	secondProbeID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	probes := []activeProbeLabels{
		{probeID: firstProbeID, labels: []domainlabel.Label{{Key: "region", Value: "tokyo"}}},
		{probeID: secondProbeID},
	}

	got := matchingProbeIDs(mustParseSelector(t, json.RawMessage(`{}`)), probes)
	if len(got) != 2 || got[0] != firstProbeID || got[1] != secondProbeID {
		t.Fatalf("expected all probes, got %#v", got)
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

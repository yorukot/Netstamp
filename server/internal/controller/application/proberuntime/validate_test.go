package proberuntime

import (
	"errors"
	"net/netip"
	"testing"

	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
)

func TestNormalizeRuntimeStatusReturnsAllFieldErrors(t *testing.T) {
	agentVersion := ""
	publicV4 := netip.MustParseAddr("2001:db8::1")
	publicV6 := netip.MustParseAddr("192.0.2.1")
	as := ""

	_, err := normalizeRuntimeStatus(RuntimeStatusInput{
		AgentVersion: &agentVersion,
		PublicV4:     &publicV4,
		PublicV6:     &publicV6,
		AS:           &as,
		Addrs:        []netip.Addr{{}},
	}, testProbeID)

	assertValidationFields(t, err, []string{"agentVersion", "publicV4", "publicV6", "as", "addrs"})
}

func TestNormalizeSubmitResultsReturnsAllFieldErrors(t *testing.T) {
	_, err := normalizeSubmitResults(SubmitResultsInput{
		Results: []RuntimeResultGroupInput{{
			CheckID: "",
			Type:    "ping",
			Ping: []PingResultInput{{
				DurationMs:  -1,
				Status:      "bad",
				SentCount:   -1,
				LossPercent: -1,
			}},
		}},
	})

	assertValidationFields(t, err, []string{
		"results[0].checkId",
		"results[0].ping[0].startedAt",
		"results[0].ping[0].finishedAt",
		"results[0].ping[0].durationMs",
		"results[0].ping[0].status",
		"results[0].ping[0].sentCount",
		"results[0].ping[0].lossPercent",
	})
}

func TestNormalizeIPFamilyCapabilitiesReturnsAllFieldErrors(t *testing.T) {
	_, err := normalizeIPFamilyCapabilities(IPFamilyCapabilitiesInput{
		BodyPresent: true,
		Families:    []string{"bad", "inet", "inet"},
	}, testProbeID)

	assertValidationFields(t, err, []string{
		"families[0]",
		"families[2]",
	})
}

func assertValidationFields(t *testing.T, err error, wantFields []string) {
	t.Helper()

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}

	fieldErrors, ok := appvalidation.FieldErrors(err)
	if !ok {
		t.Fatalf("expected field validation errors, got %v", err)
	}
	if len(fieldErrors) != len(wantFields) {
		t.Fatalf("expected %d field errors, got %d: %#v", len(wantFields), len(fieldErrors), fieldErrors)
	}
	for i, wantField := range wantFields {
		if fieldErrors[i].Field != wantField {
			t.Fatalf("expected field error %d to target %q, got %q", i, wantField, fieldErrors[i].Field)
		}
		if fieldErrors[i].Message == "" {
			t.Fatalf("expected field error %d to include a message", i)
		}
	}
}

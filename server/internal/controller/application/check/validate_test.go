package check

import (
	"errors"
	"testing"

	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
)

func TestNormalizeCreateCheckInputReturnsAllFieldErrors(t *testing.T) {
	description := ""
	packetCount := int32(0)
	packetSizeBytes := int32(0)
	timeoutMs := int32(0)
	ipFamily := "bad"

	_, err := normalizeCreateCheckInput(CreateCheckInput{
		ProjectRef:      "",
		Name:            "",
		Type:            "ping",
		Target:          "",
		Description:     &description,
		IntervalSeconds: 0,
		LabelIDs:        []string{""},
		PingConfig: &PingConfigInput{
			PacketCount:     &packetCount,
			PacketSizeBytes: &packetSizeBytes,
			TimeoutMs:       &timeoutMs,
			IPFamily:        &ipFamily,
		},
		TracerouteConfig: &TracerouteConfigInput{},
	})

	assertValidationFields(t, err, []string{
		"projectRef",
		"name",
		"target",
		"description",
		"intervalSeconds",
		"tracerouteConfig",
		"packetCount",
		"packetSizeBytes",
		"timeoutMs",
		"ipFamily",
		"labelIds",
	})
}

func TestNormalizeUpdateCheckInputReturnsAllFieldErrors(t *testing.T) {
	name := ""
	checkType := "invalid"
	target := ""
	description := ""
	interval := int32(0)
	packetCount := int32(0)
	ipFamily := "bad"
	protocol := "bad"
	labelIDs := []string{""}

	_, err := normalizeUpdateCheckInput(UpdateCheckInput{
		ProjectRef:      "",
		CheckID:         "",
		Name:            &name,
		Type:            &checkType,
		Target:          &target,
		Description:     &description,
		IntervalSeconds: &interval,
		PingConfig: &PingConfigInput{
			PacketCount: &packetCount,
			IPFamily:    &ipFamily,
		},
		TracerouteConfig: &TracerouteConfigInput{
			Protocol: &protocol,
		},
		LabelIDs: &labelIDs,
	})

	assertValidationFields(t, err, []string{
		"projectRef",
		"checkId",
		"name",
		"type",
		"target",
		"description",
		"intervalSeconds",
		"packetCount",
		"ipFamily",
		"tracerouteConfig.protocol",
		"labelIds",
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

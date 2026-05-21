package result

import (
	"errors"
	"testing"
	"time"

	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
)

func TestNormalizeQueryPingSeriesInputReturnsAllFieldErrors(t *testing.T) {
	fromMs := int64(-1)
	toMs := int64(0)
	maxDataPoints := int32(0)

	_, err := normalizeQueryPingSeriesInput(QueryPingSeriesInput{
		ProjectRef:    "",
		ProbeID:       "",
		CheckID:       "",
		FromMs:        &fromMs,
		ToMs:          &toMs,
		Metric:        "bad",
		MaxDataPoints: &maxDataPoints,
		Now:           time.Date(2026, 5, 19, 0, 0, 0, 0, time.UTC),
	})

	assertValidationFields(t, err, []string{"projectRef", "probeId", "checkId", "to", "from", "metric", "maxDataPoints"})
}

func TestNormalizeQueryTracerouteRunsInputReturnsAllFieldErrors(t *testing.T) {
	fromMs := int64(-1)
	toMs := int64(0)
	limit := int32(0)
	cursor := int64(0)

	_, err := normalizeQueryTracerouteRunsInput(QueryTracerouteRunsInput{
		ProjectRef: "",
		ProbeID:    "",
		CheckID:    "",
		FromMs:     &fromMs,
		ToMs:       &toMs,
		Limit:      &limit,
		CursorMs:   &cursor,
		Now:        time.Date(2026, 5, 19, 0, 0, 0, 0, time.UTC),
	})

	assertValidationFields(t, err, []string{"projectRef", "probeId", "checkId", "to", "from", "limit", "cursor"})
}

func TestNormalizeQueryTracerouteTopologyInputReturnsAllFieldErrors(t *testing.T) {
	fromMs := int64(-1)
	toMs := int64(0)
	limit := int32(0)

	_, err := normalizeQueryTracerouteTopologyInput(QueryTracerouteTopologyInput{
		ProjectRef: "",
		ProbeID:    "bad-probe",
		CheckID:    "bad-check",
		FromMs:     &fromMs,
		ToMs:       &toMs,
		Limit:      &limit,
		Now:        time.Date(2026, 5, 19, 0, 0, 0, 0, time.UTC),
	})

	assertValidationFields(t, err, []string{"projectRef", "probeId", "checkId", "to", "from", "limit"})
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

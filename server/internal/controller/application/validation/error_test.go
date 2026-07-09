package validation

import (
	"errors"
	"testing"

	"github.com/yorukot/spvalidator"
)

func TestCollectorAddErrorMapsValidationTagsToCodes(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "required", err: spvalidator.Required(""), want: FieldCodeRequired},
		{name: "string min", err: spvalidator.Min("", 1), want: FieldCodeValueTooShort},
		{name: "string max", err: spvalidator.Max("toolong", 1), want: FieldCodeValueTooLong},
		{name: "numeric min", err: spvalidator.Min(1, 2), want: FieldCodeValueTooSmall},
		{name: "numeric max", err: spvalidator.Max(2, 1), want: FieldCodeValueTooLarge},
		{name: "enum", err: spvalidator.OneOf("owner", "admin"), want: FieldCodeInvalidEnum},
		{name: "format", err: spvalidator.UUID("not-a-uuid"), want: FieldCodeInvalidFormat},
		{name: "fallback", err: errors.New("invalid input"), want: FieldCodeInvalidValue},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var collector Collector
			collector.AddError("field", test.err, "value")

			fieldErrors, ok := FieldErrors(collector.Err(errors.New("base")))
			if !ok || len(fieldErrors) != 1 {
				t.Fatalf("expected one field error, got %#v", fieldErrors)
			}
			if fieldErrors[0].Code != test.want {
				t.Fatalf("expected code %q, got %q", test.want, fieldErrors[0].Code)
			}
		})
	}
}

func TestNewFieldsDefaultsMissingCode(t *testing.T) {
	fieldErrors, ok := FieldErrors(NewFields(errors.New("base"), FieldError{Field: "field", Message: "invalid"}))
	if !ok || len(fieldErrors) != 1 {
		t.Fatalf("expected one field error, got %#v", fieldErrors)
	}
	if fieldErrors[0].Code != FieldCodeInvalidValue {
		t.Fatalf("expected default field code, got %q", fieldErrors[0].Code)
	}
}

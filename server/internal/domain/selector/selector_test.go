package selector

import (
	"encoding/json"
	"errors"
	"testing"

	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
)

func TestParseMatchAllSelector(t *testing.T) {
	tests := []struct {
		name string
		raw  json.RawMessage
	}{
		{name: "nil", raw: nil},
		{name: "empty", raw: json.RawMessage(``)},
		{name: "empty object", raw: json.RawMessage(`{}`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := mustParse(t, tt.raw)

			if !selector.Matches(nil) {
				t.Fatal("expected match-all selector to match empty labels")
			}
			if !selector.Matches(labels(label("region", "tokyo"))) {
				t.Fatal("expected match-all selector to match non-empty labels")
			}
			assertCanonicalJSON(t, selector, `{}`)
		})
	}
}

func TestLabelSelectorsMatch(t *testing.T) {
	probeLabels := labels(
		label("region", "tokyo"),
		label("env", "prod"),
	)

	tests := []struct {
		name      string
		raw       json.RawMessage
		wantMatch bool
	}{
		{
			name:      "eq matches exact key and value",
			raw:       json.RawMessage(`{"label":{"key":"region","op":"eq","value":"tokyo"}}`),
			wantMatch: true,
		},
		{
			name:      "eq rejects different value",
			raw:       json.RawMessage(`{"label":{"key":"region","op":"eq","value":"osaka"}}`),
			wantMatch: false,
		},
		{
			name:      "in matches any listed value",
			raw:       json.RawMessage(`{"label":{"key":"region","op":"in","values":["osaka","tokyo"]}}`),
			wantMatch: true,
		},
		{
			name:      "in rejects values outside set",
			raw:       json.RawMessage(`{"label":{"key":"region","op":"in","values":["osaka","seoul"]}}`),
			wantMatch: false,
		},
		{
			name:      "exists matches key presence",
			raw:       json.RawMessage(`{"label":{"key":"env","op":"exists"}}`),
			wantMatch: true,
		},
		{
			name:      "exists rejects missing key",
			raw:       json.RawMessage(`{"label":{"key":"zone","op":"exists"}}`),
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := mustParse(t, tt.raw)

			if got := selector.Matches(probeLabels); got != tt.wantMatch {
				t.Fatalf("expected match=%t, got %t", tt.wantMatch, got)
			}
		})
	}
}

func TestLogicalSelectorsMatch(t *testing.T) {
	probeLabels := labels(
		label("region", "tokyo"),
		label("env", "prod"),
	)

	tests := []struct {
		name      string
		raw       json.RawMessage
		wantMatch bool
	}{
		{
			name: "all matches every child",
			raw: json.RawMessage(`{"all":[
				{"label":{"key":"region","op":"eq","value":"tokyo"}},
				{"label":{"key":"env","op":"eq","value":"prod"}}
			]}`),
			wantMatch: true,
		},
		{
			name: "all rejects missing child",
			raw: json.RawMessage(`{"all":[
				{"label":{"key":"region","op":"eq","value":"tokyo"}},
				{"label":{"key":"env","op":"eq","value":"staging"}}
			]}`),
			wantMatch: false,
		},
		{
			name: "any matches one child",
			raw: json.RawMessage(`{"any":[
				{"label":{"key":"region","op":"eq","value":"osaka"}},
				{"label":{"key":"env","op":"eq","value":"prod"}}
			]}`),
			wantMatch: true,
		},
		{
			name:      "not negates child",
			raw:       json.RawMessage(`{"not":{"label":{"key":"env","op":"eq","value":"staging"}}}`),
			wantMatch: true,
		},
		{
			name: "nested logic matches",
			raw: json.RawMessage(`{"all":[
				{"any":[
					{"label":{"key":"region","op":"eq","value":"tokyo"}},
					{"label":{"key":"region","op":"eq","value":"osaka"}}
				]},
				{"not":{"label":{"key":"disabled","op":"exists"}}}
			]}`),
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := mustParse(t, tt.raw)

			if got := selector.Matches(probeLabels); got != tt.wantMatch {
				t.Fatalf("expected match=%t, got %t", tt.wantMatch, got)
			}
		})
	}
}

func TestCanonicalJSON(t *testing.T) {
	tests := []struct {
		name string
		raw  json.RawMessage
		want string
	}{
		{
			name: "trims label eq strings",
			raw:  json.RawMessage(`{"label":{"key":" region ","op":" eq ","value":" tokyo "}}`),
			want: `{"label":{"key":"region","op":"eq","value":"tokyo"}}`,
		},
		{
			name: "deduplicates in values",
			raw:  json.RawMessage(`{"label":{"key":"region","op":"in","values":[" tokyo ","osaka","tokyo"]}}`),
			want: `{"label":{"key":"region","op":"in","values":["tokyo","osaka"]}}`,
		},
		{
			name: "emits stable logical shape",
			raw: json.RawMessage(`{"all":[
				{"label":{"op":"eq","value":"tokyo","key":"region"}},
				{"label":{"values":["prod","staging"],"key":"env","op":"in"}}
			]}`),
			want: `{"all":[{"label":{"key":"region","op":"eq","value":"tokyo"}},{"label":{"key":"env","op":"in","values":["prod","staging"]}}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := mustParse(t, tt.raw)
			assertCanonicalJSON(t, selector, tt.want)
		})
	}
}

func TestParseRejectsInvalidSelectors(t *testing.T) {
	tests := []struct {
		name string
		raw  json.RawMessage
	}{
		{name: "arbitrary non ast object", raw: json.RawMessage(`{"label":"edge"}`)},
		{name: "null", raw: json.RawMessage(`null`)},
		{name: "multiple operators", raw: json.RawMessage(`{"all":[{}],"any":[{}]}`)},
		{name: "unknown top level operator", raw: json.RawMessage(`{"none":[{}]}`)},
		{name: "empty all", raw: json.RawMessage(`{"all":[]}`)},
		{name: "empty any", raw: json.RawMessage(`{"any":[]}`)},
		{name: "invalid not", raw: json.RawMessage(`{"not":[]}`)},
		{name: "null not", raw: json.RawMessage(`{"not":null}`)},
		{name: "missing label key", raw: json.RawMessage(`{"label":{"op":"eq","value":"tokyo"}}`)},
		{name: "empty label key", raw: json.RawMessage(`{"label":{"key":" ","op":"eq","value":"tokyo"}}`)},
		{name: "missing eq value", raw: json.RawMessage(`{"label":{"key":"region","op":"eq"}}`)},
		{name: "empty in values", raw: json.RawMessage(`{"label":{"key":"region","op":"in","values":[]}}`)},
		{name: "gt is not supported", raw: json.RawMessage(`{"label":{"key":"latency_ms","op":"gt","value":100}}`)},
		{name: "gte is not supported", raw: json.RawMessage(`{"label":{"key":"latency_ms","op":"gte","value":100}}`)},
		{name: "lt is not supported", raw: json.RawMessage(`{"label":{"key":"latency_ms","op":"lt","value":100}}`)},
		{name: "lte is not supported", raw: json.RawMessage(`{"label":{"key":"latency_ms","op":"lte","value":100}}`)},
		{name: "unknown label op", raw: json.RawMessage(`{"label":{"key":"region","op":"neq","value":"tokyo"}}`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.raw)
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("expected ErrInvalidInput, got %v", err)
			}
		})
	}
}

func mustParse(t *testing.T, raw json.RawMessage) Selector {
	t.Helper()

	selector, err := Parse(raw)
	if err != nil {
		t.Fatalf("parse selector: %v", err)
	}

	return selector
}

func assertCanonicalJSON(t *testing.T, selector Selector, want string) {
	t.Helper()

	got := string(selector.CanonicalJSON())
	if got != want {
		t.Fatalf("unexpected canonical JSON:\n got: %s\nwant: %s", got, want)
	}
}

func labels(values ...domainlabel.Label) []domainlabel.Label {
	return values
}

func label(key, value string) domainlabel.Label {
	return domainlabel.Label{Key: key, Value: value}
}

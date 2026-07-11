package check

import (
	"encoding/json"
	"testing"
)

func TestHTTPConfigInputTracksIPFamilyPresence(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantSet   bool
		wantValue *string
	}{
		{name: "omitted", body: `{}`, wantSet: false},
		{name: "cleared", body: `{"ipFamily":null}`, wantSet: true},
		{name: "ipv4", body: `{"ipFamily":"inet"}`, wantSet: true, wantValue: stringPointer("inet")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var input checkHTTPConfigInput
			if err := json.Unmarshal([]byte(test.body), &input); err != nil {
				t.Fatalf("decode HTTP config: %v", err)
			}
			appInput := input.appInput()
			if appInput == nil || appInput.IPFamilySet != test.wantSet {
				t.Fatalf("unexpected presence state: %#v", appInput)
			}
			if test.wantValue == nil {
				if appInput.IPFamily != nil {
					t.Fatalf("expected nil IP family, got %q", *appInput.IPFamily)
				}
				return
			}
			if appInput.IPFamily == nil || *appInput.IPFamily != *test.wantValue {
				t.Fatalf("expected IP family %q, got %#v", *test.wantValue, appInput.IPFamily)
			}
		})
	}
}

func stringPointer(value string) *string { return &value }

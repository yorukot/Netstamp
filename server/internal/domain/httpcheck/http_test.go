package httpcheck

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestConfigStatusSelectorsJSONRoundTrip(t *testing.T) {
	config := DefaultConfig()
	config.ExpectedStatusClasses = []int32{2, 4}
	config.ExpectedStatusCodes = []int32{200, 202}
	data, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Config
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded.ExpectedStatusClasses, config.ExpectedStatusClasses) || !reflect.DeepEqual(decoded.ExpectedStatusCodes, config.ExpectedStatusCodes) {
		t.Fatalf("unexpected round trip: %s", data)
	}
}

func TestVNExpectedStatusesRequiresSelectionAndCanonicalizes(t *testing.T) {
	if _, _, err := VNExpectedStatuses(nil, nil); err == nil {
		t.Fatal("expected empty selector error")
	}
	codes, classes, err := VNExpectedStatuses([]int32{204, 200, 204}, []int32{3, 2, 2})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(codes, []int32{200, 204}) || !reflect.DeepEqual(classes, []int32{2, 3}) {
		t.Fatalf("unexpected canonical statuses: %v %v", codes, classes)
	}
}

func TestConfigRejectsMalformedStatusSelectors(t *testing.T) {
	for _, input := range []string{
		`{"method":"GET","headers":[],"timeoutMs":1000,"followRedirects":true,"skipTlsVerify":false,"expectedStatuses":[{"kind":"class","class":"2ab"}]}`,
		`{"method":"GET","headers":[],"timeoutMs":1000,"followRedirects":true,"skipTlsVerify":false,"expectedStatuses":[{"kind":"code","code":200,"class":"2xx"}]}`,
	} {
		var config Config
		if err := json.Unmarshal([]byte(input), &config); err == nil {
			t.Fatalf("expected malformed selector to fail: %s", input)
		}
	}
}

func TestRedactTargetRemovesQueryValues(t *testing.T) {
	const target = "https://example.com/health?token=secret&expires=123"
	if got := RedactTarget(target); got != "https://example.com/health" {
		t.Fatalf("expected query-free target, got %q", got)
	}
	if got := RedactTarget("://invalid"); got != "" {
		t.Fatalf("expected invalid target to fail closed, got %q", got)
	}
	if got := RedactTarget("https://user:password@example.com/health?token=secret"); got != "https://example.com/health" {
		t.Fatalf("expected URL credentials to be removed, got %q", got)
	}
}

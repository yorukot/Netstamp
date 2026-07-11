package assignment

import (
	"encoding/json"
	"strings"
	"testing"

	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainhttp "github.com/yorukot/netstamp/internal/domain/httpcheck"
)

func TestProjectAssignmentResponseOmitsHTTPSecrets(t *testing.T) {
	body := "body-secret"
	config := domainhttp.DefaultConfig()
	config.Headers = []domainhttp.Header{{Name: "Authorization", Value: "header-secret"}}
	config.Body = &body
	assignment := domainassignment.Assignment{Check: &domaincheck.Check{
		Type:       domaincheck.TypeHTTP,
		Target:     "https://example.com/health?token=query-secret",
		HTTPConfig: &config,
	}}

	output := newProjectAssignmentBody(assignment)
	if output.Check == nil || output.Check.Target != "https://example.com/health" {
		t.Fatalf("expected query-free HTTP assignment target, got %#v", output.Check)
	}
	encoded, err := json.Marshal(output)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"header-secret", "body-secret", "query-secret", "httpConfig"} {
		if strings.Contains(string(encoded), secret) {
			t.Fatalf("project assignment response contains %q: %s", secret, encoded)
		}
	}
}

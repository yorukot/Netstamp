package check

import (
	"encoding/json"
	"strings"
	"testing"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainhttp "github.com/yorukot/netstamp/internal/domain/httpcheck"
)

func TestNewCheckBodyRedactsSensitiveHTTPConfigForViewer(t *testing.T) {
	body := `{"token":"body-secret"}`
	bodyContains := "response-secret"
	config := domainhttp.DefaultConfig()
	config.Headers = []domainhttp.Header{{Name: "Authorization", Value: "Bearer header-secret"}}
	config.Body = &body
	config.BodyContains = &bodyContains
	check := domaincheck.Check{
		Type:       domaincheck.TypeHTTP,
		Target:     "https://example.com/health?token=query-secret",
		HTTPConfig: &config,
	}

	output := newCheckBody(check, false)
	if output.Target != "https://example.com/health" {
		t.Fatalf("expected query-free target, got %q", output.Target)
	}
	if output.HTTPConfig == nil || !output.HTTPConfig.SensitiveFieldsRedacted {
		t.Fatalf("expected redacted HTTP config: %#v", output.HTTPConfig)
	}
	if len(output.HTTPConfig.Headers) != 0 || output.HTTPConfig.Body != nil || output.HTTPConfig.BodyContains != nil {
		t.Fatalf("expected request and assertion secrets to be omitted: %#v", output.HTTPConfig)
	}

	encoded, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("marshal redacted response: %v", err)
	}
	for _, secret := range []string{"header-secret", "body-secret", "response-secret", "query-secret"} {
		if strings.Contains(string(encoded), secret) {
			t.Fatalf("redacted response contains %q: %s", secret, encoded)
		}
	}
	if config.Headers[0].Value != "Bearer header-secret" || config.Body == nil || *config.Body != body {
		t.Fatalf("response mapping mutated the runtime config: %#v", config)
	}
}

func TestNewCheckBodyIncludesSensitiveHTTPConfigForManager(t *testing.T) {
	body := "request-body"
	config := domainhttp.DefaultConfig()
	config.Headers = []domainhttp.Header{{Name: "Authorization", Value: "Bearer secret"}}
	config.Body = &body
	check := domaincheck.Check{
		Type:       domaincheck.TypeHTTP,
		Target:     "https://example.com/health?token=secret",
		HTTPConfig: &config,
	}

	output := newCheckBody(check, true)
	if output.Target != check.Target || output.HTTPConfig == nil || output.HTTPConfig.SensitiveFieldsRedacted {
		t.Fatalf("expected full manager response: %#v", output)
	}
	if len(output.HTTPConfig.Headers) != 1 || output.HTTPConfig.Headers[0].Value != "Bearer secret" || output.HTTPConfig.Body == nil || *output.HTTPConfig.Body != body {
		t.Fatalf("expected full HTTP config: %#v", output.HTTPConfig)
	}
}

package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type openAPITestSpec struct {
	Info struct {
		Title   string `json:"title"`
		Version string `json:"version"`
	} `json:"info"`
	Servers []struct {
		URL string `json:"url"`
	} `json:"servers"`
	Paths map[string]json.RawMessage `json:"paths"`
}

func TestRunWritesOpenAPIToFile(t *testing.T) {
	output := filepath.Join(t.TempDir(), "nested", "openapi.json")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{
		"-output", output,
		"-version", "v2",
		"-server-url", "https://api.netstamp.dev/",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got %q", stdout.String())
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	assertFormattedOpenAPI(t, data)

	spec := decodeOpenAPITestSpec(t, data)
	if spec.Info.Title != "Netstamp API" {
		t.Fatalf("expected API title, got %q", spec.Info.Title)
	}
	if spec.Info.Version != "v2" {
		t.Fatalf("expected API version v2, got %q", spec.Info.Version)
	}
	assertOpenAPIServerURL(t, spec, "https://api.netstamp.dev/api/v2")
	assertOpenAPIPath(t, spec, "/auth/login")
	assertOpenAPIPath(t, spec, "/projects/{ref}/checks")
}

func TestRunWritesOpenAPIToStdout(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"-output", "-", "-version", "v1"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}

	data := stdout.Bytes()
	assertFormattedOpenAPI(t, data)

	spec := decodeOpenAPITestSpec(t, data)
	assertOpenAPIServerURL(t, spec, "/api/v1")
	assertOpenAPIPath(t, spec, "/auth/register")
}

func TestOpenAPIRequestPath(t *testing.T) {
	got := openAPIRequestPath("v3")
	if got != "/api/v3/openapi.json" {
		t.Fatalf("expected request path /api/v3/openapi.json, got %q", got)
	}
}

func decodeOpenAPITestSpec(t *testing.T, data []byte) openAPITestSpec {
	t.Helper()

	var spec openAPITestSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("decode OpenAPI: %v", err)
	}
	return spec
}

func assertFormattedOpenAPI(t *testing.T, data []byte) {
	t.Helper()

	if !json.Valid(data) {
		t.Fatalf("expected valid JSON, got %q", string(data))
	}
	if !bytes.HasSuffix(data, []byte("\n")) {
		t.Fatal("expected OpenAPI output to end with newline")
	}
	if !bytes.Contains(data, []byte("\n\t\"info\"")) {
		t.Fatal("expected OpenAPI output to be indented with tabs")
	}
}

func assertOpenAPIServerURL(t *testing.T, spec openAPITestSpec, want string) {
	t.Helper()

	if len(spec.Servers) != 1 {
		t.Fatalf("expected one server, got %d", len(spec.Servers))
	}
	if spec.Servers[0].URL != want {
		t.Fatalf("expected server URL %q, got %q", want, spec.Servers[0].URL)
	}
}

func assertOpenAPIPath(t *testing.T, spec openAPITestSpec, path string) {
	t.Helper()

	if _, ok := spec.Paths[path]; !ok {
		t.Fatalf("expected OpenAPI path %q to be registered", path)
	}
}

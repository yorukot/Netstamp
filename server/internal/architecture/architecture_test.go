package architecture_test

import (
	"bytes"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

const modulePath = "github.com/yorukot/netstamp/"

func TestDomainPackagesDoNotImportOuterLayers(t *testing.T) {
	assertNoForbiddenImports(t, "./internal/domain/...", []string{
		"internal/controller/",
		"internal/agent/",
		"internal/platform/",
		"cmd/",
	})
}

func TestApplicationPackagesDoNotImportAdapters(t *testing.T) {
	assertNoForbiddenImports(t, "./internal/controller/application/...", []string{
		"internal/controller/app",
		"internal/controller/config",
		"internal/controller/infrastructure/",
		"internal/controller/logger",
		"internal/controller/transport/",
		"internal/agent/",
		"cmd/",
	})
}

func TestPostgresRepositoriesDoNotImportApplicationOrTransport(t *testing.T) {
	assertNoForbiddenImports(t, "./internal/controller/infrastructure/postgres/...", []string{
		"internal/controller/application/",
		"internal/controller/transport/",
		"internal/controller/app",
	})
}

func assertNoForbiddenImports(t *testing.T, pattern string, forbidden []string) {
	t.Helper()

	lines := goListImports(t, pattern)
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		pkg, imports, ok := strings.Cut(line, "|")
		if !ok {
			t.Fatalf("unexpected go list output line %q", line)
		}
		for _, importPath := range strings.Fields(imports) {
			relative, ok := strings.CutPrefix(importPath, modulePath)
			if !ok {
				continue
			}
			for _, prefix := range forbidden {
				if forbiddenImport(relative, prefix) {
					t.Fatalf("%s imports forbidden package %s", pkg, importPath)
				}
			}
		}
	}
}

func forbiddenImport(relative, forbidden string) bool {
	if strings.HasSuffix(forbidden, "/") {
		return strings.HasPrefix(relative, forbidden)
	}

	return relative == forbidden || strings.HasPrefix(relative, forbidden+"/")
}

func goListImports(t *testing.T, pattern string) []string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file path")
	}
	serverDir := strings.TrimSuffix(filename, "/internal/architecture/architecture_test.go")
	cmd := exec.Command("go", "list", "-f", "{{.ImportPath}}|{{join .Imports \" \"}}", pattern)
	cmd.Dir = serverDir
	cmd.Env = append(cmd.Environ(), "GOCACHE="+t.TempDir())
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list %s: %v\n%s", pattern, err, stderr.String())
	}

	return strings.Split(strings.TrimSpace(string(output)), "\n")
}

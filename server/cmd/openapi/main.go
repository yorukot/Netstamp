package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"

	httpserver "github.com/yorukot/netstamp/internal/controller/transport/http"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flagSet := flag.NewFlagSet("openapi", flag.ContinueOnError)
	flagSet.SetOutput(stderr)

	output := flagSet.String("output", "../docs/public/openapi.json", "output file path, or - for stdout")
	apiVersion := flagSet.String("version", "v1", "API version path segment")
	serverURL := flagSet.String("server-url", "", "absolute backend origin to publish in servers[0].url")
	if err := flagSet.Parse(args); err != nil {
		return 2
	}

	data, err := generateOpenAPI(*apiVersion, *serverURL)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "generate openapi: %v\n", err)
		return 1
	}

	if *output == "-" {
		_, err = stdout.Write(data)
	} else {
		if mkdirErr := os.MkdirAll(filepath.Dir(*output), 0o755); mkdirErr != nil { //nolint:gosec // Generated OpenAPI output is a public documentation artifact.
			_, _ = fmt.Fprintf(stderr, "create output directory: %v\n", mkdirErr)
			return 1
		}
		err = os.WriteFile(*output, data, 0o644) //nolint:gosec // Generated OpenAPI output is a public documentation artifact.
	}
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "write openapi: %v\n", err)
		return 1
	}

	return 0
}

func generateOpenAPI(apiVersion, serverURL string) ([]byte, error) {
	router := httpserver.NewRouter(httpserver.Dependencies{
		APIVersion:     apiVersion,
		BackendBaseURL: serverURL,
		RequestTimeout: 10 * time.Second,
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, openAPIRequestPath(apiVersion), http.NoBody)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		return nil, fmt.Errorf("request %s: status %d", request.URL.Path, recorder.Code)
	}

	var formatted bytes.Buffer
	if err := json.Indent(&formatted, recorder.Body.Bytes(), "", "\t"); err != nil {
		return nil, fmt.Errorf("format openapi json: %w", err)
	}
	formatted.WriteByte('\n')
	return formatted.Bytes(), nil
}

func openAPIRequestPath(apiVersion string) string {
	return "/api/" + apiVersion + "/openapi.json"
}

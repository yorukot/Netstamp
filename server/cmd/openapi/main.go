package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"

	httpserver "github.com/yorukot/netstamp/internal/transport/http"
)

func main() {
	output := flag.String("output", "../docs/public/openapi.json", "output file path, or - for stdout")
	apiVersion := flag.String("version", "v1", "API version path segment")
	serverURL := flag.String("server-url", "", "absolute backend origin to publish in servers[0].url")
	flag.Parse()

	data, err := generateOpenAPI(*apiVersion, *serverURL)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "generate openapi: %v\n", err)
		os.Exit(1)
	}

	if *output == "-" {
		_, err = os.Stdout.Write(data)
	} else {
		if err := os.MkdirAll(filepath.Dir(*output), 0o755); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "create output directory: %v\n", err)
			os.Exit(1)
		}
		err = os.WriteFile(*output, data, 0o644)
	}
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "write openapi: %v\n", err)
		os.Exit(1)
	}
}

func generateOpenAPI(apiVersion, serverURL string) ([]byte, error) {
	router := httpserver.NewRouter(httpserver.Dependencies{
		APIVersion:     apiVersion,
		BackendBaseURL: serverURL,
		RequestTimeout: 10 * time.Second,
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, openAPIRequestPath(apiVersion), nil)
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

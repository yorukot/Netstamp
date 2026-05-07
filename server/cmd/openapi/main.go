package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	appauth "github.com/yorukot/netstamp/internal/application/auth"
	appproject "github.com/yorukot/netstamp/internal/application/project"
	httpserver "github.com/yorukot/netstamp/internal/transport/http"
)

type schemaVerifier struct{}

func (schemaVerifier) VerifyAccessToken(context.Context, string) (appauth.AccessTokenClaims, error) {
	return appauth.AccessTokenClaims{}, nil
}

func main() {
	output := flag.String("output", "../docs/public/openapi.json", "output file path, or - for stdout")
	apiVersion := flag.String("version", "v1", "API version path segment")
	serverURL := flag.String("server-url", "", "absolute backend origin to publish in servers[0].url")
	flag.Parse()

	spec := httpserver.NewOpenAPI(httpserver.Dependencies{
		APIVersion:     *apiVersion,
		BackendBaseURL: *serverURL,
		AuthService:    &appauth.Service{},
		AuthVerifier:   schemaVerifier{},
		ProjectService: &appproject.Service{},
	})

	data, err := json.MarshalIndent(spec, "", "\t")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "marshal openapi: %v\n", err)
		os.Exit(1)
	}
	data = append(data, '\n')

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

package openapi

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed openapi.json
var files embed.FS

func Spec(apiVersion, publicBaseURL string) ([]byte, error) {
	data, err := files.ReadFile("openapi.json")
	if err != nil {
		return nil, fmt.Errorf("read embedded openapi: %w", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("decode embedded openapi: %w", err)
	}
	spec["servers"] = []map[string]string{{"url": serverURL(apiVersion, publicBaseURL)}}

	var formatted bytes.Buffer
	encoder := json.NewEncoder(&formatted)
	encoder.SetIndent("", "\t")
	if err := encoder.Encode(spec); err != nil {
		return nil, fmt.Errorf("encode openapi: %w", err)
	}
	return formatted.Bytes(), nil
}

func serverURL(apiVersion, publicBaseURL string) string {
	basePath := "/api/" + apiVersion
	publicBaseURL = strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
	if publicBaseURL == "" {
		return basePath
	}
	return publicBaseURL + basePath
}

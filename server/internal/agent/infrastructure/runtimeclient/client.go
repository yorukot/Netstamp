package runtimeclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	agentconfig "github.com/yorukot/netstamp/internal/agent/config"
	agentruntime "github.com/yorukot/netstamp/internal/agent/runtime"
)

type RuntimeClient struct {
	baseURL    string
	apiVersion string
	probeID    string
	secret     string
	client     *http.Client
}

func New(config agentconfig.Config) *RuntimeClient {
	return &RuntimeClient{
		baseURL:    strings.TrimRight(config.ControllerURL, "/"),
		apiVersion: strings.Trim(config.APIVersion, "/"),
		probeID:    config.ProbeID,
		secret:     config.ProbeSecret,
		client: &http.Client{
			Timeout: config.HTTPTimeout,
		},
	}
}

func (c *RuntimeClient) Hello(ctx context.Context) (agentruntime.HelloResponse, error) {
	var output agentruntime.HelloResponse
	if err := c.do(ctx, http.MethodPost, "hello", nil, &output); err != nil {
		return agentruntime.HelloResponse{}, err
	}

	return output, nil
}

func (c *RuntimeClient) Heartbeat(ctx context.Context, input agentruntime.HeartbeatInput) (agentruntime.HeartbeatResponse, error) {
	var output agentruntime.HeartbeatResponse
	if err := c.do(ctx, http.MethodPost, "heartbeat", input, &output); err != nil {
		return agentruntime.HeartbeatResponse{}, err
	}

	return output, nil
}

func (c *RuntimeClient) ListAssignments(ctx context.Context) (agentruntime.AssignmentsResponse, error) {
	var output agentruntime.AssignmentsResponse
	if err := c.do(ctx, http.MethodGet, "assignments", nil, &output); err != nil {
		return agentruntime.AssignmentsResponse{}, err
	}

	return output, nil
}

func (c *RuntimeClient) SubmitResults(ctx context.Context, input agentruntime.SubmitResultsInput) (agentruntime.SubmitResultsResponse, error) {
	var output agentruntime.SubmitResultsResponse
	if err := c.do(ctx, http.MethodPost, "results", input, &output); err != nil {
		return agentruntime.SubmitResultsResponse{}, err
	}

	return output, nil
}

func (c *RuntimeClient) do(ctx context.Context, method, operation string, input, output any) error {
	var body io.Reader = http.NoBody
	if input != nil {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(input); err != nil {
			return fmt.Errorf("encode runtime request: %w", err)
		}
		body = &buf
	}

	request, err := http.NewRequestWithContext(ctx, method, c.runtimeURL(operation), body)
	if err != nil {
		return fmt.Errorf("create runtime request: %w", err)
	}
	request.Header.Set("Authorization", "Probe "+c.secret)
	request.Header.Set("Accept", "application/json")
	if input != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := c.client.Do(request)
	if err != nil {
		return fmt.Errorf("runtime request failed: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusBadRequest {
		return runtimeStatusError(response)
	}
	if output == nil || response.Body == http.NoBody {
		return nil
	}
	if err := json.NewDecoder(response.Body).Decode(output); err != nil {
		return fmt.Errorf("decode runtime response: %w", err)
	}

	return nil
}

func (c *RuntimeClient) runtimeURL(operation string) string {
	path, err := url.JoinPath("api", c.apiVersion, "runtime", "probes", c.probeID, operation)
	if err != nil {
		path = "api/" + c.apiVersion + "/runtime/probes/" + c.probeID + "/" + operation
	}

	return c.baseURL + "/" + path
}

type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	if e.Body == "" {
		return "runtime api returned status " + strconv.Itoa(e.StatusCode)
	}
	return "runtime api returned status " + strconv.Itoa(e.StatusCode) + ": " + e.Body
}

func (e *HTTPError) Is(target error) bool {
	switch target {
	case agentruntime.ErrAuthFailed:
		return e.StatusCode == http.StatusUnauthorized || e.StatusCode == http.StatusForbidden
	case agentruntime.ErrPermanentRuntimeAPI:
		return e.StatusCode >= http.StatusBadRequest && e.StatusCode < http.StatusInternalServerError
	default:
		return false
	}
}

func runtimeStatusError(response *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
	message := strings.TrimSpace(string(body))
	err := &HTTPError{
		StatusCode: response.StatusCode,
		Body:       message,
	}
	if errors.Is(err, agentruntime.ErrAuthFailed) {
		return fmt.Errorf("%w: %w", agentruntime.ErrAuthFailed, err)
	}
	if errors.Is(err, agentruntime.ErrPermanentRuntimeAPI) {
		return fmt.Errorf("%w: %w", agentruntime.ErrPermanentRuntimeAPI, err)
	}

	return err
}

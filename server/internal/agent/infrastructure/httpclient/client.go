package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	agentconfig "github.com/yorukot/netstamp/internal/agent/config"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
)

type RuntimeClient struct {
	baseURL    string
	apiVersion string
	probeID    string
	secret     string
	client     *http.Client
	tcp4Client *http.Client
	tcp6Client *http.Client
}

func New(config agentconfig.Config) *RuntimeClient {
	return &RuntimeClient{
		baseURL:    strings.TrimRight(config.ControllerURL, "/"),
		apiVersion: strings.Trim(config.APIVersion, "/"),
		probeID:    config.ProbeID,
		secret:     config.ProbeSecret,
		client: &http.Client{
			Timeout: config.HTTPTimeout.Value,
		},
		tcp4Client: forcedNetworkClient(config.HTTPTimeout.Value, "tcp4"),
		tcp6Client: forcedNetworkClient(config.HTTPTimeout.Value, "tcp6"),
	}
}

// do is a helper function that sends an HTTP request to the runtime API and decodes the response
func (c *RuntimeClient) do(ctx context.Context, method, operation string, input, output any) error {
	return c.doWithClient(ctx, c.client, method, operation, input, output)
}

func (c *RuntimeClient) doWithClient(ctx context.Context, client *http.Client, method, operation string, input, output any) error {
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

	response, err := client.Do(request)
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

func forcedNetworkClient(timeout time.Duration, network string) *http.Client {
	defaultTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		defaultTransport = &http.Transport{}
	}
	transport := defaultTransport.Clone()
	dialer := &net.Dialer{}
	transport.DialContext = func(ctx context.Context, _, address string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, address)
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

// Hello sends a hello request to the runtime API and returns the response
func (c *RuntimeClient) Hello(ctx context.Context) (HelloResponse, error) {
	var output HelloResponse
	if err := c.do(ctx, http.MethodPost, "hello", nil, &output); err != nil {
		return HelloResponse{}, err
	}

	return output, nil
}

// Heartbeat sends a heartbeat request to the runtime API and returns the response
func (c *RuntimeClient) Heartbeat(ctx context.Context, input HeartbeatInput) (HeartbeatResponse, error) {
	var output HeartbeatResponse
	if err := c.do(ctx, http.MethodPost, "heartbeat", input, &output); err != nil {
		return HeartbeatResponse{}, err
	}

	return output, nil
}

func (c *RuntimeClient) ObserveIPFamilyCapability(ctx context.Context, family domainnetwork.IPFamily) (IPFamilyCapabilitiesResponse, error) {
	var output IPFamilyCapabilitiesResponse
	if err := c.doWithClient(ctx, c.familyClient(family), http.MethodPut, "ip-family-capabilities", nil, &output); err != nil {
		return IPFamilyCapabilitiesResponse{}, err
	}

	return output, nil
}

func (c *RuntimeClient) UpdateIPFamilyCapabilities(ctx context.Context, input IPFamilyCapabilitiesInput) (IPFamilyCapabilitiesResponse, error) {
	var output IPFamilyCapabilitiesResponse
	if err := c.do(ctx, http.MethodPut, "ip-family-capabilities", input, &output); err != nil {
		return IPFamilyCapabilitiesResponse{}, err
	}

	return output, nil
}

func (c *RuntimeClient) familyClient(family domainnetwork.IPFamily) *http.Client {
	if family == domainnetwork.IPFamilyInet6 {
		return c.tcp6Client
	}

	return c.tcp4Client
}

// ListAssignments sends a request to the runtime API to list assignments and returns the response
func (c *RuntimeClient) ListAssignments(ctx context.Context) (AssignmentsResponse, error) {
	var output AssignmentsResponse
	if err := c.do(ctx, http.MethodGet, "assignments", nil, &output); err != nil {
		return AssignmentsResponse{}, err
	}

	return output, nil
}

// SubmitResults sends a request to the runtime API to submit results and returns the response
func (c *RuntimeClient) SubmitResults(ctx context.Context, input SubmitResultsInput) (SubmitResultsResponse, error) {
	var output SubmitResultsResponse
	if err := c.do(ctx, http.MethodPost, "results", input, &output); err != nil {
		return SubmitResultsResponse{}, err
	}

	return output, nil
}

// runtimeURL returns the full URL for a given operation on the runtime API
func (c *RuntimeClient) runtimeURL(operation string) string {
	path, err := url.JoinPath("api", c.apiVersion, "runtime", "probes", c.probeID, operation)
	if err != nil {
		path = "api/" + c.apiVersion + "/runtime/probes/" + c.probeID + "/" + operation
	}

	return c.baseURL + "/" + path
}

// HTTPError is an error type that holds the status code and body of an HTTP response
type HTTPError struct {
	StatusCode int
	Body       string
}

// Error returns the error message for an HTTPError
func (e *HTTPError) Error() string {
	if e.Body == "" {
		return "runtime api returned status " + strconv.Itoa(e.StatusCode)
	}
	return "runtime api returned status " + strconv.Itoa(e.StatusCode) + ": " + e.Body
}

// Is returns true if the error is an HTTPError with the same status code as the target error
func (e *HTTPError) Is(target error) bool {
	switch target {
	case ErrAuthFailed:
		return e.StatusCode == http.StatusUnauthorized || e.StatusCode == http.StatusForbidden
	case ErrPermanentRuntimeAPI:
		return e.StatusCode >= http.StatusBadRequest && e.StatusCode < http.StatusInternalServerError
	default:
		return false
	}
}

// runtimeStatusError returns an error based on the status code of an HTTP response
func runtimeStatusError(response *http.Response) error {
	body, err := io.ReadAll(io.LimitReader(response.Body, 4096))
	message := strings.TrimSpace(string(body))
	if err != nil {
		message = "failed to read response body: " + err.Error()
	}
	httpErr := &HTTPError{
		StatusCode: response.StatusCode,
		Body:       message,
	}
	if errors.Is(httpErr, ErrAuthFailed) {
		return fmt.Errorf("%w: %w", ErrAuthFailed, httpErr)
	}
	if errors.Is(httpErr, ErrPermanentRuntimeAPI) {
		return fmt.Errorf("%w: %w", ErrPermanentRuntimeAPI, httpErr)
	}

	return httpErr
}

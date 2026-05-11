package controlplane

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

const apiVersionPath = "api/v1"

type Client struct {
	baseURL    string
	probeID    string
	secret     string
	httpClient *http.Client
}

type StatusInput struct {
	AgentVersion string
	Addrs        []string
}

type HelloOutput struct {
	ServerTime                    time.Time
	HeartbeatIntervalSeconds      int32
	AssignmentPollIntervalSeconds int32
}

type SubmitResultsOutput struct {
	Accepted     bool
	ResyncNeeded bool
	StaleChecks  []string
	Assignments  []domaincheck.Assignment
}

type StatusError struct {
	StatusCode int
	Body       string
}

func (e StatusError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("controller returned HTTP %d", e.StatusCode)
	}
	return fmt.Sprintf("controller returned HTTP %d: %s", e.StatusCode, e.Body)
}

func NewClient(controllerURL, probeID, secret string, timeout time.Duration) (*Client, error) {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	baseURL, err := apiBaseURL(controllerURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		baseURL: baseURL,
		probeID: probeID,
		secret:  secret,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func (c *Client) Hello(ctx context.Context, status StatusInput) (HelloOutput, error) {
	var output helloOutputBody
	err := c.do(ctx, http.MethodPost, "hello", statusInputBody{
		AgentVersion: stringPtr(status.AgentVersion),
		Addrs:        status.Addrs,
	}, &output)
	if err != nil {
		return HelloOutput{}, err
	}

	return HelloOutput(output), nil
}

func (c *Client) Heartbeat(ctx context.Context, status StatusInput) error {
	var output heartbeatOutputBody
	return c.do(ctx, http.MethodPost, "heartbeat", statusInputBody{
		AgentVersion: stringPtr(status.AgentVersion),
		Addrs:        status.Addrs,
	}, &output)
}

func (c *Client) PollAssignments(ctx context.Context) (domainprobe.AssignmentSet, error) {
	var output assignmentsOutputBody
	if err := c.do(ctx, http.MethodGet, "assignments", nil, &output); err != nil {
		return domainprobe.AssignmentSet{}, err
	}

	assignments, err := assignmentResponsesToDomain(output.Assignments)
	if err != nil {
		return domainprobe.AssignmentSet{}, err
	}

	return domainprobe.AssignmentSet{
		ProbeID:     c.probeID,
		GeneratedAt: time.Now().UTC(),
		Assignments: assignments,
	}, nil
}

func (c *Client) SubmitResults(ctx context.Context, batch domainprobe.ResultBatch) (SubmitResultsOutput, error) {
	body := newSubmitResultsInputBody(batch)
	var output submitResultsOutputBody
	if err := c.do(ctx, http.MethodPost, "results", body, &output); err != nil {
		return SubmitResultsOutput{}, err
	}

	assignments, err := assignmentResponsesToDomain(output.Assignments)
	if err != nil {
		return SubmitResultsOutput{}, err
	}

	return SubmitResultsOutput{
		Accepted:     output.Accepted,
		ResyncNeeded: output.ResyncNeeded,
		StaleChecks:  output.StaleChecks,
		Assignments:  assignments,
	}, nil
}

func (c *Client) do(ctx context.Context, method, runtimePath string, input, output any) error {
	var body io.Reader
	if input != nil {
		data, err := json.Marshal(input)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.runtimeURL(runtimePath), body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Probe "+c.secret)
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	res, err := c.httpClient.Do(req) // #nosec G704 -- The controller URL is explicitly operator-configured for this probe.
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return statusError(res)
	}
	if output == nil {
		return nil
	}
	if err := json.NewDecoder(res.Body).Decode(output); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func (c *Client) runtimeURL(runtimePath string) string {
	value, err := url.JoinPath(c.baseURL, "probes", c.probeID, "runtime", runtimePath)
	if err != nil {
		return c.baseURL
	}
	return value
}

func statusError(res *http.Response) error {
	data, err := io.ReadAll(io.LimitReader(res.Body, 4096))
	if err != nil {
		return StatusError{StatusCode: res.StatusCode}
	}
	return StatusError{
		StatusCode: res.StatusCode,
		Body:       strings.TrimSpace(string(data)),
	}
}

func apiBaseURL(controllerURL string) (string, error) {
	trimmed := strings.TrimRight(strings.TrimSpace(controllerURL), "/")
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("invalid controller URL %q", controllerURL)
	}
	if strings.HasSuffix(trimmed, "/"+apiVersionPath) {
		return trimmed, nil
	}
	joined, err := url.JoinPath(trimmed, apiVersionPath)
	if err != nil {
		return "", err
	}

	return joined, nil
}

func stringPtr(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

type statusInputBody struct {
	AgentVersion *string  `json:"agentVersion,omitempty"`
	Addrs        []string `json:"addrs,omitempty"`
}

type helloOutputBody struct {
	ServerTime                    time.Time `json:"serverTime"`
	HeartbeatIntervalSeconds      int32     `json:"heartbeatIntervalSeconds"`
	AssignmentPollIntervalSeconds int32     `json:"assignmentPollIntervalSeconds"`
}

type heartbeatOutputBody struct {
	ServerTime time.Time `json:"serverTime"`
}

type assignmentsOutputBody struct {
	Assignments []assignmentResponse `json:"assignments"`
}

type assignmentResponse struct {
	ID              string             `json:"assignmentId"`
	ProjectID       string             `json:"projectId"`
	ProbeID         string             `json:"probeId"`
	CheckID         string             `json:"checkId"`
	CheckVersion    string             `json:"checkVersion"`
	SelectorVersion string             `json:"selectorVersion"`
	Type            domaincheck.Type   `json:"type"`
	Target          string             `json:"target"`
	IntervalSeconds int32              `json:"intervalSeconds"`
	PingConfig      pingConfigResponse `json:"pingConfig"`
}

type pingConfigResponse struct {
	PacketCount     int32   `json:"packetCount"`
	PacketSizeBytes int32   `json:"packetSizeBytes"`
	TimeoutMs       int32   `json:"timeoutMs"`
	IPFamily        *string `json:"ipFamily,omitempty"`
}

func assignmentResponsesToDomain(responses []assignmentResponse) ([]domaincheck.Assignment, error) {
	assignments := make([]domaincheck.Assignment, 0, len(responses))
	for _, response := range responses {
		assignment, err := response.toDomain()
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment)
	}

	return assignments, nil
}

func (r assignmentResponse) toDomain() (domaincheck.Assignment, error) {
	ipFamily, err := domainnetwork.ParseOptionalIPFamily(r.PingConfig.IPFamily)
	if err != nil {
		return domaincheck.Assignment{}, fmt.Errorf("decode assignment %q IP family: %w", r.ID, err)
	}

	return domaincheck.Assignment{
		ID:              r.ID,
		ProjectID:       r.ProjectID,
		ProbeID:         r.ProbeID,
		CheckID:         r.CheckID,
		CheckVersion:    r.CheckVersion,
		SelectorVersion: r.SelectorVersion,
		Type:            r.Type,
		Target:          r.Target,
		IntervalSeconds: r.IntervalSeconds,
		PingConfig: domainping.Config{
			PacketCount:     r.PingConfig.PacketCount,
			PacketSizeBytes: r.PingConfig.PacketSizeBytes,
			TimeoutMs:       r.PingConfig.TimeoutMs,
			IPFamily:        ipFamily,
		},
	}, nil
}

type submitResultsInputBody map[string]resultGroupInputBody

type resultGroupInputBody struct {
	Type    domaincheck.Type           `json:"type"`
	Detail  resultGroupDetailInputBody `json:"detail"`
	Results []pingResultInputBody      `json:"results"`
}

type resultGroupDetailInputBody struct {
	AssignmentID    string `json:"assignmentId"`
	CheckVersion    string `json:"checkVersion"`
	SelectorVersion string `json:"selectorVersion"`
}

type submitResultsOutputBody struct {
	Accepted     bool                 `json:"accepted"`
	ResyncNeeded bool                 `json:"resyncNeeded"`
	StaleChecks  []string             `json:"staleChecks"`
	Assignments  []assignmentResponse `json:"assignments"`
}

type pingResultInputBody struct {
	StartedAt     time.Time               `json:"startedAt"`
	FinishedAt    time.Time               `json:"finishedAt"`
	DurationMs    int32                   `json:"durationMs"`
	Status        domainping.Status       `json:"status"`
	SentCount     int32                   `json:"sentCount"`
	ReceivedCount int32                   `json:"receivedCount"`
	LossPercent   float64                 `json:"lossPercent"`
	RttMinMs      *float64                `json:"rttMinMs,omitempty"`
	RttAvgMs      *float64                `json:"rttAvgMs,omitempty"`
	RttMedianMs   *float64                `json:"rttMedianMs,omitempty"`
	RttMaxMs      *float64                `json:"rttMaxMs,omitempty"`
	RttStddevMs   *float64                `json:"rttStddevMs,omitempty"`
	RttSamplesMs  []float64               `json:"rttSamplesMs,omitempty"`
	ResolvedIP    *string                 `json:"resolvedIp,omitempty"`
	IPFamily      *domainnetwork.IPFamily `json:"ipFamily,omitempty"`
	Raw           map[string]any          `json:"raw,omitempty"`
	ErrorCode     *string                 `json:"errorCode,omitempty"`
	ErrorMessage  *string                 `json:"errorMessage,omitempty"`
}

func newSubmitResultsInputBody(batch domainprobe.ResultBatch) submitResultsInputBody {
	body := make(submitResultsInputBody)
	for _, result := range batch.Results {
		if result.Type != domaincheck.TypePing {
			continue
		}
		group := body[result.CheckID]
		group.Type = result.Type
		group.Detail = resultGroupDetailInputBody{
			AssignmentID:    result.AssignmentID,
			CheckVersion:    result.CheckVersion,
			SelectorVersion: result.SelectorVersion,
		}
		group.Results = append(group.Results, newPingResultInputBody(result.Ping))
		body[result.CheckID] = group
	}

	return body
}

func newPingResultInputBody(result domainping.Result) pingResultInputBody {
	return pingResultInputBody{
		StartedAt:     result.StartedAt,
		FinishedAt:    result.FinishedAt,
		DurationMs:    result.DurationMs,
		Status:        result.Status,
		SentCount:     result.SentCount,
		ReceivedCount: result.ReceivedCount,
		LossPercent:   result.LossPercent,
		RttMinMs:      result.RttMinMs,
		RttAvgMs:      result.RttAvgMs,
		RttMedianMs:   result.RttMedianMs,
		RttMaxMs:      result.RttMaxMs,
		RttStddevMs:   result.RttStddevMs,
		RttSamplesMs:  result.RttSamplesMs,
		ResolvedIP:    addrString(result.ResolvedIP),
		IPFamily:      result.IPFamily,
		Raw:           result.Raw,
		ErrorCode:     result.ErrorCode,
		ErrorMessage:  result.ErrorMessage,
	}
}

func addrString(addr *netip.Addr) *string {
	if addr == nil {
		return nil
	}

	value := addr.String()
	return &value
}

package executor

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yorukot/netstamp/internal/agent/scheduling"
	agentworker "github.com/yorukot/netstamp/internal/agent/worker"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainhttp "github.com/yorukot/netstamp/internal/domain/httpcheck"
)

const maxHTTPResponseBytes int64 = 4 * 1024 * 1024

type HTTPExecutor struct {
	dial tcpDialFunc
}

type httpTraceState struct {
	mu              sync.Mutex
	dnsDuration     time.Duration
	connectDuration time.Duration
	dnsStartedAt    time.Time
	connectStarted  map[string]time.Time
	tlsStartedAt    time.Time
	tlsDuration     time.Duration
	requestWroteAt  time.Time
	ttfbDuration    time.Duration
	resolved        pingTarget
}

func NewHTTPExecutor() *HTTPExecutor { return &HTTPExecutor{dial: defaultTCPDial} }

func (e *HTTPExecutor) Execute(ctx context.Context, req scheduling.RunRequest) agentworker.ResultEnvelope {
	return agentworker.ResultEnvelope{CheckID: req.Check.ID, Type: domaincheck.TypeHTTP, HTTP: e.execute(ctx, req)}
}

func (e *HTTPExecutor) execute(ctx context.Context, req scheduling.RunRequest) domainhttp.Result {
	startedAt := req.ScheduledAt.UTC()
	if req.Check.HTTPConfig == nil {
		return httpErrorResult(startedAt, time.Now().UTC(), domainhttp.StatusError, "missing_http_config", "http config is missing")
	}
	config := *req.Check.HTTPConfig
	timeout := time.Duration(config.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		return httpErrorResult(startedAt, time.Now().UTC(), domainhttp.StatusError, "invalid_http_config", "http timeout must be positive")
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	state := &httpTraceState{}
	transport := &http.Transport{
		Proxy:             nil,
		DisableKeepAlives: true,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.SkipTLSVerify, //nolint:gosec // This is an explicit per-check option for private/self-signed endpoints.
			MinVersion:         tls.VersionTLS12,
		},
	}
	transport.DialContext = e.dialContext(config, state)
	defer transport.CloseIdleConnections()

	redirects := int32(0)
	client := &http.Client{Transport: transport}
	client.CheckRedirect = func(request *http.Request, via []*http.Request) error {
		if !config.FollowRedirects {
			return http.ErrUseLastResponse
		}
		if len(via) > domainhttp.MaxRedirects {
			redirects = domainhttp.MaxRedirects
			return errors.New("redirect limit exceeded")
		}
		stripCrossOriginSecrets(request, via, config)
		redirects++
		return nil
	}

	request, err := newHTTPRequest(runCtx, req.Check.Target, config, state)
	if err != nil {
		return httpErrorResult(startedAt, time.Now().UTC(), domainhttp.StatusError, "invalid_http_request", err.Error())
	}
	response, err := client.Do(request)
	if err != nil {
		status, code := classifyHTTPError(runCtx, err)
		result := httpErrorResult(startedAt, time.Now().UTC(), status, code, safeHTTPErrorMessage(err))
		applyHTTPTraceResult(&result, state)
		result.RedirectCount = redirects
		return result
	}
	defer response.Body.Close()

	body, readErr := io.ReadAll(io.LimitReader(response.Body, maxHTTPResponseBytes+1))
	finishedAt := time.Now().UTC()
	truncated := int64(len(body)) > maxHTTPResponseBytes
	if truncated {
		body = body[:maxHTTPResponseBytes]
	}
	result := httpResultFromResponse(startedAt, finishedAt, response, redirects, int64(len(body)), truncated)
	applyHTTPTraceResult(&result, state)
	if readErr != nil {
		status, code := classifyHTTPError(runCtx, readErr)
		if code == "http_request_failed" {
			code = "response_read_failed"
		}
		result.Status = status
		result.ErrorCode = optionalString(code)
		result.ErrorMessage = optionalString(readErr.Error())
		return result
	}
	if !domainhttp.MatchesStatus(config, response.StatusCode) {
		result.Status = domainhttp.StatusError
		result.ErrorCode = optionalString("unexpected_status")
		result.ErrorMessage = optionalString(fmt.Sprintf("received HTTP status %d", response.StatusCode))
		return result
	}
	if config.BodyContains != nil {
		matched := bytes.Contains(body, []byte(*config.BodyContains))
		result.BodyMatched = &matched
		if !matched {
			result.Status = domainhttp.StatusError
			if truncated {
				result.ErrorCode = optionalString("response_body_limit_exceeded")
				result.ErrorMessage = optionalString("response body exceeded the matching limit")
			} else {
				result.ErrorCode = optionalString("body_assertion_failed")
				result.ErrorMessage = optionalString("response body did not contain the expected text")
			}
			return result
		}
	}
	result.Status = domainhttp.StatusSuccessful
	return result
}

func (e *HTTPExecutor) dialContext(config domainhttp.Config, state *httpTraceState) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, _, address string) (net.Conn, error) {
		dial := e.dial
		if dial == nil {
			dial = defaultTCPDial
		}
		conn, err := dial(ctx, tcpNetworkForFamily(config.IPFamily), address)
		if resolved := pingTargetFromConn(conn); resolved.addr.IsValid() {
			state.mu.Lock()
			state.resolved = resolved
			state.mu.Unlock()
		}
		return conn, err
	}
}

func stripCrossOriginSecrets(request *http.Request, via []*http.Request, config domainhttp.Config) {
	if len(via) == 0 || sameHTTPOrigin(request.URL, via[len(via)-1].URL) {
		return
	}
	for _, header := range config.Headers {
		if strings.EqualFold(header.Name, "Host") {
			request.Host = ""
			continue
		}
		request.Header.Del(header.Name)
	}
	request.Header.Del("Referer")
}

func sameHTTPOrigin(left, right *url.URL) bool {
	return strings.EqualFold(left.Scheme, right.Scheme) &&
		strings.EqualFold(left.Hostname(), right.Hostname()) &&
		effectiveHTTPPort(left) == effectiveHTTPPort(right)
}

func effectiveHTTPPort(value *url.URL) string {
	if port := value.Port(); port != "" {
		return port
	}
	if strings.EqualFold(value.Scheme, "https") {
		return "443"
	}
	return "80"
}

func newHTTPRequest(ctx context.Context, target string, config domainhttp.Config, state *httpTraceState) (*http.Request, error) {
	var body io.Reader = http.NoBody
	if config.Body != nil {
		body = strings.NewReader(*config.Body)
	}
	request, err := http.NewRequestWithContext(httptrace.WithClientTrace(ctx, newHTTPClientTrace(state)), string(config.Method), target, body)
	if err != nil {
		return nil, err
	}
	for _, header := range config.Headers {
		if strings.EqualFold(header.Name, "Host") {
			request.Host = header.Value
			continue
		}
		request.Header.Add(header.Name, header.Value)
	}
	if request.Header.Get("User-Agent") == "" {
		request.Header.Set("User-Agent", "Netstamp-Probe")
	}
	return request, nil
}

func newHTTPClientTrace(state *httpTraceState) *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		DNSStart: func(httptrace.DNSStartInfo) {
			state.mu.Lock()
			state.dnsStartedAt = time.Now()
			state.mu.Unlock()
		},
		DNSDone: func(httptrace.DNSDoneInfo) {
			state.mu.Lock()
			if !state.dnsStartedAt.IsZero() {
				state.dnsDuration += time.Since(state.dnsStartedAt)
				state.dnsStartedAt = time.Time{}
			}
			state.mu.Unlock()
		},
		ConnectStart: func(network, address string) {
			state.mu.Lock()
			if state.connectStarted == nil {
				state.connectStarted = make(map[string]time.Time)
			}
			state.connectStarted[network+"\x00"+address] = time.Now()
			state.mu.Unlock()
		},
		ConnectDone: func(network, address string, _ error) {
			state.mu.Lock()
			key := network + "\x00" + address
			if startedAt, ok := state.connectStarted[key]; ok {
				state.connectDuration += time.Since(startedAt)
				delete(state.connectStarted, key)
			}
			state.mu.Unlock()
		},
		TLSHandshakeStart: func() { state.mu.Lock(); state.tlsStartedAt = time.Now(); state.mu.Unlock() },
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			state.mu.Lock()
			if !state.tlsStartedAt.IsZero() {
				state.tlsDuration += time.Since(state.tlsStartedAt)
			}
			state.mu.Unlock()
		},
		WroteRequest: func(httptrace.WroteRequestInfo) {
			state.mu.Lock()
			state.requestWroteAt = time.Now()
			state.mu.Unlock()
		},
		GotFirstResponseByte: func() {
			state.mu.Lock()
			if !state.requestWroteAt.IsZero() {
				state.ttfbDuration += time.Since(state.requestWroteAt)
			}
			state.mu.Unlock()
		},
	}
}

func httpResultFromResponse(startedAt, finishedAt time.Time, response *http.Response, redirects int32, responseBytes int64, truncated bool) domainhttp.Result {
	statusCode := int32(response.StatusCode) //nolint:gosec // net/http accepts only three-digit response status codes.
	finalURL := domainhttp.RedactTarget(response.Request.URL.String())
	result := domainhttp.Result{
		StartedAt: startedAt, FinishedAt: finishedAt, DurationMs: durationMillis(startedAt, finishedAt),
		StatusCode: &statusCode, FinalURL: &finalURL, RedirectCount: redirects,
		ResponseBytes: &responseBytes, ResponseTruncated: truncated,
	}
	if response.TLS != nil {
		version := tlsVersionName(response.TLS.Version)
		cipher := tls.CipherSuiteName(response.TLS.CipherSuite)
		result.TLSVersion = &version
		result.TLSCipherSuite = &cipher
		if len(response.TLS.PeerCertificates) > 0 {
			notBefore := response.TLS.PeerCertificates[0].NotBefore.UTC()
			notAfter := response.TLS.PeerCertificates[0].NotAfter.UTC()
			result.CertificateNotBefore = &notBefore
			result.CertificateNotAfter = &notAfter
		}
	}
	return result
}

func applyHTTPTraceResult(result *domainhttp.Result, state *httpTraceState) {
	state.mu.Lock()
	defer state.mu.Unlock()
	result.DNSDurationMs = optionalDurationMs(state.dnsDuration)
	result.ConnectDurationMs = optionalDurationMs(state.connectDuration)
	result.TLSDurationMs = optionalDurationMs(state.tlsDuration)
	result.TTFBDurationMs = optionalDurationMs(state.ttfbDuration)
	if state.resolved.addr.IsValid() {
		addr := state.resolved.addr
		family := state.resolved.ipFamily
		result.ResolvedIP = &addr
		result.IPFamily = &family
	}
}

func optionalDurationMs(value time.Duration) *float64 {
	if value <= 0 {
		return nil
	}
	ms := durationMs(value)
	return &ms
}

func httpErrorResult(startedAt, finishedAt time.Time, status domainhttp.Status, code, message string) domainhttp.Result {
	return domainhttp.Result{StartedAt: startedAt, FinishedAt: finishedAt, DurationMs: durationMillis(startedAt, finishedAt), Status: status, ErrorCode: optionalString(code), ErrorMessage: optionalString(message)}
}

func classifyHTTPError(ctx context.Context, err error) (domainhttp.Status, string) {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || isTimeout(err) {
		return domainhttp.StatusTimeout, "http_timeout"
	}
	if errors.Is(ctx.Err(), context.Canceled) {
		return domainhttp.StatusError, "context_canceled"
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) && strings.Contains(urlErr.Error(), "redirect limit exceeded") {
		return domainhttp.StatusError, "redirect_limit_exceeded"
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return domainhttp.StatusError, "resolve_failed"
	}
	var tlsErr *tls.CertificateVerificationError
	if errors.As(err, &tlsErr) {
		return domainhttp.StatusError, "tls_verification_failed"
	}
	return domainhttp.StatusError, "http_request_failed"
}

func safeHTTPErrorMessage(err error) string {
	var urlErr *url.Error
	if !errors.As(err, &urlErr) {
		return err.Error()
	}
	redacted := *urlErr
	redacted.URL = domainhttp.RedactTarget(urlErr.URL)
	return redacted.Error()
}

func tlsVersionName(version uint16) string {
	switch version {
	case tls.VersionTLS13:
		return "TLS 1.3"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS10:
		return "TLS 1.0"
	default:
		return "TLS " + strconv.Itoa(int(version))
	}
}

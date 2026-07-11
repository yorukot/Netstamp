package executor

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/yorukot/netstamp/internal/agent/scheduling"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainhttp "github.com/yorukot/netstamp/internal/domain/httpcheck"
)

func TestHTTPExecutorRunsRequestAndAssertions(t *testing.T) {
	server := newHTTPTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("X-Netstamp-Test") != "yes" || r.Host != "health.internal" {
			t.Errorf("unexpected request: method=%s header=%q host=%q", r.Method, r.Header.Get("X-Netstamp-Test"), r.Host)
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("service healthy"))
	}))
	defer server.Close()
	body := "request"
	contains := "healthy"
	config := domainhttp.DefaultConfig()
	config.Method = domainhttp.MethodPost
	config.Body = &body
	config.BodyContains = &contains
	config.Headers = []domainhttp.Header{{Name: "X-Netstamp-Test", Value: "yes"}, {Name: "Host", Value: "health.internal"}}
	result := NewHTTPExecutor().Execute(context.Background(), scheduling.RunRequest{Check: domaincheck.Check{ID: "check-1", Type: domaincheck.TypeHTTP, Target: server.URL, HTTPConfig: &config}, ScheduledAt: time.Now().UTC()}).HTTP
	if result.Status != domainhttp.StatusSuccessful || result.StatusCode == nil || *result.StatusCode != http.StatusAccepted {
		t.Fatalf("unexpected result: %#v", result)
	}
	if result.BodyMatched == nil || !*result.BodyMatched || result.ConnectDurationMs == nil || result.TTFBDurationMs == nil {
		t.Fatalf("expected assertion and phase timings: %#v", result)
	}
}

func TestHTTPExecutorSupportsExactStatusSelection(t *testing.T) {
	server := newHTTPTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) }))
	defer server.Close()
	config := domainhttp.DefaultConfig()
	config.ExpectedStatusClasses = nil
	config.ExpectedStatusCodes = []int32{http.StatusTeapot}
	result := NewHTTPExecutor().Execute(context.Background(), scheduling.RunRequest{Check: domaincheck.Check{ID: "check-1", Type: domaincheck.TypeHTTP, Target: server.URL, HTTPConfig: &config}, ScheduledAt: time.Now().UTC()}).HTTP
	if result.Status != domainhttp.StatusSuccessful {
		t.Fatalf("expected exact status match: %#v", result)
	}
	config.ExpectedStatusCodes = []int32{http.StatusOK}
	result = NewHTTPExecutor().Execute(context.Background(), scheduling.RunRequest{Check: domaincheck.Check{ID: "check-1", Type: domaincheck.TypeHTTP, Target: server.URL, HTTPConfig: &config}, ScheduledAt: time.Now().UTC()}).HTTP
	if result.Status != domainhttp.StatusError || result.ErrorCode == nil || *result.ErrorCode != "unexpected_status" {
		t.Fatalf("expected status mismatch: %#v", result)
	}
}

func TestHTTPExecutorRedactsFinalURLQuery(t *testing.T) {
	server := newHTTPTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	defer server.Close()
	config := domainhttp.DefaultConfig()
	result := NewHTTPExecutor().Execute(context.Background(), scheduling.RunRequest{
		Check:       domaincheck.Check{ID: "check-1", Type: domaincheck.TypeHTTP, Target: server.URL + "/health?token=secret", HTTPConfig: &config},
		ScheduledAt: time.Now().UTC(),
	}).HTTP

	if result.FinalURL == nil || *result.FinalURL != server.URL+"/health" {
		t.Fatalf("expected query-free final URL, got %#v", result.FinalURL)
	}
}

func TestSafeHTTPErrorMessageRedactsURLQuery(t *testing.T) {
	message := safeHTTPErrorMessage(&url.Error{Op: "Get", URL: "https://example.com/health?token=secret", Err: errors.New("request failed")})
	if strings.Contains(message, "secret") || !strings.Contains(message, "https://example.com/health") {
		t.Fatalf("expected query-free error message, got %q", message)
	}
}

func TestHTTPExecutorTLSVerificationPolicy(t *testing.T) {
	server := newHTTPTestTLSServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	defer server.Close()
	config := domainhttp.DefaultConfig()
	strict := NewHTTPExecutor().Execute(context.Background(), scheduling.RunRequest{Check: domaincheck.Check{ID: "check-1", Type: domaincheck.TypeHTTP, Target: server.URL, HTTPConfig: &config}, ScheduledAt: time.Now().UTC()}).HTTP
	if strict.Status != domainhttp.StatusError {
		t.Fatalf("expected untrusted certificate failure: %#v", strict)
	}
	config.SkipTLSVerify = true
	skipped := NewHTTPExecutor().Execute(context.Background(), scheduling.RunRequest{Check: domaincheck.Check{ID: "check-1", Type: domaincheck.TypeHTTP, Target: server.URL, HTTPConfig: &config}, ScheduledAt: time.Now().UTC()}).HTTP
	if skipped.Status != domainhttp.StatusSuccessful || skipped.TLSDurationMs == nil || skipped.CertificateNotAfter == nil {
		t.Fatalf("expected successful TLS metadata result: %#v", skipped)
	}
}

func TestHTTPExecutorRedirectPolicy(t *testing.T) {
	final := newHTTPTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	defer final.Close()
	redirect := newHTTPTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, final.URL, http.StatusFound) }))
	defer redirect.Close()
	config := domainhttp.DefaultConfig()
	result := NewHTTPExecutor().Execute(context.Background(), scheduling.RunRequest{Check: domaincheck.Check{ID: "check-1", Type: domaincheck.TypeHTTP, Target: redirect.URL, HTTPConfig: &config}, ScheduledAt: time.Now().UTC()}).HTTP
	if result.Status != domainhttp.StatusSuccessful || result.RedirectCount != 1 || result.FinalURL == nil || *result.FinalURL != final.URL {
		t.Fatalf("unexpected redirect result: %#v", result)
	}
}

func TestHTTPExecutorStripsSecretsOnCrossOriginRedirect(t *testing.T) {
	final := newHTTPTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if value := r.Header.Get("X-Api-Key"); value != "" {
			t.Errorf("cross-origin redirect forwarded configured header %q", value)
		}
		if value := r.Header.Get("Referer"); value != "" {
			t.Errorf("cross-origin redirect forwarded Referer %q", value)
		}
		if r.Host == "secret.internal" {
			t.Errorf("cross-origin redirect forwarded configured Host %q", r.Host)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer final.Close()
	redirect := newHTTPTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, final.URL+"/final", http.StatusFound)
	}))
	defer redirect.Close()

	config := domainhttp.DefaultConfig()
	config.Headers = []domainhttp.Header{
		{Name: "X-Api-Key", Value: "header-secret"},
		{Name: "Host", Value: "secret.internal"},
	}
	result := NewHTTPExecutor().Execute(context.Background(), scheduling.RunRequest{
		Check: domaincheck.Check{
			ID:         "check-1",
			Type:       domaincheck.TypeHTTP,
			Target:     redirect.URL + "/start?token=query-secret",
			HTTPConfig: &config,
		},
		ScheduledAt: time.Now().UTC(),
	}).HTTP

	if result.Status != domainhttp.StatusSuccessful || result.RedirectCount != 1 {
		t.Fatalf("unexpected redirect result: %#v", result)
	}
}

func TestHTTPDialPreservesHostnameAndRecordsConnectedAddress(t *testing.T) {
	var gotNetwork, gotAddress string
	executor := &HTTPExecutor{dial: func(_ context.Context, network, address string) (net.Conn, error) {
		gotNetwork, gotAddress = network, address
		return newRemoteAddrTestConn(netip.MustParseAddr("192.0.2.20"), 443), nil
	}}
	state := &httpTraceState{}
	config := domainhttp.DefaultConfig()
	conn, err := executor.dialContext(config, state)(context.Background(), "tcp", "service.example:443")
	if err != nil {
		t.Fatal(err)
	}
	_ = conn.Close()

	if gotNetwork != "tcp" || gotAddress != "service.example:443" {
		t.Fatalf("expected hostname dial with automatic family fallback, got %q %q", gotNetwork, gotAddress)
	}
	if state.resolved.addr != netip.MustParseAddr("192.0.2.20") {
		t.Fatalf("expected connected remote address, got %#v", state.resolved)
	}
}

func TestHTTPExecutorRedirectLimitProducesValidResult(t *testing.T) {
	server := newHTTPTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count, _ := strconv.Atoi(r.URL.Query().Get("count"))
		http.Redirect(w, r, "/?count="+strconv.Itoa(count+1), http.StatusFound)
	}))
	defer server.Close()

	config := domainhttp.DefaultConfig()
	result := NewHTTPExecutor().Execute(context.Background(), scheduling.RunRequest{
		Check: domaincheck.Check{
			ID:         "check-1",
			Type:       domaincheck.TypeHTTP,
			Target:     server.URL + "/?count=0",
			HTTPConfig: &config,
		},
		ScheduledAt: time.Now().UTC(),
	}).HTTP

	if result.Status != domainhttp.StatusError || result.ErrorCode == nil || *result.ErrorCode != "redirect_limit_exceeded" {
		t.Fatalf("expected redirect limit error: %#v", result)
	}
	if result.RedirectCount != domainhttp.MaxRedirects {
		t.Fatalf("expected redirect count %d, got %d", domainhttp.MaxRedirects, result.RedirectCount)
	}
}

func TestHTTPExecutorClassifiesBodyReadDeadlineAsTimeout(t *testing.T) {
	server := newHTTPTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("partial"))
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		<-r.Context().Done()
	}))
	defer server.Close()

	config := domainhttp.DefaultConfig()
	config.TimeoutMs = 50
	result := NewHTTPExecutor().Execute(context.Background(), scheduling.RunRequest{
		Check:       domaincheck.Check{ID: "check-1", Type: domaincheck.TypeHTTP, Target: server.URL, HTTPConfig: &config},
		ScheduledAt: time.Now().UTC(),
	}).HTTP

	if result.Status != domainhttp.StatusTimeout || result.ErrorCode == nil || *result.ErrorCode != "http_timeout" {
		t.Fatalf("expected response body deadline to be a timeout: %#v", result)
	}
}

func newHTTPTestServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	server := newHTTPUnstartedTestServer(t, handler)
	server.Start()
	return server
}

func newHTTPTestTLSServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	server := newHTTPUnstartedTestServer(t, handler)
	server.StartTLS()
	return server
}

func newHTTPUnstartedTestServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		if errors.Is(err, syscall.EPERM) {
			t.Skipf("local listener unavailable: %v", err)
		}
		t.Fatalf("listen: %v", err)
	}
	return &httptest.Server{Listener: listener, Config: &http.Server{Handler: handler}}
}

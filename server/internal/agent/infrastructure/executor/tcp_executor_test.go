package executor

import (
	"context"
	"errors"
	"net"
	"syscall"
	"testing"
	"time"

	"github.com/yorukot/netstamp/internal/agent/scheduling"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
)

func TestTCPExecutorConnectsToLocalListener(t *testing.T) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		if errors.Is(err, syscall.EPERM) {
			t.Skipf("local tcp listener unavailable in this environment: %v", err)
		}
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	accepted := make(chan struct{})
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr == nil {
			_ = conn.Close()
		}
		close(accepted)
	}()

	addr := listener.Addr().(*net.TCPAddr)
	port := int32(addr.Port)
	inet := domainnetwork.IPFamilyInet
	config := domaintcp.Config{
		Port:      port,
		TimeoutMs: 1000,
		IPFamily:  &inet,
	}
	startedAt := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	executor := NewTCPExecutor()

	got := executor.Execute(context.Background(), scheduling.RunRequest{
		Check: domaincheck.Check{
			ID:        "check-1",
			Type:      domaincheck.TypeTCP,
			Target:    "127.0.0.1",
			TCPConfig: &config,
		},
		ScheduledAt: startedAt,
	})

	if got.CheckID != "check-1" || got.Type != domaincheck.TypeTCP {
		t.Fatalf("unexpected envelope: %#v", got)
	}
	if got.TCP.Status != domaintcp.StatusSuccessful {
		t.Fatalf("expected successful status, got %#v", got.TCP)
	}
	if got.TCP.ConnectDurationMs == nil || *got.TCP.ConnectDurationMs < 0 {
		t.Fatalf("expected connect duration, got %#v", got.TCP.ConnectDurationMs)
	}
	if got.TCP.ResolvedIP == nil || got.TCP.ResolvedIP.String() != "127.0.0.1" {
		t.Fatalf("expected resolved localhost, got %#v", got.TCP.ResolvedIP)
	}
	if got.TCP.IPFamily == nil || *got.TCP.IPFamily != domainnetwork.IPFamilyInet {
		t.Fatalf("expected inet family, got %#v", got.TCP.IPFamily)
	}
	if got.TCP.ErrorCode != nil || got.TCP.ErrorMessage != nil {
		t.Fatalf("expected no error fields, got code=%#v message=%#v", got.TCP.ErrorCode, got.TCP.ErrorMessage)
	}

	select {
	case <-accepted:
	case <-time.After(time.Second):
		t.Fatal("listener did not accept tcp connection")
	}
}

func TestTCPExecutorMissingConfigReturnsErrorResult(t *testing.T) {
	startedAt := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	executor := NewTCPExecutor()

	got := executor.Execute(context.Background(), scheduling.RunRequest{
		Check: domaincheck.Check{
			ID:   "check-1",
			Type: domaincheck.TypeTCP,
		},
		ScheduledAt: startedAt,
	})

	if got.Type != domaincheck.TypeTCP {
		t.Fatalf("expected tcp type, got %q", got.Type)
	}
	if got.TCP.Status != domaintcp.StatusError {
		t.Fatalf("expected error status, got %q", got.TCP.Status)
	}
	if got.TCP.ErrorCode == nil || *got.TCP.ErrorCode != "missing_tcp_config" {
		t.Fatalf("expected missing_tcp_config error code, got %#v", got.TCP.ErrorCode)
	}
}

func TestTCPNetworkHonorsIPFamily(t *testing.T) {
	if got := tcpNetwork(domainnetwork.IPFamilyInet); got != "tcp4" {
		t.Fatalf("expected tcp4, got %q", got)
	}
	if got := tcpNetwork(domainnetwork.IPFamilyInet6); got != "tcp6" {
		t.Fatalf("expected tcp6, got %q", got)
	}
}

func TestTCPExecutorInvalidPortReturnsErrorResult(t *testing.T) {
	startedAt := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	config := domaintcp.DefaultConfig()
	config.Port = 0
	executor := NewTCPExecutor()

	got := executor.Execute(context.Background(), scheduling.RunRequest{
		Check: domaincheck.Check{
			ID:        "check-1",
			Type:      domaincheck.TypeTCP,
			Target:    "127.0.0.1",
			TCPConfig: &config,
		},
		ScheduledAt: startedAt,
	}).TCP

	if got.Status != domaintcp.StatusError {
		t.Fatalf("expected error status, got %q", got.Status)
	}
	if got.ErrorCode == nil || *got.ErrorCode != "invalid_tcp_config" {
		t.Fatalf("expected invalid_tcp_config, got %#v", got.ErrorCode)
	}
}

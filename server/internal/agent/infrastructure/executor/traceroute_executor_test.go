package executor

import (
	"context"
	"errors"
	"math"
	"net/netip"
	"testing"
	"time"

	gotraceroute "github.com/yorukot/go-traceroute"

	"github.com/yorukot/netstamp/internal/agent/scheduling"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

func TestTracerouteExecutorMapsOptionsAndSuccessfulTrace(t *testing.T) {
	startedAt := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	finishedAt := startedAt.Add(4500 * time.Millisecond)
	inet6 := domainnetwork.IPFamilyInet6
	config := domaintraceroute.DefaultConfig()
	config.Protocol = domaintraceroute.ProtocolUDP
	config.MaxHops = 12
	config.TimeoutMs = 2000
	config.QueriesPerHop = 2
	config.PacketSizeBytes = 64
	config.Port = 33435
	config.IPFamily = &inet6
	var gotOptions gotraceroute.Options
	executor := &TracerouteExecutor{
		newRunner: func(options gotraceroute.Options) (tracerouteRunner, error) {
			gotOptions = options
			return fakeTracerouteRunner{trace: &gotraceroute.Trace{
				Destination: netip.MustParseAddr("2001:db8::10"),
				IPVersion:   gotraceroute.IPv6,
				FinishedAt:  finishedAt,
				Hops: []gotraceroute.Hop{{
					TTL: 1,
					Probes: []gotraceroute.Probe{{
						Attempt:    1,
						Addr:       netip.MustParseAddr("2001:db8::1"),
						Hostname:   "router.local",
						RTT:        20 * time.Millisecond,
						Status:     gotraceroute.StatusOK,
						ReceivedAt: startedAt.Add(20 * time.Millisecond),
					}, {
						Attempt:    2,
						Addr:       netip.MustParseAddr("2001:db8::1"),
						RTT:        40 * time.Millisecond,
						Status:     gotraceroute.StatusOK,
						ReceivedAt: startedAt.Add(40 * time.Millisecond),
					}},
				}, {
					TTL: 2,
					Probes: []gotraceroute.Probe{{
						Attempt:    1,
						Addr:       netip.MustParseAddr("2001:db8::10"),
						RTT:        60 * time.Millisecond,
						Status:     gotraceroute.StatusDestination,
						ReceivedAt: startedAt.Add(60 * time.Millisecond),
					}, {
						Attempt: 2,
						Status:  gotraceroute.StatusTimeout,
					}},
				}},
			}}, nil
		},
	}

	got := executor.Execute(context.Background(), scheduling.RunRequest{
		Check: domaincheck.Check{
			ID:               "check-1",
			Type:             domaincheck.TypeTraceroute,
			Target:           "example.com",
			TracerouteConfig: &config,
		},
		ScheduledAt: startedAt,
	})

	if got.CheckID != "check-1" || got.Type != domaincheck.TypeTraceroute {
		t.Fatalf("unexpected envelope: %#v", got)
	}
	if gotOptions.Protocol != gotraceroute.ProtocolUDP ||
		gotOptions.IPVersion != gotraceroute.IPv6 ||
		gotOptions.MaxHops != 12 ||
		gotOptions.QueriesPerHop != 2 ||
		gotOptions.Timeout != 2*time.Second ||
		gotOptions.PacketSize != 64 ||
		gotOptions.UDPBasePort != 33435 ||
		!gotOptions.ResolveNames {
		t.Fatalf("unexpected options: %#v", gotOptions)
	}
	result := got.Traceroute
	if result.Status != domaintraceroute.StatusSuccessful || !result.DestinationReached {
		t.Fatalf("unexpected result status: %#v", result)
	}
	if result.ResolvedIP == nil || *result.ResolvedIP != netip.MustParseAddr("2001:db8::10") {
		t.Fatalf("unexpected resolved ip: %#v", result.ResolvedIP)
	}
	if result.IPFamily == nil || *result.IPFamily != domainnetwork.IPFamilyInet6 {
		t.Fatalf("unexpected ip family: %#v", result.IPFamily)
	}
	if result.HopCount != 2 || len(result.Hops) != 2 {
		t.Fatalf("unexpected hops: %#v", result.Hops)
	}
	firstHop := result.Hops[0]
	if firstHop.SentCount != 2 || firstHop.ReceivedCount != 2 || firstHop.LossPercent != 0 {
		t.Fatalf("unexpected first hop counts: %#v", firstHop)
	}
	if firstHop.Hostname == nil || *firstHop.Hostname != "router.local" {
		t.Fatalf("unexpected first hop hostname: %#v", firstHop.Hostname)
	}
	assertFloatPtr(t, "traceroute min", firstHop.RttMinMs, 20)
	assertFloatPtr(t, "traceroute avg", firstHop.RttAvgMs, 30)
	assertFloatPtr(t, "traceroute median", firstHop.RttMedianMs, 30)
	assertFloatPtr(t, "traceroute max", firstHop.RttMaxMs, 40)
	assertFloatPtr(t, "traceroute stddev", firstHop.RttStddevMs, 10)
	if len(firstHop.RttSamplesMs) != 2 || firstHop.RttSamplesMs[0] != 20 || firstHop.RttSamplesMs[1] != 40 {
		t.Fatalf("unexpected rtt samples: %#v", firstHop.RttSamplesMs)
	}
	secondHop := result.Hops[1]
	if secondHop.SentCount != 2 || secondHop.ReceivedCount != 1 || secondHop.LossPercent != 50 {
		t.Fatalf("unexpected second hop counts: %#v", secondHop)
	}
	if math.Abs(float64(result.DurationMs)-4500) > 1000 {
		t.Fatalf("expected duration around scheduled-to-finished window, got %d", result.DurationMs)
	}
}

func TestTracerouteExecutorMapsTimeoutAndPermissionErrors(t *testing.T) {
	startedAt := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	config := domaintraceroute.DefaultConfig()

	tests := []struct {
		name       string
		err        error
		wantStatus domaintraceroute.Status
		wantCode   string
	}{
		{name: "timeout", err: gotraceroute.ErrTimeout, wantStatus: domaintraceroute.StatusTimeout, wantCode: "traceroute_timeout"},
		{name: "permission", err: gotraceroute.ErrPermission, wantStatus: domaintraceroute.StatusError, wantCode: "permission_denied"},
		{name: "no address", err: gotraceroute.ErrNoAddress, wantStatus: domaintraceroute.StatusError, wantCode: "resolve_failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &TracerouteExecutor{
				newRunner: func(gotraceroute.Options) (tracerouteRunner, error) {
					return fakeTracerouteRunner{err: tt.err}, nil
				},
			}
			got := executor.Execute(context.Background(), scheduling.RunRequest{
				Check: domaincheck.Check{
					ID:               "check-1",
					Type:             domaincheck.TypeTraceroute,
					Target:           "example.com",
					TracerouteConfig: &config,
				},
				ScheduledAt: startedAt,
			}).Traceroute

			if got.Status != tt.wantStatus {
				t.Fatalf("expected status %q, got %q", tt.wantStatus, got.Status)
			}
			if got.ErrorCode == nil || *got.ErrorCode != tt.wantCode {
				t.Fatalf("expected error code %q, got %#v", tt.wantCode, got.ErrorCode)
			}
		})
	}
}

func TestTracerouteExecutorMapsPartialTraceOnError(t *testing.T) {
	startedAt := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	config := domaintraceroute.DefaultConfig()
	executor := &TracerouteExecutor{
		newRunner: func(gotraceroute.Options) (tracerouteRunner, error) {
			return fakeTracerouteRunner{
				trace: &gotraceroute.Trace{Hops: []gotraceroute.Hop{{
					TTL: 1,
					Probes: []gotraceroute.Probe{{
						Addr:   netip.MustParseAddr("192.0.2.1"),
						RTT:    3 * time.Millisecond,
						Status: gotraceroute.StatusOK,
					}},
				}}},
				err: gotraceroute.ErrTimeout,
			}, nil
		},
	}

	got := executor.Execute(context.Background(), scheduling.RunRequest{
		Check: domaincheck.Check{
			ID:               "check-1",
			Type:             domaincheck.TypeTraceroute,
			Target:           "example.com",
			TracerouteConfig: &config,
		},
		ScheduledAt: startedAt,
	}).Traceroute

	if got.Status != domaintraceroute.StatusPartial {
		t.Fatalf("expected partial status, got %q", got.Status)
	}
	if got.ErrorCode == nil || *got.ErrorCode != "traceroute_timeout" {
		t.Fatalf("expected timeout error code, got %#v", got.ErrorCode)
	}
	if len(got.Hops) != 1 || got.Hops[0].ReceivedCount != 1 {
		t.Fatalf("expected partial hop response, got %#v", got.Hops)
	}
}

func TestTracerouteExecutorRejectsUnsupportedProtocol(t *testing.T) {
	startedAt := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	config := domaintraceroute.DefaultConfig()
	config.Protocol = domaintraceroute.Protocol("tcp")
	executor := &TracerouteExecutor{}

	got := executor.Execute(context.Background(), scheduling.RunRequest{
		Check: domaincheck.Check{
			ID:               "check-1",
			Type:             domaincheck.TypeTraceroute,
			Target:           "example.com",
			TracerouteConfig: &config,
		},
		ScheduledAt: startedAt,
	}).Traceroute

	if got.Status != domaintraceroute.StatusError {
		t.Fatalf("expected error status, got %q", got.Status)
	}
	if got.ErrorCode == nil || *got.ErrorCode != "invalid_traceroute_config" {
		t.Fatalf("expected invalid config error code, got %#v", got.ErrorCode)
	}
}

type fakeTracerouteRunner struct {
	trace *gotraceroute.Trace
	err   error
}

func (r fakeTracerouteRunner) Trace(context.Context, string) (*gotraceroute.Trace, error) {
	return r.trace, r.err
}

var errTracerouteFactory = errors.New("factory failed")

func TestTracerouteExecutorMapsFactoryError(t *testing.T) {
	startedAt := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	config := domaintraceroute.DefaultConfig()
	executor := &TracerouteExecutor{
		newRunner: func(gotraceroute.Options) (tracerouteRunner, error) {
			return nil, errTracerouteFactory
		},
	}

	got := executor.Execute(context.Background(), scheduling.RunRequest{
		Check: domaincheck.Check{
			ID:               "check-1",
			Type:             domaincheck.TypeTraceroute,
			Target:           "example.com",
			TracerouteConfig: &config,
		},
		ScheduledAt: startedAt,
	}).Traceroute

	if got.Status != domaintraceroute.StatusError {
		t.Fatalf("expected error status, got %q", got.Status)
	}
	if got.ErrorCode == nil || *got.ErrorCode != "traceroute_setup_failed" {
		t.Fatalf("expected setup error code, got %#v", got.ErrorCode)
	}
}

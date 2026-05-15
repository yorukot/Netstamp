package executor

import (
	"context"
	"math"
	"net"
	"net/netip"
	"testing"
	"time"

	"golang.org/x/net/icmp"

	"github.com/yorukot/netstamp/internal/agent/scheduling"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
)

func TestExecuteMissingConfigReturnsErrorResult(t *testing.T) {
	startedAt := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	executor := NewICMPPingExecutor()

	got := executor.Execute(context.Background(), scheduling.RunRequest{
		Check: domaincheck.Check{
			ID:   "check-1",
			Type: domaincheck.TypePing,
		},
		ScheduledAt: startedAt,
	})

	if got.CheckID != "check-1" {
		t.Fatalf("expected check id check-1, got %q", got.CheckID)
	}
	if got.Type != domaincheck.TypePing {
		t.Fatalf("expected ping type, got %q", got.Type)
	}
	if got.Ping.Status != domainping.StatusError {
		t.Fatalf("expected error status, got %q", got.Ping.Status)
	}
	if got.Ping.ErrorCode == nil || *got.Ping.ErrorCode != "missing_ping_config" {
		t.Fatalf("expected missing_ping_config error code, got %#v", got.Ping.ErrorCode)
	}
	if got.Ping.SentCount != 0 || got.Ping.ReceivedCount != 0 || got.Ping.LossPercent != 100 {
		t.Fatalf("unexpected packet counts: sent=%d received=%d loss=%f", got.Ping.SentCount, got.Ping.ReceivedCount, got.Ping.LossPercent)
	}
}

func TestPingResultFromStatsMapsSuccessfulStats(t *testing.T) {
	startedAt := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	finishedAt := startedAt.Add(1500 * time.Millisecond)
	resolvedIP := netip.MustParseAddr("2001:db8::1")
	ipFamily := domainnetwork.IPFamilyInet6

	got := pingResultFromStats(startedAt, finishedAt, rawPingStats{
		sentCount:     4,
		receivedCount: 3,
		resolvedIP:    &resolvedIP,
		ipFamily:      &ipFamily,
		rtts: []time.Duration{
			30 * time.Millisecond,
			10 * time.Millisecond,
			20 * time.Millisecond,
			40 * time.Millisecond,
		},
	}, domainping.StatusSuccessful, "", "")

	if got.Status != domainping.StatusSuccessful {
		t.Fatalf("expected successful status, got %q", got.Status)
	}
	if got.DurationMs != 1500 {
		t.Fatalf("expected duration 1500, got %d", got.DurationMs)
	}
	if got.SentCount != 4 || got.ReceivedCount != 3 || got.LossPercent != 25 {
		t.Fatalf("unexpected packet counts: sent=%d received=%d loss=%f", got.SentCount, got.ReceivedCount, got.LossPercent)
	}
	if got.ResolvedIP == nil || got.ResolvedIP.String() != "2001:db8::1" {
		t.Fatalf("expected resolved IPv6 address, got %#v", got.ResolvedIP)
	}
	if got.IPFamily == nil || *got.IPFamily != domainnetwork.IPFamilyInet6 {
		t.Fatalf("expected inet6 family, got %#v", got.IPFamily)
	}
	assertFloatPtr(t, "min", got.RttMinMs, 10)
	assertFloatPtr(t, "avg", got.RttAvgMs, 25)
	assertFloatPtr(t, "median", got.RttMedianMs, 25)
	assertFloatPtr(t, "max", got.RttMaxMs, 40)
	assertFloatPtr(t, "stddev", got.RttStddevMs, math.Sqrt(125))
	if len(got.RttSamplesMs) != 4 || got.RttSamplesMs[0] != 30 || got.RttSamplesMs[3] != 40 {
		t.Fatalf("unexpected rtt samples: %#v", got.RttSamplesMs)
	}
	if got.ErrorCode != nil || got.ErrorMessage != nil {
		t.Fatalf("expected nil error fields, got code=%#v message=%#v", got.ErrorCode, got.ErrorMessage)
	}
	if got.Raw["executor"] != "raw-icmp" {
		t.Fatalf("expected raw executor raw-icmp, got %#v", got.Raw)
	}
}

func TestResolverNetworkHonorsIPFamily(t *testing.T) {
	inet := domainnetwork.IPFamilyInet
	inet6 := domainnetwork.IPFamilyInet6

	tests := []struct {
		name string
		in   *domainnetwork.IPFamily
		want string
	}{
		{name: "unset", in: nil, want: "ip"},
		{name: "inet", in: &inet, want: "ip4"},
		{name: "inet6", in: &inet6, want: "ip6"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolverNetwork(tt.in); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestMakePingPayloadUsesRequestedSize(t *testing.T) {
	if got := makePingPayload(1); len(got) != 1 {
		t.Fatalf("expected payload size 1, got %d", len(got))
	}
	if got := makePingPayload(56); len(got) != 56 {
		t.Fatalf("expected payload size 56, got %d", len(got))
	}
}

func TestParseICMPEchoReplyMatchesIDSequenceAndSource(t *testing.T) {
	target := pingTarget{
		addr:     netip.MustParseAddr("192.0.2.1"),
		ipFamily: domainnetwork.IPFamilyInet,
	}
	protocol := pingProtocolForTarget(target)
	packet := echoReplyPacket(t, protocol, 1234, 7)

	sequence, ok := parseICMPEchoReply(protocol, packet, &net.IPAddr{IP: net.ParseIP("192.0.2.1")}, 1234, target.addr)
	if !ok {
		t.Fatal("expected echo reply to match")
	}
	if sequence != 7 {
		t.Fatalf("expected sequence 7, got %d", sequence)
	}

	if _, ok := parseICMPEchoReply(protocol, packet, &net.IPAddr{IP: net.ParseIP("192.0.2.1")}, 4321, target.addr); ok {
		t.Fatal("expected wrong identifier to be ignored")
	}
	if _, ok := parseICMPEchoReply(protocol, packet, &net.IPAddr{IP: net.ParseIP("192.0.2.2")}, 1234, target.addr); ok {
		t.Fatal("expected wrong source to be ignored")
	}
}

func TestDurationMillis(t *testing.T) {
	startedAt := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)

	if got := durationMillis(startedAt, startedAt.Add(2500*time.Millisecond)); got != 2500 {
		t.Fatalf("expected 2500, got %d", got)
	}
	if got := durationMillis(startedAt, startedAt.Add(-time.Second)); got != 0 {
		t.Fatalf("expected negative duration to clamp to 0, got %d", got)
	}
	if got := durationMillis(startedAt, startedAt.Add(time.Duration(math.MaxInt64))); got != math.MaxInt32 {
		t.Fatalf("expected max int32 clamp, got %d", got)
	}
}

func echoReplyPacket(t *testing.T, protocol pingProtocol, identifier, sequence int) []byte {
	t.Helper()

	packet, err := (&icmp.Message{
		Type: protocol.replyType,
		Code: 0,
		Body: &icmp.Echo{
			ID:   identifier,
			Seq:  sequence,
			Data: []byte{1, 2, 3},
		},
	}).Marshal(nil)
	if err != nil {
		t.Fatalf("marshal echo reply: %v", err)
	}

	return packet
}

func assertFloatPtr(t *testing.T, name string, got *float64, want float64) {
	t.Helper()

	if got == nil {
		t.Fatalf("expected %s=%f, got nil", name, want)
	}
	if math.Abs(*got-want) > 0.000001 {
		t.Fatalf("expected %s=%f, got %f", name, want, *got)
	}
}

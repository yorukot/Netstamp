package ping

import (
	"net"
	"syscall"
	"testing"

	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
)

func TestSampleStats(t *testing.T) {
	stats := sampleStats([]float64{30, 10, 20, 40})
	assertFloatPtr(t, stats.min, 10)
	assertFloatPtr(t, stats.avg, 25)
	assertFloatPtr(t, stats.median, 25)
	assertFloatPtr(t, stats.max, 40)
	assertFloatPtr(t, stats.stddev, 11.180339887498949)
}

func TestPingStatusAndLoss(t *testing.T) {
	if got := status(4, 4); got != domainping.StatusSuccessful {
		t.Fatalf("expected successful, got %q", got)
	}
	if got := status(4, 0); got != domainping.StatusTimeout {
		t.Fatalf("expected timeout, got %q", got)
	}
	if got := lossPercent(4, 3); got != 25 {
		t.Fatalf("expected 25%% loss, got %v", got)
	}
}

func TestNewTargetHonorsIPFamily(t *testing.T) {
	inet := domainnetwork.IPFamilyInet
	inet6 := domainnetwork.IPFamilyInet6

	if target, ok := newTarget(net.ParseIP("127.0.0.1"), &inet); !ok || target.family != domainnetwork.IPFamilyInet {
		t.Fatalf("expected IPv4 target, got %#v %v", target, ok)
	}
	if _, ok := newTarget(net.ParseIP("127.0.0.1"), &inet6); ok {
		t.Fatal("expected IPv4 target to be rejected for inet6")
	}
	if target, ok := newTarget(net.ParseIP("::1"), &inet6); !ok || target.family != domainnetwork.IPFamilyInet6 {
		t.Fatalf("expected IPv6 target, got %#v %v", target, ok)
	}
}

func TestListenPermissionErrorMessage(t *testing.T) {
	if got := listenErrorCode(syscall.EPERM); got != "raw_icmp_permission_denied" {
		t.Fatalf("unexpected error code: %q", got)
	}
	if got := listenErrorMessage(syscall.EPERM); got != "raw ICMP requires root or CAP_NET_RAW" {
		t.Fatalf("unexpected error message: %q", got)
	}
}

func assertFloatPtr(t *testing.T, got *float64, want float64) {
	t.Helper()
	if got == nil {
		t.Fatalf("expected %v, got nil", want)
	}
	const epsilon = 0.000000001
	if *got < want-epsilon || *got > want+epsilon {
		t.Fatalf("expected %v, got %v", want, *got)
	}
}

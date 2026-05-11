package ping

import (
	"context"
	"errors"
	"math"
	"net"
	"net/netip"
	"os"
	"sort"
	"syscall"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

const (
	defaultPacketCount     int32 = 4
	defaultPacketSizeBytes int32 = 56
	defaultTimeout               = 3 * time.Second
	maxDurationMs          int64 = math.MaxInt32
)

type Executor struct{}

type target struct {
	addr     net.IP
	network  string
	listen   string
	protocol int
	family   domainnetwork.IPFamily
}

func New() *Executor {
	return &Executor{}
}

func (e *Executor) Execute(ctx context.Context, assignment domaincheck.Assignment) (domainprobe.Result, error) {
	startedAt := time.Now().UTC()
	endpoint, err := resolveTarget(ctx, assignment.Target, assignment.PingConfig.IPFamily)
	if err != nil {
		return pingErrorResult(assignment, startedAt, "resolve_failed", err.Error()), nil
	}

	conn, err := icmp.ListenPacket(endpoint.network, endpoint.listen)
	if err != nil {
		return pingErrorResult(assignment, startedAt, listenErrorCode(err), listenErrorMessage(err)), nil
	}
	defer conn.Close()

	packetCount := assignment.PingConfig.PacketCount
	if packetCount <= 0 {
		packetCount = defaultPacketCount
	}
	timeout := time.Duration(assignment.PingConfig.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	payloadSize := assignment.PingConfig.PacketSizeBytes
	if payloadSize < 0 {
		payloadSize = defaultPacketSizeBytes
	}

	measurements := make([]float64, 0, packetCount)
	var sent int32
	var received int32
	var sendErr error
	identifier := os.Getpid() & 0xffff
	for seq := range packetCount {
		if err := ctx.Err(); err != nil {
			return domainprobe.Result{}, err
		}
		rtt, sentPacket, err := sendEcho(ctx, conn, endpoint, identifier, int(seq), int(payloadSize), timeout)
		if sentPacket {
			sent++
		} else if err != nil && sendErr == nil {
			sendErr = err
		}
		if err == nil {
			measurements = append(measurements, milliseconds(rtt))
			received++
		}
	}
	if sent == 0 && sendErr != nil {
		return pingErrorResult(assignment, startedAt, "send_failed", sendErr.Error()), nil
	}

	finishedAt := time.Now().UTC()
	return domainprobe.Result{
		AssignmentID:    assignment.ID,
		CheckID:         assignment.CheckID,
		CheckVersion:    assignment.CheckVersion,
		SelectorVersion: assignment.SelectorVersion,
		Type:            domaincheck.TypePing,
		Ping:            newPingResult(startedAt, finishedAt, sent, received, measurements, endpoint),
	}, nil
}

func sendEcho(ctx context.Context, conn *icmp.PacketConn, target target, identifier, seq, payloadSize int, timeout time.Duration) (time.Duration, bool, error) {
	message := icmp.Message{
		Type: echoRequestType(target.family),
		Code: 0,
		Body: &icmp.Echo{
			ID:   identifier,
			Seq:  seq,
			Data: payload(payloadSize),
		},
	}
	data, err := message.Marshal(nil)
	if err != nil {
		return 0, false, err
	}

	packetStartedAt := time.Now()
	if _, err := conn.WriteTo(data, &net.IPAddr{IP: target.addr}); err != nil {
		return 0, false, err
	}
	if err := conn.SetReadDeadline(deadline(ctx, timeout)); err != nil {
		return 0, true, err
	}

	buffer := make([]byte, 1500)
	for {
		n, _, err := conn.ReadFrom(buffer)
		if err != nil {
			return 0, true, err
		}
		reply, err := icmp.ParseMessage(target.protocol, buffer[:n])
		if err != nil {
			continue
		}
		if reply.Type != echoReplyType(target.family) {
			continue
		}
		echo, ok := reply.Body.(*icmp.Echo)
		if !ok || echo.ID != identifier || echo.Seq != seq {
			continue
		}

		return time.Since(packetStartedAt), true, nil
	}
}

func resolveTarget(ctx context.Context, name string, family *domainnetwork.IPFamily) (target, error) {
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, name)
	if err != nil {
		return target{}, err
	}
	for _, addr := range addrs {
		if selected, ok := newTarget(addr.IP, family); ok {
			return selected, nil
		}
	}

	return target{}, errors.New("no address matched requested IP family")
}

func newTarget(ip net.IP, family *domainnetwork.IPFamily) (target, bool) {
	if ip4 := ip.To4(); ip4 != nil {
		if family != nil && *family != domainnetwork.IPFamilyInet {
			return target{}, false
		}
		return target{
			addr:     ip4,
			network:  "ip4:icmp",
			listen:   "0.0.0.0",
			protocol: 1,
			family:   domainnetwork.IPFamilyInet,
		}, true
	}
	if ip16 := ip.To16(); ip16 != nil {
		if family != nil && *family != domainnetwork.IPFamilyInet6 {
			return target{}, false
		}
		return target{
			addr:     ip16,
			network:  "ip6:ipv6-icmp",
			listen:   "::",
			protocol: 58,
			family:   domainnetwork.IPFamilyInet6,
		}, true
	}

	return target{}, false
}

func newPingResult(startedAt, finishedAt time.Time, sent, received int32, samples []float64, target target) domainping.Result {
	stats := sampleStats(samples)
	resolvedIP := netip.MustParseAddr(target.addr.String())
	ipFamily := target.family
	return domainping.Result{
		StartedAt:     startedAt,
		FinishedAt:    finishedAt,
		DurationMs:    durationMs(finishedAt.Sub(startedAt)),
		Status:        status(sent, received),
		SentCount:     sent,
		ReceivedCount: received,
		LossPercent:   lossPercent(sent, received),
		RttMinMs:      stats.min,
		RttAvgMs:      stats.avg,
		RttMedianMs:   stats.median,
		RttMaxMs:      stats.max,
		RttStddevMs:   stats.stddev,
		RttSamplesMs:  samples,
		ResolvedIP:    &resolvedIP,
		IPFamily:      &ipFamily,
		Raw: map[string]any{
			"implementation": "raw_icmp",
		},
	}
}

func pingErrorResult(assignment domaincheck.Assignment, startedAt time.Time, code, message string) domainprobe.Result {
	finishedAt := time.Now().UTC()
	return domainprobe.Result{
		AssignmentID:    assignment.ID,
		CheckID:         assignment.CheckID,
		CheckVersion:    assignment.CheckVersion,
		SelectorVersion: assignment.SelectorVersion,
		Type:            domaincheck.TypePing,
		Ping: domainping.Result{
			StartedAt:     startedAt,
			FinishedAt:    finishedAt,
			DurationMs:    durationMs(finishedAt.Sub(startedAt)),
			Status:        domainping.StatusError,
			SentCount:     0,
			ReceivedCount: 0,
			LossPercent:   100,
			Raw: map[string]any{
				"implementation": "raw_icmp",
			},
			ErrorCode:    &code,
			ErrorMessage: &message,
		},
	}
}

type stats struct {
	min    *float64
	avg    *float64
	median *float64
	max    *float64
	stddev *float64
}

func sampleStats(samples []float64) stats {
	if len(samples) == 0 {
		return stats{}
	}
	sorted := append([]float64(nil), samples...)
	sort.Float64s(sorted)

	minRTT := sorted[0]
	maxRTT := sorted[len(sorted)-1]
	avg := average(sorted)
	median := percentile50(sorted)
	stddev := standardDeviation(sorted, avg)

	return stats{
		min:    &minRTT,
		avg:    &avg,
		median: &median,
		max:    &maxRTT,
		stddev: &stddev,
	}
}

func average(samples []float64) float64 {
	var sum float64
	for _, sample := range samples {
		sum += sample
	}
	return sum / float64(len(samples))
}

func percentile50(samples []float64) float64 {
	middle := len(samples) / 2
	if len(samples)%2 == 1 {
		return samples[middle]
	}

	return (samples[middle-1] + samples[middle]) / 2
}

func standardDeviation(samples []float64, avg float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, sample := range samples {
		delta := sample - avg
		sum += delta * delta
	}

	return math.Sqrt(sum / float64(len(samples)))
}

func lossPercent(sent, received int32) float64 {
	if sent <= 0 {
		return 100
	}

	return (float64(sent-received) / float64(sent)) * 100
}

func status(sent, received int32) domainping.Status {
	if sent > 0 && received == 0 {
		return domainping.StatusTimeout
	}

	return domainping.StatusSuccessful
}

func durationMs(duration time.Duration) int32 {
	value := duration.Milliseconds()
	if value > maxDurationMs {
		return int32(maxDurationMs)
	}
	if value < 0 {
		return 0
	}

	return int32(value)
}

func milliseconds(duration time.Duration) float64 {
	return float64(duration.Nanoseconds()) / float64(time.Millisecond)
}

func payload(size int) []byte {
	if size <= 0 {
		return nil
	}
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}

	return data
}

func echoRequestType(family domainnetwork.IPFamily) icmp.Type {
	if family == domainnetwork.IPFamilyInet6 {
		return ipv6.ICMPTypeEchoRequest
	}

	return ipv4.ICMPTypeEcho
}

func echoReplyType(family domainnetwork.IPFamily) icmp.Type {
	if family == domainnetwork.IPFamilyInet6 {
		return ipv6.ICMPTypeEchoReply
	}

	return ipv4.ICMPTypeEchoReply
}

func deadline(ctx context.Context, timeout time.Duration) time.Time {
	deadline := time.Now().Add(timeout)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		return ctxDeadline
	}

	return deadline
}

func listenErrorCode(err error) string {
	if errors.Is(err, syscall.EPERM) || errors.Is(err, syscall.EACCES) {
		return "raw_icmp_permission_denied"
	}

	return "raw_icmp_unavailable"
}

func listenErrorMessage(err error) string {
	if errors.Is(err, syscall.EPERM) || errors.Is(err, syscall.EACCES) {
		return "raw ICMP requires root or CAP_NET_RAW"
	}

	return err.Error()
}

package executor

import (
	"context"
	"errors"
	"math"
	"net"
	"net/netip"
	"sort"
	"time"

	probing "github.com/prometheus-community/pro-bing"

	"github.com/yorukot/netstamp/internal/agent/scheduling"
	agentworker "github.com/yorukot/netstamp/internal/agent/worker"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
)

type ProBingExecutor struct{}

func NewProBingExecutor() *ProBingExecutor {
	return &ProBingExecutor{}
}

func (e *ProBingExecutor) Execute(ctx context.Context, req scheduling.RunRequest) agentworker.ResultEnvelope {
	result := e.execute(ctx, req)
	return agentworker.ResultEnvelope{
		CheckID: req.CheckID,
		Type:    domaincheck.TypePing,
		Ping:    result,
	}
}

func (e *ProBingExecutor) execute(ctx context.Context, req scheduling.RunRequest) domainping.Result {
	startedAt := req.ScheduledAt.UTC()
	finishedAt := time.Now().UTC()
	if req.PingConfig == nil {
		return errorPingResult(startedAt, finishedAt, "missing_ping_config", "ping config is missing")
	}

	pinger := probing.New(req.Target)
	pinger.Count = int(req.PingConfig.PacketCount)
	pinger.Size = int(req.PingConfig.PacketSizeBytes)
	pinger.Timeout = time.Duration(req.PingConfig.TimeoutMs) * time.Millisecond
	pinger.ResolveTimeout = pinger.Timeout
	pinger.RecordRtts = true
	pinger.SetNetwork(pingNetwork(req.PingConfig.IPFamily))

	if err := pinger.Resolve(); err != nil {
		finishedAt = time.Now().UTC()
		return errorPingResult(startedAt, finishedAt, "resolve_failed", err.Error())
	}

	err := pinger.RunWithContext(ctx)
	finishedAt = time.Now().UTC()
	stats := pinger.Statistics()
	if err != nil && errors.Is(ctx.Err(), context.Canceled) {
		return pingResultFromStats(startedAt, finishedAt, stats, domainping.StatusError, "context_canceled", err.Error())
	}
	if err != nil && errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return pingResultFromStats(startedAt, finishedAt, stats, domainping.StatusTimeout, "context_deadline_exceeded", err.Error())
	}
	if err != nil {
		return pingResultFromStats(startedAt, finishedAt, stats, domainping.StatusError, "ping_failed", err.Error())
	}
	if stats.PacketsRecv == 0 {
		return pingResultFromStats(startedAt, finishedAt, stats, domainping.StatusTimeout, "icmp_timeout", "request timed out")
	}

	return pingResultFromStats(startedAt, finishedAt, stats, domainping.StatusSuccessful, "", "")
}

func pingNetwork(ipFamily *domainnetwork.IPFamily) string {
	if ipFamily == nil {
		return "ip"
	}
	switch *ipFamily {
	case domainnetwork.IPFamilyInet:
		return "ip4"
	case domainnetwork.IPFamilyInet6:
		return "ip6"
	default:
		return "ip"
	}
}

func pingResultFromStats(startedAt, finishedAt time.Time, stats *probing.Statistics, status domainping.Status, errorCode, errorMessage string) domainping.Result {
	sentCount := int32(0)
	receivedCount := int32(0)
	lossPercent := float64(100)
	var resolvedIP *netip.Addr
	var ipFamily *domainnetwork.IPFamily
	var samples []float64
	var minRTT, avgRTT, medianRTT, maxRTT, stddevRTT *float64

	if stats != nil {
		sentCount = int32(max(0, stats.PacketsSent))
		receivedCount = int32(max(0, stats.PacketsRecv))
		lossPercent = clamp(stats.PacketLoss, 0, 100)
		resolvedIP, ipFamily = netIPAddrToDomain(stats.IPAddr)
		samples = durationSamplesMs(stats.Rtts)
		if len(samples) > 0 {
			minRTT = durationMsPtr(stats.MinRtt)
			avgRTT = durationMsPtr(stats.AvgRtt)
			medianRTT = floatPtr(median(samples))
			maxRTT = durationMsPtr(stats.MaxRtt)
			stddevRTT = durationMsPtr(stats.StdDevRtt)
		}
	}

	return domainping.Result{
		StartedAt:     startedAt.UTC(),
		FinishedAt:    finishedAt.UTC(),
		DurationMs:    durationMillis(startedAt, finishedAt),
		Status:        status,
		SentCount:     sentCount,
		ReceivedCount: receivedCount,
		LossPercent:   lossPercent,
		RttMinMs:      minRTT,
		RttAvgMs:      avgRTT,
		RttMedianMs:   medianRTT,
		RttMaxMs:      maxRTT,
		RttStddevMs:   stddevRTT,
		RttSamplesMs:  samples,
		ResolvedIP:    resolvedIP,
		IPFamily:      ipFamily,
		Raw: map[string]any{
			"executor": "pro-bing",
		},
		ErrorCode:    optionalString(errorCode),
		ErrorMessage: optionalString(errorMessage),
	}
}

func errorPingResult(startedAt, finishedAt time.Time, errorCode, errorMessage string) domainping.Result {
	return pingResultFromStats(startedAt, finishedAt, nil, domainping.StatusError, errorCode, errorMessage)
}

func netIPAddrToDomain(value *net.IPAddr) (*netip.Addr, *domainnetwork.IPFamily) {
	if value == nil || value.IP == nil {
		return nil, nil //nolint:nilnil // Nil means the executor did not resolve an IP.
	}
	addr, ok := netip.AddrFromSlice(value.IP)
	if !ok {
		return nil, nil //nolint:nilnil // Nil means the executor did not resolve a valid IP.
	}
	addr = addr.Unmap()

	var family domainnetwork.IPFamily
	if addr.Is4() {
		family = domainnetwork.IPFamilyInet
	} else {
		family = domainnetwork.IPFamilyInet6
	}

	return &addr, &family
}

func durationSamplesMs(values []time.Duration) []float64 {
	samples := make([]float64, 0, len(values))
	for _, value := range values {
		samples = append(samples, durationMs(value))
	}

	return samples
}

func median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	copied := append([]float64(nil), values...)
	sort.Float64s(copied)
	middle := len(copied) / 2
	if len(copied)%2 == 1 {
		return copied[middle]
	}

	return (copied[middle-1] + copied[middle]) / 2
}

func durationMsPtr(value time.Duration) *float64 {
	return floatPtr(durationMs(value))
}

func durationMs(value time.Duration) float64 {
	return float64(value) / float64(time.Millisecond)
}

func durationMillis(startedAt, finishedAt time.Time) int32 {
	duration := finishedAt.Sub(startedAt)
	if duration < 0 {
		return 0
	}
	millis := duration.Milliseconds()
	if millis > math.MaxInt32 {
		return math.MaxInt32
	}

	return int32(millis)
}

func floatPtr(value float64) *float64 {
	copied := value
	return &copied
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func clamp(value, minValue, maxValue float64) float64 {
	return math.Min(maxValue, math.Max(minValue, value))
}

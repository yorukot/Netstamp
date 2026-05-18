package executor

import (
	"context"
	"errors"
	"net/netip"
	"time"

	gotraceroute "github.com/yorukot/go-traceroute"

	"github.com/yorukot/netstamp/internal/agent/scheduling"
	agentworker "github.com/yorukot/netstamp/internal/agent/worker"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

type TracerouteExecutor struct {
	newRunner tracerouteRunnerFactory
}

type tracerouteRunner interface {
	Trace(ctx context.Context, target string) (*gotraceroute.Trace, error)
}

type tracerouteRunnerFactory func(gotraceroute.Options) (tracerouteRunner, error)

func NewTracerouteExecutor() *TracerouteExecutor {
	return &TracerouteExecutor{newRunner: newGoTracerouteRunner}
}

func newGoTracerouteRunner(options gotraceroute.Options) (tracerouteRunner, error) {
	return gotraceroute.New(options)
}

func (e *TracerouteExecutor) Execute(ctx context.Context, req scheduling.RunRequest) agentworker.ResultEnvelope {
	result := e.execute(ctx, req)
	return agentworker.ResultEnvelope{
		CheckID:    req.Check.ID,
		Type:       domaincheck.TypeTraceroute,
		Traceroute: result,
	}
}

func (e *TracerouteExecutor) execute(ctx context.Context, req scheduling.RunRequest) domaintraceroute.Result {
	startedAt := req.ScheduledAt.UTC()
	finishedAt := time.Now().UTC()
	if req.Check.TracerouteConfig == nil {
		return errorTracerouteResult(startedAt, finishedAt, "missing_traceroute_config", "traceroute config is missing")
	}

	options, err := tracerouteOptions(*req.Check.TracerouteConfig)
	if err != nil {
		return errorTracerouteResult(startedAt, finishedAt, "invalid_traceroute_config", err.Error())
	}
	newRunner := e.newRunner
	if newRunner == nil {
		newRunner = newGoTracerouteRunner
	}
	runner, err := newRunner(options)
	if err != nil {
		finishedAt = time.Now().UTC()
		return errorTracerouteResult(startedAt, finishedAt, "traceroute_setup_failed", err.Error())
	}

	trace, err := runner.Trace(ctx, req.Check.Target)
	finishedAt = traceFinishedAt(trace, time.Now().UTC())
	if err != nil {
		return tracerouteResultFromError(startedAt, finishedAt, trace, err)
	}

	return tracerouteResultFromTrace(startedAt, finishedAt, trace, "", "")
}

func tracerouteOptions(config domaintraceroute.Config) (gotraceroute.Options, error) {
	protocol, err := goTracerouteProtocol(config.Protocol)
	if err != nil {
		return gotraceroute.Options{}, err
	}

	return gotraceroute.Options{
		Protocol:      protocol,
		IPVersion:     goTracerouteIPVersion(config.IPFamily),
		MaxHops:       int(config.MaxHops),
		QueriesPerHop: int(config.QueriesPerHop),
		Timeout:       time.Duration(config.TimeoutMs) * time.Millisecond,
		PacketSize:    int(config.PacketSizeBytes),
		UDPBasePort:   int(config.Port),
		ResolveNames:  false,
	}, nil
}

func goTracerouteProtocol(protocol domaintraceroute.Protocol) (gotraceroute.Protocol, error) {
	switch protocol {
	case domaintraceroute.ProtocolICMP:
		return gotraceroute.ProtocolICMP, nil
	case domaintraceroute.ProtocolUDP:
		return gotraceroute.ProtocolUDP, nil
	default:
		return 0, errors.New("unsupported traceroute protocol")
	}
}

func goTracerouteIPVersion(ipFamily *domainnetwork.IPFamily) gotraceroute.IPVersion {
	if ipFamily == nil {
		return gotraceroute.IPAny
	}
	switch *ipFamily {
	case domainnetwork.IPFamilyInet:
		return gotraceroute.IPv4
	case domainnetwork.IPFamilyInet6:
		return gotraceroute.IPv6
	default:
		return gotraceroute.IPAny
	}
}

func tracerouteResultFromError(startedAt, finishedAt time.Time, trace *gotraceroute.Trace, err error) domaintraceroute.Result {
	if hasTracerouteResponses(trace) {
		return tracerouteResultFromTrace(startedAt, finishedAt, trace, tracerouteErrorCode(err), err.Error())
	}
	status := domaintraceroute.StatusError
	if errors.Is(err, gotraceroute.ErrTimeout) || errors.Is(err, context.DeadlineExceeded) {
		status = domaintraceroute.StatusTimeout
	}

	return tracerouteResultFromTraceWithStatus(startedAt, finishedAt, trace, status, tracerouteErrorCode(err), err.Error())
}

func tracerouteErrorCode(err error) string {
	switch {
	case errors.Is(err, gotraceroute.ErrPermission):
		return "permission_denied"
	case errors.Is(err, gotraceroute.ErrNoAddress):
		return "resolve_failed"
	case errors.Is(err, gotraceroute.ErrTimeout):
		return "traceroute_timeout"
	case errors.Is(err, context.DeadlineExceeded):
		return "context_deadline_exceeded"
	case errors.Is(err, context.Canceled):
		return "context_canceled"
	default:
		return "traceroute_failed"
	}
}

func tracerouteResultFromTrace(startedAt, finishedAt time.Time, trace *gotraceroute.Trace, errorCode, errorMessage string) domaintraceroute.Result {
	return tracerouteResultFromTraceWithStatus(startedAt, finishedAt, trace, tracerouteStatus(trace), errorCode, errorMessage)
}

func tracerouteResultFromTraceWithStatus(startedAt, finishedAt time.Time, trace *gotraceroute.Trace, status domaintraceroute.Status, errorCode, errorMessage string) domaintraceroute.Result {
	hops := tracerouteHopResults(traceHops(trace))
	resolvedIP := traceDestination(trace)
	resolvedIPPtr := &resolvedIP
	if !resolvedIP.IsValid() {
		resolvedIPPtr = nil
	}
	ipFamily := tracerouteIPFamily(trace)

	return domaintraceroute.Result{
		StartedAt:          startedAt.UTC(),
		FinishedAt:         finishedAt.UTC(),
		DurationMs:         durationMillis(startedAt, finishedAt),
		Status:             status,
		ResolvedIP:         resolvedIPPtr,
		IPFamily:           ipFamily,
		DestinationReached: traceDestinationReached(trace),
		HopCount:           int32(len(hops)), //nolint:gosec // hop count is bounded by domain config.
		ErrorCode:          optionalString(errorCode),
		ErrorMessage:       optionalString(errorMessage),
		Hops:               hops,
	}
}

func tracerouteStatus(trace *gotraceroute.Trace) domaintraceroute.Status {
	if traceDestinationReached(trace) {
		return domaintraceroute.StatusSuccessful
	}
	if hasTracerouteResponses(trace) {
		return domaintraceroute.StatusPartial
	}
	return domaintraceroute.StatusTimeout
}

func traceDestinationReached(trace *gotraceroute.Trace) bool {
	for _, hop := range traceHops(trace) {
		for _, probe := range hop.Probes {
			if probe.Status == gotraceroute.StatusDestination {
				return true
			}
		}
	}
	return false
}

func hasTracerouteResponses(trace *gotraceroute.Trace) bool {
	for _, hop := range traceHops(trace) {
		for _, probe := range hop.Probes {
			if tracerouteProbeReceived(probe) {
				return true
			}
		}
	}
	return false
}

func tracerouteHopResults(hops []gotraceroute.Hop) []domaintraceroute.HopResult {
	results := make([]domaintraceroute.HopResult, 0, len(hops))
	for _, hop := range hops {
		results = append(results, tracerouteHopResult(hop))
	}
	return results
}

func tracerouteHopResult(hop gotraceroute.Hop) domaintraceroute.HopResult {
	sentCount := int32(len(hop.Probes)) //nolint:gosec // queries per hop is bounded by domain config.
	receivedCount := int32(0)
	rtts := make([]time.Duration, 0, len(hop.Probes))
	address := firstTracerouteHopAddress(hop.Probes)
	hostname := firstTracerouteHopHostname(hop.Probes)
	var errorCode, errorMessage string

	for _, probe := range hop.Probes {
		if tracerouteProbeReceived(probe) {
			receivedCount++
		}
		if tracerouteProbeHasRTT(probe) {
			rtts = append(rtts, probe.RTT)
		}
		if errorCode == "" && probe.Status == gotraceroute.StatusError {
			errorCode = "probe_error"
			errorMessage = probeErrorMessage(probe)
		}
	}
	if receivedCount == 0 && sentCount > 0 && errorCode == "" {
		errorCode = "hop_timeout"
		errorMessage = "request timed out"
	}

	lossPercent := float64(100)
	if sentCount > 0 {
		lossPercent = clamp(float64(sentCount-receivedCount)/float64(sentCount)*100, 0, 100)
	}
	samples := durationSamplesMs(rtts)
	var minRTT, avgRTT, medianRTT, maxRTT, stddevRTT *float64
	if len(samples) > 0 {
		minRTT, avgRTT, medianRTT, maxRTT, stddevRTT = rttAggregates(samples)
	}

	return domaintraceroute.HopResult{
		HopIndex:      int32(hop.TTL), //nolint:gosec // TTL is bounded by domain config.
		Address:       address,
		Hostname:      hostname,
		SentCount:     sentCount,
		ReceivedCount: receivedCount,
		LossPercent:   lossPercent,
		RttMinMs:      minRTT,
		RttAvgMs:      avgRTT,
		RttMedianMs:   medianRTT,
		RttMaxMs:      maxRTT,
		RttStddevMs:   stddevRTT,
		RttSamplesMs:  samples,
		ErrorCode:     optionalString(errorCode),
		ErrorMessage:  optionalString(errorMessage),
	}
}

func tracerouteProbeReceived(probe gotraceroute.Probe) bool {
	return probe.Status != gotraceroute.StatusTimeout &&
		probe.Status != gotraceroute.StatusError &&
		probe.Addr.IsValid()
}

func tracerouteProbeHasRTT(probe gotraceroute.Probe) bool {
	return tracerouteProbeReceived(probe) && probe.RTT > 0
}

func firstTracerouteHopAddress(probes []gotraceroute.Probe) *netip.Addr {
	for _, probe := range probes {
		if tracerouteProbeReceived(probe) {
			addr := probe.Addr
			return &addr
		}
	}
	return nil
}

func firstTracerouteHopHostname(probes []gotraceroute.Probe) *string {
	for _, probe := range probes {
		if probe.Hostname != "" {
			hostname := probe.Hostname
			return &hostname
		}
	}
	return nil
}

func probeErrorMessage(probe gotraceroute.Probe) string {
	if probe.Error != "" {
		return probe.Error
	}
	if probe.Annotation != "" {
		return probe.Annotation
	}
	return string(probe.Status)
}

func tracerouteIPFamily(trace *gotraceroute.Trace) *domainnetwork.IPFamily {
	destination := traceDestination(trace)
	if destination.IsValid() {
		family := ipFamilyForAddr(destination)
		return &family
	}
	if trace == nil {
		return nil
	}
	switch trace.IPVersion {
	case gotraceroute.IPv4:
		family := domainnetwork.IPFamilyInet
		return &family
	case gotraceroute.IPv6:
		family := domainnetwork.IPFamilyInet6
		return &family
	default:
		return nil
	}
}

func traceFinishedAt(trace *gotraceroute.Trace, fallback time.Time) time.Time {
	if trace == nil || trace.FinishedAt.IsZero() {
		return fallback
	}
	return trace.FinishedAt.UTC()
}

func errorTracerouteResult(startedAt, finishedAt time.Time, errorCode, errorMessage string) domaintraceroute.Result {
	return tracerouteResultFromTraceWithStatus(startedAt, finishedAt, nil, domaintraceroute.StatusError, errorCode, errorMessage)
}

func traceHops(trace *gotraceroute.Trace) []gotraceroute.Hop {
	if trace == nil {
		return nil
	}
	return trace.Hops
}

func traceDestination(trace *gotraceroute.Trace) netip.Addr {
	if trace == nil {
		return netip.Addr{}
	}
	return trace.Destination
}

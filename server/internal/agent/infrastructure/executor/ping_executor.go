package executor

import (
	"context"
	"errors"
	"math"
	"net"
	"net/netip"
	"os"
	"sort"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"

	"github.com/yorukot/netstamp/internal/agent/scheduling"
	agentworker "github.com/yorukot/netstamp/internal/agent/worker"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
)

const (
	icmpReadPollInterval = 50 * time.Millisecond
	protocolICMPv4       = 1
	protocolICMPv6       = 58
)

var errNoResolvedAddress = errors.New("no resolved IP address for requested family")

type ICMPPingExecutor struct{}

type pingTarget struct {
	addr     netip.Addr
	ipFamily domainnetwork.IPFamily
}

type pingProtocol struct {
	ipFamily      domainnetwork.IPFamily
	listenNetwork string
	listenAddress string
	protocol      int
	requestType   icmp.Type
	replyType     icmp.Type
}

type rawPingStats struct {
	sentCount     int32
	receivedCount int32
	rtts          []time.Duration
	resolvedIP    *netip.Addr
	ipFamily      *domainnetwork.IPFamily
}

type pingExecutionError struct {
	status  domainping.Status
	code    string
	message string
}

type icmpEchoLoop struct {
	conn        *icmp.PacketConn
	protocol    pingProtocol
	target      pingTarget
	dst         net.Addr
	config      domainping.Config
	deadline    time.Time
	stats       *rawPingStats
	identifier  int
	packetCount int
	interval    time.Duration
	baseSeq     int
	nextSendAt  time.Time
	sent        int
	sentAt      map[int]time.Time
	received    map[int]struct{}
	buffer      []byte
}

func NewICMPPingExecutor() *ICMPPingExecutor {
	return &ICMPPingExecutor{}
}

func (e *ICMPPingExecutor) Execute(ctx context.Context, req scheduling.RunRequest) agentworker.ResultEnvelope {
	result := e.execute(ctx, req)
	return agentworker.ResultEnvelope{
		CheckID: req.Check.ID,
		Type:    domaincheck.TypePing,
		Ping:    result,
	}
}

func (e *ICMPPingExecutor) execute(ctx context.Context, req scheduling.RunRequest) domainping.Result {
	startedAt := req.ScheduledAt.UTC()
	finishedAt := time.Now().UTC()
	if req.Check.PingConfig == nil {
		return errorPingResult(startedAt, finishedAt, "missing_ping_config", "ping config is missing")
	}

	stats, err := e.run(ctx, req.Check.Target, *req.Check.PingConfig)
	finishedAt = time.Now().UTC()
	if err != nil {
		var pingErr *pingExecutionError
		if errors.As(err, &pingErr) {
			return pingResultFromStats(startedAt, finishedAt, stats, pingErr.status, pingErr.code, pingErr.message)
		}
		return pingResultFromStats(startedAt, finishedAt, stats, domainping.StatusError, "ping_failed", err.Error())
	}
	if stats.receivedCount == 0 {
		return pingResultFromStats(startedAt, finishedAt, stats, domainping.StatusTimeout, "icmp_timeout", "request timed out")
	}

	return pingResultFromStats(startedAt, finishedAt, stats, domainping.StatusSuccessful, "", "")
}

func (e *ICMPPingExecutor) run(ctx context.Context, target string, config domainping.Config) (rawPingStats, error) {
	timeout := time.Duration(config.TimeoutMs) * time.Millisecond
	if config.PacketCount <= 0 || config.PacketSizeBytes <= 0 || timeout <= 0 {
		return rawPingStats{}, newPingExecutionError(domainping.StatusError, "invalid_ping_config", "ping config contains non-positive values")
	}

	deadline := time.Now().Add(timeout)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}

	resolveCtx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()
	resolved, err := resolvePingTarget(resolveCtx, target, config.IPFamily)
	if err != nil {
		if ctxErr := contextPingExecutionError(resolveCtx); ctxErr != nil {
			return rawPingStats{}, ctxErr
		}
		return rawPingStats{}, newPingExecutionError(domainping.StatusError, "resolve_failed", err.Error())
	}

	stats := newRawPingStats(resolved)
	if !time.Now().Before(deadline) {
		return stats, nil
	}

	protocol := pingProtocolForTarget(resolved)
	conn, err := icmp.ListenPacket(protocol.listenNetwork, protocol.listenAddress)
	if err != nil {
		return stats, newPingExecutionError(domainping.StatusError, "socket_open_failed", err.Error())
	}
	defer conn.Close()

	err = runICMPEchoLoop(ctx, conn, protocol, resolved, config, deadline, &stats)
	if err != nil {
		return stats, err
	}

	return stats, nil
}

func runICMPEchoLoop(ctx context.Context, conn *icmp.PacketConn, protocol pingProtocol, target pingTarget, config domainping.Config, deadline time.Time, stats *rawPingStats) error {
	loop := newICMPEchoLoop(conn, protocol, target, config, deadline, stats)
	return loop.run(ctx)
}

func newICMPEchoLoop(conn *icmp.PacketConn, protocol pingProtocol, target pingTarget, config domainping.Config, deadline time.Time, stats *rawPingStats) *icmpEchoLoop {
	packetCount := int(config.PacketCount)
	return &icmpEchoLoop{
		conn:        conn,
		protocol:    protocol,
		target:      target,
		dst:         &net.IPAddr{IP: netIPFromAddr(target.addr)},
		config:      config,
		deadline:    deadline,
		stats:       stats,
		identifier:  os.Getpid() & 0xffff,
		packetCount: packetCount,
		interval:    pingInterval(time.Duration(config.TimeoutMs)*time.Millisecond, packetCount),
		baseSeq:     int(time.Now().UnixNano() & 0xffff),
		nextSendAt:  time.Now(),
		sentAt:      make(map[int]time.Time, packetCount),
		received:    make(map[int]struct{}, packetCount),
		buffer:      make([]byte, 65535),
	}
}

func (l *icmpEchoLoop) run(ctx context.Context) error {
	for {
		if ctxErr := contextPingExecutionError(ctx); ctxErr != nil {
			return ctxErr
		}
		if err := l.sendDue(); err != nil {
			return err
		}
		if l.finished() || !time.Now().Before(l.deadline) {
			return nil
		}
		if err := l.receiveOne(ctx); err != nil {
			return err
		}
	}
}

func (l *icmpEchoLoop) sendDue() error {
	now := time.Now()
	for l.sent < l.packetCount && !now.Before(l.nextSendAt) && now.Before(l.deadline) {
		sequence := (l.baseSeq + l.sent) & 0xffff
		sendAt := time.Now()
		if err := sendICMPEcho(l.conn, l.protocol, l.dst, l.identifier, sequence, l.config.PacketSizeBytes); err != nil {
			return newPingExecutionError(domainping.StatusError, "send_failed", err.Error())
		}
		l.stats.sentCount++
		l.sentAt[sequence] = sendAt
		l.sent++
		l.nextSendAt = sendAt.Add(l.interval)
		now = time.Now()
	}

	return nil
}

func (l *icmpEchoLoop) receiveOne(ctx context.Context) error {
	readDeadline := nextReadDeadline(time.Now(), l.nextSendAt, l.deadline, l.sent < l.packetCount)
	if err := l.conn.SetReadDeadline(readDeadline); err != nil {
		return newPingExecutionError(domainping.StatusError, "receive_failed", err.Error())
	}

	n, peer, err := l.conn.ReadFrom(l.buffer)
	receivedAt := time.Now()
	if err != nil {
		return readICMPError(ctx, err)
	}

	sequence, ok := parseICMPEchoReply(l.protocol, l.buffer[:n], peer, l.identifier, l.target.addr)
	if ok {
		l.recordReply(sequence, receivedAt)
	}

	return nil
}

func readICMPError(ctx context.Context, err error) error {
	if ctxErr := contextPingExecutionError(ctx); ctxErr != nil {
		return ctxErr
	}
	if isTimeout(err) {
		return nil
	}

	return newPingExecutionError(domainping.StatusError, "receive_failed", err.Error())
}

func (l *icmpEchoLoop) recordReply(sequence int, receivedAt time.Time) {
	if _, seen := l.received[sequence]; seen {
		return
	}
	sendAt, sent := l.sentAt[sequence]
	if !sent {
		return
	}

	l.received[sequence] = struct{}{}
	l.stats.receivedCount++
	l.stats.rtts = append(l.stats.rtts, receivedAt.Sub(sendAt))
}

func (l *icmpEchoLoop) finished() bool {
	return l.sent >= l.packetCount && len(l.received) >= len(l.sentAt)
}

func resolvePingTarget(ctx context.Context, target string, ipFamily *domainnetwork.IPFamily) (pingTarget, error) {
	if addr, err := netip.ParseAddr(target); err == nil {
		addr = addr.Unmap()
		if !addrMatchesFamily(addr, ipFamily) {
			return pingTarget{}, errNoResolvedAddress
		}
		return pingTarget{addr: addr, ipFamily: ipFamilyForAddr(addr)}, nil
	}

	addrs, err := net.DefaultResolver.LookupNetIP(ctx, resolverNetwork(ipFamily), target)
	if err != nil {
		return pingTarget{}, err
	}
	for _, addr := range addrs {
		addr = addr.Unmap()
		if !addr.IsValid() || !addrMatchesFamily(addr, ipFamily) {
			continue
		}
		return pingTarget{addr: addr, ipFamily: ipFamilyForAddr(addr)}, nil
	}

	return pingTarget{}, errNoResolvedAddress
}

func resolverNetwork(ipFamily *domainnetwork.IPFamily) string {
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

func addrMatchesFamily(addr netip.Addr, ipFamily *domainnetwork.IPFamily) bool {
	if ipFamily == nil {
		return addr.Is4() || addr.Is6()
	}

	switch *ipFamily {
	case domainnetwork.IPFamilyInet:
		return addr.Is4()
	case domainnetwork.IPFamilyInet6:
		return addr.Is6()
	default:
		return false
	}
}

func ipFamilyForAddr(addr netip.Addr) domainnetwork.IPFamily {
	if addr.Is4() {
		return domainnetwork.IPFamilyInet
	}

	return domainnetwork.IPFamilyInet6
}

func pingProtocolForTarget(target pingTarget) pingProtocol {
	if target.ipFamily == domainnetwork.IPFamilyInet6 {
		return pingProtocol{
			ipFamily:      domainnetwork.IPFamilyInet6,
			listenNetwork: "ip6:ipv6-icmp",
			listenAddress: "::",
			protocol:      protocolICMPv6,
			requestType:   ipv6.ICMPTypeEchoRequest,
			replyType:     ipv6.ICMPTypeEchoReply,
		}
	}

	return pingProtocol{
		ipFamily:      domainnetwork.IPFamilyInet,
		listenNetwork: "ip4:icmp",
		listenAddress: "0.0.0.0",
		protocol:      protocolICMPv4,
		requestType:   ipv4.ICMPTypeEcho,
		replyType:     ipv4.ICMPTypeEchoReply,
	}
}

func sendICMPEcho(conn *icmp.PacketConn, protocol pingProtocol, dst net.Addr, identifier, sequence int, packetSizeBytes int32) error {
	msg := icmp.Message{
		Type: protocol.requestType,
		Code: 0,
		Body: &icmp.Echo{
			ID:   identifier,
			Seq:  sequence,
			Data: makePingPayload(packetSizeBytes),
		},
	}
	raw, err := msg.Marshal(nil)
	if err != nil {
		return err
	}

	_, err = conn.WriteTo(raw, dst)
	return err
}

func makePingPayload(packetSizeBytes int32) []byte {
	payload := make([]byte, int(packetSizeBytes))
	for i := range payload {
		payload[i] = 1
	}

	return payload
}

func parseICMPEchoReply(protocol pingProtocol, packet []byte, peer net.Addr, identifier int, target netip.Addr) (int, bool) {
	from, ok := netAddrToDomain(peer)
	if !ok || from != target {
		return 0, false
	}
	packet, ok = stripIPHeader(protocol, packet)
	if !ok {
		return 0, false
	}

	msg, err := icmp.ParseMessage(protocol.protocol, packet)
	if err != nil || msg.Type != protocol.replyType {
		return 0, false
	}
	body, ok := msg.Body.(*icmp.Echo)
	if !ok || body.ID != identifier {
		return 0, false
	}

	return body.Seq, true
}

func stripIPHeader(protocol pingProtocol, packet []byte) ([]byte, bool) {
	if len(packet) == 0 {
		return packet, true
	}
	if protocol.ipFamily == domainnetwork.IPFamilyInet6 {
		if packet[0]>>4 != 6 {
			return packet, true
		}
		header, err := ipv6.ParseHeader(packet)
		if err != nil || header.NextHeader != protocolICMPv6 || len(packet) < ipv6.HeaderLen {
			return nil, false
		}
		return packet[ipv6.HeaderLen:], true
	}

	if packet[0]>>4 != 4 {
		return packet, true
	}
	header, err := ipv4.ParseHeader(packet)
	if err != nil || header.Protocol != protocolICMPv4 || len(packet) < header.Len {
		return nil, false
	}

	return packet[header.Len:], true
}

func nextReadDeadline(now, nextSendAt, deadline time.Time, hasMoreSends bool) time.Time {
	readDeadline := deadline
	if hasMoreSends && nextSendAt.Before(readDeadline) {
		readDeadline = nextSendAt
	}
	if pollDeadline := now.Add(icmpReadPollInterval); pollDeadline.Before(readDeadline) {
		readDeadline = pollDeadline
	}
	if !readDeadline.After(now) {
		return now
	}

	return readDeadline
}

func pingInterval(timeout time.Duration, count int) time.Duration {
	if count <= 1 {
		return timeout
	}

	interval := timeout / time.Duration(count)
	if interval <= 0 {
		return time.Millisecond
	}

	return interval
}

func pingResultFromStats(startedAt, finishedAt time.Time, stats rawPingStats, status domainping.Status, errorCode, errorMessage string) domainping.Result {
	sentCount := stats.sentCount
	receivedCount := stats.receivedCount
	lossPercent := float64(100)
	var minRTT, avgRTT, medianRTT, maxRTT, stddevRTT *float64

	if sentCount > 0 {
		lossPercent = clamp(float64(sentCount-receivedCount)/float64(sentCount)*100, 0, 100)
	}
	samples := durationSamplesMs(stats.rtts)
	if len(samples) > 0 {
		minRTT, avgRTT, medianRTT, maxRTT, stddevRTT = rttAggregates(samples)
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
		ResolvedIP:    cloneAddr(stats.resolvedIP),
		IPFamily:      cloneIPFamily(stats.ipFamily),
		Raw: map[string]any{
			"executor": "raw-icmp",
		},
		ErrorCode:    optionalString(errorCode),
		ErrorMessage: optionalString(errorMessage),
	}
}

func errorPingResult(startedAt, finishedAt time.Time, errorCode, errorMessage string) domainping.Result {
	return pingResultFromStats(startedAt, finishedAt, rawPingStats{}, domainping.StatusError, errorCode, errorMessage)
}

func newRawPingStats(target pingTarget) rawPingStats {
	addr := target.addr
	ipFamily := target.ipFamily
	return rawPingStats{
		resolvedIP: &addr,
		ipFamily:   &ipFamily,
	}
}

func newPingExecutionError(status domainping.Status, code, message string) *pingExecutionError {
	return &pingExecutionError{
		status:  status,
		code:    code,
		message: message,
	}
}

func contextPingExecutionError(ctx context.Context) *pingExecutionError {
	err := ctx.Err()
	if err == nil {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return newPingExecutionError(domainping.StatusTimeout, "context_deadline_exceeded", err.Error())
	}

	return newPingExecutionError(domainping.StatusError, "context_canceled", err.Error())
}

func (e *pingExecutionError) Error() string {
	return e.message
}

func netIPFromAddr(addr netip.Addr) net.IP {
	if addr.Is4() {
		raw := addr.As4()
		return net.IPv4(raw[0], raw[1], raw[2], raw[3])
	}

	raw := addr.As16()
	return net.IP(raw[:])
}

func netAddrToDomain(value net.Addr) (netip.Addr, bool) {
	var ip net.IP
	switch addr := value.(type) {
	case *net.IPAddr:
		ip = addr.IP
	case *net.UDPAddr:
		ip = addr.IP
	default:
		return netip.Addr{}, false
	}
	if ip == nil {
		return netip.Addr{}, false
	}

	parsed, ok := netip.AddrFromSlice(ip)
	if !ok {
		return netip.Addr{}, false
	}

	return parsed.Unmap(), true
}

func durationSamplesMs(values []time.Duration) []float64 {
	samples := make([]float64, 0, len(values))
	for _, value := range values {
		samples = append(samples, durationMs(value))
	}

	return samples
}

func rttAggregates(samples []float64) (*float64, *float64, *float64, *float64, *float64) {
	copied := append([]float64(nil), samples...)
	sort.Float64s(copied)

	minRTT := copied[0]
	maxRTT := copied[len(copied)-1]
	avgRTT := average(copied)
	medianRTT := medianSorted(copied)
	stddevRTT := stddev(copied, avgRTT)

	return floatPtr(minRTT), floatPtr(avgRTT), floatPtr(medianRTT), floatPtr(maxRTT), floatPtr(stddevRTT)
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var sum float64
	for _, value := range values {
		sum += value
	}

	return sum / float64(len(values))
}

func medianSorted(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	middle := len(values) / 2
	if len(values)%2 == 1 {
		return values[middle]
	}

	return (values[middle-1] + values[middle]) / 2
}

func stddev(values []float64, avg float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var sum float64
	for _, value := range values {
		delta := value - avg
		sum += delta * delta
	}

	return math.Sqrt(sum / float64(len(values)))
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

	return int32(millis) //nolint:gosec // millis is clamped to math.MaxInt32 above.
}

func cloneAddr(value *netip.Addr) *netip.Addr {
	if value == nil {
		return nil
	}

	copied := *value
	return &copied
}

func cloneIPFamily(value *domainnetwork.IPFamily) *domainnetwork.IPFamily {
	if value == nil {
		return nil
	}

	copied := *value
	return &copied
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

func isTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

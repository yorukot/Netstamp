package traceroute

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"time"

	"github.com/yorukot/spvalidator"

	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
)

const (
	DefaultProtocol        Protocol = ProtocolICMP
	DefaultMaxHops         int32    = 30
	DefaultTimeoutMs       int32    = 3000
	DefaultQueriesPerHop   int32    = 3
	DefaultPacketSizeBytes int32    = 56
	DefaultPort            int32    = 33434
	MaxHopsLimit           int32    = 64
	MaxQueriesPerHop       int32    = 10
	MaxPacketSizeBytes     int32    = 65507
	MaxTimeoutMs           int32    = 60000
	MaxPort                int32    = 65535
)

var (
	ErrInvalidConfig = errors.New("traceroute config invalid")
	ErrInvalidResult = errors.New("traceroute result invalid")
)

type Protocol string

const (
	ProtocolICMP Protocol = "icmp"
	ProtocolUDP  Protocol = "udp"
)

type Status string

const (
	StatusSuccessful Status = "successful"
	StatusTimeout    Status = "timeout"
	StatusError      Status = "error"
	StatusPartial    Status = "partial"
)

type Config = TracerouteConfig

type TracerouteConfig struct {
	Protocol        Protocol                `json:"protocol"`
	MaxHops         int32                   `json:"maxHops"`
	TimeoutMs       int32                   `json:"timeoutMs"`
	QueriesPerHop   int32                   `json:"queriesPerHop"`
	PacketSizeBytes int32                   `json:"packetSizeBytes"`
	Port            int32                   `json:"port"`
	IPFamily        *domainnetwork.IPFamily `json:"ipFamily,omitempty"`
}

func DefaultConfig() Config {
	return Config{
		Protocol:        DefaultProtocol,
		MaxHops:         DefaultMaxHops,
		TimeoutMs:       DefaultTimeoutMs,
		QueriesPerHop:   DefaultQueriesPerHop,
		PacketSizeBytes: DefaultPacketSizeBytes,
		Port:            DefaultPort,
	}
}

func VNConfigProtocol(protocol Protocol) (Protocol, error) {
	switch Protocol(strings.TrimSpace(string(protocol))) {
	case ProtocolICMP:
		return ProtocolICMP, nil
	case ProtocolUDP:
		return ProtocolUDP, nil
	default:
		return "", errors.New("invalid traceroute protocol")
	}
}

func VNConfigMaxHops(maxHops int32) (int32, error) {
	if err := spvalidator.Min(maxHops, 1); err != nil {
		return 0, err
	}
	if err := spvalidator.Max(maxHops, MaxHopsLimit); err != nil {
		return 0, err
	}
	return maxHops, nil
}

func VNConfigTimeoutMs(timeoutMs int32) (int32, error) {
	if err := spvalidator.Min(timeoutMs, 1); err != nil {
		return 0, err
	}
	if err := spvalidator.Max(timeoutMs, MaxTimeoutMs); err != nil {
		return 0, err
	}
	return timeoutMs, nil
}

func VNConfigQueriesPerHop(queriesPerHop int32) (int32, error) {
	if err := spvalidator.Min(queriesPerHop, 1); err != nil {
		return 0, err
	}
	if err := spvalidator.Max(queriesPerHop, MaxQueriesPerHop); err != nil {
		return 0, err
	}
	return queriesPerHop, nil
}

func VNConfigPacketSizeBytes(packetSizeBytes int32) (int32, error) {
	if err := spvalidator.Min(packetSizeBytes, 1); err != nil {
		return 0, err
	}
	if err := spvalidator.Max(packetSizeBytes, MaxPacketSizeBytes); err != nil {
		return 0, err
	}
	return packetSizeBytes, nil
}

func VNConfigPort(port int32) (int32, error) {
	if err := spvalidator.Min(port, 1); err != nil {
		return 0, err
	}
	if err := spvalidator.Max(port, MaxPort); err != nil {
		return 0, err
	}
	return port, nil
}

func VNConfigIPFamily(ipFamily *domainnetwork.IPFamily) (*domainnetwork.IPFamily, error) {
	if ipFamily == nil {
		return nil, nil //nolint:nilnil // Nil means the caller did not provide an IP family override.
	}

	if *ipFamily != domainnetwork.IPFamilyInet && *ipFamily != domainnetwork.IPFamilyInet6 {
		return nil, fmt.Errorf("invalid ip family: %s", *ipFamily)
	}

	return ipFamily, nil
}

func VNResultTime(value time.Time) (time.Time, error) {
	if value.IsZero() {
		return time.Time{}, errors.New("must be provided")
	}
	return value.UTC(), nil
}

func VNResultDurationMs(durationMs int32) (int32, error) {
	if err := spvalidator.Min(durationMs, 0); err != nil {
		return 0, err
	}
	return durationMs, nil
}

func VNResultStatus(status Status) (Status, error) {
	switch Status(strings.TrimSpace(string(status))) {
	case StatusSuccessful:
		return StatusSuccessful, nil
	case StatusTimeout:
		return StatusTimeout, nil
	case StatusError:
		return StatusError, nil
	case StatusPartial:
		return StatusPartial, nil
	default:
		return "", errors.New("invalid traceroute status")
	}
}

func VNResultHopCount(hopCount int32) (int32, error) {
	if err := spvalidator.Min(hopCount, 0); err != nil {
		return 0, err
	}
	return hopCount, nil
}

func VNResultHopIndex(hopIndex int32) (int32, error) {
	if err := spvalidator.Min(hopIndex, 1); err != nil {
		return 0, err
	}
	if err := spvalidator.Max(hopIndex, MaxHopsLimit); err != nil {
		return 0, err
	}
	return hopIndex, nil
}

func VNResultSentCount(sentCount int32) (int32, error) {
	if err := spvalidator.Min(sentCount, 0); err != nil {
		return 0, err
	}
	return sentCount, nil
}

func VNResultReceivedCount(receivedCount, sentCount int32) (int32, error) {
	if err := spvalidator.Min(receivedCount, 0); err != nil {
		return 0, err
	}
	if receivedCount > sentCount {
		return 0, errors.New("must be less than or equal to sent count")
	}
	return receivedCount, nil
}

func VNResultLossPercent(lossPercent float64) (float64, error) {
	if err := spvalidator.Min(lossPercent, 0.0); err != nil {
		return 0, err
	}
	if err := spvalidator.Max(lossPercent, 100.0); err != nil {
		return 0, err
	}
	return lossPercent, nil
}

func VNResultOptionalRTT(value *float64) (*float64, error) {
	if value == nil {
		return nil, nil //nolint:nilnil // Nil means the result did not include this RTT aggregate.
	}
	if err := spvalidator.Min(*value, 0.0); err != nil {
		return nil, err
	}

	copied := *value
	return &copied, nil
}

func VNResultRTTSamples(values []float64) ([]float64, error) {
	copied := make([]float64, 0, len(values))
	for _, value := range values {
		if err := spvalidator.Min(value, 0.0); err != nil {
			return nil, err
		}
		copied = append(copied, value)
	}
	return copied, nil
}

func VNResultOptionalText(value *string) (*string, error) {
	if value == nil {
		return nil, nil //nolint:nilnil // Nil means the result did not include optional text.
	}

	trimmed := strings.TrimSpace(*value)
	if err := spvalidator.Required(trimmed); err != nil {
		return nil, err
	}

	return &trimmed, nil
}

type ResultStorageInput struct {
	ProbeStorageID     int64
	CheckStorageID     int64
	StartedAt          time.Time
	FinishedAt         time.Time
	DurationMs         int32
	Status             Status
	ResolvedIP         *netip.Addr
	IPFamily           *domainnetwork.IPFamily
	DestinationReached bool
	HopCount           int32
	ErrorCode          *string
	ErrorMessage       *string
	Hops               []HopStorageInput
}

type Result struct {
	StartedAt          time.Time
	FinishedAt         time.Time
	DurationMs         int32
	Status             Status
	ResolvedIP         *netip.Addr
	IPFamily           *domainnetwork.IPFamily
	DestinationReached bool
	HopCount           int32
	ErrorCode          *string
	ErrorMessage       *string
	Hops               []HopResult
}

type HopStorageInput struct {
	HopIndex      int32
	Address       *netip.Addr
	Hostname      *string
	SentCount     int32
	ReceivedCount int32
	LossPercent   float64
	RttMinMs      *float64
	RttAvgMs      *float64
	RttMedianMs   *float64
	RttMaxMs      *float64
	RttStddevMs   *float64
	RttSamplesMs  []float64
	ErrorCode     *string
	ErrorMessage  *string
}

type HopResult struct {
	HopIndex      int32
	Address       *netip.Addr
	Hostname      *string
	SentCount     int32
	ReceivedCount int32
	LossPercent   float64
	RttMinMs      *float64
	RttAvgMs      *float64
	RttMedianMs   *float64
	RttMaxMs      *float64
	RttStddevMs   *float64
	RttSamplesMs  []float64
	ErrorCode     *string
	ErrorMessage  *string
}

type RunQuery struct {
	ProjectID string
	ProbeID   string
	CheckID   string
	From      time.Time
	To        time.Time
	Limit     int32
	Cursor    *time.Time
}

type RunResult struct {
	Runs       []Run
	NextCursor *time.Time
}

type Run struct {
	StartedAt          time.Time
	FinishedAt         time.Time
	DurationMs         int32
	Status             Status
	ResolvedIP         *netip.Addr
	IPFamily           *domainnetwork.IPFamily
	DestinationReached bool
	HopCount           int32
	ErrorCode          *string
	ErrorMessage       *string
	Hops               []Hop
}

type Hop struct {
	HopIndex      int32
	Address       *netip.Addr
	Hostname      *string
	SentCount     int32
	ReceivedCount int32
	LossPercent   float64
	RttMinMs      *float64
	RttAvgMs      *float64
	RttMedianMs   *float64
	RttMaxMs      *float64
	RttStddevMs   *float64
	RttSamplesMs  []float64
	ErrorCode     *string
	ErrorMessage  *string
}

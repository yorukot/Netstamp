package ping

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/yorukot/spvalidator"

	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
)

const (
	DefaultPacketCount     int32 = 4
	DefaultPacketSizeBytes int32 = 56
	DefaultTimeoutMs       int32 = 3000
	MaxPacketSizeBytes     int32 = 65507
)

var (
	ErrInvalidConfig = errors.New("ping config invalid")
	ErrInvalidResult = errors.New("ping result invalid")
)

type Status string

const (
	StatusSuccessful Status = "successful"
	StatusTimeout    Status = "timeout"
	StatusError      Status = "error"
)

type Config struct {
	PacketCount     int32                   `json:"packetCount"`
	PacketSizeBytes int32                   `json:"packetSizeBytes"`
	TimeoutMs       int32                   `json:"timeoutMs"`
	IPFamily        *domainnetwork.IPFamily `json:"ipFamily,omitempty"`
}

func DefaultConfig() Config {
	return Config{
		PacketCount:     DefaultPacketCount,
		PacketSizeBytes: DefaultPacketSizeBytes,
		TimeoutMs:       DefaultTimeoutMs,
	}
}

func VNConfigPacketCount(packetCount int32) (int32, error) {
	err := spvalidator.Min(packetCount, 1)
	if err != nil {
		return 0, err
	}

	return packetCount, nil
}

func VNConfigPacketSizeBytes(packetSizeBytes int32) (int32, error) {
	err := spvalidator.Min(packetSizeBytes, 1)
	if err != nil {
		return 0, err
	}

	err = spvalidator.Max(packetSizeBytes, MaxPacketSizeBytes)
	if err != nil {
		return 0, err
	}

	return packetSizeBytes, nil
}

func VNConfigTimeoutMs(timeoutMs int32) (int32, error) {
	err := spvalidator.Min(timeoutMs, 1)
	if err != nil {
		return 0, err
	}

	return timeoutMs, nil
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

type Result struct {
	StartedAt     time.Time
	FinishedAt    time.Time
	DurationMs    int32
	Status        Status
	SentCount     int32
	ReceivedCount int32
	LossPercent   float64
	RttMinMs      *float64
	RttAvgMs      *float64
	RttMedianMs   *float64
	RttMaxMs      *float64
	RttStddevMs   *float64
	RttSamplesMs  []float64
	ResolvedIP    *netip.Addr
	IPFamily      *domainnetwork.IPFamily
	Raw           map[string]any
	ErrorCode     *string
	ErrorMessage  *string
}

type ResultStorageInput struct {
	ProjectID     string
	ProbeID       string
	CheckID       string
	StartedAt     time.Time
	FinishedAt    time.Time
	DurationMs    int32
	Status        Status
	SentCount     int32
	ReceivedCount int32
	LossPercent   float64
	RttMinMs      *float64
	RttAvgMs      *float64
	RttMedianMs   *float64
	RttMaxMs      *float64
	RttStddevMs   *float64
	RttSamplesMs  []float64
	ResolvedIP    *netip.Addr
	IPFamily      *domainnetwork.IPFamily
	Raw           json.RawMessage
	ErrorCode     *string
	ErrorMessage  *string
}

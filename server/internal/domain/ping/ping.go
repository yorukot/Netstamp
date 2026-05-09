package ping

import (
	"encoding/json"
	"errors"
	"net/netip"
	"time"

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
	PacketCount     int32
	PacketSizeBytes int32
	TimeoutMs       int32
	IPFamily        *domainnetwork.IPFamily
}

type ResultStorageInput struct {
	ExternalID    string
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

type ResultWriteStatus string

const (
	ResultWriteAccepted  ResultWriteStatus = "accepted"
	ResultWriteDuplicate ResultWriteStatus = "duplicate"
)

type ResultWriteOutcome struct {
	ExternalID string
	Status     ResultWriteStatus
}

type VersionPayload struct {
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

func NewConfig(packetCount, packetSizeBytes, timeoutMs *int32, ipFamilyValue *string) (Config, error) {
	config := DefaultConfig()

	if packetCount != nil {
		if err := ValidatePacketCount(*packetCount); err != nil {
			return Config{}, err
		}
		config.PacketCount = *packetCount
	}
	if packetSizeBytes != nil {
		if err := ValidatePacketSizeBytes(*packetSizeBytes); err != nil {
			return Config{}, err
		}
		config.PacketSizeBytes = *packetSizeBytes
	}
	if timeoutMs != nil {
		if err := ValidateTimeoutMs(*timeoutMs); err != nil {
			return Config{}, err
		}
		config.TimeoutMs = *timeoutMs
	}
	ipFamily, err := domainnetwork.ParseOptionalIPFamily(ipFamilyValue)
	if err != nil {
		return Config{}, ErrInvalidConfig
	}
	config.IPFamily = ipFamily

	return config, nil
}

func ValidatePacketCount(value int32) error {
	if value <= 0 {
		return ErrInvalidConfig
	}

	return nil
}

func ValidatePacketSizeBytes(value int32) error {
	if value < 0 || value > MaxPacketSizeBytes {
		return ErrInvalidConfig
	}

	return nil
}

func ValidateTimeoutMs(value int32) error {
	if value <= 0 {
		return ErrInvalidConfig
	}

	return nil
}

func ConfigVersionPayload(config Config) VersionPayload {
	return VersionPayload(config)
}

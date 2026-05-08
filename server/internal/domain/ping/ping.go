package ping

import (
	"errors"

	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
)

const (
	DefaultPacketCount     int32 = 4
	DefaultPacketSizeBytes int32 = 56
	DefaultTimeoutMs       int32 = 3000
	MaxPacketSizeBytes     int32 = 65507
)

var ErrInvalidConfig = errors.New("ping config invalid")

type Config struct {
	PacketCount     int32
	PacketSizeBytes int32
	TimeoutMs       int32
	IPFamily        *domainnetwork.IPFamily
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

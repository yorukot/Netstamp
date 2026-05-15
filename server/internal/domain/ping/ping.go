package ping

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
	default:
		return "", errors.New("invalid ping status")
	}
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
	ErrorCode     *string
	ErrorMessage  *string
}

type ResultStorageInput struct {
	ProbeStorageID int64
	CheckStorageID int64
	StartedAt      time.Time
	FinishedAt     time.Time
	DurationMs     int32
	Status         Status
	SentCount      int32
	ReceivedCount  int32
	LossPercent    float64
	RttMinMs       *float64
	RttAvgMs       *float64
	RttMedianMs    *float64
	RttMaxMs       *float64
	RttStddevMs    *float64
	RttSamplesMs   []float64
	ResolvedIP     *netip.Addr
	IPFamily       *domainnetwork.IPFamily
	ErrorCode      *string
	ErrorMessage   *string
}

type SeriesQuery struct {
	ProjectID     string
	ProbeID       string
	CheckID       string
	From          time.Time
	To            time.Time
	Metric        string
	MaxDataPoints int32
}

type SeriesResolution string

const (
	SeriesResolutionRaw    SeriesResolution = "raw"
	SeriesResolutionBucket SeriesResolution = "bucket"
)

type SeriesResult struct {
	Points      []SeriesPoint
	Resolution  SeriesResolution
	TotalPoints int64
}

type SeriesPoint struct {
	Timestamp time.Time
	Value     float64
}

package tcp

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
	DefaultPort      int32 = 443
	DefaultTimeoutMs int32 = 3000
	MaxPort          int32 = 65535
)

var (
	ErrInvalidConfig = errors.New("tcp config invalid")
	ErrInvalidResult = errors.New("tcp result invalid")
)

type Status string

const (
	StatusSuccessful Status = "successful"
	StatusTimeout    Status = "timeout"
	StatusError      Status = "error"
)

type Config struct {
	Port      int32                   `json:"port"`
	TimeoutMs int32                   `json:"timeoutMs"`
	IPFamily  *domainnetwork.IPFamily `json:"ipFamily,omitempty"`
}

func DefaultConfig() Config {
	return Config{
		Port:      DefaultPort,
		TimeoutMs: DefaultTimeoutMs,
	}
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

func VNConfigTimeoutMs(timeoutMs int32) (int32, error) {
	if err := spvalidator.Min(timeoutMs, 1); err != nil {
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

func VNResultConnectDurationMs(connectDurationMs *float64) (*float64, error) {
	if connectDurationMs == nil {
		return nil, nil //nolint:nilnil // Nil means no TCP connect timing was recorded.
	}
	if err := spvalidator.Min(*connectDurationMs, 0.0); err != nil {
		return nil, err
	}

	copied := *connectDurationMs
	return &copied, nil
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
		return "", errors.New("invalid tcp status")
	}
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
	StartedAt         time.Time
	FinishedAt        time.Time
	DurationMs        int32
	Status            Status
	ConnectDurationMs *float64
	ResolvedIP        *netip.Addr
	IPFamily          *domainnetwork.IPFamily
	ErrorCode         *string
	ErrorMessage      *string
}

type ResultStorageInput struct {
	ProbeStorageID    int64
	CheckStorageID    int64
	StartedAt         time.Time
	FinishedAt        time.Time
	DurationMs        int32
	Status            Status
	ConnectDurationMs *float64
	ResolvedIP        *netip.Addr
	IPFamily          *domainnetwork.IPFamily
	ErrorCode         *string
	ErrorMessage      *string
}

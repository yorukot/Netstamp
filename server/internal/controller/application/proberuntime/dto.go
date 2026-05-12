package proberuntime

import (
	"encoding/json"
	"net/netip"
	"time"

	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

const (
	DefaultHeartbeatIntervalSeconds      int32 = 30
	DefaultAssignmentPollIntervalSeconds int32 = 30
	DefaultMaxConcurrentChecks           int32 = 16
	DefaultInitialBackoffSeconds         int32 = 1
	DefaultMaxBackoffSeconds             int32 = 30
	DefaultMaxAttempts                   int32 = 5
	DefaultMinimumSupportedAgentVersion        = "0.1.0"
	MaxResultGroupBatchSize              int   = 100
)

type RuntimeAuthInput struct {
	ProbeID    string
	Credential string
}

type RuntimeStatusInput struct {
	RuntimeAuthInput
	AgentVersion *string
	PublicV4     *netip.Addr
	PublicV6     *netip.Addr
	AS           *string
	Addrs        []netip.Addr
}

type HelloOutput struct {
	ServerTime                   time.Time
	MinimumSupportedAgentVersion string
	Config                       domainprobe.RuntimeConfig
}

type HeartbeatOutput struct {
	ServerTime time.Time
}

type ListAssignmentsOutput struct {
	ServerTime  time.Time
	Config      domainprobe.RuntimeConfig
	Assignments []domainassignment.Assignment
}

type SubmitResultsInput struct {
	RuntimeAuthInput
	Results []RuntimeResultGroupInput
}

type RuntimeResultGroupInput struct {
	CheckID string
	Type    string
	Ping    []PingResultInput
}

type PingResultInput struct {
	StartedAt     time.Time
	FinishedAt    time.Time
	DurationMs    int32
	Status        string
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
	IPFamily      *string
	Raw           json.RawMessage
	ErrorCode     *string
	ErrorMessage  *string
}

type SubmitResultsOutput struct {
	Accepted   int
	ServerTime time.Time
}

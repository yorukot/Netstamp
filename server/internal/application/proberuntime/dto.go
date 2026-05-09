package proberuntime

import (
	"encoding/json"
	"net/netip"
	"time"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
)

const (
	DefaultHeartbeatIntervalSeconds      int32 = 30
	DefaultAssignmentPollIntervalSeconds int32 = 30
	MaxPingResultBatchSize               int   = 500
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
	Addrs        []netip.Addr
}

type HelloOutput struct {
	ServerTime                    time.Time
	HeartbeatIntervalSeconds      int32
	AssignmentPollIntervalSeconds int32
}

type HeartbeatOutput struct {
	ServerTime time.Time
}

type ListAssignmentsOutput struct {
	Assignments []domaincheck.Assignment
}

type SubmitPingResultsInput struct {
	RuntimeAuthInput
	Results []PingResultInput
}

type PingResultInput struct {
	ResultID      string
	CheckID       string
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
	IPFamily      *domainnetwork.IPFamily
	Raw           json.RawMessage
	ErrorCode     *string
	ErrorMessage  *string
}

type PingResultOutcomeStatus string

const (
	PingResultOutcomeAccepted  PingResultOutcomeStatus = "accepted"
	PingResultOutcomeDuplicate PingResultOutcomeStatus = "duplicate"
	PingResultOutcomeRejected  PingResultOutcomeStatus = "rejected"
)

type PingResultOutcome struct {
	ResultID string
	Status   PingResultOutcomeStatus
	Error    *string
}

type SubmitPingResultsOutput struct {
	Results []PingResultOutcome
}

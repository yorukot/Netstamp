package proberuntime

import (
	"encoding/json"
	"net/netip"
	"time"

	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
)

const (
	DefaultHeartbeatIntervalSeconds      int32 = 30
	DefaultAssignmentPollIntervalSeconds int32 = 30
	MaxResultBatchSize                   int   = 500
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
	ServerTime                    time.Time
	HeartbeatIntervalSeconds      int32
	AssignmentPollIntervalSeconds int32
}

type HeartbeatOutput struct {
	ServerTime time.Time
}

type ListAssignmentsOutput struct {
	Assignments []domainassignment.Assignment
}

type SubmitResultsOutput struct {
	Accepted     bool
	ResyncNeeded bool
	StaleChecks  []string
	Assignments  []domainassignment.Assignment
}

type SubmitResultsInput struct {
	RuntimeAuthInput
	Groups []ResultGroupInput
}

type ResultGroupInput struct {
	CheckID         string
	Type            domaincheck.Type
	AssignmentID    string
	CheckVersion    string
	SelectorVersion string
	PingResults     []PingResultInput
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
	IPFamily      *domainnetwork.IPFamily
	Raw           json.RawMessage
	ErrorCode     *string
	ErrorMessage  *string
}

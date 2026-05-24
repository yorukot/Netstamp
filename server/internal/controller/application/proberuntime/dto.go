package proberuntime

import (
	"net/netip"
	"time"

	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
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
	CheckID    string
	Type       string
	Ping       []PingResultInput
	TCP        []TCPResultInput
	Traceroute []TracerouteResultInput
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
	ErrorCode     *string
	ErrorMessage  *string
}

type TCPResultInput struct {
	StartedAt         time.Time
	FinishedAt        time.Time
	DurationMs        int32
	Status            string
	ConnectDurationMs *float64
	ResolvedIP        *netip.Addr
	IPFamily          *string
	ErrorCode         *string
	ErrorMessage      *string
}

type SubmitResultsOutput struct {
	Accepted   int
	ServerTime time.Time
}

type TracerouteResultInput struct {
	StartedAt          time.Time
	FinishedAt         time.Time
	DurationMs         int32
	Status             string
	ResolvedIP         *netip.Addr
	IPFamily           *string
	DestinationReached bool
	HopCount           int32
	Hops               []TracerouteHopInput
	ErrorCode          *string
	ErrorMessage       *string
}

type TracerouteHopInput struct {
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

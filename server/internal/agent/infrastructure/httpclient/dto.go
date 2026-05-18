package httpclient

import (
	"net/netip"
	"time"

	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

type HelloResponse struct {
	ServerTime                   time.Time                 `json:"serverTime"`
	MinimumSupportedAgentVersion string                    `json:"minimumSupportedAgentVersion"`
	Config                       domainprobe.RuntimeConfig `json:"config"`
}

type HeartbeatInput struct {
	AgentVersion *string      `json:"agentVersion,omitempty"`
	PublicV4     *netip.Addr  `json:"publicV4,omitempty"`
	PublicV6     *netip.Addr  `json:"publicV6,omitempty"`
	AS           *string      `json:"as,omitempty"`
	Addrs        []netip.Addr `json:"addrs,omitempty"`
}

type HeartbeatResponse struct {
	ServerTime time.Time `json:"serverTime"`
}

type AssignmentsResponse struct {
	ServerTime  time.Time                     `json:"serverTime"`
	Config      domainprobe.RuntimeConfig     `json:"config"`
	Assignments []domainassignment.Assignment `json:"assignments"`
}

type SubmitResultsInput struct {
	Results []RuntimeResultGroup `json:"results"`
}

type RuntimeResultGroup struct {
	CheckID    string                 `json:"checkId"`
	Type       domaincheck.Type       `json:"type"`
	Ping       []PingResultBody       `json:"ping,omitempty"`
	Traceroute []TracerouteResultBody `json:"traceroute,omitempty"`
}

type PingResultBody struct {
	StartedAt     time.Time               `json:"startedAt"`
	FinishedAt    time.Time               `json:"finishedAt"`
	DurationMs    int32                   `json:"durationMs"`
	Status        domainping.Status       `json:"status"`
	SentCount     int32                   `json:"sentCount"`
	ReceivedCount int32                   `json:"receivedCount"`
	LossPercent   float64                 `json:"lossPercent"`
	RttMinMs      *float64                `json:"rttMinMs,omitempty"`
	RttAvgMs      *float64                `json:"rttAvgMs,omitempty"`
	RttMedianMs   *float64                `json:"rttMedianMs,omitempty"`
	RttMaxMs      *float64                `json:"rttMaxMs,omitempty"`
	RttStddevMs   *float64                `json:"rttStddevMs,omitempty"`
	RttSamplesMs  []float64               `json:"rttSamplesMs,omitempty"`
	ResolvedIP    *netip.Addr             `json:"resolvedIp,omitempty"`
	IPFamily      *domainnetwork.IPFamily `json:"ipFamily,omitempty"`
	ErrorCode     *string                 `json:"errorCode,omitempty"`
	ErrorMessage  *string                 `json:"errorMessage,omitempty"`
}

type SubmitResultsResponse struct {
	Accepted   int       `json:"accepted"`
	ServerTime time.Time `json:"serverTime"`
}

type TracerouteResultBody struct {
	StartedAt          time.Time               `json:"startedAt"`
	FinishedAt         time.Time               `json:"finishedAt"`
	DurationMs         int32                   `json:"durationMs"`
	Status             domaintraceroute.Status `json:"status"`
	ResolvedIP         *netip.Addr             `json:"resolvedIp,omitempty"`
	IPFamily           *domainnetwork.IPFamily `json:"ipFamily,omitempty"`
	DestinationReached bool                    `json:"destinationReached"`
	HopCount           int32                   `json:"hopCount"`
	Hops               []TracerouteHopBody     `json:"hops,omitempty"`
	ErrorCode          *string                 `json:"errorCode,omitempty"`
	ErrorMessage       *string                 `json:"errorMessage,omitempty"`
}

type TracerouteHopBody struct {
	HopIndex      int32       `json:"hopIndex"`
	Address       *netip.Addr `json:"address,omitempty"`
	Hostname      *string     `json:"hostname,omitempty"`
	SentCount     int32       `json:"sentCount"`
	ReceivedCount int32       `json:"receivedCount"`
	LossPercent   float64     `json:"lossPercent"`
	RttMinMs      *float64    `json:"rttMinMs,omitempty"`
	RttAvgMs      *float64    `json:"rttAvgMs,omitempty"`
	RttMedianMs   *float64    `json:"rttMedianMs,omitempty"`
	RttMaxMs      *float64    `json:"rttMaxMs,omitempty"`
	RttStddevMs   *float64    `json:"rttStddevMs,omitempty"`
	RttSamplesMs  []float64   `json:"rttSamplesMs,omitempty"`
	ErrorCode     *string     `json:"errorCode,omitempty"`
	ErrorMessage  *string     `json:"errorMessage,omitempty"`
}

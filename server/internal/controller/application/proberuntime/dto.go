package proberuntime

import (
	"net/netip"
	"time"

	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
)

const (
	DefaultHeartbeatIntervalSeconds      int32 = 30
	DefaultAssignmentPollIntervalSeconds int32 = 30
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

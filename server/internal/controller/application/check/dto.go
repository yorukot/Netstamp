package check

import (
	"encoding/json"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

type ListChecksInput struct {
	CurrentUserID string
	ProjectRef    string
}

type GetCheckInput struct {
	CurrentUserID string
	ProjectRef    string
	CheckID       string
}

type CreateCheckInput struct {
	CurrentUserID   string
	ProjectRef      string
	Name            string
	Type            string
	Target          string
	Selector        map[string]any
	Description     *string
	IntervalSeconds int32
	LabelIDs        []string

	PingConfig       *PingConfigInput
	TCPConfig        *TCPConfigInput
	TracerouteConfig *TracerouteConfigInput
}

type UpdateCheckInput struct {
	CurrentUserID    string
	ProjectRef       string
	CheckID          string
	Name             *string
	Type             *string
	Target           *string
	Selector         map[string]any
	Description      *string
	IntervalSeconds  *int32
	PingConfig       *PingConfigInput
	TCPConfig        *TCPConfigInput
	TracerouteConfig *TracerouteConfigInput
	LabelIDs         *[]string
}

type PingConfigInput struct {
	PacketCount     *int32
	PacketSizeBytes *int32
	TimeoutMs       *int32
	IPFamily        *string
}

type TracerouteConfigInput struct {
	Protocol        *string
	MaxHops         *int32
	TimeoutMs       *int32
	QueriesPerHop   *int32
	PacketSizeBytes *int32
	Port            *int32
	IPFamily        *string
}

type TCPConfigInput struct {
	Port      *int32
	TimeoutMs *int32
	IPFamily  *string
}

type normalizedCreateCheckInput struct {
	projectRef       string
	name             string
	checkType        domaincheck.Type
	target           string
	selector         json.RawMessage
	description      *string
	intervalSeconds  int32
	pingConfig       *domainping.Config
	tcpConfig        *domaintcp.Config
	tracerouteConfig *domaintraceroute.Config
	labelIDs         []string
}

type normalizedUpdateCheckInput struct {
	projectRef       string
	checkID          string
	name             *string
	checkType        *domaincheck.Type
	target           *string
	selector         json.RawMessage
	description      *string
	intervalSeconds  *int32
	pingConfig       updatePingConfigPatch
	tcpConfig        updateTCPConfigPatch
	tracerouteConfig updateTracerouteConfigPatch
	replaceLabels    bool
	labelIDs         []string
}

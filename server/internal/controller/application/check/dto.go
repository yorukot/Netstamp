package check

import (
	"encoding/json"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
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
	PingConfig      *PingConfigInput
	LabelIDs        []string
}

type UpdateCheckInput struct {
	CurrentUserID   string
	ProjectRef      string
	CheckID         string
	Name            *string
	Type            *string
	Target          *string
	Selector        map[string]any
	Description     *string
	IntervalSeconds *int32
	PingConfig      *PingConfigInput
	LabelIDs        *[]string
}

type PingConfigInput struct {
	PacketCount     *int32
	PacketSizeBytes *int32
	TimeoutMs       *int32
	IPFamily        *string
}

type normalizedCreateCheckInput struct {
	name            string
	checkType       domaincheck.Type
	target          string
	selector        json.RawMessage
	description     *string
	intervalSeconds int32
	pingConfig      domainping.Config
	labelIDs        []string
}

type normalizedUpdateCheckInput struct {
	name            *string
	checkType       *domaincheck.Type
	target          *string
	selector        json.RawMessage
	description     *string
	intervalSeconds *int32
	packetCount     *int32
	packetSizeBytes *int32
	timeoutMs       *int32
	ipFamily        *domainnetwork.IPFamily
	replaceLabels   bool
	labelIDs        []string
}

package check

import (
	"encoding/json"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
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
	IntervalSeconds int
	PacketCount     *int
	PacketSizeBytes *int
	TimeoutMs       *int
	IPFamily        *string
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
	IntervalSeconds *int
	PacketCount     *int
	PacketSizeBytes *int
	TimeoutMs       *int
	IPFamily        *string
	LabelIDs        *[]string
}

type normalizedCreateCheckInput struct {
	name            string
	checkType       domaincheck.Type
	target          string
	selector        json.RawMessage
	description     *string
	intervalSeconds int
	pingConfig      domaincheck.PingConfig
	labelIDs        []string
}

type normalizedUpdateCheckInput struct {
	name            *string
	checkType       *domaincheck.Type
	target          *string
	selector        json.RawMessage
	description     *string
	intervalSeconds *int
	packetCount     *int
	packetSizeBytes *int
	timeoutMs       *int
	ipFamily        *domaincheck.IPFamily
	replaceLabels   bool
	labelIDs        []string
}
